package core

import "github.com/jamesroutley/mal/impls/go/src/types"

func add(args ...types.MalType) (types.MalType, error) {
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
