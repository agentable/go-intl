package listformat

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestListFormatResolvedOptionsDefaults(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got := format.ResolvedOptions()
	if got.Locale.String() != "en" {
		t.Fatalf("ResolvedOptions().Locale = %q, want en", got.Locale.String())
	}
	if got.Type != Conjunction {
		t.Fatalf("ResolvedOptions().Type = %q, want %q", got.Type, Conjunction)
	}
	if got.Style != LongStyle {
		t.Fatalf("ResolvedOptions().Style = %q, want %q", got.Style, LongStyle)
	}
}

func TestListFormatRejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	loc := intltest.Locale(t, "en-US")
	tests := []struct {
		name string
		opts Options
	}{
		{
			name: "locale matcher",
			opts: Options{LocaleMatcher: LocaleMatcher("bad")},
		},
		{
			name: "type",
			opts: Options{Type: Type("bad")},
		},
		{
			name: "style",
			opts: Options{Style: Style("bad")},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
		})
	}
}

func TestListFormatFormatConjunctionLong(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	if got := format.Format([]string{"A", "B", "C"}); got != "A, B, and C" {
		t.Fatalf("Format(A, B, C) = %q, want %q", got, "A, B, and C")
	}
}

func TestListFormatFormatBoundaryLengths(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name string
		list []string
		want string
	}{
		{name: "empty", list: nil, want: ""},
		{name: "single", list: []string{"A"}, want: "A"},
		{name: "pair", list: []string{"A", "B"}, want: "A and B"},
		{name: "many", list: []string{"A", "B", "C", "D"}, want: "A, B, C, and D"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := format.Format(tc.list); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.list, got, tc.want)
			}
		})
	}
}

func TestListFormatFormatTreatsElementPlaceholdersAsText(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	if got := format.Format([]string{"{1}", "B"}); got != "{1} and B" {
		t.Fatalf("Format({1}, B) = %q, want %q", got, "{1} and B")
	}
}

func TestListFormatFormatEqualsFormatToPartsJoin(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name string
		list []string
	}{
		{name: "empty", list: nil},
		{name: "single", list: []string{"A"}},
		{name: "pair", list: []string{"A", "B"}},
		{name: "many", list: []string{"A", "B", "C", "D"}},
		{name: "placeholder text", list: []string{"{1}", "B"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := format.Format(tc.list)
			if want := listPartsText(format.FormatToParts(tc.list)); got != want {
				t.Fatalf("Format(%v) = %q, want joined FormatToParts %q", tc.list, got, want)
			}
		})
	}
}

func TestListFormatFormatUsesPartsOwner(t *testing.T) {
	t.Parallel()

	parsed, err := parser.ParseFile(token.NewFileSet(), "format.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	format := findMethodDecl(parsed, "Format")
	if format == nil {
		t.Fatal("Format method not found")
	}
	if !methodCalls(format, "FormatToParts") {
		t.Fatal("Format must derive output from FormatToParts to avoid string/parts drift")
	}
}

func TestListFormatFormatToPartsPair(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got := format.FormatToParts([]string{"A", "B"})
	want := []Part{
		{Type: PartElement, Value: "A"},
		{Type: PartLiteral, Value: " and "},
		{Type: PartElement, Value: "B"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(A, B) = %#v, want %#v", got, want)
	}
}

func TestListFormatFormatToPartsBoundaryLengths(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name string
		list []string
		want []Part
	}{
		{name: "empty", list: nil, want: nil},
		{
			name: "single",
			list: []string{"A"},
			want: []Part{{Type: PartElement, Value: "A"}},
		},
		{
			name: "many",
			list: []string{"A", "B", "C"},
			want: []Part{
				{Type: PartElement, Value: "A"},
				{Type: PartLiteral, Value: ", "},
				{Type: PartElement, Value: "B"},
				{Type: PartLiteral, Value: ", and "},
				{Type: PartElement, Value: "C"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := format.FormatToParts(tc.list); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("FormatToParts(%v) = %#v, want %#v", tc.list, got, tc.want)
			}
		})
	}
}

func listPartsText(parts []Part) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func findMethodDecl(file *ast.File, name string) *ast.FuncDecl {
	for _, decl := range file.Decls {
		decl, ok := decl.(*ast.FuncDecl)
		if ok && decl.Recv != nil && decl.Name.Name == name {
			return decl
		}
	}
	return nil
}

func methodCalls(fn *ast.FuncDecl, name string) bool {
	found := false
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if ok && selector.Sel.Name == name {
			found = true
			return false
		}
		return true
	})
	return found
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "fr-FR"), intltest.Locale(t, "en-US"), intltest.Locale(t, "zh-Hans-CN")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"fr-FR", "en-US", "zh-Hans-CN"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}
