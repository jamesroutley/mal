package types

import (
	"fmt"
	"strconv"
	"strings"
)

type MalType interface {
	String() string
}

type EnvType interface {
	Set(string, MalType)
	Get(string) (MalType, error)
	Find(string) (EnvType, error)
}

type MalList struct {
	Items []MalType
}

func (l *MalList) String() string {
	itemStrings := make([]string, len(l.Items))
	for i, item := range l.Items {
		itemStrings[i] = item.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(itemStrings, " "))
}

type MalInt struct {
	Value int
}

func (i *MalInt) String() string {
	return strconv.Itoa(i.Value)
}

type MalSymbol struct {
	Value string
}

func (s *MalSymbol) String() string {
	return s.Value
}

type MalFunction struct {
	Func              func(args ...MalType) (MalType, error)
	TailCallOptimised bool
	AST               MalType
	Params            []*MalSymbol
	Env               EnvType
}

func (f *MalFunction) String() string {
	return "#<function>"
}

type MalBoolean struct {
	Value bool
}

func (b *MalBoolean) String() string {
	if b.Value {
		return "true"
	}
	return "false"
}

type MalNil struct{}

func (n *MalNil) String() string {
	return "nil"
}
