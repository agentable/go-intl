package codegen

import (
	"bytes"
	"errors"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"
)

func TestStringTableAddDeduplicatesRefs(t *testing.T) {
	t.Parallel()

	var zero StringRef
	if got := zero.String(); got != "" {
		t.Fatalf("zero StringRef.String() = %q, want empty string", got)
	}

	table := NewStringTable()
	first := table.Add("alpha")
	second := table.Add("beta")
	again := table.Add("alpha")

	if got := first.String(); got != "alpha" {
		t.Errorf(`first.String() = %q, want "alpha"`, got)
	}
	if got := second.String(); got != "beta" {
		t.Errorf(`second.String() = %q, want "beta"`, got)
	}
	if again != first {
		t.Errorf("Add(%q) returned %+v, want original ref %+v", "alpha", again, first)
	}
	if got := table.data; got != "alphabeta" {
		t.Errorf("StringTable data after duplicate add = %q, want %q", got, "alphabeta")
	}
}

func TestStringTableEmitDataConst(t *testing.T) {
	t.Parallel()

	want := `quote:" slash:\ newline:` + "\n" + strings.Repeat("x", 70)
	table := NewStringTable()
	table.Add(want)

	var buf bytes.Buffer
	if err := table.EmitDataConst(&buf, "_data"); err != nil {
		t.Fatalf("EmitDataConst() error = %v", err)
	}

	got := constValue(t, "package data\n"+buf.String(), "_data")
	if got != want {
		t.Errorf("emitted _data = %q, want %q", got, want)
	}

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if len(lines) < 4 {
		t.Fatalf("EmitDataConst() lines = %q, want chunked const output", lines)
	}
	for _, line := range lines[1 : len(lines)-1] {
		if len(line) > 80 {
			t.Errorf("EmitDataConst() payload line length = %d, want <= 80: %q", len(line), line)
		}
	}
}

func TestStringTableEmitDataConstPropagatesWriterError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("write stopped")
	table := NewStringTable()
	table.Add("value")

	err := table.EmitDataConst(errorWriter{err: wantErr}, "_data")
	if !errors.Is(err, wantErr) {
		t.Fatalf("EmitDataConst() error = %v, want %v", err, wantErr)
	}
}

func constValue(t *testing.T, src, name string) string {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "data.go", src, 0)
	if err != nil {
		t.Fatalf("parser.ParseFile() error = %v\n%s", err, src)
	}
	info := types.Info{Defs: map[*ast.Ident]types.Object{}}
	if _, err := new(types.Config).Check("data", fset, []*ast.File{file}, &info); err != nil {
		t.Fatalf("types.Config.Check() error = %v\n%s", err, src)
	}
	for ident, obj := range info.Defs {
		if ident.Name != name {
			continue
		}
		c, ok := obj.(*types.Const)
		if !ok {
			t.Fatalf("%s object = %T, want *types.Const", name, obj)
		}
		return constant.StringVal(c.Val())
	}
	t.Fatalf("%s const not found in generated source:\n%s", name, src)
	return ""
}

type errorWriter struct {
	err error
}

func (w errorWriter) Write([]byte) (int, error) {
	return 0, w.err
}
