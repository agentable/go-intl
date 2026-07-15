package datetimeformat

import (
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestDateTimeRangeTextAndPartsShareExecutionDecisions(t *testing.T) {
	t.Parallel()

	base := time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC)
	for _, localeName := range []string{"en-US", "fr-FR", "zh-CN"} {
		for _, tc := range []struct {
			name       string
			options    Options
			start, end time.Time
		}{
			{name: "date interval", options: Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)}, start: base, end: base.AddDate(0, 0, 2)},
			{name: "time interval", options: Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Second: stringPtr(TwoDigitFieldStyle)}, start: base, end: base.Add(2*time.Hour + 3*time.Minute)},
			{name: "date-time interval", options: Options{DateStyle: stringPtr(MediumDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle)}, start: base, end: base.Add(2 * time.Hour)},
			{name: "fallback", options: Options{DateStyle: stringPtr(FullDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle)}, start: base, end: base.AddDate(0, 1, 0)},
			{name: "reversed", options: Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)}, start: base.AddDate(1, 0, 0), end: base},
			{name: "equal", options: Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)}, start: base, end: base},
		} {
			t.Run(localeName+"/"+tc.name, func(t *testing.T) {
				t.Parallel()

				format, err := New(locale.List{intltest.Locale(t, localeName)}, tc.options)
				if err != nil {
					t.Fatal(err)
				}
				text, err := format.FormatRange(tc.start, tc.end)
				if err != nil {
					t.Fatal(err)
				}
				parts, err := format.FormatRangeToParts(tc.start, tc.end)
				if err != nil {
					t.Fatal(err)
				}
				if joined := joinRangePartValues(parts); joined != text {
					t.Fatalf("joined range parts = %q, want text %q; parts=%#v", joined, text, parts)
				}
				for _, part := range parts {
					switch part.Source {
					case SourceStartRange, SourceEndRange, SourceShared:
					default:
						t.Fatalf("range part has invalid source: %#v", part)
					}
				}
			})
		}
	}
}
