package listformat

import (
	"errors"
	"strings"
	"testing"
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
