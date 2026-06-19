package durationformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

// Verifies the typed PartType constants match the strings emitted via the
// embedded NumberFormat passthrough and list-pattern literals.
func TestFormatToPartsTypedConstants(t *testing.T) {
	t.Parallel()
	df, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Hours:   LongUnitStyle,
		Minutes: LongUnitStyle,
	})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	parts, err := df.FormatToParts(Duration{Hours: 1, Minutes: 30})
	if err != nil {
		t.Fatalf("FormatToParts err = %v", err)
	}
	seenUnit, seenInteger := false, false
	for _, part := range parts {
		switch part.Type {
		case PartLiteral, PartInteger, PartGroup, PartDecimal, PartFraction,
			PartPlusSign, PartMinusSign, PartInfinity, PartNaN, PartUnit:
			if part.Type == PartUnit {
				seenUnit = true
			}
			if part.Type == PartInteger {
				seenInteger = true
			}
		default:
			t.Fatalf("unexpected PartType %q in parts: %+v", part.Type, parts)
		}
	}
	if !seenUnit {
		t.Fatalf("expected at least one PartUnit in parts: %+v", parts)
	}
	if !seenInteger {
		t.Fatalf("expected at least one PartInteger in parts: %+v", parts)
	}
}
