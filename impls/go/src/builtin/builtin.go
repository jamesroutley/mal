package builtin

import (
	"fmt"

	"github.com/jamesroutley/mal/impls/go/src/types"
)

var Add = &types.MalFunction{
	Func: func(args ...types.MalType) (types.MalType, error) {
		if err := ValidateNArgs(2, args); err != nil {
			return nil, err
		}
		numbers, err := ArgsToMalInt(args)
		if err != nil {
			return nil, err
		}
		return &types.MalInt{
			Value: numbers[0].Value + numbers[1].Value,
		}, nil
	},
}

var Subtract = &types.MalFunction{
	Func: func(args ...types.MalType) (types.MalType, error) {
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
	},
}

var Multiply = &types.MalFunction{
	Func: func(args ...types.MalType) (types.MalType, error) {
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
	},
}

var Divide = &types.MalFunction{
	Func: func(args ...types.MalType) (types.MalType, error) {
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
	},
}

func ValidateNArgs(n int, args []types.MalType) error {
	if actual := len(args); actual != n {
		return fmt.Errorf("Function takes %d args, got %d", n, actual)
	}
	return nil
}

func ArgsToMalInt(args []types.MalType) ([]*types.MalInt, error) {
	numbers := make([]*types.MalInt, len(args))
	for i, arg := range args {
		number, ok := arg.(*types.MalInt)
		if !ok {
			return nil, fmt.Errorf("Could not cast type to int")
		}
		numbers[i] = number
	}
	return numbers, nil
}
