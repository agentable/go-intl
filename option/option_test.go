package option_test

import (
	"testing"

	"github.com/agentable/go-intl/option"
)

func TestInt(t *testing.T) {
	t.Parallel()
	p := option.Int(7)
	if p == nil || *p != 7 {
		t.Fatalf("Int(7) = %v, want pointer to 7", p)
	}
}

func TestBool(t *testing.T) {
	t.Parallel()
	p := option.Bool(true)
	if p == nil || *p != true {
		t.Fatalf("Bool(true) = %v, want pointer to true", p)
	}
}

// stringKind is a formatter-style typed enum, to exercise the ~string generic.
type stringKind string

func TestString(t *testing.T) {
	t.Parallel()
	p := option.String(stringKind("decimal"))
	if p == nil || *p != "decimal" {
		t.Fatalf("String(decimal) = %v, want pointer to \"decimal\"", p)
	}
}
