package durationformat

import (
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

// TestFormatEqualsConcatOfParts locks the ECMA-402 invariant that Format is the
// concatenation of FormatToParts' values. Format now derives from the parts walk,
// so this can never drift; the cases cover fractional, negative, mixed-style, and
// numeric/digital shapes that previously lived only in hand-written expectations.
func TestFormatEqualsConcatOfParts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		opts Options
		dur  Duration
	}{
		{name: "default long", opts: Options{}, dur: Duration{Hours: 1, Minutes: 30, Seconds: 15}},
		{name: "negative", opts: Options{}, dur: Duration{Hours: -1, Minutes: -30}},
		{name: "digital numeric", opts: Options{Style: stringPtr(DigitalStyle)}, dur: Duration{Hours: 1, Minutes: 2, Seconds: 3}},
		{name: "fractional seconds", opts: Options{Style: stringPtr(DigitalStyle)}, dur: Duration{Seconds: 1, Milliseconds: 234}},
		{name: "mixed styles", opts: Options{Years: stringPtr(LongUnitStyle), Months: stringPtr(NarrowUnitStyle), Days: stringPtr(ShortUnitStyle)}, dur: Duration{Years: 1, Months: 2, Days: 3}},
		{name: "single unit", opts: Options{}, dur: Duration{Days: 5}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			got, err := f.Format(tc.dur)
			if err != nil {
				t.Fatal(err)
			}
			parts, err := f.FormatToParts(tc.dur)
			if err != nil {
				t.Fatal(err)
			}
			var b strings.Builder
			for _, p := range parts {
				b.WriteString(p.Value)
			}
			if want := b.String(); got != want {
				t.Errorf("Format = %q, concat(FormatToParts) = %q", got, want)
			}
		})
	}
}
