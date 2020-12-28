package environment

import (
	"fmt"
	"log"

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

func NewChildEnv(parent *Env, binds []*types.MalSymbol, exprs []types.MalType) *Env {
	env := &Env{
		Outer: parent,
		Data:  map[string]types.MalType{},
	}
	if len(binds) != len(exprs) {
		// TODO: return this?
		log.Fatal("can't create env - num binds != num exprs")
	}
	for i := range binds {
		env.Set(binds[i].Value, exprs[i])
	}
	return env
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
