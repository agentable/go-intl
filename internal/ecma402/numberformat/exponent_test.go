package ecma402nf

import (
	"testing"

	"github.com/agentable/go-intl/internal/decimal"
)

func TestScientificExponent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		in          string
		maxFrac     int
		engineering bool
		want        int
	}{
		{name: "zero", in: "0", maxFrac: 3, want: 0},
		{name: "scientific positive integer", in: "12345", maxFrac: 3, want: 4},
		{name: "scientific small decimal", in: "0.0123", maxFrac: 5, want: -2},
		{name: "scientific negative input", in: "-987", maxFrac: 3, want: 2},
		{name: "engineering positive integer", in: "12345", maxFrac: 3, engineering: true, want: 3},
		{name: "engineering small decimal", in: "0.0123", maxFrac: 5, engineering: true, want: -3},
		{name: "engineering negative magnitude", in: "0.000123", maxFrac: 7, engineering: true, want: -6},
		// Rounding carry: mantissa rounds up a magnitude, so the exponent is
		// re-derived from magnitude+1.
		{name: "scientific carry 999", in: "999", maxFrac: 0, want: 3},
		{name: "scientific carry 999500", in: "999500", maxFrac: 0, want: 6},
		{name: "engineering carry 999500", in: "999500", maxFrac: 0, engineering: true, want: 6},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			d, err := decimal.ParseString(tc.in)
			if err != nil {
				t.Fatal(err)
			}
			opts := DigitOptions{
				MinimumIntegerDigits:  1,
				MaximumFractionDigits: tc.maxFrac,
				RoundingIncrement:     1,
				RoundingMode:          "halfExpand",
				RoundingPriority:      "auto",
				TrailingZeroDisplay:   "auto",
			}
			got, ok := ScientificExponent(d, inferredResolvedDigits(opts), tc.engineering)
			if !ok {
				t.Fatalf("ScientificExponent(%s, %t) ok = false, want true", tc.in, tc.engineering)
			}
			if got != tc.want {
				t.Fatalf("ScientificExponent(%s, %t) = %d, want %d", tc.in, tc.engineering, got, tc.want)
			}
		})
	}
}

func TestScientificExponentRejectsNonFinite(t *testing.T) {
	t.Parallel()

	for _, d := range []decimal.Decimal{decimal.NaNValue, decimal.PosInfinity, decimal.NegInfinity} {
		t.Run(d.String(), func(t *testing.T) {
			t.Parallel()

			if got, ok := ScientificExponent(d, inferredResolvedDigits(DigitOptions{MaximumFractionDigits: 3, RoundingIncrement: 1}), false); ok {
				t.Fatalf("ScientificExponent(%s, false) = %d, true; want ok false", d.String(), got)
			}
		})
	}
}
