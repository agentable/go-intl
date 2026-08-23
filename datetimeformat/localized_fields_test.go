package datetimeformat

import (
	"reflect"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestDateTimeFormatLocalizedFieldWidths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		locale  string
		options Options
		date    time.Time
		want    string
	}{
		{
			name:   "abbreviated weekday and numeric month",
			locale: "en-US-u-nu-arab",
			options: Options{
				Weekday:  stringPtr(ShortFieldStyle),
				Month:    stringPtr(TwoDigitMonthStyle),
				Day:      stringPtr(TwoDigitFieldStyle),
				TimeZone: stringPtr("UTC"),
			},
			date: time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC),
			want: "Fri, ٠٥/٠٨",
		},
		{
			name:   "narrow standard period",
			locale: "en-US",
			options: Options{
				Hour:      stringPtr(NumericFieldStyle),
				DayPeriod: stringPtr(NarrowFieldStyle),
				Hour12:    boolPtr(true),
				TimeZone:  stringPtr("UTC"),
			},
			date: time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC),
			want: "12 mi",
		},
		{
			name:   "short fixed-offset name falls back to GMT",
			locale: "fr-FR",
			options: Options{
				Hour:         stringPtr(NumericFieldStyle),
				TimeZone:     stringPtr("+05:30"),
				TimeZoneName: stringPtr(ShortTimeZoneName),
			},
			date: time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC),
			want: "17 h UTC+5:30",
		},
		{
			name:   "long generic non-metazone location",
			locale: "fr-FR",
			options: Options{
				Hour:         stringPtr(NumericFieldStyle),
				TimeZone:     stringPtr("Antarctica/DumontDUrville"),
				TimeZoneName: stringPtr(LongGenericTimeZoneName),
			},
			date: time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC),
			want: "22 h heure : Dumont-d’Urville",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, tc.locale)}, tc.options)
			if err != nil {
				t.Fatalf("New() error = %v", err)
			}
			if got := format.Format(tc.date); got != tc.want {
				t.Fatalf("Format() = %q, want %q", got, tc.want)
			}
			parts := format.FormatToParts(tc.date)
			var joined string
			for _, part := range parts {
				joined += part.Value
			}
			if joined != tc.want {
				t.Fatalf("joined FormatToParts() = %q, want %q; parts=%#v", joined, tc.want, parts)
			}
		})
	}
}

func TestDateTimeFormatLocalizedRangeNamesKeepSources(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "fr-FR")}, Options{
		Weekday:      stringPtr(LongFieldStyle),
		Month:        stringPtr(LongMonthStyle),
		Day:          stringPtr(NumericFieldStyle),
		Hour:         stringPtr(NumericFieldStyle),
		Minute:       stringPtr(TwoDigitFieldStyle),
		TimeZone:     stringPtr("Europe/Paris"),
		TimeZoneName: stringPtr(LongTimeZoneName),
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	start := time.Date(2026, time.March, 28, 22, 30, 0, 0, time.UTC)
	end := time.Date(2026, time.March, 29, 1, 30, 0, 0, time.UTC)
	text := format.FormatRange(start, end)
	parts := format.FormatRangeToParts(start, end)
	var joined string
	seen := map[RangeSource]bool{}
	for _, part := range parts {
		joined += part.Value
		seen[part.Source] = true
	}
	if joined != text {
		t.Fatalf("joined FormatRangeToParts() = %q, want FormatRange() %q; parts=%#v", joined, text, parts)
	}
	if !reflect.DeepEqual(seen, map[RangeSource]bool{SourceStartRange: true, SourceShared: true, SourceEndRange: true}) {
		t.Fatalf("range sources = %#v, want start/shared/end; parts=%#v", seen, parts)
	}
}
