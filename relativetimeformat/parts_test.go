package relativetimeformat

import (
	"testing"

	"github.com/agentable/go-intl/locale"
)

// Verifies the typed PartType constants match the strings emitted via the
// embedded NumberFormat passthrough, so consumer code can switch on the
// typed value instead of comparing raw strings.
func TestFormatToPartsTypedConstants(t *testing.T) {
	t.Parallel()
	rtf, err := New(locale.List{locale.MustParse("en")}, Options{Numeric: NumericAlways})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	parts, err := rtf.FormatInt64ToParts(-3, Day)
	if err != nil {
		t.Fatalf("FormatInt64ToParts err = %v", err)
	}
	wantTypes := map[PartType]bool{}
	for _, part := range parts {
		switch part.Type {
		case PartLiteral, PartInteger, PartGroup, PartDecimal, PartFraction,
			PartPlusSign, PartMinusSign, PartInfinity, PartNaN:
			wantTypes[part.Type] = true
		default:
			t.Fatalf("unexpected PartType %q in parts: %+v", part.Type, parts)
		}
	}
	if !wantTypes[PartInteger] {
		t.Fatalf("expected at least one PartInteger in parts: %+v", parts)
	}
}
