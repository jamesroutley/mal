package types

import (
	"fmt"
	"strconv"
	"strings"
)

type MalType interface {
	String() string
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
	Func func(args ...MalType) (MalType, error)
}

func (f *MalFunction) String() string {
	return "function"
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
