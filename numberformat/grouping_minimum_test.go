package numberformat

import (
	"testing"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/intltest"
)

func TestLocaleGroupingMinimum(t *testing.T) {
	t.Parallel()
	for _, tag := range []string{"es", "pl", "en", "hi"} {
		for _, policy := range []string{"auto", "always", "min2", "false"} {
			t.Run(tag+"/"+policy, func(t *testing.T) {
				t.Parallel()
				f, err := New(intltest.LocaleList(t, tag), Options{UseGrouping: new(policy), MaximumFractionDigits: new(0)})
				if err != nil {
					t.Fatal(err)
				}
				sep := ","
				if tag == "es" {
					sep = "."
				}
				if tag == "pl" {
					sep = "\u00a0"
				}
				for _, tc := range []struct {
					value  float64
					digits string
				}{{999, "999"}, {1000, "1000"}, {9999, "9999"}, {10000, "10000"}, {999.5, "1000"}, {9999.5, "10000"}, {123456, "123456"}} {
					want := tc.digits
					threshold := 4
					if policy == "min2" || policy == "auto" && (tag == "es" || tag == "pl") {
						threshold = 5
					}
					if policy != "false" && len(want) >= threshold {
						if tag == "hi" && len(want) == 6 {
							want = "1,23,456"
						} else {
							want = want[:len(want)-3] + sep + want[len(want)-3:]
						}
					}
					got := f.Format(Float(tc.value))
					if got != want {
						t.Errorf("Format(%v) = %q, want %q", tc.value, got, want)
					}
					parts := f.FormatToParts(Float(tc.value))
					joined := ""
					for _, p := range parts {
						joined += p.Value
					}
					if joined != want {
						t.Errorf("parts = %q, want %q", joined, want)
					}
				}
				parts, err := f.FormatRangeToParts(Int(1000), Int(10000))
				if err != nil {
					t.Fatal(err)
				}
				text, err := f.FormatRange(Int(1000), Int(10000))
				if err != nil {
					t.Fatal(err)
				}
				joined := ""
				start, end := false, false
				for _, p := range parts {
					joined += p.Value
					start = start || p.Source == "startRange"
					end = end || p.Source == "endRange"
				}
				if joined != text || !start || !end {
					t.Fatalf("range partition = %v, text %q", parts, text)
				}
			})
		}
	}
}

func TestGroupingWithThreeDigitMinimum(t *testing.T) {
	t.Parallel()
	// A three-digit minimum needs six displayed digits before the first separator.
	grouping := digitGrouping{primary: 3, secondary: 3, minimum: 3}
	symbols := cldrnumber.NumberSymbols{Group: ","}
	for _, tc := range []struct{ in, want string }{{"1000", "1000"}, {"10000", "10000"}, {"100000", "100,000"}, {"1000000", "1,000,000"}} {
		if got := partsText(appendDecimalParts(nil, tc.in, grouping, symbols)); got != tc.want {
			t.Errorf("grouping %q = %q, want %q", tc.in, got, tc.want)
		}
	}
}
