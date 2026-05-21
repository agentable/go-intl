package durationformat

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/locale"
)

func TestDurationFormatResolvedOptionsDefault(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	got := format.ResolvedOptions()
	want := ResolvedOptions{
		Locale:              locale.MustParse("en"),
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

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{
		locale.MustParse("en-US"),
		locale.MustParse("hi"),
		locale.MustParse("zh-Hans-CN"),
	}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := locale.List{locale.MustParse("en-US"), locale.MustParse("hi"), locale.MustParse("zh-Hans-CN")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}

	_, err = SupportedLocalesOf(requested, Options{LocaleMatcher: LocaleMatcher("bad")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
	}

	_, err = SupportedLocalesOf(requested, Options{Style: Style("bad")})
	if err != nil {
		t.Fatalf("SupportedLocalesOf(invalid formatting option) error = %v, want nil", err)
	}
}

func TestDurationFormatFormatDefaultShort(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{})
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

	format, err := New(locale.List{locale.MustParse("en")}, Options{})
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

	format, err := New(locale.List{locale.MustParse("en")}, Options{Style: DigitalStyle})
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

func TestDurationFormatRejectsInvalidDurationValues(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}

	tests := []struct {
		name     string
		duration Duration
	}{
		{name: "mixed signs", duration: Duration{Hours: 1, Minutes: -1}},
		{name: "years too large", duration: Duration{Years: 1 << 32}},
		{name: "months too negative", duration: Duration{Months: -(1 << 32)}},
		{name: "normalized seconds too large", duration: Duration{Days: 1 << 40}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if _, err := format.Format(tc.duration); !errors.Is(err, intlerr.ErrInvalidValue) {
				t.Fatalf("Format(%s) error = %v, want intlerr.ErrInvalidValue", tc.name, err)
			}
			if _, err := format.FormatToParts(tc.duration); !errors.Is(err, intlerr.ErrInvalidValue) {
				t.Fatalf("FormatToParts(%s) error = %v, want intlerr.ErrInvalidValue", tc.name, err)
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

			_, err := New(locale.List{locale.MustParse("en")}, Options{FractionalDigits: intPtr(tc.digits)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(fractionalDigits=%d) error = %v, want intlerr.ErrInvalidOption", tc.digits, err)
			}
		})
	}
}

func TestDurationFormatFractionalDigitsExplicitZero(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParseList("en"), Options{FractionalDigits: intPtr(0)})
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

	en := locale.List{locale.MustParse("en")}
	tests := []struct {
		name string
		opts Options
	}{
		{name: "locale matcher", opts: Options{LocaleMatcher: LocaleMatcher("bad")}},
		{name: "style", opts: Options{Style: Style("bad")}},
		{name: "numbering system", opts: Options{NumberingSystem: "bad!"}},
		{name: "date unit numeric style", opts: Options{Years: NumericUnitStyle}},
		{name: "unit display", opts: Options{HoursDisplay: Display("sometimes")}},
		{name: "fractional unit always display", opts: Options{Milliseconds: NumericUnitStyle, MillisecondsDisplay: AlwaysDisplay}},
		{name: "fractional chain broken", opts: Options{Milliseconds: NumericUnitStyle, Microseconds: ShortUnitStyle}},
		{name: "numeric chain broken", opts: Options{Hours: NumericUnitStyle, Minutes: LongUnitStyle}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(en, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s) error = %v, want intlerr.ErrInvalidOption", tc.name, err)
			}
		})
	}
}

func TestDurationFormatSubsecondRollupIsExact(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{
		Style:        NarrowStyle,
		Milliseconds: NumericUnitStyle,
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

	format, err := New(locale.List{locale.MustParse("en")}, Options{Style: DigitalStyle})
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

func TestDurationFormatSubsecondRollupCarriesToParentUnit(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{
		Style:        NarrowStyle,
		Milliseconds: NumericUnitStyle,
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

func TestDurationFormatFormatToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{})
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

func TestDurationFormatRejectsMixedSigns(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{})
	if err != nil {
		t.Fatalf("New(en) error = %v", err)
	}
	if _, err := format.Format(Duration{Hours: 1, Minutes: -1}); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("Format(mixed signs) error = %v, want intlerr.ErrInvalidValue", err)
	}
}

func TestDurationFormatRejectsInvalidFractionalDigitsAboveMax(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{locale.MustParse("en")}, Options{FractionalDigits: intPtr(10)})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(fractionalDigits=10) error = %v, want intlerr.ErrInvalidOption", err)
	}
}
