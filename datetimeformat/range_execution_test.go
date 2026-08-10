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
			{name: "cross-day date interval", options: Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle), TimeZone: stringPtr("UTC")}, start: base, end: base.AddDate(0, 0, 2)},
			{name: "time interval", options: Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Second: stringPtr(TwoDigitFieldStyle), TimeZone: stringPtr("UTC")}, start: base, end: base.Add(2*time.Hour + 3*time.Minute)},
			{name: "date-time interval", options: Options{DateStyle: stringPtr(MediumDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), TimeZone: stringPtr("UTC")}, start: base, end: base.Add(2 * time.Hour)},
			{name: "cross-month fallback", options: Options{DateStyle: stringPtr(FullDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), TimeZone: stringPtr("UTC")}, start: base, end: base.AddDate(0, 1, 0)},
			{name: "cross-year fallback", options: Options{DateStyle: stringPtr(MediumDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), TimeZone: stringPtr("UTC")}, start: base, end: base.AddDate(1, 0, 0)},
			{name: "DST boundary", options: Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), TimeZone: stringPtr("America/New_York")}, start: time.Date(2026, time.March, 8, 6, 30, 0, 0, time.UTC), end: time.Date(2026, time.March, 8, 7, 30, 0, 0, time.UTC)},
			{name: "fixed offset", options: Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), TimeZone: stringPtr("+05:30")}, start: base, end: base.Add(2 * time.Hour)},
			{name: "reversed", options: Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle), TimeZone: stringPtr("UTC")}, start: base.AddDate(1, 0, 0), end: base},
			{name: "equal", options: Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle), TimeZone: stringPtr("UTC")}, start: base, end: base},
		} {
			t.Run(localeName+"/"+tc.name, func(t *testing.T) {
				t.Parallel()

				format, err := New(locale.List{intltest.Locale(t, localeName)}, tc.options)
				if err != nil {
					t.Fatal(err)
				}
				text := format.FormatRange(tc.start, tc.end)
				parts := format.FormatRangeToParts(tc.start, tc.end)
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

func TestDateTimeRangeProgramsCompileInConstructor(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Year:     stringPtr(NumericFieldStyle),
		Month:    stringPtr(ShortMonthStyle),
		Day:      stringPtr(NumericFieldStyle),
		Hour:     stringPtr(NumericFieldStyle),
		Minute:   stringPtr(TwoDigitFieldStyle),
		TimeZone: stringPtr("UTC"),
	})
	if err != nil {
		t.Fatal(err)
	}
	record := format.pattern.rangeRecord
	for field, program := range record.dateFields {
		if program != nil && len(program) == 0 {
			t.Errorf("date interval field %d has no compiled program", field)
		}
	}
	for field, program := range record.timeFields {
		if program != nil && len(program) == 0 {
			t.Errorf("time interval field %d has no compiled program", field)
		}
	}
	if len(record.fallback) == 0 {
		t.Error("range fallback has no compiled program")
	}
}
