package ecma402nf

import (
	"testing"

	"github.com/agentable/go-intl/internal/decimal"
)

// TestCanUseRoundedStringAgreesWithFullPath guards against silent drift between
// the canUseRoundedString fast path and the general fixed path. Input literal
// scale is not part of an Intl mathematical value, so trailing-zero spellings are
// subject to the same byte-equality requirement as every other input.
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
		"1.0", "1.00", "1.50", "-0.00",
	}
	for _, in := range inputs {
		d, err := decimal.ParseString(in)
		if err != nil {
			t.Fatalf("ParseString(%q): %v", in, err)
		}
		formatted := FormatNumericToString(d, opts).Formatted
		full, _ := formatFixedCandidate(d, opts)
		if formatted != full {
			t.Errorf("FormatNumericToString(%q) = %q, full path = %q; fast-path eligibility = %t", in, formatted, full, canUseRoundedString(d, opts))
		}
	}
}

func TestFormatNumericToStringIgnoresInputLiteralScale(t *testing.T) {
	t.Parallel()

	inputs := []string{"1", "1.0", "1.00", "1.50", "-0.00"}
	tests := []struct {
		name string
		opts DigitOptions
		want []string
	}{
		{
			name: "fraction digits",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 0, MaximumFractionDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			want: []string{"1", "1", "1", "1.5", "-0"},
		},
		{
			name: "significant digits",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumSignificantDigits: 1, MaximumSignificantDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "auto"},
			want: []string{"1", "1", "1", "1.5", "-0"},
		},
		{
			name: "more precision",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 2, MaximumFractionDigits: 2, MinimumSignificantDigits: 1, MaximumSignificantDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "morePrecision", TrailingZeroDisplay: "auto"},
			want: []string{"1", "1", "1", "1.5", "-0"},
		},
		{
			name: "less precision",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 2, MaximumFractionDigits: 2, MinimumSignificantDigits: 1, MaximumSignificantDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "lessPrecision", TrailingZeroDisplay: "auto"},
			want: []string{"1.00", "1.00", "1.00", "1.50", "-0.00"},
		},
		{
			name: "strip if integer",
			opts: DigitOptions{MinimumIntegerDigits: 1, MinimumFractionDigits: 2, MaximumFractionDigits: 3, RoundingIncrement: 1, RoundingMode: "halfExpand", RoundingPriority: "auto", TrailingZeroDisplay: "stripIfInteger"},
			want: []string{"1", "1", "1", "1.50", "-0"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for i, input := range inputs {
				d, err := decimal.ParseString(input)
				if err != nil {
					t.Fatalf("ParseString(%q): %v", input, err)
				}
				if got := FormatNumericToString(d, tc.opts).Formatted; got != tc.want[i] {
					t.Errorf("FormatNumericToString(%q) = %q, want %q", input, got, tc.want[i])
				}
			}
		})
	}
}
