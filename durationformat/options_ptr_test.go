package durationformat

import (
	"testing"

	"github.com/agentable/go-intl/locale"
)

func intPtr(v int) *int {
	return &v
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	digits := 0
	format, err := New(locale.List{locale.MustParse("en")}, Options{FractionalDigits: &digits})
	if err != nil {
		t.Fatal(err)
	}
	digits = 9

	if got := format.ResolvedOptions().FractionalDigits; got == nil || *got != 0 {
		t.Fatalf("ResolvedOptions().FractionalDigits = %v, want 0", got)
	}
}
