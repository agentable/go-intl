package datetimeformat

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intlerr"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/tz"
	"github.com/agentable/go-intl/locale"
)

func TestMain(m *testing.M) {
	oldLocal := time.Local
	time.Local = time.UTC
	code := m.Run()
	time.Local = oldLocal
	os.Exit(code)
}

func assertOptionError(t *testing.T, err error, kind, name, value, loc string) {
	t.Helper()

	wantKind := kind
	switch kind {
	case "invalid":
		wantKind = "invalidOption"
	case "unsupported":
		wantKind = "unsupportedOption"
	}
	optErr, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("error = %T, want OptionError", err)
	}
	if optErr.Owner != "datetimeformat" || string(optErr.Kind) != wantKind || optErr.Name != name || optErr.Value != value || optErr.Locale != loc {
		t.Fatalf("OptionError = %+v, want owner=datetimeformat kind=%q name=%q value=%q locale=%q", optErr, kind, name, value, loc)
	}
}

func TestDateTimeFormatDefaultResolvedOptions(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := resolved.Locale.String(), "en"; got != want {
		t.Fatalf("ResolvedOptions().Locale = %q, want %q", got, want)
	}
	if got, want := resolved.Calendar, "gregory"; got != want {
		t.Fatalf("ResolvedOptions().Calendar = %q, want %q", got, want)
	}
	if got, want := resolved.NumberingSystem, "latn"; got != want {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, want)
	}
	if resolved.Year != "numeric" || resolved.Month != "numeric" || resolved.Day != "numeric" {
		t.Fatalf("ResolvedOptions() date fields = year:%q month:%q day:%q, want numeric/numeric/numeric", resolved.Year, resolved.Month, resolved.Day)
	}
}

func TestDateTimeFormatDefaultFormatUsesDateFields(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	if got, want := format.Format(date), "5/8/2026"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
	wantParts := []Part{
		{Type: PartMonth, Value: "5"},
		{Type: PartLiteral, Value: "/"},
		{Type: PartDay, Value: "8"},
		{Type: PartLiteral, Value: "/"},
		{Type: PartYear, Value: "2026"},
	}
	if got := format.FormatToParts(date); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatDefaultFormatUsesCLDRDateOrder(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("zh-Hans-CN")}, Options{})
	if err != nil {
		t.Fatalf("New(zh-Hans-CN) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := cldrdate.GregorianFor(format.cldrLoc)
	if got, want := gregorian.AvailableFormats["yMd"], "y/M/d"; got != want {
		t.Fatalf("CLDR yMd pattern = %q, want %q", got, want)
	}
	if got, want := format.Format(date), "2026/5/8"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
	wantParts := []Part{
		{Type: PartYear, Value: "2026"},
		{Type: PartLiteral, Value: "/"},
		{Type: PartMonth, Value: "5"},
		{Type: PartLiteral, Value: "/"},
		{Type: PartDay, Value: "8"},
	}
	if got := format.FormatToParts(date); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatResolvedOptionsNumberingSystem(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US-u-nu-latn")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US-u-nu-latn) error = %v", err)
	}
	if got, want := format.ResolvedOptions().NumberingSystem, "latn"; got != want {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, want)
	}
}

func TestDateTimeFormatFallsBackToDateDataLocale(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("zh-Hans-CN")}, Options{Hour: NumericFieldStyle, DayPeriod: LongFieldStyle})
	if err != nil {
		t.Fatalf("New(zh-Hans-CN) error = %v", err)
	}
	parts := format.FormatToParts(time.Date(2026, 5, 10, 8, 0, 0, 0, time.UTC))
	if len(parts) == 0 || parts[0].Value == "" || parts[0].Value == "in the morning" {
		t.Fatalf("FormatToParts() = %#v, want zh-Hans day-period data", parts)
	}
}

func TestDateTimeFormatUsesSafeNumberingFallback(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("fr-FR")}, Options{})
	if err != nil {
		t.Fatalf("New(fr-FR) error = %v", err)
	}
	if got, want := format.ResolvedOptions().NumberingSystem, "latn"; got != want {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRejectsUnsupportedCalendarOption(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US")
	_, err := New(locale.List{loc}, Options{Calendar: "buddhist"})
	if !errors.Is(err, intlerr.ErrUnsupportedOption) {
		t.Fatalf("New(WithCalendar(buddhist)) error = %v, want intlerr.ErrUnsupportedOption", err)
	}
	assertOptionError(t, err, "unsupported", "calendar", "buddhist", loc.String())
}

func TestDateTimeFormatRejectsUnsupportedLocaleCalendar(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US-u-ca-buddhist")
	_, err := New(locale.List{loc}, Options{})
	if !errors.Is(err, intlerr.ErrUnsupportedOption) {
		t.Fatalf("New(en-US-u-ca-buddhist) error = %v, want intlerr.ErrUnsupportedOption", err)
	}
}

func TestDateTimeFormatSupportsISO8601Calendar(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Calendar: "iso8601"})
	if err != nil {
		t.Fatalf("New(calendar=iso8601) error = %v", err)
	}
	if got := format.ResolvedOptions().Calendar; got != "iso8601" {
		t.Fatalf("ResolvedOptions().Calendar = %q, want iso8601", got)
	}
}

func TestDateTimeFormatNumberingSystemOptionLocalizesDigits(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{NumberingSystem: "arab"})
	if err != nil {
		t.Fatalf("New(numberingSystem=arab) error = %v", err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "arab" {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want arab", got)
	}
	if got := format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)); got != "٥/٨/٢٠٢٦" {
		t.Fatalf("Format() = %q, want Arabic-Indic digits", got)
	}
}

func TestDateTimeFormatRejectsInvalidStringOptions(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US")
	for _, tt := range []struct {
		name  string
		value string
		opt   Options
	}{
		{name: "month", value: "wide", opt: Options{Month: MonthStyle("wide")}},
		{name: "era", value: "numeric", opt: Options{Era: FieldStyle("numeric")}},
		{name: "localeMatcher", value: "fast", opt: Options{LocaleMatcher: LocaleMatcher("fast")}},
		{name: "formatMatcher", value: "fast", opt: Options{FormatMatcher: FormatMatcher("fast")}},
		{name: "dateStyle", value: "tiny", opt: Options{DateStyle: Style("tiny")}},
		{name: "timeStyle", value: "tiny", opt: Options{TimeStyle: Style("tiny")}},
		{name: "weekday", value: "numeric", opt: Options{Weekday: FieldStyle("numeric")}},
		{name: "dayPeriod", value: "numeric", opt: Options{DayPeriod: FieldStyle("numeric")}},
		{name: "hourCycle", value: "h99", opt: Options{HourCycle: HourCycle("h99")}},
		{name: "timeZoneName", value: "city", opt: Options{TimeZoneName: TimeZoneName("city")}},
		{name: "calendar", value: "bad!", opt: Options{Calendar: "bad!"}},
		{name: "numberingSystem", value: "bad!", opt: Options{NumberingSystem: "bad!"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, tt.opt)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s=%q) error = %v, want intlerr.ErrInvalidOption", tt.name, tt.value, err)
			}
			assertOptionError(t, err, "invalid", tt.name, tt.value, loc.String())
		})
	}
}

func TestDateTimeFormatRejectsInvalidFractionalSecondDigits(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US")
	for _, digits := range []int{-1, 0, 4} {
		t.Run(fmt.Sprint(digits), func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, Options{FractionalSecondDigits: intPtr(digits)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(fractionalSecondDigits=%d) error = %v, want intlerr.ErrInvalidOption", digits, err)
			}
			assertOptionError(t, err, "invalid", "fractionalSecondDigits", fmt.Sprint(digits), loc.String())
		})
	}
}

func TestDateTimeFormatRejectsDateStyleWithGranularField(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: MediumDateTimeStyle, Year: NumericFieldStyle})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(dateStyle+year) error = %v, want intlerr.ErrInvalidOption", err)
	}
	assertOptionError(t, err, "invalid", "dateStyle/timeStyle", "year", "en-US")
}

func TestDateTimeFormatRejectsStyleWithAnyGranularField(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US")
	tests := []struct {
		name string
		opts Options
		want string
	}{
		{name: "dateStyle era", opts: Options{DateStyle: MediumDateTimeStyle, Era: ShortFieldStyle}, want: "era"},
		{name: "timeStyle hour", opts: Options{TimeStyle: ShortDateTimeStyle, Hour: NumericFieldStyle}, want: "hour"},
		{name: "dateStyle timezone name", opts: Options{DateStyle: FullDateTimeStyle, TimeZoneName: LongTimeZoneName}, want: "timeZoneName"},
		{name: "timeStyle fractional second", opts: Options{TimeStyle: ShortDateTimeStyle, FractionalSecondDigits: intPtr(3)}, want: "fractionalSecondDigits"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			assertOptionError(t, err, "invalid", "dateStyle/timeStyle", tc.want, loc.String())
		})
	}
}

func TestDateTimeFormatResolvedOptionsReturnsStableSnapshots(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}
	first := format.ResolvedOptions()
	second := format.ResolvedOptions()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("ResolvedOptions() changed between calls: first = %#v, second = %#v", first, second)
	}
	if got, want := second.Year, NumericStyle("numeric"); got != want {
		t.Fatalf("ResolvedOptions().Year = %q, want %q", got, want)
	}
	if got, want := second.Month, MonthStyle("short"); got != want {
		t.Fatalf("ResolvedOptions().Month = %q, want %q", got, want)
	}
	first.Year = "2-digit"
	first.Month = "long"
	third := format.ResolvedOptions()
	if reflect.DeepEqual(first, third) {
		t.Fatalf("ResolvedOptions() returned mutable internal state: got %#v", third)
	}
	if !reflect.DeepEqual(second, third) {
		t.Fatalf("ResolvedOptions() changed after mutating prior snapshot: second = %#v, third = %#v", second, third)
	}
}

func TestDateTimeFormatCanonicalizesTimeZoneLink(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeZone: "US/Eastern"})
	if err != nil {
		t.Fatalf("New(Options{TimeZone: US/Eastern}) error = %v", err)
	}
	if got, want := format.ResolvedOptions().TimeZone, "America/New_York"; got != want {
		t.Fatalf("ResolvedOptions().TimeZone = %q, want %q", got, want)
	}
}

func TestDateTimeFormatPreservesFixedOffsetTimeZone(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in   string
		want string
	}{
		{in: "+05:30", want: "+05:30"},
		{in: "+0530", want: "+05:30"},
		{in: "+05", want: "+05:00"},
		{in: "+14:00", want: "+14:00"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeZone: tc.in})
			if err != nil {
				t.Fatalf("New(Options{TimeZone: %s}) error = %v", tc.in, err)
			}
			if got := format.ResolvedOptions().TimeZone; got != tc.want {
				t.Fatalf("ResolvedOptions().TimeZone = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDateTimeFormatRejectsUnsupportedTimeZone(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US")
	for _, timeZone := range []string{"Mars/Olympus", "+14:01"} {
		t.Run(timeZone, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, Options{TimeZone: timeZone})
			if !errors.Is(err, intlerr.ErrUnsupportedOption) {
				t.Fatalf("New(Options{TimeZone: %s}) error = %v, want intlerr.ErrUnsupportedOption", timeZone, err)
			}
			assertOptionError(t, err, "unsupported", "timeZone", timeZone, loc.String())
		})
	}
}

func TestDateTimeFormatAllowsEmptyTimeZone(t *testing.T) {
	restore := tz.OverrideDefaultForTest("America/New_York")
	t.Cleanup(restore)

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeZone: ""})
	if err != nil {
		t.Fatalf("New(Options{TimeZone: empty}) error = %v", err)
	}
	if got, want := format.ResolvedOptions().TimeZone, "America/New_York"; got != want {
		t.Fatalf("ResolvedOptions().TimeZone = %q, want %q", got, want)
	}
	instant := time.Date(2026, time.May, 8, 1, 30, 0, 0, time.FixedZone("input", 9*3600))
	if got, want := format.Format(instant), "5/7/2026"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatHour12FalseResolvesHourCycle(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Hour: NumericFieldStyle, Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(WithHour+Options{Hour12: boolPtr(false)}) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := resolved.Hour, NumericStyle("numeric"); got != want {
		t.Fatalf("ResolvedOptions().Hour = %q, want %q", got, want)
	}
	if got, want := resolved.HourCycle, HourCycle("h23"); got != want {
		t.Fatalf("ResolvedOptions().HourCycle = %q, want %q", got, want)
	}
	if resolved.Hour12 == nil || *resolved.Hour12 {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want pointer to false", resolved.Hour12)
	}
}

func TestDateTimeFormatHour12OverridesHourCycle(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Hour: NumericFieldStyle, HourCycle: H11HourCycle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(WithHourCycle+Options{Hour12: boolPtr(true)}) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := resolved.HourCycle, HourCycle("h12"); got != want {
		t.Fatalf("ResolvedOptions().HourCycle = %q, want %q", got, want)
	}
	if resolved.Hour12 == nil || !*resolved.Hour12 {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want pointer to true", resolved.Hour12)
	}
}

func TestDateTimeFormatUsesLocaleHourCycleExtension(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US-u-hc-h23")}, Options{Hour: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(en-US-u-hc-h23, WithHour) error = %v", err)
	}
	if got, want := format.ResolvedOptions().HourCycle, HourCycle("h23"); got != want {
		t.Fatalf("ResolvedOptions().HourCycle = %q, want %q", got, want)
	}
}

func TestDateTimeFormatHourPatternValuesAtMidnight(t *testing.T) {
	t.Parallel()

	midnight := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		hourCycle HourCycle
		field     rune
		want      int
	}{
		{name: "h11 starts at zero", hourCycle: H11HourCycle, field: 'K', want: 0},
		{name: "h12 starts at twelve", hourCycle: H12HourCycle, field: 'h', want: 12},
		{name: "h23 starts at zero", hourCycle: H23HourCycle, field: 'H', want: 0},
		{name: "h24 starts at twenty four", hourCycle: H24HourCycle, field: 'k', want: 24},
	}
	for _, tc := range tests {
		f := DateTimeFormat{resolved: ResolvedOptions{HourCycle: tc.hourCycle}}
		if got := f.hourPatternValue(tc.field, midnight); got != tc.want {
			t.Fatalf("%s hourPatternValue(%q, midnight) = %d, want %d", tc.name, tc.field, got, tc.want)
		}
	}
}

func TestDateTimeFormatFormatDateOnly(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if want := "May 8, 2026"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatShortMonthDateUsesCLDRPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("zh-Hans-CN")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(zh-Hans short date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := cldrdate.GregorianFor(format.cldrLoc)
	if got, want := gregorian.AvailableFormats["yMMMd"], "y年M月d日"; got != want {
		t.Fatalf("CLDR yMMMd pattern = %q, want %q", got, want)
	}
	if got, want := format.Format(date), "2026年5月8日"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
	wantParts := []Part{
		{Type: PartYear, Value: "2026"},
		{Type: PartLiteral, Value: "年"},
		{Type: PartMonth, Value: "5"},
		{Type: PartLiteral, Value: "月"},
		{Type: PartDay, Value: "8"},
		{Type: PartLiteral, Value: "日"},
	}
	if parts := format.FormatToParts(date); !reflect.DeepEqual(parts, wantParts) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, wantParts)
	}
}

func TestDateTimeFormatFormatEqualsFormatToPartsJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		date time.Time
	}{
		{
			name: "date",
			opts: Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle},
			date: time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "date time connector",
			opts: Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle, Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle},
			date: time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC),
		},
		{
			name: "24 hour skips day period",
			opts: Options{Hour: NumericFieldStyle, DayPeriod: ShortFieldStyle, Hour12: boolPtr(false)},
			date: time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{locale.MustParse("en-US")}, tc.opts)
			if err != nil {
				t.Fatalf("New(%s) error = %v", tc.name, err)
			}
			var joined strings.Builder
			for _, part := range format.FormatToParts(tc.date) {
				joined.WriteString(part.Value)
			}
			if got, want := format.Format(tc.date), joined.String(); got != want {
				t.Fatalf("Format() = %q, want joined FormatToParts %q", got, want)
			}
		})
	}
}

func TestDateTimeFormatFormatToPartsDateFields(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Weekday: LongFieldStyle, Month: LongMonthStyle, Day: NumericFieldStyle, Year: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(long date fields) error = %v", err)
	}
	parts := format.FormatToParts(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	want := []Part{
		{Type: PartWeekday, Value: "Friday"},
		{Type: PartLiteral, Value: ", "},
		{Type: PartMonth, Value: "May"},
		{Type: PartLiteral, Value: " "},
		{Type: PartDay, Value: "8"},
		{Type: PartLiteral, Value: ", "},
		{Type: PartYear, Value: "2026"},
	}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, want)
	}
}

func TestDateTimeFormatFormatToPartsEra(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Era: ShortFieldStyle, Year: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(era+year) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := resolved.Era, ShortFieldStyle; got != want {
		t.Fatalf("ResolvedOptions().Era = %q, want %q", got, want)
	}
	parts := format.FormatToParts(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	want := []Part{
		{Type: PartYear, Value: "2026"},
		{Type: PartLiteral, Value: " "},
		{Type: PartEra, Value: "AD"},
	}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, want)
	}
}

func TestDateTimeFormatFormatToPartsBCEWideEra(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Era: LongFieldStyle, Year: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(wide era+year) error = %v", err)
	}
	parts := format.FormatToParts(time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC))
	var era, year string
	for _, part := range parts {
		switch part.Type {
		case PartEra:
			era = part.Value
		case PartYear:
			year = part.Value
		case PartLiteral, PartMonth, PartDay, PartHour, PartMinute, PartSecond, PartDayPeriod, PartTimeZoneName, PartWeekday, PartFractionalSecondDigits, PartRelatedYear, PartYearName, PartUnknown:
		}
	}
	if era == "" || era == "Anno Domini" {
		t.Fatalf("FormatToParts(BCE).era = %q, want BC era name in %#v", era, parts)
	}
	if year != "0" {
		t.Fatalf("FormatToParts(BCE).year = %q, want 0 in %#v", year, parts)
	}
}

func TestDateTimeFormatFormatsNarrowDateFieldNames(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Weekday: NarrowFieldStyle, Month: NarrowMonthStyle, Day: NumericFieldStyle, Year: NumericFieldStyle, Era: NarrowFieldStyle})
	if err != nil {
		t.Fatalf("New(narrow date fields) error = %v", err)
	}
	if got, want := format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)), "F, M 8, 2026 A"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatUsesCLDRDateFieldNames(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("zh-Hans-CN")}, Options{Weekday: LongFieldStyle, Month: LongMonthStyle, Day: NumericFieldStyle, Year: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(zh-Hans date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := cldrdate.GregorianFor(format.cldrLoc)
	wantWeekday := gregorian.Weekdays.Wide[int(date.Weekday())]
	if wantWeekday == "" || wantWeekday == "Friday" {
		t.Fatalf("generated CLDR weekday=%q, want non-English value", wantWeekday)
	}

	parts := format.FormatToParts(date)
	var gotWeekday, gotMonth string
	for _, part := range parts {
		switch part.Type {
		case PartWeekday:
			gotWeekday = part.Value
		case PartMonth:
			gotMonth = part.Value
		case PartYear, PartDay, PartHour, PartMinute, PartSecond, PartEra, PartDayPeriod, PartTimeZoneName, PartLiteral, PartFractionalSecondDigits, PartRelatedYear, PartYearName, PartUnknown:
		}
	}
	if gotWeekday != wantWeekday || gotMonth != "5" {
		t.Fatalf("FormatToParts() weekday/month = %q/%q, want CLDR weekday %q and pattern month 5 in %#v", gotWeekday, gotMonth, wantWeekday, parts)
	}
}

func TestDateTimeFormatExplicitTimeComponentsUseCLDRPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		opts   Options
		want   string
		parts  []Part
	}{
		{
			name:   "en 12 hour",
			locale: "en-US",
			opts:   Options{Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: boolPtr(true)},
			want:   "9:07\u202fAM",
			parts: []Part{
				{Type: PartHour, Value: "9"},
				{Type: PartLiteral, Value: ":"},
				{Type: PartMinute, Value: "07"},
				{Type: PartLiteral, Value: "\u202f"},
				{Type: PartDayPeriod, Value: "AM"},
			},
		},
		{
			name:   "zh 24 hour with seconds",
			locale: "zh-Hans-CN",
			opts:   Options{Hour: TwoDigitFieldStyle, Minute: TwoDigitFieldStyle, Second: TwoDigitFieldStyle, Hour12: boolPtr(false)},
			want:   "09:07:06",
			parts: []Part{
				{Type: PartHour, Value: "09"},
				{Type: PartLiteral, Value: ":"},
				{Type: PartMinute, Value: "07"},
				{Type: PartLiteral, Value: ":"},
				{Type: PartSecond, Value: "06"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{locale.MustParse(tt.locale)}, tt.opts)
			if err != nil {
				t.Fatalf("New(%s) error = %v", tt.locale, err)
			}
			date := time.Date(2026, time.May, 8, 9, 7, 6, 0, time.UTC)
			if got := format.Format(date); got != tt.want {
				t.Fatalf("Format() = %q, want %q", got, tt.want)
			}
			if parts := format.FormatToParts(date); !reflect.DeepEqual(parts, tt.parts) {
				t.Fatalf("FormatToParts() = %#v, want %#v", parts, tt.parts)
			}
		})
	}
}

func TestDateTimeFormatFormatToPartsTimeFieldsWithFractionalSeconds(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Hour: TwoDigitFieldStyle, Minute: TwoDigitFieldStyle, Second: TwoDigitFieldStyle, FractionalSecondDigits: intPtr(3), Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(time fields) error = %v", err)
	}
	parts := format.FormatToParts(time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC))
	want := []Part{
		{Type: PartHour, Value: "09"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartMinute, Value: "07"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartSecond, Value: "06"},
		{Type: PartLiteral, Value: "."},
		{Type: PartFractionalSecondDigits, Value: "123"},
	}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, want)
	}
}

func TestDateTimeFormatFormatFlexibleDayPeriod(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("zh")}, Options{Hour: NumericFieldStyle, DayPeriod: LongFieldStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(day period fields) error = %v", err)
	}
	parts := format.FormatToParts(time.Date(2026, time.May, 8, 15, 0, 0, 0, time.UTC))
	want := []Part{
		{Type: PartDayPeriod, Value: "下午"},
		{Type: PartHour, Value: "3"},
		{Type: PartLiteral, Value: "时"},
	}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, want)
	}
}

func TestDateTimeFormatFormatsTimeZoneName(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeZone: "America/New_York", Hour: NumericFieldStyle, TimeZoneName: LongGenericTimeZoneName})
	if err != nil {
		t.Fatalf("New(timezone name fields) error = %v", err)
	}
	parts := format.FormatToParts(time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC))
	want := []Part{
		{Type: PartHour, Value: "7"},
		{Type: PartLiteral, Value: "\u202f"},
		{Type: PartDayPeriod, Value: "AM"},
		{Type: PartLiteral, Value: " "},
		{Type: PartTimeZoneName, Value: "Eastern Time"},
	}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, want)
	}
}

func TestDateTimeFormatFormatsLocalizedTimeZoneNameForms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style TimeZoneName
		want  string
	}{
		{name: "short specific", style: ShortTimeZoneName, want: "7\u202fAM EST"},
		{name: "long specific", style: LongTimeZoneName, want: "7\u202fAM Eastern Standard Time"},
		{name: "default utc", style: ShortTimeZoneName, want: "12\u202fPM GMT"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			timeZone := "America/New_York"
			date := time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC)
			if tc.name == "default utc" {
				timeZone = ""
			}
			format, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeZone: timeZone, Hour: NumericFieldStyle, TimeZoneName: tc.style})
			if err != nil {
				t.Fatalf("New(timezone name fields) error = %v", err)
			}
			if got := format.Format(date); got != tc.want {
				t.Fatalf("Format() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDateTimeFormatFormatsOffsetTimeZoneName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		locale   string
		timeZone string
		style    TimeZoneName
		want     string
	}{
		{name: "en short whole hour", locale: "en-US", timeZone: "+02:00", style: ShortOffsetTimeZoneName, want: "2\u202fAM GMT+2"},
		{name: "en long whole hour", locale: "en-US", timeZone: "+02:00", style: LongOffsetTimeZoneName, want: "2\u202fAM GMT+02:00"},
		{name: "en short zero", locale: "en-US", timeZone: "+00:00", style: ShortOffsetTimeZoneName, want: "12\u202fAM GMT"},
		{name: "en long zero", locale: "en-US", timeZone: "+00:00", style: LongOffsetTimeZoneName, want: "12\u202fAM GMT+00:00"},
		{name: "en half hour", locale: "en-US", timeZone: "-03:30", style: ShortOffsetTimeZoneName, want: "8\u202fPM GMT-3:30"},
		{name: "en quarter hour", locale: "en-US", timeZone: "+05:45", style: ShortOffsetTimeZoneName, want: "5\u202fAM GMT+5:45"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{locale.MustParse(tc.locale)}, Options{TimeZone: tc.timeZone, Hour: NumericFieldStyle, TimeZoneName: tc.style})
			if err != nil {
				t.Fatalf("New(offset timezone name) error = %v", err)
			}
			if got := format.Format(time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)); got != tc.want {
				t.Fatalf("Format() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDateTimeFormatFormatsDifferentTimeZones(t *testing.T) {
	t.Parallel()

	newYork, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeZone: "America/New_York", Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle, Hour: TwoDigitFieldStyle, Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(New_York) error = %v", err)
	}
	shanghai, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeZone: "Asia/Shanghai", Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle, Hour: TwoDigitFieldStyle, Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(Shanghai) error = %v", err)
	}
	instant := time.Date(2026, time.May, 8, 2, 0, 0, 0, time.UTC)
	if got, want := newYork.Format(instant), "May 7, 2026, 22"; got != want {
		t.Fatalf("New_York Format() = %q, want %q", got, want)
	}
	if got, want := shanghai.Format(instant), "May 8, 2026, 10"; got != want {
		t.Fatalf("Shanghai Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatIgnoresMonotonicClock(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Hour: TwoDigitFieldStyle, Minute: TwoDigitFieldStyle, Second: TwoDigitFieldStyle, FractionalSecondDigits: intPtr(3), Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(time fields) error = %v", err)
	}
	withMonotonic := time.Now()
	withoutMonotonic := withMonotonic.Round(0)
	if got, want := format.Format(withMonotonic), format.Format(withoutMonotonic); got != want {
		t.Fatalf("Format(monotonic) = %q, want Format(Round(0)) %q", got, want)
	}
}

func TestDateTimeFormatConcurrentFormatCalls(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	want := "May 8, 2026"
	errCh := make(chan string, 16)
	for range 16 {
		go func() {
			if got := format.Format(date); got != want {
				errCh <- got
				return
			}
			parts := format.FormatToParts(date)
			if len(parts) == 0 {
				errCh <- "empty parts"
				return
			}
			errCh <- ""
		}()
	}
	for range 16 {
		if got := <-errCh; got != "" {
			t.Fatalf("concurrent format result = %q, want %q", got, want)
		}
	}
}

func TestDateTimeFormatDateStyleFull(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: FullDateTimeStyle})
	if err != nil {
		t.Fatalf("New(DateStyle=full) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if want := "Friday, May 8, 2026"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatDateStyleUsesCLDRPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("zh-Hans-CN")}, Options{DateStyle: FullDateTimeStyle})
	if err != nil {
		t.Fatalf("New(zh-Hans-CN dateStyle full) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := cldrdate.GregorianFor(format.cldrLoc)
	if got, want := format.Format(date), "2026年5月8日星期五"; got != want {
		t.Fatalf("Format() = %q, want %q from CLDR pattern %q", got, want, gregorian.DateFormats[0])
	}
	parts := format.FormatToParts(date)
	wantParts := []Part{
		{Type: PartYear, Value: "2026"},
		{Type: PartLiteral, Value: "年"},
		{Type: PartMonth, Value: "5"},
		{Type: PartLiteral, Value: "月"},
		{Type: PartDay, Value: "8"},
		{Type: PartLiteral, Value: "日"},
		{Type: PartWeekday, Value: gregorian.Weekdays.Wide[int(date.Weekday())]},
	}
	if !reflect.DeepEqual(parts, wantParts) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, wantParts)
	}
}

func TestDateTimeFormatDateStyleShortUsesCLDRPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: ShortDateTimeStyle})
	if err != nil {
		t.Fatalf("New(en-US dateStyle short) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	if got, want := format.Format(date), "5/8/26"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
	wantParts := []Part{
		{Type: PartMonth, Value: "5"},
		{Type: PartLiteral, Value: "/"},
		{Type: PartDay, Value: "8"},
		{Type: PartLiteral, Value: "/"},
		{Type: PartYear, Value: "26"},
	}
	if got := format.FormatToParts(date); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatTimeStyleShort(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeStyle: ShortDateTimeStyle, Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(TimeStyle=short) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC))
	if want := "09:07"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatTimeStyleMediumUsesCLDRPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("zh-Hans-CN")}, Options{TimeStyle: MediumDateTimeStyle, Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(zh-Hans timeStyle medium) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 9, 7, 6, 0, time.UTC)
	gregorian := cldrdate.GregorianFor(format.cldrLoc)
	if got, want := gregorian.TimeFormats[2], "HH:mm:ss"; got != want {
		t.Fatalf("CLDR medium time pattern = %q, want %q", got, want)
	}
	if got, want := format.Format(date), "09:07:06"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
	wantParts := []Part{
		{Type: PartHour, Value: "09"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartMinute, Value: "07"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartSecond, Value: "06"},
	}
	if parts := format.FormatToParts(date); !reflect.DeepEqual(parts, wantParts) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, wantParts)
	}
}

func TestDateTimeFormatTimeStyleLongUsesCLDRTimeZonePattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeStyle: LongDateTimeStyle, TimeZone: "America/New_York", Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(en-US timeStyle long) error = %v", err)
	}
	date := time.Date(2026, time.January, 8, 12, 7, 6, 0, time.UTC)
	gregorian := cldrdate.GregorianFor(format.cldrLoc)
	if got, want := gregorian.TimeFormats[1], "h:mm:ss\u202fa z"; got != want {
		t.Fatalf("CLDR long time pattern = %q, want %q", got, want)
	}
	if got, want := format.Format(date), "07:07:06 EST"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
	wantParts := []Part{
		{Type: PartHour, Value: "07"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartMinute, Value: "07"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartSecond, Value: "06"},
		{Type: PartLiteral, Value: " "},
		{Type: PartTimeZoneName, Value: "EST"},
	}
	if parts := format.FormatToParts(date); !reflect.DeepEqual(parts, wantParts) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, wantParts)
	}
}

func TestDateTimeFormatLongDateAndFullTimeStyles(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: LongDateTimeStyle, TimeStyle: FullDateTimeStyle, TimeZone: "America/New_York", Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(long dateStyle+full timeStyle) error = %v", err)
	}
	date := time.Date(2026, time.January, 8, 12, 7, 6, 0, time.UTC)
	if got, want := format.Format(date), "January 8, 2026 at 07:07:06 Eastern Standard Time"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatCombinesDateAndTimeStyles(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(dateStyle+timeStyle) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC))
	if want := "May 8, 2026, 09:07"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatCombinesFullDateAndTimeStylesWithAtConnector(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: FullDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(full dateStyle+short timeStyle) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC))
	if want := "Friday, May 8, 2026 at 9:07\u202fAM"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatCombinesLongMonthDateAndTimeWithAtConnector(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: LongMonthStyle, Day: NumericFieldStyle, Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(long month date+time fields) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC))
	if want := "May 8, 2026 at 9:07\u202fAM"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatCombinesDateAndTimeStylesWithCLDRConnector(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("zh-Hans-CN")}, Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(zh-Hans dateStyle+timeStyle) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	gregorian := cldrdate.GregorianFor(format.cldrLoc)
	pattern := gregorian.DateTimeAtFormats[2]
	if pattern == "" || !strings.Contains(pattern, "{1}") || !strings.Contains(pattern, "{0}") {
		t.Fatalf("CLDR medium dateTime at pattern = %q, want date/time connector", pattern)
	}

	got := format.Format(date)
	if strings.Contains(got, ",") {
		t.Fatalf("Format() = %q, want CLDR connector %q without hard-coded comma", got, pattern)
	}
	if want := "2026年5月8日 09:07"; got != want {
		t.Fatalf("Format() = %q, want %q from CLDR connector %q", got, want, pattern)
	}
	wantParts := []Part{
		{Type: PartYear, Value: "2026"},
		{Type: PartLiteral, Value: "年"},
		{Type: PartMonth, Value: "5"},
		{Type: PartLiteral, Value: "月"},
		{Type: PartDay, Value: "8"},
		{Type: PartLiteral, Value: "日 "},
		{Type: PartHour, Value: "09"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartMinute, Value: "07"},
	}
	if parts := format.FormatToParts(date); !reflect.DeepEqual(parts, wantParts) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, wantParts)
	}
}

func TestDateTimeFormatBasicMatcherWithComponents(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{FormatMatcher: BasicFormatMatcher, Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(FormatMatcher=basic) error = %v", err)
	}
	if got, want := format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)), "May 8, 2026"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatStyleResolvedOptionsSuppressGranularFields(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle})
	if err != nil {
		t.Fatalf("New(dateStyle+timeStyle) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := resolved.DateStyle, Style("medium"); got != want {
		t.Fatalf("ResolvedOptions().DateStyle = %q, want %q", got, want)
	}
	if got, want := resolved.TimeStyle, Style("short"); got != want {
		t.Fatalf("ResolvedOptions().TimeStyle = %q, want %q", got, want)
	}
	if resolved.Year != "" || resolved.Month != "" || resolved.Day != "" || resolved.Hour != "" || resolved.Minute != "" {
		t.Fatalf("ResolvedOptions() granular fields = year:%q month:%q day:%q hour:%q minute:%q, want suppressed", resolved.Year, resolved.Month, resolved.Day, resolved.Hour, resolved.Minute)
	}
}

func TestDateTimeFormatRangeEqualInstantsUsesSingleFormat(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	if got, want := format.FormatRange(date, date), format.Format(date); got != want {
		t.Fatalf("FormatRange(equal) = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeToPartsEqualInstantsAreShared(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	parts := format.FormatRangeToParts(date, date)
	want := []RangePart{
		{Type: PartMonth, Value: "May", Source: SourceShared},
		{Type: PartLiteral, Value: " ", Source: SourceShared},
		{Type: PartDay, Value: "8", Source: SourceShared},
		{Type: PartLiteral, Value: ", ", Source: SourceShared},
		{Type: PartYear, Value: "2026", Source: SourceShared},
	}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("FormatRangeToParts(equal) = %#v, want %#v", parts, want)
	}
}

func TestDateTimeFormatRangeUsesIntervalPatternForDifferentDays(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
	if got, want := format.FormatRange(start, end), "May 8\u2009–\u200910, 2026"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
	wantParts := []RangePart{
		{Type: PartMonth, Value: "May", Source: SourceShared},
		{Type: PartLiteral, Value: " ", Source: SourceShared},
		{Type: PartDay, Value: "8", Source: SourceStartRange},
		{Type: PartLiteral, Value: "\u2009–\u2009", Source: SourceShared},
		{Type: PartDay, Value: "10", Source: SourceEndRange},
		{Type: PartLiteral, Value: ", ", Source: SourceShared},
		{Type: PartYear, Value: "2026", Source: SourceShared},
	}
	if got := format.FormatRangeToParts(start, end); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatRangeToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatRangeUsesIntervalPatternForDifferentMonths(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)
	if got, want := format.FormatRange(start, end), "May 8\u2009–\u2009Jun 10, 2026"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeUsesIntervalPatternForDifferentYears(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, time.June, 10, 0, 0, 0, 0, time.UTC)
	if got, want := format.FormatRange(start, end), "May 8, 2026\u2009–\u2009Jun 10, 2027"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeUsesTimeIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(time fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 10, 7, 0, 0, time.UTC)
	if got, want := format.FormatRange(start, end), "9:07\u2009–\u200910:07\u202fAM"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
	wantParts := []RangePart{
		{Type: PartHour, Value: "9", Source: SourceStartRange},
		{Type: PartLiteral, Value: ":", Source: SourceStartRange},
		{Type: PartMinute, Value: "07", Source: SourceStartRange},
		{Type: PartLiteral, Value: "\u2009–\u2009", Source: SourceShared},
		{Type: PartHour, Value: "10", Source: SourceEndRange},
		{Type: PartLiteral, Value: ":", Source: SourceEndRange},
		{Type: PartMinute, Value: "07", Source: SourceEndRange},
		{Type: PartLiteral, Value: "\u202f", Source: SourceShared},
		{Type: PartDayPeriod, Value: "AM", Source: SourceShared},
	}
	if got := format.FormatRangeToParts(start, end); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatRangeToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatRangeUsesLowerTimeDiffFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  Options
		start time.Time
		end   time.Time
		want  string
	}{
		{
			name:  "minute",
			opts:  Options{Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: boolPtr(true)},
			start: time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC),
			end:   time.Date(2026, time.May, 8, 9, 8, 0, 0, time.UTC),
			want:  "9:07\u2009–\u20099:08\u202fAM",
		},
		{
			name:  "second",
			opts:  Options{Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Second: TwoDigitFieldStyle, Hour12: boolPtr(true)},
			start: time.Date(2026, time.May, 8, 9, 7, 6, 0, time.UTC),
			end:   time.Date(2026, time.May, 8, 9, 7, 8, 0, time.UTC),
			want:  "9:07:06\u202fAM\u2009–\u20099:07:08\u202fAM",
		},
		{
			name:  "fractional second",
			opts:  Options{Hour: TwoDigitFieldStyle, Minute: TwoDigitFieldStyle, Second: TwoDigitFieldStyle, FractionalSecondDigits: intPtr(3), Hour12: boolPtr(false)},
			start: time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC),
			end:   time.Date(2026, time.May, 8, 9, 7, 6, 456_000_000, time.UTC),
			want:  "09:07:06.123\u2009–\u200909:07:06.456",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{locale.MustParse("en-US")}, tc.opts)
			if err != nil {
				t.Fatalf("New(time fields) error = %v", err)
			}
			if got := format.FormatRange(tc.start, tc.end); got != tc.want {
				t.Fatalf("FormatRange() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDateTimeFormatRangeUsesFlexibleDayPeriodIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Hour: NumericFieldStyle, DayPeriod: LongFieldStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(dayPeriod fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 15, 0, 0, 0, time.UTC)
	if got, want := format.FormatRange(start, end), "8 in the morning\u2009–\u20093 in the afternoon"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeCombinesSharedDateWithTimeIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle, Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(date+time fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 10, 7, 0, 0, time.UTC)
	if got, want := format.FormatRange(start, end), "May 8, 2026, 9:07\u2009–\u200910:07\u202fAM"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeDateTimeStyleUsesTimeIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(dateStyle+timeStyle) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 10, 7, 0, 0, time.UTC)
	if got, want := format.FormatRange(start, end), "May 8, 2026, 9:07\u2009–\u200910:07\u202fAM"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeDateStyleUsesIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: FullDateTimeStyle})
	if err != nil {
		t.Fatalf("New(dateStyle) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)
	want := "Friday, May 8\u2009–\u2009Wednesday, June 10, 2026"
	if got := format.FormatRange(start, end); got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
	wantParts := []RangePart{
		{Type: PartWeekday, Value: "Friday", Source: SourceStartRange},
		{Type: PartLiteral, Value: ", ", Source: SourceStartRange},
		{Type: PartMonth, Value: "May", Source: SourceStartRange},
		{Type: PartLiteral, Value: " ", Source: SourceStartRange},
		{Type: PartDay, Value: "8", Source: SourceStartRange},
		{Type: PartLiteral, Value: "\u2009–\u2009", Source: SourceShared},
		{Type: PartWeekday, Value: "Wednesday", Source: SourceEndRange},
		{Type: PartLiteral, Value: ", ", Source: SourceEndRange},
		{Type: PartMonth, Value: "June", Source: SourceEndRange},
		{Type: PartLiteral, Value: " ", Source: SourceEndRange},
		{Type: PartDay, Value: "10", Source: SourceEndRange},
		{Type: PartLiteral, Value: ", ", Source: SourceShared},
		{Type: PartYear, Value: "2026", Source: SourceShared},
	}
	if got := format.FormatRangeToParts(start, end); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatRangeToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatRangeFallbackPartsRemainJoined(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: FullDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(dateStyle+timeStyle) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 10, 10, 7, 0, 0, time.UTC)
	want := "Friday, May 8, 2026 at 9:07\u202fAM\u2009–\u2009Wednesday, June 10, 2026 at 10:07\u202fAM"
	if got := format.FormatRange(start, end); got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
	wantParts := []RangePart{
		{Type: PartWeekday, Value: "Friday", Source: SourceStartRange},
		{Type: PartLiteral, Value: ", ", Source: SourceStartRange},
		{Type: PartMonth, Value: "May", Source: SourceStartRange},
		{Type: PartLiteral, Value: " ", Source: SourceStartRange},
		{Type: PartDay, Value: "8", Source: SourceStartRange},
		{Type: PartLiteral, Value: ", ", Source: SourceStartRange},
		{Type: PartYear, Value: "2026", Source: SourceStartRange},
		{Type: PartLiteral, Value: " at ", Source: SourceStartRange},
		{Type: PartHour, Value: "9", Source: SourceStartRange},
		{Type: PartLiteral, Value: ":", Source: SourceStartRange},
		{Type: PartMinute, Value: "07", Source: SourceStartRange},
		{Type: PartLiteral, Value: "\u202f", Source: SourceStartRange},
		{Type: PartDayPeriod, Value: "AM", Source: SourceStartRange},
		{Type: PartLiteral, Value: "\u2009–\u2009", Source: SourceShared},
		{Type: PartWeekday, Value: "Wednesday", Source: SourceEndRange},
		{Type: PartLiteral, Value: ", ", Source: SourceEndRange},
		{Type: PartMonth, Value: "June", Source: SourceEndRange},
		{Type: PartLiteral, Value: " ", Source: SourceEndRange},
		{Type: PartDay, Value: "10", Source: SourceEndRange},
		{Type: PartLiteral, Value: ", ", Source: SourceEndRange},
		{Type: PartYear, Value: "2026", Source: SourceEndRange},
		{Type: PartLiteral, Value: " at ", Source: SourceEndRange},
		{Type: PartHour, Value: "10", Source: SourceEndRange},
		{Type: PartLiteral, Value: ":", Source: SourceEndRange},
		{Type: PartMinute, Value: "07", Source: SourceEndRange},
		{Type: PartLiteral, Value: "\u202f", Source: SourceEndRange},
		{Type: PartDayPeriod, Value: "AM", Source: SourceEndRange},
	}
	if got := format.FormatRangeToParts(start, end); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatRangeToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatRangeReversedPreservesInputOrder(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	if got, want := format.FormatRange(start, end), "May 10\u2009–\u20098, 2026"; got != want {
		t.Fatalf("FormatRange(reversed) = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeEqualsJoinedRangeParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
	parts := format.FormatRangeToParts(start, end)
	var joined strings.Builder
	var hasStart, hasShared, hasEnd bool
	for _, part := range parts {
		joined.WriteString(part.Value)
		hasStart = hasStart || part.Source == SourceStartRange
		hasShared = hasShared || part.Source == SourceShared
		hasEnd = hasEnd || part.Source == SourceEndRange
	}
	if got, want := joined.String(), format.FormatRange(start, end); got != want {
		t.Fatalf("joined FormatRangeToParts values = %q, want FormatRange() %q", got, want)
	}
	if !hasStart || !hasShared || !hasEnd {
		t.Fatalf("FormatRangeToParts sources start=%v shared=%v end=%v, want all present in %#v", hasStart, hasShared, hasEnd, parts)
	}
}

func ExampleDateTimeFormat_Format() {
	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)))

	// Output:
	// May 8, 2026
}

func ExampleDateTimeFormat_Format_timezone() {
	format, err := New(locale.List{locale.MustParse("en-US")}, Options{TimeZone: "America/New_York", Hour: NumericFieldStyle, TimeZoneName: LongGenericTimeZoneName})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format(time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC)))

	// Output:
	// 7 AM Eastern Time
}

func ExampleDateTimeFormat_FormatToParts() {
	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Weekday: LongFieldStyle, Month: LongMonthStyle, Day: NumericFieldStyle, Year: NumericFieldStyle})
	if err != nil {
		panic(err)
	}

	for _, part := range format.FormatToParts(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)) {
		fmt.Printf("%s=%q\n", part.Type, part.Value)
	}

	// Output:
	// weekday="Friday"
	// literal=", "
	// month="May"
	// literal=" "
	// day="8"
	// literal=", "
	// year="2026"
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{
		locale.MustParse("fr-FR"),
		locale.MustParse("en-US-u-hc-h23"),
		locale.MustParse("ban"),
	}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"fr-FR", "en-US-u-hc-h23"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}

func TestSupportedLocalesOfFiltersUnsupportedLocaleCalendars(t *testing.T) {
	t.Parallel()

	requested := locale.MustParseList("en-US-u-ca-buddhist", "en-US-u-ca-iso8601", "en-US")
	got, err := SupportedLocalesOf(requested, Options{})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"en-US-u-ca-iso8601", "en-US"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got.Strings(), want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}

func TestSupportedLocalesOfErrors(t *testing.T) {
	t.Parallel()

	if _, err := SupportedLocalesOf(nil, Options{LocaleMatcher: LocaleMatcher("fast")}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
	}
}
