package numberformat

import (
	"testing"

	"github.com/agentable/go-intl/locale"
)

func intPtr(v int) *int {
	return &v
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	digits := 2
	format, err := New(locale.List{locale.MustParse("en")}, Options{
		MinimumFractionDigits: &digits,
		MaximumFractionDigits: &digits,
	})
	if err != nil {
		t.Fatal(err)
	}
	digits = 0

	if got := format.Format(Int(1)); got != "1.00" {
		t.Fatalf("FormatInt64(1) = %q, want 1.00", got)
	}
}
