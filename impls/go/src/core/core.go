package core

import (
	"fmt"

	"github.com/jamesroutley/mal/impls/go/src/types"
)

type NamespaceItem struct {
	Symbol *types.MalSymbol
	Func   *types.MalFunction
}

var Namespace []*NamespaceItem

func register(symbol string, f func(...types.MalType) (types.MalType, error)) {
	item := &NamespaceItem{
		Symbol: &types.MalSymbol{Value: symbol},
		Func:   &types.MalFunction{Func: f},
	}
	Namespace = append(Namespace, item)
}

func init() {
	register("+", add)
	register("-", subtract)
	register("*", multiply)
	register("/", divide)
	register("prn", prn)
	register("list", list)
	register("list?", isList)
	register("empty?", isEmpty)
	register("count", count)
	register("=", equals)
	register("<", lt)
	register("<=", lte)
	register(">", gt)
	register(">=", gte)
}

func ValidateNArgs(n int, args []types.MalType) error {
	if actual := len(args); actual != n {
		return fmt.Errorf("function takes %d args, got %d", n, actual)
	}
	return nil
}

func ArgsToMalInt(args []types.MalType) ([]*types.MalInt, error) {
	numbers := make([]*types.MalInt, len(args))
	for i, arg := range args {
		number, ok := arg.(*types.MalInt)
		if !ok {
			return nil, fmt.Errorf("could not cast type to int")
		}
		numbers[i] = number
	}
	return numbers, nil
}
