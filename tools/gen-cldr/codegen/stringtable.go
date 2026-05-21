// Package codegen renders generated Go source for internal/cldr/.
package codegen

import (
	"fmt"
	"io"
	"strconv"
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
	return r.table.data[r.start : r.start+r.length]
}

func (r StringRef) GoLiteral() string {
	return fmt.Sprintf("sliceRef{start: %d, length: %d}", r.start, r.length)
}

type StringTable struct {
	data string
	refs map[string]StringRef
}

func NewStringTable() *StringTable {
	return &StringTable{refs: make(map[string]StringRef)}
}

func (t *StringTable) Add(s string) StringRef {
	if ref, ok := t.refs[s]; ok {
		return ref
	}
	ref := StringRef{start: uint32(len(t.data)), length: uint32(len(s)), table: t}
	t.data += s
	t.refs[s] = ref
	return ref
}

func (t *StringTable) Emit(w io.Writer) error {
	return t.EmitPackage(w, "cldr")
}

func (t *StringTable) EmitPackage(w io.Writer, packageName string) error {
	if _, err := fmt.Fprintf(w, "package %s\n", packageName); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "type sliceRef struct{ start, length uint32 }"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "func (r sliceRef) string() string { return _data[r.start : r.start+r.length] }"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "const _data = \"\" +"); err != nil {
		return err
	}
	for start := 0; start < len(t.data); start += 64 {
		end := min(start+64, len(t.data))
		if _, err := fmt.Fprintf(w, "\t%s +\n", strconv.Quote(t.data[start:end])); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(w, "\t\"\"")
	return err
}
