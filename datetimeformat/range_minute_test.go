package datetimeformat

import (
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

// TestFormatRangeNumericMinute pins C3: AdjustFieldTypes must not rewrite
// minute/second widths (FormatJS BestFitFormatMatcher continues past them), so
// the interval path keeps the zero-padded minute the single-format path already
// produces. Before the fix the range printed "9:5 – 9:7 AM"; V8 prints
// "9:05 – 9:07 AM".
func TestFormatRangeNumericMinute(t *testing.T) {
	t.Parallel()

	f, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Hour:     stringPtr(NumericFieldStyle),
		Minute:   stringPtr(NumericFieldStyle),
		TimeZone: stringPtr("UTC"),
	})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 1, 1, 9, 5, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 9, 7, 0, 0, time.UTC)

	got := f.FormatRange(start, end)
	const want = "9:05\u2009\u2013\u20099:07 AM"
	if got != want {
		t.Errorf("FormatRange = %q, want %q", got, want)
	}
}
