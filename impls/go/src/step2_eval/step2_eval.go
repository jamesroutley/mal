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

	// fmt.Println(env)

	// fmt.Println(Rep("(+ 1 100)", env))
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

func Read(s string) (types.MalType, error) {
	return reader.ReadStr(s)
}

func Eval(ast types.MalType, env *environment.Env) (types.MalType, error) {
	switch tok := ast.(type) {
	case *types.MalList:
		if len(tok.Items) == 0 {
			return ast, nil
		}

		reader.DebugType(tok)

		evaluated, err := evalAST(tok, env)
		if err != nil {
			return nil, err
		}

		evaluatedList, ok := evaluated.(*types.MalList)
		if !ok {
			return nil, fmt.Errorf("List did not evaluate to a list")
		}

		reader.DebugType(evaluatedList)

		function, ok := evaluatedList.Items[0].(*types.MalFunction)
		if !ok {
			return nil, fmt.Errorf("First item in list isn't a function")
		}

		return function.Func(evaluatedList.Items[1:len(evaluatedList.Items)]...)
	}
	return evalAST(ast, env)
}

func Print(s types.MalType) string {
	return printer.PrStr(s)
}

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
