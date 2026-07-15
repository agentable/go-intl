package pluralrules

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func intPtr(v int) *int {
	return &v
}

func stringPtr[T ~string](v T) *string {
	value := string(v)
	return &value
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	digits := 2
	typ := string(Ordinal)
	notation := string(CompactNotation)
	compactDisplay := string(LongCompactDisplay)
	roundingMode := string(TruncRoundingMode)
	rules, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Type:                  &typ,
		MinimumFractionDigits: &digits,
		Notation:              &notation,
		CompactDisplay:        &compactDisplay,
		RoundingMode:          &roundingMode,
	})
	if err != nil {
		t.Fatal(err)
	}
	typ = string(Cardinal)
	digits = 0
	notation = string(StandardNotation)
	compactDisplay = string(ShortCompactDisplay)
	roundingMode = string(HalfExpandRoundingMode)

	resolved := rules.ResolvedOptions()
	if resolved.Type != Ordinal ||
		resolved.Notation != CompactNotation ||
		resolved.CompactDisplay == nil ||
		*resolved.CompactDisplay != LongCompactDisplay ||
		resolved.RoundingMode != TruncRoundingMode {
		t.Fatalf("ResolvedOptions() = %#v, want copied string option values", resolved)
	}
	if got := rules.Select(Int(int64(1))); got != One {
		t.Fatalf("SelectInt(1) = %s, want %s", got, One)
	}
}
