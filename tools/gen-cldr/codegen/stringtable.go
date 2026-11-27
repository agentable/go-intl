// Package codegen renders generated Go source for internal/cldr/.
package codegen

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

type StringRef struct {
	start  uint32
	length uint32
	table  *StringTable
}

func (r StringRef) String() string {
	if r.table == nil {
		return ""
	}
	return r.table.data.String()[r.start : r.start+r.length]
}

// StringTable interns strings for one generated domain. Use it through its pointer.
type StringTable struct {
	data strings.Builder
	refs map[string]StringRef
}

func NewStringTable() *StringTable {
	return &StringTable{refs: make(map[string]StringRef)}
}

func (t *StringTable) Add(s string) StringRef {
	if ref, ok := t.refs[s]; ok {
		return ref
	}
	ref := StringRef{start: uint32(t.data.Len()), length: uint32(len(s)), table: t}
	t.data.WriteString(s)
	t.refs[s] = ref
	return ref
}

// EmitDataConst writes only the string table body as a chunked const with the
// given name. It emits no package clause, type, or method: the caller owns
// those. This is the const-only payload form a domain data.go uses, where the
// sliceRef type and string() method live in the hand-written decode.go instead.
func (t *StringTable) EmitDataConst(w io.Writer, name string) error {
	return emitStringConst(w, name, t.data.String())
}

// emitStringConst writes a string value as a chunked "" + "…" const so the
// generated source line length stays bounded regardless of payload size.
func emitStringConst(w io.Writer, name, value string) error {
	if _, err := fmt.Fprintf(w, "const %s = \"\" +\n", name); err != nil {
		return err
	}
	for start := 0; start < len(value); start += 64 {
		end := min(start+64, len(value))
		if _, err := fmt.Fprintf(w, "\t%s +\n", strconv.Quote(value[start:end])); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "\t\"\"")
	return err
}
