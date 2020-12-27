package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jamesroutley/mal/impls/go/src/printer"
	"github.com/jamesroutley/mal/impls/go/src/reader"
	"github.com/jamesroutley/mal/impls/go/src/types"
)

func main() {
	for {
		fmt.Printf("user> ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		line = strings.TrimSuffix(line, "\n")
		output, err := (Rep(line))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(output)
	}
}

func Read(s string) (types.MalType, error) {
	return reader.ReadStr(s)
}

func Eval(s types.MalType) types.MalType {
	return s
}

func Print(s types.MalType) string {
	return printer.PrStr(s)
}

func Rep(s string) (string, error) {
	t, err := Read(s)
	if err != nil {
		return "", err
	}
	t = Eval(t)
	s = Print(t)
	return s, nil
}
