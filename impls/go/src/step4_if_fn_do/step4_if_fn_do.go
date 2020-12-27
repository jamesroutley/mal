package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jamesroutley/mal/impls/go/src/builtin"
	"github.com/jamesroutley/mal/impls/go/src/environment"
	"github.com/jamesroutley/mal/impls/go/src/printer"
	"github.com/jamesroutley/mal/impls/go/src/reader"
	"github.com/jamesroutley/mal/impls/go/src/types"
)

func main() {
	env := environment.NewEnv()
	env.Set("+", builtin.Add)
	env.Set("-", builtin.Subtract)
	env.Set("*", builtin.Multiply)
	env.Set("/", builtin.Divide)

	// code := "(def! (a 1 b (+ a 5)) (+ b 4))"
	// ast, err := Read(code)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// reader.DebugType(ast)

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
	case "def!", "let*":
		return operator, items[1:], true
	}

	return nil, nil, false
}

func evalSpecialForm(
	operator *types.MalSymbol, args []types.MalType, env *environment.Env,
) (types.MalType, error) {
	switch operator.Value {
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
	}

	return nil, fmt.Errorf("unexpected special form: %s", operator.Value)
}
