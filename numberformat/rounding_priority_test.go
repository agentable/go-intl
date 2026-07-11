package numberformat_test

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

// TestRoundingPriorityTieBreak pins the ECMA-402 roundingPriority tie-break to
// the normative RoundingMagnitude comparison (numberformat.html §"FormatNumericToString"),
// not numeric distance. When the fixed and significant candidates round to the
// same value (a distance tie) the spec still discriminates by which candidate
// rounds at the deeper decimal place. Values verified against Node 26.
func TestRoundingPriorityTieBreak(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		priority string
		maxSig   int
		minFrac  int
		maxFrac  int
		value    float64
		want     string
	}{
		// Distance tie: 1.2 == 1.200. morePrecision must keep the deeper-rounding
		// fixed candidate ("1.200"); the old abs-distance code returned "1.2".
		{name: "tie morePrecision", priority: "morePrecision", maxSig: 2, minFrac: 3, maxFrac: 3, value: 1.2, want: "1.200"},
		{name: "tie lessPrecision", priority: "lessPrecision", maxSig: 2, minFrac: 3, maxFrac: 3, value: 1.2, want: "1.2"},
		// Non-tie: significant is strictly more precise; exercises the other branch.
		{name: "sig-more morePrecision", priority: "morePrecision", maxSig: 5, maxFrac: 1, value: 1.23456, want: "1.2346"},
		{name: "sig-more lessPrecision", priority: "lessPrecision", maxSig: 5, maxFrac: 1, value: 1.23456, want: "1.2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts := numberformat.Options{
				MaximumSignificantDigits: intptr(tc.maxSig),
				MaximumFractionDigits:    intptr(tc.maxFrac),
				RoundingPriority:         strptr(tc.priority),
			}
			if tc.minFrac != 0 {
				opts.MinimumFractionDigits = intptr(tc.minFrac)
			}
			nf, err := numberformat.New(locale.List{intltest.Locale(t, "en")}, opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := nf.Format(numberformat.Float(tc.value)); got != tc.want {
				t.Errorf("Format(%v) [%s maxSig=%d maxFrac=%d minFrac=%d] = %q, want %q",
					tc.value, tc.priority, tc.maxSig, tc.maxFrac, tc.minFrac, got, tc.want)
			}
		})
	}
}

func intptr(v int) *int       { return &v }
func strptr(v string) *string { return &v }
