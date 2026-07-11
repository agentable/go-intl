package ecma402nf

import (
	"testing"

	"github.com/agentable/go-intl/internal/decimal"
)

// TestCanUseRoundedStringAgreesWithFullPath guards against silent drift between
// the canUseRoundedString fast path and the general fixed path. For every input
// where the fast path fires and the decimal carries no trailing fraction zeros,
// the two must produce byte-identical output; if someone edits the rounding or
// trimming logic without updating the fast path, this fails.
//
// Trailing-zero inputs are excluded on purpose: the fast path preserves them
// (PluralRules operands depend on the source visible fraction digits), while the
// general path would trim them. That single intentional difference is the reason
// the fast path exists and cannot be deleted.
func TestCanUseRoundedStringAgreesWithFullPath(t *testing.T) {
	t.Parallel()

	opts := DigitOptions{
		MinimumIntegerDigits:  1,
		MinimumFractionDigits: 0,
		MaximumFractionDigits: 5,
		RoundingIncrement:     1,
		RoundingMode:          "halfExpand",
		RoundingPriority:      "auto",
		TrailingZeroDisplay:   "auto",
	}

	inputs := []string{
		"0", "1", "-1", "42", "1.5", "-2.25", "123.456", "0.001",
		"1000000", "-0.5", "3.14159", "999.99", "0.12345",
	}
	for _, in := range inputs {
		d, err := decimal.ParseString(in)
		if err != nil {
			t.Fatalf("ParseString(%q): %v", in, err)
		}
		if !canUseRoundedString(d, opts) {
			t.Fatalf("canUseRoundedString(%q) = false, expected fast path to fire", in)
		}
		fast := FormatNumericToString(d, opts).Formatted
		full, _ := formatFixedCandidate(d, opts)
		if fast != full {
			t.Errorf("fast path %q != full path %q for %q", fast, full, in)
		}
	}
}
