package pluralrules

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func intPtr(v int) *int {
	return &v
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	digits := 2
	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumFractionDigits: &digits})
	if err != nil {
		t.Fatal(err)
	}
	digits = 0

	if got := mustCategory(rules.Select(Int(int64(1)))); got != Other {
		t.Fatalf("SelectInt(1) = %s, want %s", got, Other)
	}
}
