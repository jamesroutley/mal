package core

import (
	"fmt"

	"github.com/jamesroutley/mal/impls/go/src/types"
)

func add(args ...types.MalType) (types.MalType, error) {
	switch a := args[0].(type) {
	case *types.MalInt:
		sum := a.Value
		for _, arg := range args[1:] {
			b, ok := arg.(*types.MalInt)
			if !ok {
				return nil, fmt.Errorf("addition between different types")
			}
			sum += b.Value
		}
		return &types.MalInt{
			Value: sum,
		}, nil
	case *types.MalString:
		sum := a.Value
		for _, arg := range args[1:] {
			b, ok := arg.(*types.MalString)
			if !ok {
				return nil, fmt.Errorf("addition between different types")
			}
			sum += b.Value
		}
		return &types.MalString{
			Value: sum,
		}, nil
	}
	return nil, fmt.Errorf("unsupported first arg to +")
}

func subtract(args ...types.MalType) (types.MalType, error) {
	if err := ValidateNArgs(2, args); err != nil {
		return nil, err
	}
	numbers, err := ArgsToMalInt(args)
	if err != nil {
		return nil, err
	}
	return &types.MalInt{
		Value: numbers[0].Value - numbers[1].Value,
	}, nil
}

func multiply(args ...types.MalType) (types.MalType, error) {
	if err := ValidateNArgs(2, args); err != nil {
		return nil, err
	}
	numbers, err := ArgsToMalInt(args)
	if err != nil {
		return nil, err
	}
	return &types.MalInt{
		Value: numbers[0].Value * numbers[1].Value,
	}, nil
}

func divide(args ...types.MalType) (types.MalType, error) {
	if err := ValidateNArgs(2, args); err != nil {
		return nil, err
	}
	numbers, err := ArgsToMalInt(args)
	if err != nil {
		return nil, err
	}
	return &types.MalInt{
		Value: numbers[0].Value / numbers[1].Value,
	}, nil
}
