package datetimeformat

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/locale"
)

func TestMain(m *testing.M) {
	oldLocal := time.Local
	time.Local = time.UTC
	code := m.Run()
	time.Local = oldLocal
	os.Exit(code)
}

func TestDateTimeFormatDefaultResolvedOptions(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"))
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

	format, err := New(locale.MustParse("en-US"))
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

	format, err := New(locale.MustParse("zh-Hans-CN"))
	if err != nil {
		t.Fatalf("New(zh-Hans-CN) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := cldr.GregorianFor(format.cldrLoc)
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

	format, err := New(locale.MustParse("en-US-u-nu-latn"))
	if err != nil {
		t.Fatalf("New(en-US-u-nu-latn) error = %v", err)
	}
	if got, want := format.ResolvedOptions().NumberingSystem, "latn"; got != want {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, want)
	}
}

func TestDateTimeFormatFallsBackToDateDataLocale(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("zh-Hans-CN"), Options{Hour: NumericFieldStyle, DayPeriod: LongFieldStyle})
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

	format, err := New(locale.MustParse("fr-FR"))
	if err != nil {
		t.Fatalf("New(fr-FR) error = %v", err)
	}
	if got, want := format.ResolvedOptions().NumberingSystem, "latn"; got != want {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, want)
	}
}

func TestDateTimeFormatUnsupportedCalendarOptionFallsBack(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US")
	format, err := New(loc, Options{Calendar: "buddhist"})
	if err != nil {
		t.Fatalf("New(WithCalendar(buddhist)) error = %v", err)
	}
	if got := format.ResolvedOptions().Calendar; got != "gregory" {
		t.Fatalf("ResolvedOptions().Calendar = %q, want gregory fallback", got)
	}
}

func TestDateTimeFormatUnsupportedLocaleCalendarFallsBack(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US-u-ca-buddhist")
	format, err := New(loc)
	if err != nil {
		t.Fatalf("New(en-US-u-ca-buddhist) error = %v", err)
	}
	resolved := format.ResolvedOptions()
	if resolved.Calendar != "gregory" || strings.Contains(resolved.Locale.String(), "ca-buddhist") {
		t.Fatalf("ResolvedOptions() = %#v, want gregory without unsupported calendar extension", resolved)
	}
}

func TestDateTimeFormatNumberingSystemOptionLocalizesDigits(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{NumberingSystem: "arab"})
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
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(loc, tt.opt)
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New(%s=%q) error = %v, want ErrInvalidOption", tt.name, tt.value, err)
			}
			if !strings.Contains(err.Error(), tt.name) || !strings.Contains(err.Error(), tt.value) {
				t.Fatalf("New(%s=%q) error = %v, want option name and value in message", tt.name, tt.value, err)
			}
		})
	}
}

func TestDateTimeFormatCanonicalKeyOmitsDefaults(t *testing.T) {
	t.Parallel()

	if got := CanonicalKey(); got != "" {
		t.Fatalf("CanonicalKey() = %q, want empty default key", got)
	}
	if got := CanonicalKey(Options{FormatMatcher: BasicFormatMatcher}); !strings.Contains(got, "formatMatcher=basic") {
		t.Fatalf("CanonicalKey(FormatMatcher=basic) = %q, want formatMatcher", got)
	}
	if got := CanonicalKey(Options{LocaleMatcher: LookupLocaleMatcher}); !strings.Contains(got, "localeMatcher=lookup") {
		t.Fatalf("CanonicalKey(LocaleMatcher=lookup) = %q, want localeMatcher", got)
	}
}

func TestDateTimeFormatRejectsInvalidFractionalSecondDigits(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US")
	for _, digits := range []int{-1, 4} {
		t.Run(string(rune('0'+digits)), func(t *testing.T) {
			t.Parallel()

			_, err := New(loc, Options{FractionalSecondDigits: FractionalSecondDigits(digits)})
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New(fractionalSecondDigits=%d) error = %v, want ErrInvalidOption", digits, err)
			}
			if !strings.Contains(err.Error(), "fractionalSecondDigits") || !strings.Contains(err.Error(), fmt.Sprint(digits)) {
				t.Fatalf("New(fractionalSecondDigits=%d) error = %v, want option name and value in message", digits, err)
			}
		})
	}
}

func TestDateTimeFormatRejectsDateStyleWithGranularField(t *testing.T) {
	t.Parallel()

	_, err := New(locale.MustParse("en-US"), Options{DateStyle: MediumDateTimeStyle, Year: NumericFieldStyle})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New(dateStyle+year) error = %v, want ErrInvalidOption", err)
	}
	if !strings.Contains(err.Error(), "dateStyle") || !strings.Contains(err.Error(), "year") {
		t.Fatalf("New(dateStyle+year) error = %v, want dateStyle and granular field in message", err)
	}
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
		{name: "timeStyle fractional second", opts: Options{TimeStyle: ShortDateTimeStyle, FractionalSecondDigits: FractionalSecondDigits(3)}, want: "fractionalSecondDigits"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(loc, tc.opts)
			if !errors.Is(err, ErrInvalidOption) {
				t.Fatalf("New() error = %v, want ErrInvalidOption", err)
			}
			if !strings.Contains(err.Error(), "dateStyle/timeStyle") || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("New() error = %v, want style and %s context", err, tc.want)
			}
		})
	}
}

func TestDateTimeFormatResolvedOptionsReturnsStableSnapshots(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{TimeZone: "US/Eastern"})
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
	} {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.MustParse("en-US"), Options{TimeZone: tc.in})
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
	_, err := New(loc, Options{TimeZone: "Mars/Olympus"})
	if !errors.Is(err, ErrUnsupportedTimeZone) {
		t.Fatalf("New(Options{TimeZone: Mars/Olympus}) error = %v, want ErrUnsupportedTimeZone", err)
	}
	if !strings.Contains(err.Error(), "Mars/Olympus") || !strings.Contains(err.Error(), loc.String()) {
		t.Fatalf("New(Options{TimeZone: Mars/Olympus}) error = %v, want timezone and locale in message", err)
	}
}

func TestDateTimeFormatAllowsEmptyTimeZone(t *testing.T) {
	oldLocal := time.Local
	local, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	time.Local = local
	t.Cleanup(func() {
		time.Local = oldLocal
	})
	format, err := New(locale.MustParse("en-US"), Options{TimeZone: ""})
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

	format, err := New(locale.MustParse("en-US"), Options{Hour: NumericFieldStyle, Hour12: Hour12(false)})
	if err != nil {
		t.Fatalf("New(WithHour+Options{Hour12: Hour12(false)}) error = %v", err)
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

	format, err := New(locale.MustParse("en-US"), Options{Hour: NumericFieldStyle, HourCycle: H11HourCycle, Hour12: Hour12(true)})
	if err != nil {
		t.Fatalf("New(WithHourCycle+Options{Hour12: Hour12(true)}) error = %v", err)
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

	format, err := New(locale.MustParse("en-US-u-hc-h23"), Options{Hour: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(en-US-u-hc-h23, WithHour) error = %v", err)
	}
	if got, want := format.ResolvedOptions().HourCycle, HourCycle("h23"); got != want {
		t.Fatalf("ResolvedOptions().HourCycle = %q, want %q", got, want)
	}
}

func TestDateTimeFormatFormatDateOnly(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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

	format, err := New(locale.MustParse("zh-Hans-CN"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(zh-Hans short date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := cldr.GregorianFor(format.cldrLoc)
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	var joined strings.Builder
	for _, part := range format.FormatToParts(date) {
		joined.WriteString(part.Value)
	}
	if got, want := joined.String(), format.Format(date); got != want {
		t.Fatalf("joined FormatToParts values = %q, want Format() %q", got, want)
	}
}

func TestDateTimeFormatFormatToPartsDateFields(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Weekday: LongFieldStyle, Month: LongMonthStyle, Day: NumericFieldStyle, Year: NumericFieldStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{Era: ShortFieldStyle, Year: NumericFieldStyle})
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

func TestDateTimeFormatUsesCLDRDateFieldNames(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("zh-Hans-CN"), Options{Weekday: LongFieldStyle, Month: LongMonthStyle, Day: NumericFieldStyle, Year: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(zh-Hans date fields) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := cldr.GregorianFor(format.cldrLoc)
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
		case PartYear, PartDay, PartHour, PartMinute, PartSecond, PartEra, PartDayPeriod, PartTimeZoneName, PartLiteral, PartFractionalSecondDigits:
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
			opts:   Options{Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: Hour12(true)},
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
			opts:   Options{Hour: TwoDigitFieldStyle, Minute: TwoDigitFieldStyle, Second: TwoDigitFieldStyle, Hour12: Hour12(false)},
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

			format, err := New(locale.MustParse(tt.locale), tt.opts)
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

	format, err := New(locale.MustParse("en-US"), Options{Hour: TwoDigitFieldStyle, Minute: TwoDigitFieldStyle, Second: TwoDigitFieldStyle, FractionalSecondDigits: FractionalSecondDigits(3), Hour12: Hour12(false)})
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

	format, err := New(locale.MustParse("zh"), Options{Hour: NumericFieldStyle, DayPeriod: LongFieldStyle, Hour12: Hour12(true)})
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

	format, err := New(locale.MustParse("en-US"), Options{TimeZone: "America/New_York", Hour: NumericFieldStyle, TimeZoneName: LongGenericTimeZoneName})
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

			format, err := New(locale.MustParse(tc.locale), Options{TimeZone: tc.timeZone, Hour: NumericFieldStyle, TimeZoneName: tc.style})
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

	newYork, err := New(locale.MustParse("en-US"), Options{TimeZone: "America/New_York", Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle, Hour: TwoDigitFieldStyle, Hour12: Hour12(false)})
	if err != nil {
		t.Fatalf("New(New_York) error = %v", err)
	}
	shanghai, err := New(locale.MustParse("en-US"), Options{TimeZone: "Asia/Shanghai", Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle, Hour: TwoDigitFieldStyle, Hour12: Hour12(false)})
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

	format, err := New(locale.MustParse("en-US"), Options{Hour: TwoDigitFieldStyle, Minute: TwoDigitFieldStyle, Second: TwoDigitFieldStyle, FractionalSecondDigits: FractionalSecondDigits(3), Hour12: Hour12(false)})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{DateStyle: FullDateTimeStyle})
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

	format, err := New(locale.MustParse("zh-Hans-CN"), Options{DateStyle: FullDateTimeStyle})
	if err != nil {
		t.Fatalf("New(zh-Hans-CN dateStyle full) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)
	gregorian := cldr.GregorianFor(format.cldrLoc)
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

	format, err := New(locale.MustParse("en-US"), Options{DateStyle: ShortDateTimeStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{TimeStyle: ShortDateTimeStyle, Hour12: Hour12(false)})
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

	format, err := New(locale.MustParse("zh-Hans-CN"), Options{TimeStyle: MediumDateTimeStyle, Hour12: Hour12(false)})
	if err != nil {
		t.Fatalf("New(zh-Hans timeStyle medium) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 9, 7, 6, 0, time.UTC)
	gregorian := cldr.GregorianFor(format.cldrLoc)
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

	format, err := New(locale.MustParse("en-US"), Options{TimeStyle: LongDateTimeStyle, TimeZone: "America/New_York", Hour12: Hour12(false)})
	if err != nil {
		t.Fatalf("New(en-US timeStyle long) error = %v", err)
	}
	date := time.Date(2026, time.January, 8, 12, 7, 6, 0, time.UTC)
	gregorian := cldr.GregorianFor(format.cldrLoc)
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

func TestDateTimeFormatCombinesDateAndTimeStyles(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: Hour12(false)})
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

	format, err := New(locale.MustParse("en-US"), Options{DateStyle: FullDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: Hour12(true)})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: LongMonthStyle, Day: NumericFieldStyle, Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: Hour12(true)})
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

	format, err := New(locale.MustParse("zh-Hans-CN"), Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: Hour12(false)})
	if err != nil {
		t.Fatalf("New(zh-Hans dateStyle+timeStyle) error = %v", err)
	}
	date := time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC)
	gregorian := cldr.GregorianFor(format.cldrLoc)
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

	format, err := New(locale.MustParse("en-US"), Options{FormatMatcher: BasicFormatMatcher, Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		t.Fatalf("New(FormatMatcher=basic) error = %v", err)
	}
	if got, want := format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)), "May 8, 2026"; got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDateTimeFormatStyleResolvedOptionsSuppressGranularFields(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: Hour12(true)})
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

func TestDateTimeFormatRangeCombinesSharedDateWithTimeIntervalPattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle, Hour: NumericFieldStyle, Minute: TwoDigitFieldStyle, Hour12: Hour12(true)})
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

	format, err := New(locale.MustParse("en-US"), Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: Hour12(true)})
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

	format, err := New(locale.MustParse("en-US"), Options{DateStyle: FullDateTimeStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{DateStyle: FullDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: Hour12(true)})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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

	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
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
	format, err := New(locale.MustParse("en-US"), Options{Year: NumericFieldStyle, Month: ShortMonthStyle, Day: NumericFieldStyle})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)))

	// Output:
	// May 8, 2026
}

func ExampleDateTimeFormat_Format_timezone() {
	format, err := New(locale.MustParse("en-US"), Options{TimeZone: "America/New_York", Hour: NumericFieldStyle, TimeZoneName: LongGenericTimeZoneName})
	if err != nil {
		panic(err)
	}

	fmt.Println(format.Format(time.Date(2026, time.January, 8, 12, 0, 0, 0, time.UTC)))

	// Output:
	// 7 AM Eastern Time
}

func ExampleDateTimeFormat_FormatToParts() {
	format, err := New(locale.MustParse("en-US"), Options{Weekday: LongFieldStyle, Month: LongMonthStyle, Day: NumericFieldStyle, Year: NumericFieldStyle})
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

func TestDateTimeFormatRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	_, err := New(locale.MustParse("en-US"), Options{DateStyle: ShortDateTimeStyle}, Options{TimeStyle: ShortDateTimeStyle})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := []locale.Locale{
		locale.MustParse("fr-FR"),
		locale.MustParse("en-US-u-hc-h23"),
		locale.MustParse("ban"),
	}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"en-US-u-hc-h23"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}
