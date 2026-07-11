package numberformat_test

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

// TestScientificExponentRoundingCarry pins the ECMA-402 ComputeExponent carry
// recheck: when digit rounding pushes the mantissa to 10 (999 -> "10E2"), the
// exponent must be re-derived from the incremented magnitude ("1E3"). Values
// verified against Node 26. This is the scientific/engineering analogue of the
// carry the compact path already handled.
func TestScientificExponentRoundingCarry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		notation numberformat.Notation
		maxFrac  int
		value    float64
		want     string
	}{
		{name: "sci 999 carries to 1E3", notation: numberformat.ScientificNotation, maxFrac: 0, value: 999, want: "1E3"},
		{name: "sci 999500 carries to 1E6", notation: numberformat.ScientificNotation, maxFrac: 0, value: 999500, want: "1E6"},
		{name: "eng 999500 carries to 1E6", notation: numberformat.EngineeringNotation, maxFrac: 0, value: 999500, want: "1E6"},
		{name: "sci 123 no carry", notation: numberformat.ScientificNotation, maxFrac: 0, value: 123, want: "1E2"},
		{name: "eng 12345 no carry", notation: numberformat.EngineeringNotation, maxFrac: 0, value: 12345, want: "12E3"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			notation := string(tc.notation)
			maxFrac := tc.maxFrac
			nf, err := numberformat.New(locale.List{intltest.Locale(t, "en")}, numberformat.Options{
				Notation:              &notation,
				MaximumFractionDigits: &maxFrac,
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := nf.Format(numberformat.Float(tc.value)); got != tc.want {
				t.Errorf("Format(%v) [%s maxFrac=%d] = %q, want %q", tc.value, tc.notation, tc.maxFrac, got, tc.want)
			}
		})
	}
}
