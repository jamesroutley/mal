package core

import (
	"fmt"
	"log"
	"reflect"

	"github.com/jamesroutley/mal/impls/go/src/printer"
	"github.com/jamesroutley/mal/impls/go/src/types"
)

func prn(args ...types.MalType) (types.MalType, error) {
	fmt.Println(printer.PrStr(args[0]))
	return &types.MalNil{}, nil
}

func list(args ...types.MalType) (types.MalType, error) {
	return &types.MalList{
		Items: args,
	}, nil
}

func isList(args ...types.MalType) (types.MalType, error) {
	_, ok := args[0].(*types.MalList)
	return &types.MalBoolean{
		Value: ok,
	}, nil
}

func isEmpty(args ...types.MalType) (types.MalType, error) {
	list, ok := args[0].(*types.MalList)
	if !ok {
		return nil, fmt.Errorf("first argument to empty? isn't a list")
	}

	return &types.MalBoolean{
		Value: len(list.Items) == 0,
	}, nil
}

func count(args ...types.MalType) (types.MalType, error) {
	if _, ok := args[0].(*types.MalNil); ok {
		return &types.MalInt{
			Value: 0,
		}, nil
	}
	list, ok := args[0].(*types.MalList)
	if !ok {
		return nil, fmt.Errorf("first argument to count isn't a list")
	}

	return &types.MalInt{
		Value: len(list.Items),
	}, nil
}

func equals(args ...types.MalType) (types.MalType, error) {
	if len(args) != 2 {
		return nil, fmt.Errorf("equals requires 2 args - got %d", len(args))
	}

	return &types.MalBoolean{
		Value: equalsInternal(args[0], args[1]),
	}, nil
}

func equalsInternal(aa types.MalType, bb types.MalType) bool {
	if reflect.TypeOf(aa) != reflect.TypeOf(bb) {
		return false
	}

	switch a := aa.(type) {
	case *types.MalList:
		b := bb.(*types.MalList)
		if len(a.Items) != len(b.Items) {
			return false
		}

		for i := range a.Items {
			if !equalsInternal(a.Items[i], b.Items[i]) {
				return false
			}
		}

	case *types.MalInt:
		b := bb.(*types.MalInt)
		return a.Value == b.Value

	case *types.MalBoolean:
		b := bb.(*types.MalBoolean)
		return a.Value == b.Value

	case *types.MalNil:
		// Nils don't have values, so they're always equal
		return true

	default:
		log.Fatalf("equals unimplemented for type %T", a)
	}

	return true
}

func lt(args ...types.MalType) (types.MalType, error) {
	if err := ValidateNArgs(2, args); err != nil {
		return nil, err
	}
	numbers, err := ArgsToMalInt(args)
	if err != nil {
		return nil, err
	}

	return &types.MalBoolean{
		Value: numbers[0].Value < numbers[1].Value,
	}, nil
}

func lte(args ...types.MalType) (types.MalType, error) {
	if err := ValidateNArgs(2, args); err != nil {
		return nil, err
	}
	numbers, err := ArgsToMalInt(args)
	if err != nil {
		return nil, err
	}

	return &types.MalBoolean{
		Value: numbers[0].Value <= numbers[1].Value,
	}, nil
}

func gt(args ...types.MalType) (types.MalType, error) {
	if err := ValidateNArgs(2, args); err != nil {
		return nil, err
	}
	numbers, err := ArgsToMalInt(args)
	if err != nil {
		return nil, err
	}

	return &types.MalBoolean{
		Value: numbers[0].Value > numbers[1].Value,
	}, nil
}

func gte(args ...types.MalType) (types.MalType, error) {
	if err := ValidateNArgs(2, args); err != nil {
		return nil, err
	}
	numbers, err := ArgsToMalInt(args)
	if err != nil {
		return nil, err
	}

	return &types.MalBoolean{
		Value: numbers[0].Value >= numbers[1].Value,
	}, nil
}

// func list(args ...types.MalType) (types.MalType, error) {

// }
