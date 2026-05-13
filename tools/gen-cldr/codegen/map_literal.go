package codegen

import (
	"maps"
	"slices"
	"strconv"
	"strings"
)

func stringKeyedMapLiteral[V any](typ string, values map[string]V, valueLiteral func(V) string) string {
	if len(values) == 0 {
		return "nil"
	}
	var b strings.Builder
	b.WriteString(typ)
	b.WriteString("{")
	for _, key := range slices.Sorted(maps.Keys(values)) {
		b.WriteString(strconv.Quote(key))
		b.WriteString(": ")
		b.WriteString(valueLiteral(values[key]))
		b.WriteString(", ")
	}
	b.WriteString("}")
	return b.String()
}
