package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jamesroutley/mal/impls/go/src/core"
	"github.com/jamesroutley/mal/impls/go/src/environment"
	"github.com/jamesroutley/mal/impls/go/src/printer"
	"github.com/jamesroutley/mal/impls/go/src/reader"
	"github.com/jamesroutley/mal/impls/go/src/types"
)

func main() {
	env := environment.NewEnv()
	for _, item := range core.Namespace {
		env.Set(item.Symbol.Value, item.Func)
	}

	// Builtin functions defined in lisp
	_, err := Rep("(def! not (fn* (a) (if a false true)))", env)
	if err != nil {
		log.Fatal(err)
	}

	// code := "((fn* (a) a) true)"
	// ast, err := Read(code)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// reader.DebugType(ast)
	// fmt.Println(Eval(ast, env))

	// return

	for {
		fmt.Printf("user> ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		line = strings.TrimSuffix(line, "\n")
		output, err := Rep(line, env)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(output)
	}
}

// Read tokenizes and parses source code
func Read(s string) (types.MalType, error) {
	return reader.ReadStr(s)
}

// Eval evaulates a piece of parsed code.
// The way code is evaluated depends on its structure.
//
// 1. Special forms - these are language-level features, which behave
// differently to normal functions. These include `def!` and `let*`. For
// example, certain elements in the argument list might be evaluated
// differently (or not at all)
// 2. Symbols: evaluated to their corresponding value in the environment `env`
// 3. Lists: by default, they're treated as function calls - each item is
// evaluated, and the first item (the function itself) is called with the rest
// of the items as arguments.
func Eval(ast types.MalType, env *environment.Env) (types.MalType, error) {
	if operator, args, ok := isSpecialForm(ast); ok {
		return evalSpecialForm(operator, args, env)
	}

	list, ok := ast.(*types.MalList)
	if !ok {
		return evalAST(ast, env)
	}
	if len(list.Items) == 0 {
		return ast, nil
	}

	evaluated, err := evalAST(list, env)
	if err != nil {
		return nil, err
	}

	evaluatedList, ok := evaluated.(*types.MalList)
	if !ok {
		return nil, fmt.Errorf("list did not evaluate to a list")
	}

	function, ok := evaluatedList.Items[0].(*types.MalFunction)
	if !ok {
		return nil, fmt.Errorf("first item in list isn't a function")
	}
	return function.Func(evaluatedList.Items[1:]...)
}

// Print prints the AST as a human readable string. It's not inteded for debugging
func Print(s types.MalType) string {
	return printer.PrStr(s)
}

// Rep - read, evaluate, print
func Rep(s string, env *environment.Env) (string, error) {
	t, err := Read(s)
	if err != nil {
		return "", err
	}
	t, err = Eval(t, env)
	if err != nil {
		return "", err
	}
	s = Print(t)
	return s, nil
}

// evalAST implements the evaluation rules for normal expressions. Any special
// cases are handed above us, in the Eval function. This function is an
// implementation detail of Eval, and shoulnd't be called apart from by it.
func evalAST(ast types.MalType, env *environment.Env) (types.MalType, error) {
	switch tok := ast.(type) {
	case *types.MalSymbol:
		value, err := env.Get(tok.Value)
		if err != nil {
			return nil, err
		}
		return value, nil
	case *types.MalList:
		items := make([]types.MalType, len(tok.Items))
		for i, item := range tok.Items {
			evaluated, err := Eval(item, env)
			if err != nil {
				return nil, err
			}
			items[i] = evaluated
		}
		return &types.MalList{
			Items: items,
		}, nil
	}
	return ast, nil
}

func isSpecialForm(ast types.MalType) (operator *types.MalSymbol, args []types.MalType, ok bool) {
	tok, ok := ast.(*types.MalList)
	if !ok {
		return nil, nil, false
	}
	items := tok.Items
	if len(items) == 0 {
		return nil, nil, false
	}

	operator, ok = items[0].(*types.MalSymbol)
	if !ok {
		return nil, nil, false
	}

	switch operator.Value {
	case "def!", "let*", "if", "do", "fn*":
		return operator, items[1:], true
	}

	return nil, nil, false
}

func evalSpecialForm(
	operator *types.MalSymbol, args []types.MalType, env *environment.Env,
) (types.MalType, error) {
	switch operator.Value {

	// Assigns a value to a symbol in the current environment
	// e.g:
	//
	// > (def a 10)
	// 10
	// > a
	// 10
	case "def!":
		if len(args) != 2 {
			return nil, fmt.Errorf("def! takes 2 args")
		}
		key, ok := args[0].(*types.MalSymbol)
		if !ok {
			return nil, fmt.Errorf("def!: first arg isn't a symbol")
		}
		value, err := Eval(args[1], env)
		if err != nil {
			return nil, err
		}
		env.Set(key.Value, value)
		return value, nil

	// Creates a new environment with certain variables set, then evaluates a
	// statement in that environment.
	// e.g:
	//
	// (let* (a 1 b (+ a 1)) b)
	// 2 ; a == b, b == a+1 == 2
	case "let*":
		if len(args) != 2 {
			return nil, fmt.Errorf("let* takes 2 args")
		}
		bindingList, ok := args[0].(*types.MalList)
		if !ok {
			return nil, fmt.Errorf("let*: first arg isn't a list")
		}
		if len(bindingList.Items)%2 != 0 {
			return nil, fmt.Errorf("let*: first arg doesn't have an even number of items")
		}

		childEnv := env.ChildEnv()
		for i := 0; i < len(bindingList.Items); i += 2 {
			key, ok := bindingList.Items[i].(*types.MalSymbol)
			if !ok {
				return nil, fmt.Errorf("let*: binding list: arg %d isn't a symbol", i)
			}
			value, err := Eval(bindingList.Items[i+1], childEnv)
			if err != nil {
				return nil, err
			}
			childEnv.Set(key.Value, value)
		}

		// Finally, evaluate the last arg, in the context of the newly
		// created environment
		return Eval(args[1], childEnv)

	// Evaluates the elements in the arg list and returns the final result.
	case "do":
		var result types.MalType
		for _, arg := range args {
			var err error
			result, err = Eval(arg, env)
			if err != nil {
				return nil, err
			}
		}
		return result, nil

	// Evaluate first param. If not `nil` or `false`, return the evaluated
	// second param. If it is, return the evaluated third param, or `nil` if
	// none is supplied
	case "if":
		if numArgs := len(args); numArgs != 2 && numArgs != 3 {
			return nil, fmt.Errorf("if statements must have two or three arguments, got %d", numArgs)
		}
		condition, err := Eval(args[0], env)
		if err != nil {
			return nil, err
		}
		if IsTruthy(condition) {
			return Eval(args[1], env)
		}

		if len(args) == 3 {
			return Eval(args[2], env)
		}

		return &types.MalNil{}, nil

	case "fn*":
		if len(args) != 2 {
			return nil, fmt.Errorf("fn* statements must have two arguments, got %d", len(args))
		}
		arguments, ok := args[0].(*types.MalList)
		if !ok {
			return nil, fmt.Errorf("fn* statements must have a list as the first arg")
		}
		binds := make([]*types.MalSymbol, len(arguments.Items))
		for i, a := range arguments.Items {
			bind, ok := a.(*types.MalSymbol)
			if !ok {
				// TODO: improve this
				return nil, fmt.Errorf("fn* statements must have a list of symbols as the first arg")
			}
			binds[i] = bind
		}

		return &types.MalFunction{
			Func: func(exprs ...types.MalType) (types.MalType, error) {
				childEnv := environment.NewChildEnv(
					env, binds, exprs,
				)
				return Eval(args[1], childEnv)
			},
		}, nil
	}

	// XXX: if you add a case here, you also need to add it to `isSpecialForm`

	return nil, fmt.Errorf("unexpected special form: %s", operator.Value)
}

// IsTruthy returns a type's truthiness. Currently: it's falsy if the type is
// `nil` or the boolean 'false'. All other values are truthy.
func IsTruthy(t types.MalType) bool {
	switch token := t.(type) {
	case *types.MalNil:
		return false
	case *types.MalBoolean:
		return token.Value
	}
	return true

}
