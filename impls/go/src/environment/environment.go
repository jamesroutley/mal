package environment

import (
	"fmt"

	"github.com/jamesroutley/mal/impls/go/src/types"
)

type Env struct {
	Outer *Env
	Data  map[string]types.MalType
}

func NewEnv() *Env {
	return &Env{
		Outer: nil,
		Data:  map[string]types.MalType{},
	}
}

func (e *Env) Set(key string, value types.MalType) {
	e.Data[key] = value
}

func (e *Env) Find(key string) (*Env, error) {
	if e == nil {
		return nil, fmt.Errorf("`%s` is undefined", key)
	}
	if _, ok := e.Data[key]; ok {
		return e, nil
	}
	return e.Outer.Find(key)
}

func (e *Env) Get(key string) (types.MalType, error) {
	env, err := e.Find(key)
	if err != nil {
		return nil, err
	}
	return env.Data[key], nil
}

func (e *Env) ChildEnv() *Env {
	return &Env{
		Outer: e,
		Data:  map[string]types.MalType{},
	}
}
