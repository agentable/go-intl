package listformat

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
)

// TestCompileListTemplatePanicsOnMalformed locks the Must* invariant: the
// generator guarantees embedded list patterns are well-formed, so a parse
// failure is corrupt data and must fail loudly rather than degrade to empty
// output. Well-formed input must never panic.
func TestCompileListTemplatePanicsOnMalformed(t *testing.T) {
	t.Parallel()

	defer func() {
		recovered := recover()
		err, ok := recovered.(error)
		if !ok {
			t.Fatalf("panic value = %#v, want error", recovered)
		}
		if !strings.HasPrefix(err.Error(), "listformat: malformed embedded CLDR list pattern: ") {
			t.Fatalf("panic error = %q, want listformat attribution", err)
		}
		if errors.Unwrap(err) == nil {
			t.Fatalf("panic error = %v, want wrapped parser error", err)
		}
	}()
	compileListTemplate("{") // unmatched placeholder: PartitionPattern rejects it
}

func TestCompileListTemplateAcceptsWellFormed(t *testing.T) {
	t.Parallel()

	if got := compileListTemplate("{0} and {1}"); len(got) == 0 {
		t.Fatal("compileListTemplate returned empty pattern for valid input")
	}
}

func TestListFormatPreservesTemplateOrderAndParts(t *testing.T) {
	t.Parallel()
	format := &ListFormat{templates: compileListTemplates(cldrlist.ListPattern{
		Pair: "P{1}/{0}Q", Start: "S{0}<{1}>s", Middle: "M{1}|{0}m", End: "E{1}+{0}e",
	}, "en")}
	for _, tc := range []struct {
		list  []string
		text  string
		parts []Part
	}{
		{nil, "", []Part{}},
		{[]string{"東京"}, "東京", []Part{{PartElement, "東京"}}},
		{[]string{"", "東京"}, "P東京/Q", []Part{{PartLiteral, "P"}, {PartElement, "東京"}, {PartLiteral, "/"}, {PartElement, ""}, {PartLiteral, "Q"}}},
		{[]string{"A", "B", "東京"}, "SA<E東京+Be>s", []Part{
			{PartLiteral, "S"}, {PartElement, "A"}, {PartLiteral, "<"}, {PartLiteral, "E"}, {PartElement, "東京"}, {PartLiteral, "+"}, {PartElement, "B"}, {PartLiteral, "e"}, {PartLiteral, ">s"},
		}},
		{[]string{"A", "B", "C", "東京"}, "SA<ME東京+Ce|Bm>s", []Part{
			{PartLiteral, "S"}, {PartElement, "A"}, {PartLiteral, "<"}, {PartLiteral, "M"}, {PartLiteral, "E"}, {PartElement, "東京"}, {PartLiteral, "+"}, {PartElement, "C"}, {PartLiteral, "e"}, {PartLiteral, "|"}, {PartElement, "B"}, {PartLiteral, "m"}, {PartLiteral, ">s"},
		}},
	} {
		if got := format.Format(tc.list); got != tc.text {
			t.Errorf("Format(%q) = %q, want %q", tc.list, got, tc.text)
		}
		if got := format.FormatToParts(tc.list); !reflect.DeepEqual(got, tc.parts) {
			t.Errorf("FormatToParts(%q) = %#v, want %#v", tc.list, got, tc.parts)
		}
	}
}
