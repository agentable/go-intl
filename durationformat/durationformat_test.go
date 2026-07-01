package durationformat

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
)

func assertDurationInvalidValue(t testing.TB, err error, name, value, loc, expected string) {
	t.Helper()

	if !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("error = %v, want intlerr.ErrInvalidValue", err)
	}
	testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "durationformat", name, value, loc)
	testcontract.AssertErrorExpected(t, err, expected)
}

func TestDurationFormatResolvedOptionsDefault(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got := format.ResolvedOptions()
	want := ResolvedOptions{
		Locale:              intltest.Locale(t, "en"),
		NumberingSystem:     "latn",
		Style:               ShortStyle,
		Years:               ShortUnitStyle,
		YearsDisplay:        AutoDisplay,
		Months:              ShortUnitStyle,
		MonthsDisplay:       AutoDisplay,
		Weeks:               ShortUnitStyle,
		WeeksDisplay:        AutoDisplay,
		Days:                ShortUnitStyle,
		DaysDisplay:         AutoDisplay,
		Hours:               ShortUnitStyle,
		HoursDisplay:        AutoDisplay,
		Minutes:             ShortUnitStyle,
		MinutesDisplay:      AutoDisplay,
		Seconds:             ShortUnitStyle,
		SecondsDisplay:      AutoDisplay,
		Milliseconds:        ShortUnitStyle,
		MillisecondsDisplay: AutoDisplay,
		Microseconds:        ShortUnitStyle,
		MicrosecondsDisplay: AutoDisplay,
		Nanoseconds:         ShortUnitStyle,
		NanosecondsDisplay:  AutoDisplay,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ResolvedOptions() = %#v, want %#v", got, want)
	}
}

func TestDurationFormatResolvedOptionsPreservesUnitOverrides(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Years:               stringPtr(LongUnitStyle),
		YearsDisplay:        stringPtr(AlwaysDisplay),
		Months:              stringPtr(NarrowUnitStyle),
		MonthsDisplay:       stringPtr(AutoDisplay),
		Weeks:               stringPtr(ShortUnitStyle),
		WeeksDisplay:        stringPtr(AlwaysDisplay),
		Days:                stringPtr(LongUnitStyle),
		DaysDisplay:         stringPtr(AutoDisplay),
		Hours:               stringPtr(NarrowUnitStyle),
		HoursDisplay:        stringPtr(AlwaysDisplay),
		Minutes:             stringPtr(LongUnitStyle),
		MinutesDisplay:      stringPtr(AutoDisplay),
		Seconds:             stringPtr(ShortUnitStyle),
		SecondsDisplay:      stringPtr(AlwaysDisplay),
		Milliseconds:        stringPtr(NarrowUnitStyle),
		MillisecondsDisplay: stringPtr(AutoDisplay),
		Microseconds:        stringPtr(LongUnitStyle),
		MicrosecondsDisplay: stringPtr(AlwaysDisplay),
		Nanoseconds:         stringPtr(ShortUnitStyle),
		NanosecondsDisplay:  stringPtr(AutoDisplay),
	})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got := format.ResolvedOptions()
	tests := []struct {
		name        string
		gotStyle    UnitStyle
		wantStyle   UnitStyle
		gotDisplay  Display
		wantDisplay Display
	}{
		{"years", got.Years, LongUnitStyle, got.YearsDisplay, AlwaysDisplay},
		{"months", got.Months, NarrowUnitStyle, got.MonthsDisplay, AutoDisplay},
		{"weeks", got.Weeks, ShortUnitStyle, got.WeeksDisplay, AlwaysDisplay},
		{"days", got.Days, LongUnitStyle, got.DaysDisplay, AutoDisplay},
		{"hours", got.Hours, NarrowUnitStyle, got.HoursDisplay, AlwaysDisplay},
		{"minutes", got.Minutes, LongUnitStyle, got.MinutesDisplay, AutoDisplay},
		{"seconds", got.Seconds, ShortUnitStyle, got.SecondsDisplay, AlwaysDisplay},
		{"milliseconds", got.Milliseconds, NarrowUnitStyle, got.MillisecondsDisplay, AutoDisplay},
		{"microseconds", got.Microseconds, LongUnitStyle, got.MicrosecondsDisplay, AlwaysDisplay},
		{"nanoseconds", got.Nanoseconds, ShortUnitStyle, got.NanosecondsDisplay, AutoDisplay},
	}
	for _, tc := range tests {
		if tc.gotStyle != tc.wantStyle {
			t.Errorf("%s style = %q, want %q", tc.name, tc.gotStyle, tc.wantStyle)
		}
		if tc.gotDisplay != tc.wantDisplay {
			t.Errorf("%s display = %q, want %q", tc.name, tc.gotDisplay, tc.wantDisplay)
		}
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "en-US"), intltest.Locale(t, "hi"), intltest.Locale(t, "zh-Hans-CN")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf()", got, []string{"en-US", "hi", "zh-Hans-CN"})

	for _, matcher := range []string{"bad", ""} {
		t.Run(matcher, func(t *testing.T) {
			t.Parallel()
			_, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(matcher)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "durationformat", intlerr.InvalidOption, "localeMatcher", matcher, "en-US")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}

	_, err = SupportedLocalesOf(requested, Options{Style: stringPtr("bad")})
	if err != nil {
		t.Fatalf("SupportedLocalesOf(invalid formatting option) error = %v, want nil", err)
	}
}

func TestDurationFormatFormatDefaultShort(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{
		Years:        1,
		Months:       2,
		Days:         3,
		Hours:        4,
		Minutes:      5,
		Seconds:      6,
		Milliseconds: 7,
	})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "1 yr, 2 mths, 3 days, 4 hr, 5 min, 6 sec, 7 ms"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatFormatEmptyDuration(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{})
	if err != nil {
		t.Fatalf("Format(empty) error = %v", err)
	}
	if got != "" {
		t.Fatalf("Format(empty) = %q, want empty string", got)
	}
	parts, err := format.FormatToParts(Duration{})
	if err != nil {
		t.Fatalf("FormatToParts(empty) error = %v", err)
	}
	if parts != nil {
		t.Fatalf("FormatToParts(empty) = %#v, want nil", parts)
	}
}

func TestDurationFormatFormatDigital(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}

	got, err := format.Format(Duration{Years: 1})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "1 yr, 0:00:00"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatFormatToPartsDigital(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}

	got, err := format.FormatToParts(Duration{Hours: 1, Minutes: 2, Seconds: 3})
	if err != nil {
		t.Fatalf("FormatToParts() error = %v", err)
	}
	want := []Part{
		{Type: PartInteger, Value: "1", Unit: Hour},
		{Type: PartLiteral, Value: ":"},
		{Type: PartInteger, Value: "02", Unit: Minute},
		{Type: PartLiteral, Value: ":"},
		{Type: PartInteger, Value: "03", Unit: Second},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", got, want)
	}
}

func TestDurationFormatUsesNumberingSystemForEmbeddedNumberFormat(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		NumberingSystem: stringPtr("arab"),
		Style:           stringPtr(DigitalStyle),
	})
	if err != nil {
		t.Fatalf("New(en, arab digital) error = %v", err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "arab" {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want %q", got, "arab")
	}

	got, err := format.Format(Duration{Hours: 1, Minutes: 2, Seconds: 3})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "١:٠٢:٠٣"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatRejectsInvalidDurationValues(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	tests := []struct {
		name         string
		duration     Duration
		wantName     string
		wantValue    string
		wantExpected string
	}{
		{
			name:         "mixed signs",
			duration:     Duration{Hours: 1, Minutes: -1},
			wantName:     "duration",
			wantValue:    "mixed signs",
			wantExpected: "all non-zero duration fields to have the same sign",
		},
		{
			name:         "years too large",
			duration:     Duration{Years: 1 << 32},
			wantName:     "years",
			wantValue:    "4294967296",
			wantExpected: "an absolute value less than 2^32",
		},
		{
			name:         "months too negative",
			duration:     Duration{Months: -(1 << 32)},
			wantName:     "months",
			wantValue:    "-4294967296",
			wantExpected: "an absolute value less than 2^32",
		},
		{
			name:         "normalized seconds too large",
			duration:     Duration{Days: 1 << 40},
			wantName:     "duration",
			wantValue:    "normalized seconds",
			wantExpected: "normalized day and smaller fields below 1e9 * 2^53 nanoseconds",
		},
		{
			name:         "normalized seconds too large from seconds",
			duration:     Duration{Seconds: 1 << 53},
			wantName:     "duration",
			wantValue:    "normalized seconds",
			wantExpected: "normalized day and smaller fields below 1e9 * 2^53 nanoseconds",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := format.Format(tc.duration); err == nil {
				t.Fatalf("Format(%s) error = nil, want intlerr.ErrInvalidValue", tc.name)
			} else {
				assertDurationInvalidValue(t, err, tc.wantName, tc.wantValue, "en", tc.wantExpected)
			}
			if _, err := format.FormatToParts(tc.duration); err == nil {
				t.Fatalf("FormatToParts(%s) error = nil, want intlerr.ErrInvalidValue", tc.name)
			} else {
				assertDurationInvalidValue(t, err, tc.wantName, tc.wantValue, "en", tc.wantExpected)
			}
		})
	}
}

func TestDurationFormatRejectsInvalidFractionalDigits(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		digits int
	}{
		{name: "negative", digits: -1},
		{name: "above max", digits: 10},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en")}, Options{FractionalDigits: intPtr(tc.digits)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(fractionalDigits=%d) error = %v, want intlerr.ErrInvalidOption", tc.digits, err)
			}
		})
	}
}

func TestDurationFormatFractionalDigitsExplicitZero(t *testing.T) {
	t.Parallel()

	format, err := New(intltest.LocaleList(t, "en"), Options{FractionalDigits: intPtr(0)})
	if err != nil {
		t.Fatalf("New(fractionalDigits=0) error = %v, want nil", err)
	}
	resolved := format.ResolvedOptions()
	if resolved.FractionalDigits == nil || *resolved.FractionalDigits != 0 {
		t.Fatalf("ResolvedOptions().FractionalDigits = %v, want pointer to 0", resolved.FractionalDigits)
	}
}

func TestDurationFormatRejectsInvalidOptions(t *testing.T) {
	t.Parallel()

	en := locale.List{intltest.Locale(t, "en")}
	tests := []struct {
		name         string
		opts         Options
		wantName     string
		wantValue    string
		wantExpected string
	}{
		{name: "locale matcher", opts: Options{LocaleMatcher: stringPtr("bad")}, wantName: "localeMatcher", wantValue: "bad"},
		{name: "locale matcher empty", opts: Options{LocaleMatcher: stringPtr("")}, wantName: "localeMatcher", wantValue: ""},
		{name: "style", opts: Options{Style: stringPtr("bad")}, wantName: "style", wantValue: "bad"},
		{name: "style empty", opts: Options{Style: stringPtr("")}, wantName: "style", wantValue: ""},
		{name: "numbering system", opts: Options{NumberingSystem: stringPtr("bad!")}, wantName: "numberingSystem", wantValue: "bad!"},
		{name: "numbering system empty", opts: Options{NumberingSystem: stringPtr("")}, wantName: "numberingSystem", wantValue: ""},
		{
			name:         "date unit numeric style",
			opts:         Options{Years: stringPtr(NumericUnitStyle)},
			wantName:     "years",
			wantValue:    "numeric",
			wantExpected: `one of "long", "short", "narrow"`,
		},
		{
			name:      "date unit empty style",
			opts:      Options{Years: stringPtr("")},
			wantName:  "years",
			wantValue: "",
		},
		{
			name:         "unit display",
			opts:         Options{HoursDisplay: stringPtr("sometimes")},
			wantName:     "hoursDisplay",
			wantValue:    "sometimes",
			wantExpected: `one of "auto", "always"`,
		},
		{
			name:      "unit display empty",
			opts:      Options{HoursDisplay: stringPtr("")},
			wantName:  "hoursDisplay",
			wantValue: "",
		},
		{
			name:         "fractional unit always display",
			opts:         Options{Milliseconds: stringPtr(NumericUnitStyle), MillisecondsDisplay: stringPtr(AlwaysDisplay)},
			wantName:     "millisecondsDisplay",
			wantValue:    "always",
			wantExpected: "auto display when formatting subsecond units as a fractional part",
		},
		{
			name:         "fractional chain broken",
			opts:         Options{Milliseconds: stringPtr(NumericUnitStyle), Microseconds: stringPtr(ShortUnitStyle)},
			wantName:     "microseconds",
			wantValue:    "short",
			wantExpected: "fractional style while continuing a subsecond fractional chain",
		},
		{
			name:         "numeric chain broken",
			opts:         Options{Hours: stringPtr(NumericUnitStyle), Minutes: stringPtr(LongUnitStyle)},
			wantName:     "minutes",
			wantValue:    "long",
			wantExpected: "numeric, 2-digit, or fractional style while continuing a digital time chain",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(en, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s) error = %v, want intlerr.ErrInvalidOption", tc.name, err)
			}
			testcontract.AssertOptionError(t, err, "durationformat", intlerr.InvalidOption, tc.wantName, tc.wantValue, "en")
			if tc.wantExpected != "" {
				testcontract.AssertOptionExpected(t, err, tc.wantExpected)
			}
		})
	}
}

func TestDurationFormatSubsecondRollupIsExact(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:        stringPtr(NarrowStyle),
		Milliseconds: stringPtr(NumericUnitStyle),
	})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{Seconds: 1, Milliseconds: 473})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "1.473s"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatDigitalNegativeSignAppearsOnce(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}
	got, err := format.Format(Duration{Hours: -1, Minutes: -2, Seconds: -3})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if got != "-1:02:03" {
		t.Fatalf("Format() = %q, want -1:02:03", got)
	}
}

func TestDurationFormatDigitalNegativeZeroLeadingUnit(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}
	got, err := format.Format(Duration{Minutes: -1})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if got != "-0:01:00" {
		t.Fatalf("Format() = %q, want -0:01:00", got)
	}
}

func TestDurationFormatDigitalFormatsSubsecondOnly(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(DigitalStyle)})
	if err != nil {
		t.Fatalf("New(en, digital) error = %v", err)
	}
	got, err := format.Format(Duration{Milliseconds: 473})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if got != "0:00:00.473" {
		t.Fatalf("Format() = %q, want 0:00:00.473", got)
	}
}

func TestDurationFormatSubsecondRollupCarriesToParentUnit(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:        stringPtr(NarrowStyle),
		Milliseconds: stringPtr(NumericUnitStyle),
	})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{Seconds: 1, Milliseconds: 1_000})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "2s"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatSubsecondRollupFromMicrosecondsToMilliseconds(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:        stringPtr(NarrowStyle),
		Microseconds: stringPtr(NumericUnitStyle),
	})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.Format(Duration{
		Milliseconds: 1,
		Microseconds: 2,
		Nanoseconds:  3,
	})
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	const want = "1.002003ms"
	if got != want {
		t.Fatalf("Format() = %q, want %q", got, want)
	}
}

func TestDurationFormatSubsecondRollupToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Style:        stringPtr(NarrowStyle),
		Milliseconds: stringPtr(NumericUnitStyle),
	})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.FormatToParts(Duration{
		Seconds:      1,
		Milliseconds: 2,
		Microseconds: 3,
		Nanoseconds:  4,
	})
	if err != nil {
		t.Fatalf("FormatToParts() error = %v", err)
	}
	want := []Part{
		{Type: PartInteger, Value: "1", Unit: Second},
		{Type: PartDecimal, Value: ".", Unit: Second},
		{Type: PartFraction, Value: "002003004", Unit: Second},
		{Type: PartUnit, Value: "s", Unit: Second},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", got, want)
	}
}

func TestDurationFormatFormatToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got, err := format.FormatToParts(Duration{Hours: 1, Minutes: 2})
	if err != nil {
		t.Fatalf("FormatToParts() error = %v", err)
	}
	want := []Part{
		{Type: PartInteger, Value: "1", Unit: Hour},
		{Type: PartLiteral, Value: " ", Unit: Hour},
		{Type: PartUnit, Value: "hr", Unit: Hour},
		{Type: PartLiteral, Value: ", "},
		{Type: PartInteger, Value: "2", Unit: Minute},
		{Type: PartLiteral, Value: " ", Unit: Minute},
		{Type: PartUnit, Value: "min", Unit: Minute},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts() = %#v, want %#v", got, want)
	}
}

func TestDurationFormatListFormatPartsFlattensGroups(t *testing.T) {
	t.Parallel()

	listType := string(listformat.Unit)
	listStyle := string(listformat.LongStyle)
	formatter, err := listformat.New(locale.List{intltest.Locale(t, "en")}, listformat.Options{
		Type:  &listType,
		Style: &listStyle,
	})
	if err != nil {
		t.Fatalf("listformat.New(en, unit long) error = %v", err)
	}

	got := durationListFormatParts(formatter, [][]Part{
		{
			{Type: PartInteger, Value: "1", Unit: Hour},
			{Type: PartLiteral, Value: " ", Unit: Hour},
			{Type: PartUnit, Value: "hr", Unit: Hour},
		},
		{
			{Type: PartInteger, Value: "2", Unit: Minute},
			{Type: PartLiteral, Value: " ", Unit: Minute},
			{Type: PartUnit, Value: "min", Unit: Minute},
		},
		{
			{Type: PartInteger, Value: "3", Unit: Second},
			{Type: PartLiteral, Value: " ", Unit: Second},
			{Type: PartUnit, Value: "sec", Unit: Second},
		},
	})
	want := []Part{
		{Type: PartInteger, Value: "1", Unit: Hour},
		{Type: PartLiteral, Value: " ", Unit: Hour},
		{Type: PartUnit, Value: "hr", Unit: Hour},
		{Type: PartLiteral, Value: ", "},
		{Type: PartInteger, Value: "2", Unit: Minute},
		{Type: PartLiteral, Value: " ", Unit: Minute},
		{Type: PartUnit, Value: "min", Unit: Minute},
		{Type: PartLiteral, Value: ", "},
		{Type: PartInteger, Value: "3", Unit: Second},
		{Type: PartLiteral, Value: " ", Unit: Second},
		{Type: PartUnit, Value: "sec", Unit: Second},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("durationListFormatParts() = %#v, want %#v", got, want)
	}
}

func TestDurationFormatFormatEqualsFormatToPartsJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		options  Options
		duration Duration
	}{
		{
			name:    "default short",
			options: Options{},
			duration: Duration{
				Years:        1,
				Months:       2,
				Days:         3,
				Hours:        4,
				Minutes:      5,
				Seconds:      6,
				Milliseconds: 7,
			},
		},
		{
			name:     "empty",
			options:  Options{},
			duration: Duration{},
		},
		{
			name:     "digital",
			options:  Options{Style: stringPtr(DigitalStyle)},
			duration: Duration{Hours: 1, Minutes: 2, Seconds: 3},
		},
		{
			name:     "negative digital",
			options:  Options{Style: stringPtr(DigitalStyle)},
			duration: Duration{Hours: -1, Minutes: -2, Seconds: -3},
		},
		{
			name: "fractional subsecond",
			options: Options{
				Style:        stringPtr(NarrowStyle),
				Milliseconds: stringPtr(NumericUnitStyle),
			},
			duration: Duration{Seconds: 1, Milliseconds: 473},
		},
		{
			name: "explicit fractional digits",
			options: Options{
				Style:            stringPtr(NarrowStyle),
				Milliseconds:     stringPtr(NumericUnitStyle),
				FractionalDigits: intPtr(2),
			},
			duration: Duration{Seconds: 1, Milliseconds: 230},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, tc.options)
			if err != nil {
				t.Fatalf("New(en) error = %v", err)
			}

			got, err := format.Format(tc.duration)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			parts, err := format.FormatToParts(tc.duration)
			if err != nil {
				t.Fatalf("FormatToParts() error = %v", err)
			}
			if text := durationPartsText(parts); text != got {
				t.Fatalf("Format() = %q, joined FormatToParts() = %q", got, text)
			}
		})
	}
}

func TestDurationFormatRejectsMixedSigns(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}
	duration := Duration{Hours: 1, Minutes: -1}
	if _, err := format.Format(duration); err == nil {
		t.Fatal("Format(mixed signs) error = nil, want intlerr.ErrInvalidValue")
	} else {
		assertDurationInvalidValue(t, err, "duration", "mixed signs", "en", expectedDurationMixedSigns)
	}
	if _, err := format.FormatToParts(duration); err == nil {
		t.Fatal("FormatToParts(mixed signs) error = nil, want intlerr.ErrInvalidValue")
	} else {
		assertDurationInvalidValue(t, err, "duration", "mixed signs", "en", expectedDurationMixedSigns)
	}
}

func TestDurationFormatRejectsInvalidFractionalDigitsAboveMax(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en")}, Options{FractionalDigits: intPtr(10)})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(fractionalDigits=10) error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func durationPartsText(parts []Part) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}
