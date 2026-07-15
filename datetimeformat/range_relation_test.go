package datetimeformat

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestDateTimeFormatRangeCollapsesWhenSelectedYearRecordMakesLowerFieldsIrrelevant(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Year:     stringPtr(NumericFieldStyle),
		TimeZone: stringPtr("UTC"),
	})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)

	if got, want := mustFormatRange(t, format, start, end), "2026"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
	wantParts := []RangePart{{Type: PartYear, Value: "2026", Source: SourceShared}}
	if got := mustFormatRangeToParts(t, format, start, end); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatRangeToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatRangeDateRecordUsesPresenceAndSemanticValues(t *testing.T) {
	t.Parallel()

	t.Run("month record stops before day", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Month:    stringPtr(ShortMonthStyle),
			TimeZone: stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)

		if got, want := mustFormatRange(t, format, start, end), "May"; got != want {
			t.Fatalf("FormatRange() = %q, want %q", got, want)
		}
		assertRangePartsShared(t, mustFormatRangeToParts(t, format, start, end))
	})

	t.Run("narrow month labels do not define equality", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Month:    stringPtr(NarrowMonthStyle),
			TimeZone: stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)

		if got, single := mustFormatRange(t, format, start, end), format.Format(start); got == single {
			t.Fatalf("FormatRange() = single endpoint %q for distinct semantic months", got)
		}
		assertRangePartsHaveEndpoints(t, mustFormatRangeToParts(t, format, start, end))
	})

	t.Run("era difference remains relevant before year pattern", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Era:      stringPtr(ShortFieldStyle),
			Year:     stringPtr(NumericFieldStyle),
			TimeZone: stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC)

		parts := mustFormatRangeToParts(t, format, start, end)
		assertRangePartsHaveEndpoints(t, parts)
		if got, want := joinRangePartValues(parts), mustFormatRange(t, format, start, end); got != want {
			t.Fatalf("joined parts = %q, want FormatRange() %q", got, want)
		}
	})

	t.Run("localized year controls equality", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Year:     stringPtr(NumericFieldStyle),
			TimeZone: stringPtr("America/Los_Angeles"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2027, time.January, 1, 0, 30, 0, 0, time.UTC)
		end := time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC)

		if got, want := mustFormatRange(t, format, start, end), "2026"; got != want {
			t.Fatalf("FormatRange() = %q, want localized year %q", got, want)
		}
		assertRangePartsShared(t, mustFormatRangeToParts(t, format, start, end))
	})

	t.Run("reversed equal endpoints remain shared", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Year:     stringPtr(NumericFieldStyle),
			TimeZone: stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.March, 8, 0, 0, 0, 0, time.UTC)

		if got, want := mustFormatRange(t, format, start, end), format.Format(start); got != want {
			t.Fatalf("FormatRange() = %q, want first endpoint %q", got, want)
		}
		assertRangePartsShared(t, mustFormatRangeToParts(t, format, start, end))
	})
}

func TestDateTimeFormatRangeCollapsesWhenSelectedHourRecordMakesLowerFieldsIrrelevant(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Hour:     stringPtr(NumericFieldStyle),
		Hour12:   boolPtr(true),
		TimeZone: stringPtr("UTC"),
	})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.May, 8, 9, 5, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 9, 55, 30, 0, time.UTC)

	if got, want := mustFormatRange(t, format, start, end), format.Format(start); got != want {
		t.Fatalf("FormatRange() = %q, want first endpoint %q", got, want)
	}
	assertRangePartsShared(t, mustFormatRangeToParts(t, format, start, end))
}

func TestDateTimeFormatRangeTimeRecordUsesLocalizedSemanticValues(t *testing.T) {
	t.Parallel()

	t.Run("standard period reaches hour", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Hour:     stringPtr(NumericFieldStyle),
			Hour12:   boolPtr(true),
			TimeZone: stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.May, 8, 10, 0, 0, 0, time.UTC)

		assertRangePartsHaveEndpoints(t, mustFormatRangeToParts(t, format, start, end))
	})

	t.Run("standard period difference remains relevant", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Hour:     stringPtr(NumericFieldStyle),
			Hour12:   boolPtr(true),
			TimeZone: stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.May, 8, 15, 0, 0, 0, time.UTC)

		assertRangePartsHaveEndpoints(t, mustFormatRangeToParts(t, format, start, end))
	})

	t.Run("flexible period reaches hour", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Hour:      stringPtr(NumericFieldStyle),
			DayPeriod: stringPtr(LongFieldStyle),
			Hour12:    boolPtr(true),
			TimeZone:  stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.May, 8, 8, 0, 0, 0, time.UTC)
		end := time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC)

		assertRangePartsHaveEndpoints(t, mustFormatRangeToParts(t, format, start, end))
	})

	t.Run("fractional equality uses resolved precision", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Second:                 stringPtr(NumericFieldStyle),
			FractionalSecondDigits: intPtr(1),
			TimeZone:               stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC)
		end := time.Date(2026, time.May, 8, 9, 7, 6, 199_000_000, time.UTC)

		if got, want := mustFormatRange(t, format, start, end), format.Format(start); got != want {
			t.Fatalf("FormatRange() = %q, want equal-at-precision endpoint %q", got, want)
		}
		assertRangePartsShared(t, mustFormatRangeToParts(t, format, start, end))

		different := time.Date(2026, time.May, 8, 9, 7, 6, 223_000_000, time.UTC)
		assertRangePartsHaveEndpoints(t, mustFormatRangeToParts(t, format, start, different))
	})

	t.Run("minute-only request follows selected record", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Minute:   stringPtr(NumericFieldStyle),
			TimeZone: stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.May, 8, 9, 7, 1, 0, time.UTC)
		end := time.Date(2026, time.May, 8, 9, 7, 59, 0, time.UTC)

		if got, want := mustFormatRange(t, format, start, end), format.Format(start); got != want {
			t.Fatalf("FormatRange() = %q, want selected-format endpoint %q", got, want)
		}
		assertRangePartsShared(t, mustFormatRangeToParts(t, format, start, end))
	})

	t.Run("DST fold compares localized fields", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			Hour:     stringPtr(NumericFieldStyle),
			Minute:   stringPtr(TwoDigitFieldStyle),
			Hour12:   boolPtr(true),
			TimeZone: stringPtr("America/New_York"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.November, 1, 5, 30, 0, 0, time.UTC)
		end := time.Date(2026, time.November, 1, 6, 30, 0, 0, time.UTC)

		if got, want := mustFormatRange(t, format, start, end), format.Format(start); got != want {
			t.Fatalf("FormatRange() = %q, want repeated local time %q", got, want)
		}
		assertRangePartsShared(t, mustFormatRangeToParts(t, format, start, end))
	})

	t.Run("date-time short record stops before second", func(t *testing.T) {
		t.Parallel()

		format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
			DateStyle: stringPtr(MediumDateTimeStyle),
			TimeStyle: stringPtr(ShortDateTimeStyle),
			Hour12:    boolPtr(true),
			TimeZone:  stringPtr("UTC"),
		})
		if err != nil {
			t.Fatal(err)
		}
		start := time.Date(2026, time.May, 8, 9, 7, 1, 0, time.UTC)
		end := time.Date(2026, time.May, 8, 9, 7, 59, 0, time.UTC)

		if got, want := mustFormatRange(t, format, start, end), format.Format(start); got != want {
			t.Fatalf("FormatRange() = %q, want first date-time endpoint %q", got, want)
		}
		assertRangePartsShared(t, mustFormatRangeToParts(t, format, start, end))
	})
}

func TestDateTimeFormatRangeSelectsFlexiblePeriodBySemanticIdentity(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "fr")}, Options{
		Hour:      stringPtr(NumericFieldStyle),
		DayPeriod: stringPtr(LongFieldStyle),
		Hour12:    boolPtr(true),
		TimeZone:  stringPtr("UTC"),
	})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.May, 8, 2, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 8, 0, 0, 0, time.UTC)

	parts := mustFormatRangeToParts(t, format, start, end)
	if got, want := joinRangePartValues(parts), "2 du matin\u2009–\u20098 du matin"; got != want {
		t.Fatalf("joined parts = %q, want flexible-period interval %q", got, want)
	}
	var periodSources []RangeSource
	for _, part := range parts {
		if part.Type == PartDayPeriod {
			periodSources = append(periodSources, part.Source)
		}
	}
	wantSources := []RangeSource{SourceStartRange, SourceEndRange}
	if !reflect.DeepEqual(periodSources, wantSources) {
		t.Fatalf("day-period sources = %#v, want semantic period endpoints %#v; parts=%#v", periodSources, wantSources, parts)
	}
	if got, want := joinRangePartValues(parts), mustFormatRange(t, format, start, end); got != want {
		t.Fatalf("joined parts = %q, want FormatRange() %q", got, want)
	}
}

func TestDateTimeFormatRangeSelectsStandardPeriodBeforeHour(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Hour:     stringPtr(NumericFieldStyle),
		Hour12:   boolPtr(true),
		TimeZone: stringPtr("UTC"),
	})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 15, 0, 0, 0, time.UTC)

	if got, want := mustFormatRange(t, format, start, end), "9 AM\u2009–\u20093 PM"; got != want {
		t.Fatalf("FormatRange() = %q, want standard-period interval %q", got, want)
	}
	parts := mustFormatRangeToParts(t, format, start, end)
	var periodSources []RangeSource
	for _, part := range parts {
		if part.Type == PartDayPeriod {
			periodSources = append(periodSources, part.Source)
		}
	}
	wantSources := []RangeSource{SourceStartRange, SourceEndRange}
	if !reflect.DeepEqual(periodSources, wantSources) {
		t.Fatalf("day-period sources = %#v, want %#v; parts=%#v", periodSources, wantSources, parts)
	}
}

func TestDateTimeFormatRangeFallbackAddsOmittedDifferingDateFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options Options
		want    string
	}{
		{
			name: "time only adds short date",
			options: Options{
				Hour:     stringPtr(NumericFieldStyle),
				Hour12:   boolPtr(true),
				TimeZone: stringPtr("UTC"),
			},
			want: "5/8/2026, 9 AM\u2009–\u20095/10/2026, 9 AM",
		},
		{
			name: "date time adds missing day",
			options: Options{
				Month:    stringPtr(ShortMonthStyle),
				Hour:     stringPtr(NumericFieldStyle),
				Hour12:   boolPtr(true),
				TimeZone: stringPtr("UTC"),
			},
			want: "May 8, 9 AM\u2009–\u2009May 10, 9 AM",
		},
	}
	start := time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 10, 9, 0, 0, 0, time.UTC)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en-US")}, tc.options)
			if err != nil {
				t.Fatal(err)
			}
			parts := mustFormatRangeToParts(t, format, start, end)
			if got := joinRangePartValues(parts); got != tc.want {
				t.Fatalf("joined FormatRangeToParts() = %q, want %q; parts=%#v", got, tc.want, parts)
			}
			assertRangePartsHaveEndpoints(t, parts)
			if got := mustFormatRange(t, format, start, end); got != tc.want {
				t.Fatalf("FormatRange() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDateTimeFormatRangeFallbackAccumulatesLessSignificantDateFields(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Month:    stringPtr(ShortMonthStyle),
		Hour:     stringPtr(NumericFieldStyle),
		Hour12:   boolPtr(true),
		TimeZone: stringPtr("UTC"),
	})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC)
	end := time.Date(2027, time.May, 10, 9, 0, 0, 0, time.UTC)
	const want = "May 8, 2026, 9 AM\u2009–\u2009May 10, 2027, 9 AM"

	parts := mustFormatRangeToParts(t, format, start, end)
	if got := joinRangePartValues(parts); got != want {
		t.Fatalf("joined FormatRangeToParts() = %q, want %q; parts=%#v", got, want, parts)
	}
	assertRangePartsHaveEndpoints(t, parts)
	if got := mustFormatRange(t, format, start, end); got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func assertRangePartsShared(t *testing.T, parts []RangePart) {
	t.Helper()
	for _, part := range parts {
		if part.Source != SourceShared {
			t.Fatalf("range part = %#v, want source %q in %#v", part, SourceShared, parts)
		}
	}
}

func assertRangePartsHaveEndpoints(t *testing.T, parts []RangePart) {
	t.Helper()
	var sources strings.Builder
	for _, part := range parts {
		sources.WriteString(string(part.Source))
	}
	if !strings.Contains(sources.String(), string(SourceStartRange)) || !strings.Contains(sources.String(), string(SourceEndRange)) {
		t.Fatalf("range parts lack both endpoint sources: %#v", parts)
	}
}
