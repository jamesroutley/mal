package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
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
		fmt.Println(Rep(line))
	}
}

func Read(s string) string {
	return s
}

func Eval(s string) string {
	return s
}

func Print(s string) string {
	return s
}

func Rep(s string) string {
	s = Read(s)
	s = Eval(s)
	s = Print(s)
	return s
}
