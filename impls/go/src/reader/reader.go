package reader

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jamesroutley/mal/impls/go/src/types"
)

type Reader struct {
	Tokens   []string
	Position int
}

func NewReader(tokens []string) *Reader {
	return &Reader{
		Tokens:   tokens,
		Position: 0,
	}
}

func (r *Reader) Peek() (string, error) {
	if r.Position == len(r.Tokens) {
		return "", fmt.Errorf("EOF")
	}
	return strings.Trim(r.Tokens[r.Position], " ,\n\t"), nil
}

func (r *Reader) Next() (string, error) {
	if r.Position == len(r.Tokens) {
		return "", fmt.Errorf("EOF")
	}
	current := strings.Trim(r.Tokens[r.Position], " ,\n\t")
	r.Position++
	return current, nil
}

func ReadStr(s string) (types.MalType, error) {
	tokens := Tokenize(s)
	reader := NewReader(tokens)
	return ReadForm(reader)
}

func Tokenize(s string) []string {
	re := regexp.MustCompile(`[\s,]*(~@|[\[\]{}()'` + "`" + `~^@]|"(?:\\.|[^\\"])*"?|;.*|[^\s\[\]{}('"` + "`" + `,;)]*)`)
	return re.FindAllString(s, -1)
}

func ReadForm(reader *Reader) (types.MalType, error) {
	token, err := reader.Peek()
	if err != nil {
		return nil, err
	}
	switch token {
	case "(":
		// Increment the position pointer
		_, err = reader.Next()
		if err != nil {
			return nil, err
		}
		return ReadList(reader)
	default:
		return ReadAtom(reader)
	}
}

func ReadList(reader *Reader) (types.MalType, error) {
	var items []types.MalType
	for {
		// TODO: error case when we hit file without closing bracket
		tok, err := reader.Peek()
		if err != nil {
			return nil, err
		}
		if tok == ")" {
			// Increment the position pointer
			_, err := reader.Next()
			if err != nil {
				return nil, err
			}
			return &types.MalList{
				Items: items,
			}, nil
		}
		item, err := ReadForm(reader)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
}

func ReadAtom(reader *Reader) (types.MalType, error) {
	token, err := reader.Next()
	if err != nil {
		return nil, err
	}

	if num, err := strconv.Atoi(token); err == nil {
		return &types.MalInt{
			Value: num,
		}, nil
	}

	if token == "true" {
		return &types.MalBoolean{Value: true}, nil
	}
	if token == "false" {
		return &types.MalBoolean{Value: false}, nil
	}

	if token == "nil" {
		return &types.MalNil{}, nil
	}

	return &types.MalSymbol{
		Value: token,
	}, nil
}

func DebugType(m types.MalType) {
	fmt.Println(debugType(m, 0))
}

func debugType(m types.MalType, indent int) string {
	switch tok := m.(type) {
	case *types.MalList:
		itemStrings := make([]string, len(tok.Items))
		for i, item := range tok.Items {
			itemStrings[i] = debugType(item, 0)
		}
		return fmt.Sprintf("(%s)", strings.Join(itemStrings, " "))
	case *types.MalInt:
		return fmt.Sprintf("int:%d ", tok.Value)
	case *types.MalSymbol:
		return fmt.Sprintf("symbol:`%s` ", tok.Value)
	case *types.MalFunction:
		return fmt.Sprintf("function")
	default:
		panic("debug print not implemented for this type")
	}
}

func DebugTokens(s string) {
	tokens := Tokenize(s)
	for _, tok := range tokens {
		fmt.Printf("`%s`\n", tok)
	}
}
