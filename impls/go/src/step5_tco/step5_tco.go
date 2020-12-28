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
top:
	// Some special forms are tail call optimised. Instead of recusively
	// calling Eval, they return a new `ast` and `env`, and we loop back to
	// the top of this function.
	if operator, args, ok := isTCOSpecialForm(ast); ok {
		newAST, newEnv, err := evalTCOSpecialForm(operator, args, env)
		if err != nil {
			return nil, err
		}
		ast = newAST
		env = newEnv
		// XXX: The other option here is to wrap this function body if a while
		// loop, and `continue` here. They've equivalent because all other
		// branches return. Using a goto seems somewhat nicer though??
		goto top
	}

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

	// Apply phase - evaluate all elements in the list, then call the first
	// as a function, with the rest as arguments
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

	if !function.TailCallOptimised {
		return function.Func(evaluatedList.Items[1:]...)
	}

	// Function is tail call optimised.
	// Construct the correct environment it should be run in
	childEnv := environment.NewChildEnv(
		function.Env.(*environment.Env), function.Params, evaluatedList.Items[1:],
	)

	ast = function.AST
	env = childEnv
	goto top
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

func isTCOSpecialForm(ast types.MalType) (operator *types.MalSymbol, args []types.MalType, ok bool) {
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
	case "let*", "if", "do":
		return operator, items[1:], true
	}

	return nil, nil, false
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
	case "fn*", "def!":
		return operator, items[1:], true
	}

	return nil, nil, false
}

// Some special forms end in an evaluation. We could implement this by
// recusively calling `Eval` (it's recusive because evalTCOSpecialForm is
// called by Eval), but that can lead to stack overflow issues. Instead, we
// tail call optimise by returning a new AST to evaluate, and a new environment
// to invaluate it in. Eval loops back to the beginning of the function and
// re-runs itself using these new params.
func evalTCOSpecialForm(
	operator *types.MalSymbol, args []types.MalType, env *environment.Env,
) (newAST types.MalType, newEnv *environment.Env, err error) {
	switch operator.Value {

	// Creates a new environment with certain variables set, then evaluates a
	// statement in that environment.
	// e.g:
	//
	// > (let* (a 1 b (+ a 1)) b)
	// 2 ; a == b, b == a+1 == 2
	case "let*":
		if len(args) != 2 {
			return nil, nil, fmt.Errorf("let* takes 2 args")
		}
		bindingList, ok := args[0].(*types.MalList)
		if !ok {
			return nil, nil, fmt.Errorf("let*: first arg isn't a list")
		}
		if len(bindingList.Items)%2 != 0 {
			return nil, nil, fmt.Errorf("let*: first arg doesn't have an even number of items")
		}

		childEnv := env.ChildEnv()
		for i := 0; i < len(bindingList.Items); i += 2 {
			key, ok := bindingList.Items[i].(*types.MalSymbol)
			if !ok {
				return nil, nil, fmt.Errorf("let*: binding list: arg %d isn't a symbol", i)
			}
			value, err := Eval(bindingList.Items[i+1], childEnv)
			if err != nil {
				return nil, nil, err
			}
			childEnv.Set(key.Value, value)
		}

		// Finally, return the last arg as the new AST to be evaluated, and the
		// newly constructed env as the environment
		return args[1], childEnv, nil

	// Evaluates the elements in the arg list and returns the final result.
	// For TCO, we eval all but the last argument here, then return the last
	// argument to be evaluated in the main Eval loop.
	case "do":
		for _, arg := range args[:len(args)-1] {
			var err error
			_, err = Eval(arg, env)
			if err != nil {
				return nil, nil, err
			}
		}
		return args[len(args)-1], env, nil

	// Evaluate first param. If not `nil` or `false`, return the second param
	// to be evaluated. If it is, return the third param to be evaluated, or
	// `nil` if none is supplied. If none is supplied, the `nil` value is
	// evalulated, but just evaluates to `nil`.
	case "if":
		if numArgs := len(args); numArgs != 2 && numArgs != 3 {
			return nil, nil, fmt.Errorf("if statements must have two or three arguments, got %d", numArgs)
		}
		condition, err := Eval(args[0], env)
		if err != nil {
			return nil, nil, err
		}
		if IsTruthy(condition) {
			return args[1], env, nil
		}

		if len(args) == 3 {
			return args[2], env, nil
		}

		return &types.MalNil{}, env, nil

	default:
		return nil, nil, fmt.Errorf("unexpected tail call optimised special form: %s", operator.Value)
	}
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

	// Create a new function.
	//
	// e.g:
	// > (def! add1 (fn* (a) (+ a 1)))
	// #<function>
	// > (add1 2)
	// 3
	case "fn*":
		if len(args) != 2 {
			return nil, fmt.Errorf("fn* statements must have two arguments, got %d", len(args))
		}

		// arguments is the first argument supplied to the fn* function (e.g.
		// `(a)` in the example above)
		arguments, ok := args[0].(*types.MalList)
		if !ok {
			return nil, fmt.Errorf("fn* statements must have a list as the first arg")
		}
		// Cast it from a list of MalType to a list of MalSymbol
		binds := make([]*types.MalSymbol, len(arguments.Items))
		for i, a := range arguments.Items {
			bind, ok := a.(*types.MalSymbol)
			if !ok {
				// TODO: improve this - say which argument isn't a symbol
				return nil, fmt.Errorf("fn* statements must have a list of symbols as the first arg")
			}
			binds[i] = bind
		}

		// TODO: recomment this
		return &types.MalFunction{
			TailCallOptimised: true,
			AST:               args[1],
			Params:            binds,
			Env:               env,
			// This Go function is what's run when the Lisp function is
			// run. When the Lisp function is run, we create a new environment,
			// which binds the Lisp function's arguments to the parameters
			// defined when the function was defined.
			Func: func(exprs ...types.MalType) (types.MalType, error) {
				childEnv := environment.NewChildEnv(
					env, binds, exprs,
				)
				return Eval(args[1], childEnv)
			},
		}, nil

	// XXX: if you add a case here, you also need to add it to `isSpecialForm`

	default:
		return nil, fmt.Errorf("unexpected special form: %s", operator.Value)
	}
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
