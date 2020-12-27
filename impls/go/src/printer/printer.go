package printer

import "github.com/jamesroutley/mal/impls/go/src/types"

func PrStr(t types.MalType) string {
	return t.String()
}
