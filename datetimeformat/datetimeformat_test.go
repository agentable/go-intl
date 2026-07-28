package datetimeformat

import (
	"encoding/json"
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
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
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

func mustFormatRange(t *testing.T, f *DateTimeFormat, start, end time.Time) string {
	t.Helper()

	out, err := f.FormatRange(start, end)
	if err != nil {
		t.Fatalf("FormatRange(%v, %v) error = %v", start, end, err)
	}
	return out
}

func mustFormatRangeToParts(t *testing.T, f *DateTimeFormat, start, end time.Time) []RangePart {
	t.Helper()

	parts, err := f.FormatRangeToParts(start, end)
	if err != nil {
		t.Fatalf("FormatRangeToParts(%v, %v) error = %v", start, end, err)
	}
	return parts
}

func joinPartValues(parts []Part) string {
	var joined strings.Builder
	for _, part := range parts {
		joined.WriteString(part.Value)
	}
	return joined.String()
}

func joinRangePartValues(parts []RangePart) string {
	var joined strings.Builder
	for _, part := range parts {
		joined.WriteString(part.Value)
	}
	return joined.String()
}

func mustGregorianForDateLocale(t *testing.T, tag string) cldrdate.Gregorian {
	t.Helper()

	loc, ok := cldrdate.ResolveLocale(tag)
	if !ok {
		t.Fatalf("cldrdate.ResolveLocale(%q) = false", tag)
	}
	return cldrdate.GregorianFor(loc)
}

func TestDateTimeFormatDefaultResolvedOptions(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := resolved.Locale.String(), "en-US"; got != want {
		t.Fatalf("ResolvedOptions().Locale = %q, want %q", got, want)
	}
	if got, want := resolved.Calendar, "gregory"; got != want {
		t.Fatalf("ResolvedOptions().Calendar = %q, want %q", got, want)
	}
	if got, want := resolved.NumberingSystem, "latn"; got != want {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, want)
	}
	if ecma402.ResolvedScalarValue(resolved.Year) != "numeric" || ecma402.ResolvedScalarValue(resolved.Month) != "numeric" || ecma402.ResolvedScalarValue(resolved.Day) != "numeric" {
		t.Fatalf("ResolvedOptions() date fields = year:%v month:%v day:%v, want numeric/numeric/numeric", resolved.Year, resolved.Month, resolved.Day)
	}
	if resolved.HourCycle != nil || resolved.Hour12 != nil {
		t.Fatalf("ResolvedOptions() hour-cycle fields = hourCycle:%v hour12:%v, want nil/nil", resolved.HourCycle, resolved.Hour12)
	}
	if resolved.FractionalSecondDigits != nil {
		t.Fatalf("ResolvedOptions().FractionalSecondDigits = %v, want nil", resolved.FractionalSecondDigits)
	}
}

func TestDateTimeFormatDefaultFormatUsesDateFields(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
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

	format, err := New(locale.List{intltest.Locale(t, "zh-Hans-CN")}, Options{})
	if err != nil {
		t.Fatalf("New(zh-Hans-CN) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := mustGregorianForDateLocale(t, "zh-Hans-CN")
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

	format, err := New(locale.List{intltest.Locale(t, "en-US-u-nu-latn")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US-u-nu-latn) error = %v", err)
	}
	if got, want := format.ResolvedOptions().NumberingSystem, "latn"; got != want {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, want)
	}
}

func TestDateTimeFormatFallsBackToDateDataLocale(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "zh-Hans-CN")}, Options{Hour: stringPtr(NumericFieldStyle), DayPeriod: stringPtr(LongFieldStyle)})
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

	format, err := New(locale.List{intltest.Locale(t, "fr-FR")}, Options{})
	if err != nil {
		t.Fatalf("New(fr-FR) error = %v", err)
	}
	if got, want := format.ResolvedOptions().NumberingSystem, "latn"; got != want {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, want)
	}
}

func TestDateTimeFormatFallsBackForUnsupportedCalendarOption(t *testing.T) {
	t.Parallel()

	loc := intltest.Locale(t, "en-US")
	format, err := New(locale.List{loc}, Options{Calendar: stringPtr("buddhist")})
	if err != nil {
		t.Fatalf("New(calendar=buddhist) error = %v", err)
	}
	if got := format.ResolvedOptions().Calendar; got != "gregory" {
		t.Fatalf("ResolvedOptions().Calendar = %q, want gregory fallback", got)
	}
}

func TestDateTimeFormatFallsBackForUnsupportedLocaleCalendar(t *testing.T) {
	t.Parallel()

	loc := intltest.Locale(t, "en-US-u-ca-buddhist")
	format, err := New(locale.List{loc}, Options{})
	if err != nil {
		t.Fatalf("New(en-US-u-ca-buddhist) error = %v", err)
	}
	if got := format.ResolvedOptions().Calendar; got != "gregory" {
		t.Fatalf("ResolvedOptions().Calendar = %q, want gregory fallback", got)
	}
}

func TestDateTimeFormatSupportsISO8601Calendar(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Calendar: stringPtr("iso8601")})
	if err != nil {
		t.Fatalf("New(calendar=iso8601) error = %v", err)
	}
	if got := format.ResolvedOptions().Calendar; got != "iso8601" {
		t.Fatalf("ResolvedOptions().Calendar = %q, want iso8601", got)
	}
}

func TestSupportedLocalesOfPreservesCalendarExtensions(t *testing.T) {
	t.Parallel()

	requested := locale.List{
		intltest.Locale(t, "en-u-ca-gregory"),
		intltest.Locale(t, "en-u-ca-buddhist"),
		intltest.Locale(t, "en-u-ca-iso8601"),
	}
	got, err := SupportedLocalesOf(requested, Options{})
	if err != nil {
		t.Fatal(err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf", got, []string{"en-u-ca-gregory", "en-u-ca-buddhist", "en-u-ca-iso8601"})
}

func TestDateTimeFormatNumberingSystemOptionLocalizesDigits(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{NumberingSystem: stringPtr("arab")})
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

	loc := intltest.Locale(t, "en-US")
	for _, tt := range []struct {
		name       string
		optionName string
		value      string
		opt        Options
	}{
		{name: "month", optionName: "month", value: "wide", opt: Options{Month: stringPtr("wide")}},
		{name: "month empty", optionName: "month", value: "", opt: Options{Month: stringPtr("")}},
		{name: "era", optionName: "era", value: "numeric", opt: Options{Era: stringPtr("numeric")}},
		{name: "era empty", optionName: "era", value: "", opt: Options{Era: stringPtr("")}},
		{name: "localeMatcher", optionName: "localeMatcher", value: "fast", opt: Options{LocaleMatcher: stringPtr("fast")}},
		{name: "localeMatcher empty", optionName: "localeMatcher", value: "", opt: Options{LocaleMatcher: stringPtr("")}},
		{name: "formatMatcher", optionName: "formatMatcher", value: "fast", opt: Options{FormatMatcher: stringPtr("fast")}},
		{name: "formatMatcher empty", optionName: "formatMatcher", value: "", opt: Options{FormatMatcher: stringPtr("")}},
		{name: "dateStyle", optionName: "dateStyle", value: "tiny", opt: Options{DateStyle: stringPtr("tiny")}},
		{name: "dateStyle empty", optionName: "dateStyle", value: "", opt: Options{DateStyle: stringPtr("")}},
		{name: "timeStyle", optionName: "timeStyle", value: "tiny", opt: Options{TimeStyle: stringPtr("tiny")}},
		{name: "timeStyle empty", optionName: "timeStyle", value: "", opt: Options{TimeStyle: stringPtr("")}},
		{name: "weekday", optionName: "weekday", value: "numeric", opt: Options{Weekday: stringPtr("numeric")}},
		{name: "weekday empty", optionName: "weekday", value: "", opt: Options{Weekday: stringPtr("")}},
		{name: "dayPeriod", optionName: "dayPeriod", value: "numeric", opt: Options{DayPeriod: stringPtr("numeric")}},
		{name: "dayPeriod empty", optionName: "dayPeriod", value: "", opt: Options{DayPeriod: stringPtr("")}},
		{name: "year", optionName: "year", value: "long", opt: Options{Year: stringPtr("long")}},
		{name: "year empty", optionName: "year", value: "", opt: Options{Year: stringPtr("")}},
		{name: "day empty", optionName: "day", value: "", opt: Options{Day: stringPtr("")}},
		{name: "hour empty", optionName: "hour", value: "", opt: Options{Hour: stringPtr("")}},
		{name: "minute empty", optionName: "minute", value: "", opt: Options{Minute: stringPtr("")}},
		{name: "second empty", optionName: "second", value: "", opt: Options{Second: stringPtr("")}},
		{name: "hourCycle", optionName: "hourCycle", value: "h99", opt: Options{HourCycle: stringPtr("h99")}},
		{name: "hourCycle empty", optionName: "hourCycle", value: "", opt: Options{HourCycle: stringPtr("")}},
		{name: "timeZoneName", optionName: "timeZoneName", value: "city", opt: Options{TimeZoneName: stringPtr("city")}},
		{name: "timeZoneName empty", optionName: "timeZoneName", value: "", opt: Options{TimeZoneName: stringPtr("")}},
		{name: "calendar", optionName: "calendar", value: "bad!", opt: Options{Calendar: stringPtr("bad!")}},
		{name: "calendar empty", optionName: "calendar", value: "", opt: Options{Calendar: stringPtr("")}},
		{name: "numberingSystem", optionName: "numberingSystem", value: "bad!", opt: Options{NumberingSystem: stringPtr("bad!")}},
		{name: "numberingSystem empty", optionName: "numberingSystem", value: "", opt: Options{NumberingSystem: stringPtr("")}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, tt.opt)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s=%q) error = %v, want intlerr.ErrInvalidOption", tt.name, tt.value, err)
			}
			testcontract.AssertOptionError(t, err, "datetimeformat", intlerr.InvalidOption, tt.optionName, tt.value, loc.String())
		})
	}
}

func TestDateTimeFormatRejectsInvalidFractionalSecondDigits(t *testing.T) {
	t.Parallel()

	loc := intltest.Locale(t, "en-US")
	for _, digits := range []int{-1, 0, 4} {
		t.Run(fmt.Sprint(digits), func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, Options{FractionalSecondDigits: intPtr(digits)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(fractionalSecondDigits=%d) error = %v, want intlerr.ErrInvalidOption", digits, err)
			}
			testcontract.AssertOptionError(t, err, "datetimeformat", intlerr.InvalidOption, "fractionalSecondDigits", fmt.Sprint(digits), loc.String())
		})
	}
}

func TestDateTimeFormatRejectsDateStyleWithGranularField(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(MediumDateTimeStyle), Year: stringPtr(NumericFieldStyle)})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(dateStyle+year) error = %v, want intlerr.ErrInvalidOption", err)
	}
	testcontract.AssertOptionError(t, err, "datetimeformat", intlerr.InvalidOption, "dateStyle/timeStyle", "year", "en-US")
	testcontract.AssertOptionExpected(t, err, "no explicit component options when dateStyle or timeStyle is set")
}

func TestDateTimeFormatRejectsStyleWithAnyGranularField(t *testing.T) {
	t.Parallel()

	loc := intltest.Locale(t, "en-US")
	tests := []struct {
		name string
		opts Options
		want string
	}{
		{name: "dateStyle era", opts: Options{DateStyle: stringPtr(MediumDateTimeStyle), Era: stringPtr(ShortFieldStyle)}, want: "era"},
		{name: "timeStyle hour", opts: Options{TimeStyle: stringPtr(ShortDateTimeStyle), Hour: stringPtr(NumericFieldStyle)}, want: "hour"},
		{name: "dateStyle timezone name", opts: Options{DateStyle: stringPtr(FullDateTimeStyle), TimeZoneName: stringPtr(LongTimeZoneName)}, want: "timeZoneName"},
		{name: "timeStyle fractional second", opts: Options{TimeStyle: stringPtr(ShortDateTimeStyle), FractionalSecondDigits: intPtr(3)}, want: "fractionalSecondDigits"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "datetimeformat", intlerr.InvalidOption, "dateStyle/timeStyle", tc.want, loc.String())
		})
	}
}

func TestDateTimeFormatResolvedOptionsReturnsStableSnapshots(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle)})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}
	first := format.ResolvedOptions()
	second := format.ResolvedOptions()
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("ResolvedOptions() changed between calls: first = %#v, second = %#v", first, second)
	}
	if got, want := ecma402.ResolvedScalarValue(second.Year), NumericStyle("numeric"); got != want {
		t.Fatalf("ResolvedOptions().Year = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(second.Month), MonthStyle("short"); got != want {
		t.Fatalf("ResolvedOptions().Month = %q, want %q", got, want)
	}
	if first.Year == nil || first.Month == nil {
		t.Fatal("ResolvedOptions() omitted Year/Month, want snapshot scalars")
	}
	*first.Year = TwoDigitFieldStyle
	*first.Month = LongMonthStyle
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

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "legacy iana", in: "US/Eastern", want: "America/New_York"},
		{name: "cldr alias", in: "America/Montreal", want: "America/Toronto"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{TimeZone: stringPtr(tc.in)})
			if err != nil {
				t.Fatalf("New(Options{TimeZone: %s}) error = %v", tc.in, err)
			}
			if got := format.ResolvedOptions().TimeZone; got != tc.want {
				t.Fatalf("ResolvedOptions().TimeZone = %q, want %q", got, tc.want)
			}
		})
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

			format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{TimeZone: stringPtr(tc.in)})
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

	loc := intltest.Locale(t, "en-US")
	for _, timeZone := range []string{"Mars/Olympus", "+24:00"} {
		t.Run(timeZone, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{loc}, Options{TimeZone: stringPtr(timeZone)})
			if !errors.Is(err, intlerr.ErrUnsupportedOption) {
				t.Fatalf("New(Options{TimeZone: %s}) error = %v, want intlerr.ErrUnsupportedOption", timeZone, err)
			}
			if !errors.Is(err, tz.ErrUnsupportedTimeZone) {
				t.Fatalf("New(Options{TimeZone: %s}) error = %v, want internal time-zone cause", timeZone, err)
			}
			testcontract.AssertOptionError(t, err, "datetimeformat", intlerr.UnsupportedOption, "timeZone", timeZone, loc.String())
			testcontract.AssertOptionExpected(t, err, timeZoneExpected)
		})
	}
}

func TestDateTimeFormatUsesDefaultTimeZoneWhenOmitted(t *testing.T) {
	for _, tc := range []struct {
		name            string
		defaultTimeZone string
		wantTimeZone    string
		instant         time.Time
		wantFormat      string
	}{
		{
			name:            "iana",
			defaultTimeZone: "America/New_York",
			wantTimeZone:    "America/New_York",
			instant:         time.Date(2026, time.May, 8, 1, 30, 0, 0, time.FixedZone("input", 9*3600)),
			wantFormat:      "5/7/2026",
		},
		{
			name:            "fixed offset",
			defaultTimeZone: "+0530",
			wantTimeZone:    "+05:30",
			instant:         time.Date(2026, time.May, 7, 20, 0, 0, 0, time.UTC),
			wantFormat:      "5/8/2026",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			restore, err := tz.OverrideDefaultForTest(tc.defaultTimeZone)
			if err != nil {
				t.Fatalf("OverrideDefaultForTest() error = %v", err)
			}
			t.Cleanup(restore)

			format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
			if err != nil {
				t.Fatalf("New(Options{}) error = %v", err)
			}
			if got := format.ResolvedOptions().TimeZone; got != tc.wantTimeZone {
				t.Fatalf("ResolvedOptions().TimeZone = %q, want %q", got, tc.wantTimeZone)
			}
			if got := format.Format(tc.instant); got != tc.wantFormat {
				t.Fatalf("Format() = %q, want %q", got, tc.wantFormat)
			}
		})
	}
}

func TestDateTimeFormatRejectsEmptyTimeZone(t *testing.T) {
	t.Parallel()

	loc := intltest.Locale(t, "en-US")
	_, err := New(locale.List{loc}, Options{TimeZone: stringPtr("")})
	if !errors.Is(err, intlerr.ErrUnsupportedOption) {
		t.Fatalf("New(Options{TimeZone: empty}) error = %v, want intlerr.ErrUnsupportedOption", err)
	}
	testcontract.AssertOptionError(t, err, "datetimeformat", intlerr.UnsupportedOption, "timeZone", "", loc.String())
	testcontract.AssertOptionExpected(t, err, timeZoneExpected)
}

func TestDateTimeFormatHour12FalseResolvesHourCycle(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Hour: stringPtr(NumericFieldStyle), Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(WithHour+Options{Hour12: boolPtr(false)}) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := ecma402.ResolvedScalarValue(resolved.Hour), NumericStyle("2-digit"); got != want {
		t.Fatalf("ResolvedOptions().Hour = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.HourCycle), HourCycle("h23"); got != want {
		t.Fatalf("ResolvedOptions().HourCycle = %q, want %q", got, want)
	}
	if resolved.Hour12 == nil || *resolved.Hour12 {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want pointer to false", resolved.Hour12)
	}
}

func TestDateTimeFormatHour12OverridesHourCycle(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Hour: stringPtr(NumericFieldStyle), HourCycle: stringPtr(H11HourCycle), Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(WithHourCycle+Options{Hour12: boolPtr(true)}) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := ecma402.ResolvedScalarValue(resolved.HourCycle), HourCycle("h12"); got != want {
		t.Fatalf("ResolvedOptions().HourCycle = %q, want %q", got, want)
	}
	if resolved.Hour12 == nil || !*resolved.Hour12 {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want pointer to true", resolved.Hour12)
	}
}

func TestDateTimeFormatUsesLocaleHourCycleExtension(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US-u-hc-h23")}, Options{Hour: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(en-US-u-hc-h23, WithHour) error = %v", err)
	}
	if got, want := ecma402.ResolvedScalarValue(format.ResolvedOptions().HourCycle), HourCycle("h23"); got != want {
		t.Fatalf("ResolvedOptions().HourCycle = %q, want %q", got, want)
	}
	if got := format.ResolvedOptions().Hour12; got == nil || *got {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want pointer to false", got)
	}
}

func TestDateTimeFormatHourPatternValuesAtMidnight(t *testing.T) {
	t.Parallel()

	midnight := gregoryLocalTime(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	tests := []struct {
		name       string
		uses24Hour bool
		field      rune
		want       int
	}{
		{name: "h11 starts at zero", uses24Hour: false, field: 'K', want: 0},
		{name: "h12 starts at twelve", uses24Hour: false, field: 'h', want: 12},
		{name: "h23 starts at zero", uses24Hour: true, field: 'H', want: 0},
		{name: "h24 starts at twenty four", uses24Hour: true, field: 'k', want: 24},
	}
	for _, tc := range tests {
		if got := hourPatternValue(tc.field, midnight, tc.uses24Hour); got != tc.want {
			t.Fatalf("%s hourPatternValue(%q, midnight) = %d, want %d", tc.name, tc.field, got, tc.want)
		}
	}
}

func TestDateTimeFormatFormatDateOnly(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
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

	format, err := New(locale.List{intltest.Locale(t, "zh-Hans-CN")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(zh-Hans short date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := mustGregorianForDateLocale(t, "zh-Hans-CN")
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
		name   string
		locale string
		opts   Options
		date   time.Time
	}{
		{
			name:   "date",
			locale: "en-US",
			opts:   Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)},
			date:   time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "date time connector",
			locale: "en-US",
			opts:   Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle), Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle)},
			date:   time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC),
		},
		{
			name:   "quoted style connector",
			locale: "en-US",
			opts:   Options{DateStyle: stringPtr(LongDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), TimeZone: stringPtr("UTC")},
			date:   time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC),
		},
		{
			name:   "non ASCII literals",
			locale: "zh-CN",
			opts:   Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(NumericMonthStyle), Day: stringPtr(NumericFieldStyle)},
			date:   time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC),
		},
		{
			name:   "fractional second",
			locale: "en-US",
			opts:   Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Second: stringPtr(TwoDigitFieldStyle), FractionalSecondDigits: intPtr(3), TimeZone: stringPtr("UTC")},
			date:   time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC),
		},
		{
			name:   "short timezone name",
			locale: "en-US",
			opts:   Options{Hour: stringPtr(NumericFieldStyle), TimeZoneName: stringPtr(ShortTimeZoneName), TimeZone: stringPtr("America/Los_Angeles")},
			date:   time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC),
		},
		{
			name:   "long timezone name",
			locale: "en-US",
			opts:   Options{Hour: stringPtr(NumericFieldStyle), TimeZoneName: stringPtr(LongTimeZoneName), TimeZone: stringPtr("America/Los_Angeles")},
			date:   time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC),
		},
		{
			name:   "24 hour skips day period",
			locale: "en-US",
			opts:   Options{Hour: stringPtr(NumericFieldStyle), DayPeriod: stringPtr(ShortFieldStyle), Hour12: boolPtr(false)},
			date:   time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC),
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, tc.locale)}, tc.opts)
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Weekday: stringPtr(LongFieldStyle), Month: stringPtr(LongMonthStyle), Day: stringPtr(NumericFieldStyle), Year: stringPtr(NumericFieldStyle)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Era: stringPtr(ShortFieldStyle), Year: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(era+year) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := ecma402.ResolvedScalarValue(resolved.Era), ShortFieldStyle; got != want {
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Era: stringPtr(LongFieldStyle), Year: stringPtr(NumericFieldStyle)})
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
		case PartLiteral, PartMonth, PartDay, PartHour, PartMinute, PartSecond, PartDayPeriod, PartTimeZoneName, PartWeekday, PartFractionalSecond, PartRelatedYear, PartYearName, PartUnknown:
		}
	}
	if era == "" || era == "Anno Domini" {
		t.Fatalf("FormatToParts(BCE).era = %q, want BC era name in %#v", era, parts)
	}
	if year != "1" {
		t.Fatalf("FormatToParts(BCE).year = %q, want 1 in %#v", year, parts)
	}
}

func TestDateTimeFormatFormatsNarrowDateFieldNames(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Weekday: stringPtr(NarrowFieldStyle), Month: stringPtr(NarrowMonthStyle), Day: stringPtr(NumericFieldStyle), Year: stringPtr(NumericFieldStyle), Era: stringPtr(NarrowFieldStyle)})
	if err != nil {
		t.Fatalf("New(narrow date fields) error = %v", err)
	}
	if got, want := format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)), "F, M 8, 2026 A"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatUsesCLDRDateFieldNames(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "zh-Hans-CN")}, Options{Weekday: stringPtr(LongFieldStyle), Month: stringPtr(LongMonthStyle), Day: stringPtr(NumericFieldStyle), Year: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(zh-Hans date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := mustGregorianForDateLocale(t, "zh-Hans-CN")
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
		case PartYear, PartDay, PartHour, PartMinute, PartSecond, PartEra, PartDayPeriod, PartTimeZoneName, PartLiteral, PartFractionalSecond, PartRelatedYear, PartYearName, PartUnknown:
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
			opts:   Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(true)},
			want:   "9:07 AM",
			parts: []Part{
				{Type: PartHour, Value: "9"},
				{Type: PartLiteral, Value: ":"},
				{Type: PartMinute, Value: "07"},
				{Type: PartLiteral, Value: " "},
				{Type: PartDayPeriod, Value: "AM"},
			},
		},
		{
			name:   "zh 24 hour with seconds",
			locale: "zh-Hans-CN",
			opts:   Options{Hour: stringPtr(TwoDigitFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Second: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(false)},
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

			format, err := New(locale.List{intltest.Locale(t, tt.locale)}, tt.opts)
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Hour: stringPtr(TwoDigitFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Second: stringPtr(TwoDigitFieldStyle), FractionalSecondDigits: intPtr(3), Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(time fields) error = %v", err)
	}
	if got := format.ResolvedOptions().FractionalSecondDigits; got == nil || *got != 3 {
		t.Fatalf("ResolvedOptions().FractionalSecondDigits = %v, want 3", got)
	}
	parts := format.FormatToParts(time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC))
	want := []Part{
		{Type: PartHour, Value: "09"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartMinute, Value: "07"},
		{Type: PartLiteral, Value: ":"},
		{Type: PartSecond, Value: "06"},
		{Type: PartLiteral, Value: "."},
		{Type: PartFractionalSecond, Value: "123"},
	}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, want)
	}
}

func TestDateTimeFormatFractionalSecondPartUsesECMA402Type(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Second:                 stringPtr(TwoDigitFieldStyle),
		FractionalSecondDigits: intPtr(1),
	})
	if err != nil {
		t.Fatal(err)
	}
	parts := format.FormatToParts(time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC))
	for _, part := range parts {
		if part.Value == "1" {
			if got, want := part.Type, PartType("fractionalSecond"); got != want {
				t.Fatalf("fractional-second part type = %q, want %q", got, want)
			}
			return
		}
	}
	t.Fatalf("FormatToParts() = %#v, want fractional-second part", parts)
}

func TestDateTimeFormatFractionalSecondPartWidthsAndNumberingSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		digits          int
		numberingSystem string
		want            string
	}{
		{name: "one digit", digits: 1, want: "1"},
		{name: "two digits", digits: 2, want: "12"},
		{name: "three digits", digits: 3, want: "123"},
		{name: "arab digits", digits: 3, numberingSystem: "arab", want: "١٢٣"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			options := Options{
				Hour:                   stringPtr(TwoDigitFieldStyle),
				Minute:                 stringPtr(TwoDigitFieldStyle),
				Second:                 stringPtr(TwoDigitFieldStyle),
				FractionalSecondDigits: intPtr(tc.digits),
				Hour12:                 boolPtr(false),
				TimeZone:               stringPtr("UTC"),
			}
			if tc.numberingSystem != "" {
				options.NumberingSystem = stringPtr(tc.numberingSystem)
			}
			format, err := New(locale.List{intltest.Locale(t, "en-US")}, options)
			if err != nil {
				t.Fatal(err)
			}
			input := time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC)
			parts := format.FormatToParts(input)
			var fractional Part
			for _, part := range parts {
				if part.Type == PartFractionalSecond {
					fractional = part
				}
				if part.Type == PartType("fractionalSecondDigits") {
					t.Fatalf("FormatToParts() emitted obsolete type in %#v", parts)
				}
			}
			if fractional.Value != tc.want {
				t.Fatalf("fractional-second part = %#v, want value %q", fractional, tc.want)
			}
			if got, want := format.Format(input), joinPartValues(parts); got != want {
				t.Fatalf("Format() = %q, want joined parts %q", got, want)
			}
			record, err := json.Marshal(fractional)
			if err != nil {
				t.Fatal(err)
			}
			if got, want := string(record), `{"type":"fractionalSecond","value":"`+tc.want+`"}`; got != want {
				t.Fatalf("fractional-second JSON = %s, want %s", got, want)
			}
		})
	}
}

func TestDateTimeFormatRangeFractionalSecondPartUsesECMA402Type(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Hour:                   stringPtr(TwoDigitFieldStyle),
		Minute:                 stringPtr(TwoDigitFieldStyle),
		Second:                 stringPtr(TwoDigitFieldStyle),
		FractionalSecondDigits: intPtr(3),
		Hour12:                 boolPtr(false),
		TimeZone:               stringPtr("UTC"),
	})
	if err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC)
	end := time.Date(2026, time.May, 8, 9, 7, 6, 456_000_000, time.UTC)
	parts := mustFormatRangeToParts(t, format, start, end)
	var fractional []RangePart
	for _, part := range parts {
		if part.Type == PartFractionalSecond {
			fractional = append(fractional, part)
		}
		if part.Type == PartType("fractionalSecondDigits") {
			t.Fatalf("FormatRangeToParts() emitted obsolete type in %#v", parts)
		}
	}
	want := []RangePart{
		{Type: PartFractionalSecond, Value: "123", Source: SourceStartRange},
		{Type: PartFractionalSecond, Value: "456", Source: SourceEndRange},
	}
	if !reflect.DeepEqual(fractional, want) {
		t.Fatalf("fractional-second range parts = %#v, want %#v", fractional, want)
	}
	if got, want := mustFormatRange(t, format, start, end), joinRangePartValues(parts); got != want {
		t.Fatalf("FormatRange() = %q, want joined parts %q", got, want)
	}
}

func TestDateTimeFormatFormatFlexibleDayPeriod(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "zh")}, Options{Hour: stringPtr(NumericFieldStyle), DayPeriod: stringPtr(LongFieldStyle), Hour12: boolPtr(true)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{TimeZone: stringPtr("America/New_York"), Hour: stringPtr(NumericFieldStyle), TimeZoneName: stringPtr(LongGenericTimeZoneName)})
	if err != nil {
		t.Fatalf("New(timezone name fields) error = %v", err)
	}
	parts := format.FormatToParts(time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC))
	want := []Part{
		{Type: PartHour, Value: "7"},
		{Type: PartLiteral, Value: " "},
		{Type: PartDayPeriod, Value: "AM"},
		{Type: PartLiteral, Value: " "},
		{Type: PartTimeZoneName, Value: "Eastern Time"},
	}
	if !reflect.DeepEqual(parts, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", parts, want)
	}
}

func TestDateTimeFormatUsesLocalizedRegionFormatForLocationTimeZoneName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		locale string
		want   string
	}{
		{locale: "en", want: "Dumont d’Urville Station Time"},
		{locale: "fr", want: "heure : Dumont-d’Urville"},
		{locale: "zh", want: "迪蒙·迪维尔时间"},
	}
	for _, tc := range tests {
		t.Run(tc.locale, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, tc.locale)}, Options{
				TimeZone:     stringPtr("Antarctica/DumontDUrville"),
				Hour:         stringPtr(NumericFieldStyle),
				TimeZoneName: stringPtr(ShortGenericTimeZoneName),
			})
			if err != nil {
				t.Fatalf("New(location timezone name) error = %v", err)
			}
			parts := format.FormatToParts(time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC))
			for _, part := range parts {
				if part.Type == PartTimeZoneName {
					if part.Value != tc.want {
						t.Fatalf("timeZoneName part = %q, want %q", part.Value, tc.want)
					}
					return
				}
			}
			t.Fatal("FormatToParts() did not return a timeZoneName part")
		})
	}
}

func TestDateTimeFormatFormatsLocalizedTimeZoneNameForms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		style TimeZoneName
		want  string
	}{
		{name: "short specific", style: ShortTimeZoneName, want: "7 AM EST"},
		{name: "long specific", style: LongTimeZoneName, want: "7 AM Eastern Standard Time"},
		{name: "default utc", style: ShortTimeZoneName, want: "12 PM GMT"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var timeZone *string
			date := time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC)
			if tc.name != "default utc" {
				timeZone = stringPtr("America/New_York")
			}
			format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{TimeZone: timeZone, Hour: stringPtr(NumericFieldStyle), TimeZoneName: stringPtr(tc.style)})
			if err != nil {
				t.Fatalf("New(timezone name fields) error = %v", err)
			}
			if got := format.Format(date); got != tc.want {
				t.Fatalf("Format() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDateTimeFormatUsesHistoricalTransitionDSTFlagForSpecificName(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-GB")}, Options{
		TimeZone:     stringPtr("Europe/London"),
		Hour:         stringPtr(NumericFieldStyle),
		TimeZoneName: stringPtr(LongTimeZoneName),
	})
	if err != nil {
		t.Fatalf("New(historical timezone name) error = %v", err)
	}
	parts := format.FormatToParts(time.Date(1941, time.January, 15, 12, 0, 0, 0, time.UTC))
	for _, part := range parts {
		if part.Type == PartTimeZoneName {
			if got, want := part.Value, "British Summer Time"; got != want {
				t.Fatalf("timeZoneName part = %q, want %q", got, want)
			}
			return
		}
	}
	t.Fatal("FormatToParts() did not return a timeZoneName part")
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
		{name: "en short whole hour", locale: "en-US", timeZone: "+02:00", style: ShortOffsetTimeZoneName, want: "2 AM GMT+2"},
		{name: "en long whole hour", locale: "en-US", timeZone: "+02:00", style: LongOffsetTimeZoneName, want: "2 AM GMT+02:00"},
		{name: "en short zero", locale: "en-US", timeZone: "+00:00", style: ShortOffsetTimeZoneName, want: "12 AM GMT"},
		{name: "en long zero", locale: "en-US", timeZone: "+00:00", style: LongOffsetTimeZoneName, want: "12 AM GMT+00:00"},
		{name: "en half hour", locale: "en-US", timeZone: "-03:30", style: ShortOffsetTimeZoneName, want: "8 PM GMT-3:30"},
		{name: "en quarter hour", locale: "en-US", timeZone: "+05:45", style: ShortOffsetTimeZoneName, want: "5 AM GMT+5:45"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, tc.locale)}, Options{TimeZone: stringPtr(tc.timeZone), Hour: stringPtr(NumericFieldStyle), TimeZoneName: stringPtr(tc.style)})
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

	newYork, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{TimeZone: stringPtr("America/New_York"), Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle), Hour: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(New_York) error = %v", err)
	}
	shanghai, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{TimeZone: stringPtr("Asia/Shanghai"), Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle), Hour: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(false)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Hour: stringPtr(TwoDigitFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Second: stringPtr(TwoDigitFieldStyle), FractionalSecondDigits: intPtr(3), Hour12: boolPtr(false)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(FullDateTimeStyle)})
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

	format, err := New(locale.List{intltest.Locale(t, "zh-Hans-CN")}, Options{DateStyle: stringPtr(FullDateTimeStyle)})
	if err != nil {
		t.Fatalf("New(zh-Hans-CN dateStyle full) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := mustGregorianForDateLocale(t, "zh-Hans-CN")
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(ShortDateTimeStyle)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{TimeStyle: stringPtr(ShortDateTimeStyle), Hour12: boolPtr(false)})
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

	format, err := New(locale.List{intltest.Locale(t, "zh-Hans-CN")}, Options{TimeStyle: stringPtr(MediumDateTimeStyle), Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(zh-Hans timeStyle medium) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 9, 7, 6, 0, time.UTC)
	gregorian := mustGregorianForDateLocale(t, "zh-Hans-CN")
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{TimeStyle: stringPtr(LongDateTimeStyle), TimeZone: stringPtr("America/New_York"), Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(en-US timeStyle long) error = %v", err)
	}
	date := time.Date(2026, time.January, 8, 12, 7, 6, 0, time.UTC)
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(LongDateTimeStyle), TimeStyle: stringPtr(FullDateTimeStyle), TimeZone: stringPtr("America/New_York"), Hour12: boolPtr(false)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(MediumDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), Hour12: boolPtr(false)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(FullDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(full dateStyle+short timeStyle) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC))
	if want := "Friday, May 8, 2026 at 9:07 AM"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatCombinesLongMonthDateAndTimeWithAtConnector(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(LongMonthStyle), Day: stringPtr(NumericFieldStyle), Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(long month date+time fields) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC))
	if want := "May 8, 2026 at 9:07 AM"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatCombinesDateAndTimeStylesWithCLDRConnector(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "zh-Hans-CN")}, Options{DateStyle: stringPtr(MediumDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(zh-Hans dateStyle+timeStyle) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	gregorian := mustGregorianForDateLocale(t, "zh-Hans-CN")
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{FormatMatcher: stringPtr(BasicFormatMatcher), Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(FormatMatcher=basic) error = %v", err)
	}
	if got, want := format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)), "May 8, 2026"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatResolvedComponentsFollowSelectedPattern(t *testing.T) {
	t.Parallel()

	for _, matcher := range []FormatMatcher{BasicFormatMatcher, BestFitFormatMatcher} {
		t.Run(string(matcher), func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
				FormatMatcher: stringPtr(matcher),
				Hour:          stringPtr(NumericFieldStyle),
				Minute:        stringPtr(NumericFieldStyle),
				TimeZone:      stringPtr("UTC"),
			})
			if err != nil {
				t.Fatal(err)
			}
			resolved := format.ResolvedOptions()
			if got, want := ecma402.ResolvedScalarValue(resolved.Hour), NumericFieldStyle; got != want {
				t.Fatalf("ResolvedOptions().Hour = %q, want %q", got, want)
			}
			if got, want := ecma402.ResolvedScalarValue(resolved.Minute), TwoDigitFieldStyle; got != want {
				t.Fatalf("ResolvedOptions().Minute = %q, want selected pattern width %q", got, want)
			}
			if got, want := format.Format(time.Date(2026, time.January, 1, 9, 5, 0, 0, time.UTC)), "9:05 AM"; got != want {
				t.Fatalf("Format() = %q, want %q", got, want)
			}
		})
	}
}

func TestDateTimeFormatStyleResolvedOptionsSuppressGranularFields(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(MediumDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle)})
	if err != nil {
		t.Fatalf("New(dateStyle+timeStyle) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if got, want := ecma402.ResolvedScalarValue(resolved.DateStyle), Style("medium"); got != want {
		t.Fatalf("ResolvedOptions().DateStyle = %q, want %q", got, want)
	}
	if got, want := ecma402.ResolvedScalarValue(resolved.TimeStyle), Style("short"); got != want {
		t.Fatalf("ResolvedOptions().TimeStyle = %q, want %q", got, want)
	}
	if resolved.Year != nil || resolved.Month != nil || resolved.Day != nil || resolved.Hour != nil || resolved.Minute != nil {
		t.Fatalf("ResolvedOptions() granular fields = year:%v month:%v day:%v hour:%v minute:%v, want suppressed", resolved.Year, resolved.Month, resolved.Day, resolved.Hour, resolved.Minute)
	}
}

func TestDateTimeFormatRangeEqualInstantsUsesSingleFormat(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	if got, want := mustFormatRange(t, format, date, date), format.Format(date); got != want {
		t.Fatalf("FormatRange(equal) = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeToPartsEqualInstantsAreShared(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	parts := mustFormatRangeToParts(t, format, date, date)
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
	if got, want := mustFormatRange(t, format, start, end), "May 8\u2009–\u200910, 2026"; got != want {
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
	if got := mustFormatRangeToParts(t, format, start, end); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatRangeToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatRangeUsesIntervalPatternForDifferentMonths(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)
	if got, want := mustFormatRange(t, format, start, end), "May 8\u2009–\u2009Jun 10, 2026"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeUsesIntervalPatternForDifferentYears(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, time.June, 10, 0, 0, 0, 0, time.UTC)
	if got, want := mustFormatRange(t, format, start, end), "May 8, 2026\u2009–\u2009Jun 10, 2027"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeUsesTimeIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(time fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 10, 7, 0, 0, time.UTC)
	if got, want := mustFormatRange(t, format, start, end), "9:07\u2009–\u200910:07 AM"; got != want {
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
		{Type: PartLiteral, Value: " ", Source: SourceShared},
		{Type: PartDayPeriod, Value: "AM", Source: SourceShared},
	}
	if got := mustFormatRangeToParts(t, format, start, end); !reflect.DeepEqual(got, wantParts) {
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
			opts:  Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(true)},
			start: time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC),
			end:   time.Date(2026, time.May, 8, 9, 8, 0, 0, time.UTC),
			want:  "9:07\u2009–\u20099:08 AM",
		},
		{
			name:  "second",
			opts:  Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Second: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(true)},
			start: time.Date(2026, time.May, 8, 9, 7, 6, 0, time.UTC),
			end:   time.Date(2026, time.May, 8, 9, 7, 8, 0, time.UTC),
			want:  "9:07:06 AM\u2009–\u20099:07:08 AM",
		},
		{
			name:  "fractional second",
			opts:  Options{Hour: stringPtr(TwoDigitFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Second: stringPtr(TwoDigitFieldStyle), FractionalSecondDigits: intPtr(3), Hour12: boolPtr(false)},
			start: time.Date(2026, time.May, 8, 9, 7, 6, 123_000_000, time.UTC),
			end:   time.Date(2026, time.May, 8, 9, 7, 6, 456_000_000, time.UTC),
			want:  "09:07:06.123\u2009–\u200909:07:06.456",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en-US")}, tc.opts)
			if err != nil {
				t.Fatalf("New(time fields) error = %v", err)
			}
			if got := mustFormatRange(t, format, tc.start, tc.end); got != tc.want {
				t.Fatalf("FormatRange() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDateTimeFormatRangeUsesFlexibleDayPeriodIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Hour: stringPtr(NumericFieldStyle), DayPeriod: stringPtr(LongFieldStyle), Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(dayPeriod fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 8, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 15, 0, 0, 0, time.UTC)
	if got, want := mustFormatRange(t, format, start, end), "8 in the morning\u2009–\u20093 in the afternoon"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeCombinesSharedDateWithTimeIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle), Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(date+time fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 10, 7, 0, 0, time.UTC)
	if got, want := mustFormatRange(t, format, start, end), "May 8, 2026, 9:07\u2009–\u200910:07 AM"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeDateTimeStyleUsesTimeIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(MediumDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(dateStyle+timeStyle) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 10, 7, 0, 0, time.UTC)
	if got, want := mustFormatRange(t, format, start, end), "May 8, 2026, 9:07\u2009–\u200910:07 AM"; got != want {
		t.Fatalf("FormatRange() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatRangeDateStyleUsesIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(FullDateTimeStyle)})
	if err != nil {
		t.Fatalf("New(dateStyle) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 10, 0, 0, 0, 0, time.UTC)
	want := "Friday, May 8\u2009–\u2009Wednesday, June 10, 2026"
	if got := mustFormatRange(t, format, start, end); got != want {
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
	if got := mustFormatRangeToParts(t, format, start, end); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatRangeToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatRangeFallbackPartsRemainJoined(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{DateStyle: stringPtr(FullDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), Hour12: boolPtr(true)})
	if err != nil {
		t.Fatalf("New(dateStyle+timeStyle) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	end := time.Date(2026, time.June, 10, 10, 7, 0, 0, time.UTC)
	want := "Friday, May 8, 2026 at 9:07 AM\u2009–\u2009Wednesday, June 10, 2026 at 10:07 AM"
	if got := mustFormatRange(t, format, start, end); got != want {
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
		{Type: PartLiteral, Value: " ", Source: SourceStartRange},
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
		{Type: PartLiteral, Value: " ", Source: SourceEndRange},
		{Type: PartDayPeriod, Value: "AM", Source: SourceEndRange},
	}
	if got := mustFormatRangeToParts(t, format, start, end); !reflect.DeepEqual(got, wantParts) {
		t.Fatalf("FormatRangeToParts() = %#v, want %#v", got, wantParts)
	}
}

func TestDateTimeFormatRangePreservesReversedInputOrder(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2027, time.June, 10, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	got, err := format.FormatRange(start, end)
	if err != nil {
		t.Fatalf("FormatRange(reversed) error = %v", err)
	}
	if want := "Jun 10, 2027\u2009–\u2009May 8, 2026"; got != want {
		t.Fatalf("FormatRange(reversed) = %q, want %q", got, want)
	}
	parts, err := format.FormatRangeToParts(start, end)
	if err != nil {
		t.Fatalf("FormatRangeToParts(reversed) error = %v", err)
	}
	if joined := joinRangePartValues(parts); joined != got {
		t.Fatalf("joined FormatRangeToParts(reversed) = %q, want %q", joined, got)
	}
	var startValues, endValues strings.Builder
	for _, part := range parts {
		switch part.Source {
		case SourceStartRange:
			startValues.WriteString(part.Value)
		case SourceEndRange:
			endValues.WriteString(part.Value)
		case SourceShared:
		}
	}
	if got, want := startValues.String(), "Jun 10, 2027"; got != want {
		t.Fatalf("startRange values = %q, want first argument %q", got, want)
	}
	if got, want := endValues.String(), "May 8, 2026"; got != want {
		t.Fatalf("endRange values = %q, want second argument %q", got, want)
	}
}

func TestDateTimeFormatRangePreservesReversedTimeAndFallbackOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		opts      Options
		start     time.Time
		end       time.Time
		want      string
		wantStart string
		wantEnd   string
	}{
		{
			name:      "same-day time interval",
			opts:      Options{Hour: stringPtr(NumericFieldStyle), Minute: stringPtr(TwoDigitFieldStyle), Hour12: boolPtr(true)},
			start:     time.Date(2026, time.May, 8, 10, 7, 0, 0, time.UTC),
			end:       time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC),
			want:      "10:07\u2009–\u20099:07 AM",
			wantStart: "10:07",
			wantEnd:   "9:07",
		},
		{
			name:      "date-time fallback",
			opts:      Options{DateStyle: stringPtr(FullDateTimeStyle), TimeStyle: stringPtr(ShortDateTimeStyle), Hour12: boolPtr(true)},
			start:     time.Date(2026, time.June, 10, 10, 7, 0, 0, time.UTC),
			end:       time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC),
			want:      "Wednesday, June 10, 2026 at 10:07 AM\u2009–\u2009Friday, May 8, 2026 at 9:07 AM",
			wantStart: "Wednesday, June 10, 2026 at 10:07 AM",
			wantEnd:   "Friday, May 8, 2026 at 9:07 AM",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			format, err := New(locale.List{intltest.Locale(t, "en-US")}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			got, err := format.FormatRange(tc.start, tc.end)
			if err != nil {
				t.Fatalf("FormatRange(reversed) error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("FormatRange(reversed) = %q, want %q", got, tc.want)
			}
			parts, err := format.FormatRangeToParts(tc.start, tc.end)
			if err != nil {
				t.Fatalf("FormatRangeToParts(reversed) error = %v", err)
			}
			if joined := joinRangePartValues(parts); joined != got {
				t.Fatalf("joined FormatRangeToParts(reversed) = %q, want %q", joined, got)
			}
			var startValues, endValues strings.Builder
			for _, part := range parts {
				switch part.Source {
				case SourceStartRange:
					startValues.WriteString(part.Value)
				case SourceEndRange:
					endValues.WriteString(part.Value)
				case SourceShared:
				}
			}
			if got := startValues.String(); got != tc.wantStart {
				t.Fatalf("startRange values = %q, want first argument %q", got, tc.wantStart)
			}
			if got := endValues.String(); got != tc.wantEnd {
				t.Fatalf("endRange values = %q, want second argument %q", got, tc.wantEnd)
			}
		})
	}
}

func TestDateTimeFormatRangeEqualsJoinedRangeParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	start := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 10, 0, 0, 0, 0, time.UTC)
	parts := mustFormatRangeToParts(t, format, start, end)
	var joined strings.Builder
	var hasStart, hasShared, hasEnd bool
	for _, part := range parts {
		joined.WriteString(part.Value)
		hasStart = hasStart || part.Source == SourceStartRange
		hasShared = hasShared || part.Source == SourceShared
		hasEnd = hasEnd || part.Source == SourceEndRange
	}
	if got, want := joined.String(), mustFormatRange(t, format, start, end); got != want {
		t.Fatalf("joined FormatRangeToParts values = %q, want FormatRange() %q", got, want)
	}
	if !hasStart || !hasShared || !hasEnd {
		t.Fatalf("FormatRangeToParts sources start=%v shared=%v end=%v, want all present in %#v", hasStart, hasShared, hasEnd, parts)
	}
}

func ExampleDateTimeFormat_Format() {
	format, err := New(mustLocaleList("en-US"), Options{Year: stringPtr(NumericFieldStyle), Month: stringPtr(ShortMonthStyle), Day: stringPtr(NumericFieldStyle)})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)))

	// Output:
	// May 8, 2026
}

func ExampleDateTimeFormat_Format_timezone() {
	format, err := New(mustLocaleList("en-US"), Options{TimeZone: stringPtr("America/New_York"), Hour: stringPtr(NumericFieldStyle), TimeZoneName: stringPtr(LongGenericTimeZoneName)})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format(time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC)))

	// Output:
	// 7 AM Eastern Time
}

func ExampleDateTimeFormat_FormatToParts() {
	format, err := New(mustLocaleList("en-US"), Options{Weekday: stringPtr(LongFieldStyle), Month: stringPtr(LongMonthStyle), Day: stringPtr(NumericFieldStyle), Year: stringPtr(NumericFieldStyle)})
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

func mustLocaleList(tags ...string) locale.List {
	locales, err := locale.ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "fr-FR"), intltest.Locale(t, "en-US-u-hc-h23"), intltest.Locale(t, "ban")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf()", got, []string{"fr-FR", "en-US-u-hc-h23"})
}

func TestSupportedLocalesOfPreservesUnsupportedLocaleCalendars(t *testing.T) {
	t.Parallel()

	requested := intltest.LocaleList(t, "en-US-u-ca-buddhist", "en-US-u-ca-iso8601", "en-US")
	got, err := SupportedLocalesOf(requested, Options{})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf()", got, []string{"en-US-u-ca-buddhist", "en-US-u-ca-iso8601", "en-US"})
}

func TestSupportedLocalesOfErrors(t *testing.T) {
	t.Parallel()

	for _, matcher := range []string{"fast", ""} {
		t.Run(matcher, func(t *testing.T) {
			t.Parallel()
			_, err := SupportedLocalesOf(nil, Options{LocaleMatcher: stringPtr(matcher)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "datetimeformat", intlerr.InvalidOption, "localeMatcher", matcher, "en")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}
}
