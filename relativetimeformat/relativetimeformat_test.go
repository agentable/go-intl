package relativetimeformat

import (
	"errors"
	"math"
	"slices"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"

	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

func TestRelativeTimeFormatResolvedOptionsDefaults(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got := format.ResolvedOptions()
	if got.Locale.String() != "en" {
		t.Fatalf("ResolvedOptions().Locale = %q, want en", got.Locale.String())
	}
	if got.Style != LongStyle {
		t.Fatalf("ResolvedOptions().Style = %q, want %q", got.Style, LongStyle)
	}
	if got.Numeric != NumericAlways {
		t.Fatalf("ResolvedOptions().Numeric = %q, want %q", got.Numeric, NumericAlways)
	}
	if got.NumberingSystem == "" {
		t.Fatal("ResolvedOptions().NumberingSystem is empty")
	}
}

func TestRelativeTimeFormatUnsupportedLocaleFallsBackToDefaultData(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "ban")}, Options{})
	if err != nil {
		t.Fatalf("New(ban) error = %v", err)
	}
	if got := format.ResolvedOptions().Locale.String(); got == "ban" {
		t.Fatalf("ResolvedOptions().Locale = %q, want supported fallback", got)
	}
	if got, err := format.Format(Int(int64(1)), Day); err != nil {
		t.Fatalf("Format(1, Day) error = %v", err)
	} else if got == "" {
		t.Fatal("Format(1, Day) returned empty fallback output")
	}
}

func TestRelativeTimeFormatFormatValuePastSecond(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.Format(Int(int64(-1)), Second)
	if err != nil {
		t.Fatalf("Format(-1, Second) error = %v", err)
	}
	if got != "1 second ago" {
		t.Fatalf("Format(-1, Second) = %q, want %q", got, "1 second ago")
	}
}

func TestSingularUnitAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input Unit
		want  Unit
		ok    bool
	}{
		{input: Second, want: Second, ok: true},
		{input: Unit("seconds"), want: Second, ok: true},
		{input: Minute, want: Minute, ok: true},
		{input: Unit("minutes"), want: Minute, ok: true},
		{input: Hour, want: Hour, ok: true},
		{input: Unit("hours"), want: Hour, ok: true},
		{input: Day, want: Day, ok: true},
		{input: Unit("days"), want: Day, ok: true},
		{input: Week, want: Week, ok: true},
		{input: Unit("weeks"), want: Week, ok: true},
		{input: Month, want: Month, ok: true},
		{input: Unit("months"), want: Month, ok: true},
		{input: Quarter, want: Quarter, ok: true},
		{input: Unit("quarters"), want: Quarter, ok: true},
		{input: Year, want: Year, ok: true},
		{input: Unit("years"), want: Year, ok: true},
		{input: Unit("monthes")},
		{input: Unit("bad")},
	}
	for _, tc := range tests {
		t.Run(string(tc.input), func(t *testing.T) {
			t.Parallel()

			got, ok := singularUnit(tc.input)
			if ok != tc.ok {
				t.Fatalf("singularUnit(%q) ok = %v, want %v", tc.input, ok, tc.ok)
			}
			if got != tc.want {
				t.Fatalf("singularUnit(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestRelativeTimeFormatFormatFloatValueFutureDay(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.Format(Float(1.5), Day)
	if err != nil {
		t.Fatalf("Format(1.5, Day) error = %v", err)
	}
	if got != "in 1.5 days" {
		t.Fatalf("Format(1.5, Day) = %q, want %q", got, "in 1.5 days")
	}
}

func TestRelativeTimeFormatFormatValuePastMinute(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.Format(mustDecimalValue(t, "-2"), Minute)
	if err != nil {
		t.Fatalf("Format(-2, Minute) error = %v", err)
	}
	if got != "2 minutes ago" {
		t.Fatalf("Format(-2, Minute) = %q, want %q", got, "2 minutes ago")
	}
}

func TestRelativeTimeFormatNumberingSystemOptionLocalizesDigits(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{NumberingSystem: stringPtr("arab")})
	if err != nil {
		t.Fatalf("New(en-US, numberingSystem=arab) error = %v", err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "arab" {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want arab", got)
	}

	got, err := format.Format(Int(int64(2)), Week)
	if err != nil {
		t.Fatalf("Format(2, Week) error = %v", err)
	}
	if got != "in ٢ weeks" {
		t.Fatalf("Format(2, Week) = %q, want Arabic-Indic digit", got)
	}
}

func TestCompileRelativeTimeFieldsFallsBackToLongStyleData(t *testing.T) {
	t.Parallel()

	fields, err := compileRelativeTimeFields(cldrrelativetime.RelativeTimeFields{
		string(Day): {
			string(LongStyle): {
				Future: map[string]string{"one": "in {0} day", "other": "in {0} days"},
				Past:   map[string]string{"one": "{0} day ago", "other": "{0} days ago"},
			},
		},
	}, ShortStyle)
	if err != nil {
		t.Fatalf("compile relative time fields: %v", err)
	}
	field, ok := fields[Day]
	if !ok {
		t.Fatal("compileRelativeTimeFields(short) omitted day field with long fallback data")
	}
	parts := relativeTimePatternParts(field.future.pattern(pluralrules.One), Day, []numberformat.Part{
		{Type: numberformat.PartInteger, Value: "1"},
	})
	if got := relativePartsText(parts); got != "in 1 day" {
		t.Fatalf("compiled long fallback pattern = %q, want in 1 day", got)
	}
}

func TestCompileRelativeTimePatternSetFillsMissingCategoriesFromOther(t *testing.T) {
	t.Parallel()

	patterns, err := compileRelativeTimePatternSet(map[string]string{
		"other": "in {0} days",
	})
	if err != nil {
		t.Fatalf("compile relative time pattern set: %v", err)
	}
	if len(patterns[pluralrules.One]) == 0 {
		t.Fatal("compileRelativeTimePatternSet left missing one category empty; want constructor-time other fallback")
	}
	parts := relativeTimePatternParts(patterns.pattern(pluralrules.One), Day, []numberformat.Part{
		{Type: numberformat.PartInteger, Value: "1"},
	})
	if got := relativePartsText(parts); got != "in 1 days" {
		t.Fatalf("fallback pattern text = %q, want in 1 days", got)
	}
}

func TestRelativeTimePatternText(t *testing.T) {
	t.Parallel()

	pattern, err := compileRelativeTimePattern("in {0} days")
	if err != nil {
		t.Fatalf("compileRelativeTimePattern() error = %v", err)
	}
	if got := relativeTimePatternText(pattern, "1.5"); got != "in 1.5 days" {
		t.Fatalf("relativeTimePatternText() = %q, want %q", got, "in 1.5 days")
	}
}

func TestRelativeTimeFormatNumericAutoLiteral(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Numeric: stringPtr(NumericAuto)})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.Format(Int(int64(-1)), Day)
	if err != nil {
		t.Fatalf("Format(-1, Day) error = %v", err)
	}
	if got != "yesterday" {
		t.Fatalf("Format(-1, Day) = %q, want %q", got, "yesterday")
	}
}

func TestRelativeTimeFormatNumericAutoLiteralToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Numeric: stringPtr(NumericAuto)})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	parts, err := format.FormatToParts(Int(int64(-1)), Day)
	if err != nil {
		t.Fatalf("FormatToParts(-1, Day) error = %v", err)
	}
	if len(parts) != 1 || parts[0].Type != PartLiteral || parts[0].Value != "yesterday" || parts[0].Unit != "" {
		t.Fatalf("FormatToParts(-1, Day) = %#v, want literal yesterday without unit", parts)
	}
}

func TestRelativeTimeFormatNumericAutoDecimalNegativeZeroLiteral(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Numeric: stringPtr(NumericAuto)})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	value := mustDecimalValue(t, "-0")
	got, err := format.Format(value, Day)
	if err != nil {
		t.Fatalf("Format(-0, Day) error = %v", err)
	}
	if got != "today" {
		t.Fatalf("Format(-0, Day) = %q, want %q", got, "today")
	}

	parts, err := format.FormatToParts(value, Day)
	if err != nil {
		t.Fatalf("FormatToParts(-0, Day) error = %v", err)
	}
	if len(parts) != 1 || parts[0].Type != PartLiteral || parts[0].Value != "today" || parts[0].Unit != "" {
		t.Fatalf("FormatToParts(-0, Day) = %#v, want literal today without unit", parts)
	}
}

func TestRelativeTimeFormatPluralUnitAliases(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name  string
		value int
		unit  Unit
		want  string
	}{
		{name: "seconds", value: -2, unit: Second, want: "2 seconds ago"},
		{name: "minutes", value: 2, unit: Minute, want: "in 2 minutes"},
		{name: "hours", value: -2, unit: Hour, want: "2 hours ago"},
		{name: "days", value: 2, unit: Day, want: "in 2 days"},
		{name: "weeks", value: -2, unit: Week, want: "2 weeks ago"},
		{name: "months", value: 2, unit: Month, want: "in 2 months"},
		{name: "quarters", value: 1, unit: Quarter, want: "in 1 quarter"},
		{name: "years", value: -2, unit: Year, want: "2 years ago"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := format.Format(Int(int64(tc.value)), tc.unit)
			if err != nil {
				t.Fatalf("Format(%d, %s) error = %v", tc.value, tc.unit, err)
			}
			if got != tc.want {
				t.Fatalf("Format(%d, %s) = %q, want %q", tc.value, tc.unit, got, tc.want)
			}
		})
	}
}

func TestRelativeTimeFormatFormatValueToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatToParts(Int(int64(-1)), Second)
	if err != nil {
		t.Fatalf("FormatToParts(-1, Second) error = %v", err)
	}
	want := []Part{
		{Type: PartInteger, Value: "1", Unit: Second},
		{Type: PartLiteral, Value: " second ago"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("FormatToParts(-1, Second) = %#v, want %#v", got, want)
	}
}

func TestRelativeTimeFormatFormatValue(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.Format(Uint(uint64(2)), Week)
	if err != nil {
		t.Fatalf("Format(2, Week) error = %v", err)
	}
	if got != "in 2 weeks" {
		t.Fatalf("Format(2, Week) = %q, want %q", got, "in 2 weeks")
	}

	got, err = format.Format(Uint(^uint64(0)), Year)
	if err != nil {
		t.Fatalf("Format(max, Year) error = %v", err)
	}
	if !strings.HasPrefix(got, "in 18,446,744,073,709,551,615 years") {
		t.Fatalf("Format(max, Year) = %q, want exact uint64 magnitude", got)
	}
}

func TestRelativeTimeFormatFormatUnsignedToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	parts, err := format.FormatToParts(Uint(uint64(2)), Week)
	if err != nil {
		t.Fatalf("FormatToParts(2, Week) error = %v", err)
	}
	if got := relativePartsText(parts); got != "in 2 weeks" {
		t.Fatalf("FormatToParts(2, Week) text = %q, want %q", got, "in 2 weeks")
	}
	if len(parts) != 3 || parts[1].Type != PartInteger || parts[1].Unit != Week {
		t.Fatalf("FormatToParts(2, Week) = %#v, want integer part tagged with singular week", parts)
	}

	maxParts, err := format.FormatToParts(Uint(^uint64(0)), Year)
	if err != nil {
		t.Fatalf("FormatToParts(max, Year) error = %v", err)
	}
	if got := relativePartsText(maxParts); !strings.HasPrefix(got, "in 18,446,744,073,709,551,615 years") {
		t.Fatalf("FormatToParts(max, Year) text = %q, want exact uint64 magnitude", got)
	}
}

func TestRelativeTimeFormatFormatFloatValueToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatToParts(Float(1.5), Day)
	if err != nil {
		t.Fatalf("FormatToParts(1.5, Day) error = %v", err)
	}
	want := []Part{
		{Type: PartLiteral, Value: "in "},
		{Type: PartInteger, Value: "1", Unit: Day},
		{Type: PartDecimal, Value: ".", Unit: Day},
		{Type: PartFraction, Value: "5", Unit: Day},
		{Type: PartLiteral, Value: " days"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("FormatToParts(1.5, Day) = %#v, want %#v", got, want)
	}
}

func TestRelativeTimeFormatFormatValueToPartsNegativeZero(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	parts, err := format.FormatToParts(mustDecimalValue(t, "-0"), Day)
	if err != nil {
		t.Fatalf("FormatToParts(-0, Day) error = %v", err)
	}
	if got := relativePartsText(parts); got != "0 days ago" {
		t.Fatalf("FormatToParts(-0, Day) text = %q, want %q", got, "0 days ago")
	}
	if len(parts) < 1 || parts[0].Type != PartInteger || parts[0].Unit != Day {
		t.Fatalf("FormatToParts(-0, Day) = %#v, want integer part tagged with singular day", parts)
	}
}

func TestRelativeTimeFormatFormatEqualsFormatToPartsJoin(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Numeric: stringPtr(NumericAuto)})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name   string
		format func() (string, error)
		parts  func() ([]Part, error)
	}{
		{
			name: "int singular past",
			format: func() (string, error) {
				return format.Format(Int(int64(-1)), Second)
			},
			parts: func() ([]Part, error) {
				return format.FormatToParts(Int(int64(-1)), Second)
			},
		},
		{
			name: "int64 auto literal",
			format: func() (string, error) {
				return format.Format(Int(-1), Day)
			},
			parts: func() ([]Part, error) {
				return format.FormatToParts(Int(-1), Day)
			},
		},
		{
			name: "uint plural future",
			format: func() (string, error) {
				return format.Format(Uint(uint64(2)), Week)
			},
			parts: func() ([]Part, error) {
				return format.FormatToParts(Uint(uint64(2)), Week)
			},
		},
		{
			name: "uint64 max",
			format: func() (string, error) {
				return format.Format(Uint(^uint64(0)), Year)
			},
			parts: func() ([]Part, error) {
				return format.FormatToParts(Uint(^uint64(0)), Year)
			},
		},
		{
			name: "float fraction future",
			format: func() (string, error) {
				return format.Format(Float(1.5), Day)
			},
			parts: func() ([]Part, error) {
				return format.FormatToParts(Float(1.5), Day)
			},
		},
		{
			name: "decimal negative zero",
			format: func() (string, error) {
				return format.Format(mustDecimalValue(t, "-0"), Day)
			},
			parts: func() ([]Part, error) {
				return format.FormatToParts(mustDecimalValue(t, "-0"), Day)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.format()
			if err != nil {
				t.Fatalf("format error = %v", err)
			}
			parts, err := tc.parts()
			if err != nil {
				t.Fatalf("parts error = %v", err)
			}
			if text := relativePartsText(parts); text != got {
				t.Fatalf("joined parts = %q, want format output %q; parts=%#v", text, got, parts)
			}
		})
	}
}

func relativePartsText(parts []Part) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "fr-FR"), intltest.Locale(t, "en-US"), intltest.Locale(t, "zh-Hans-CN")}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: stringPtr(LookupLocaleMatcher)})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	testcontract.AssertLocaleListStrings(t, "SupportedLocalesOf()", got, []string{"fr-FR", "en-US", "zh-Hans-CN"})
}

func TestRelativeTimeFormatErrors(t *testing.T) {
	t.Parallel()

	constructorErrors := []struct {
		name         string
		options      Options
		wantName     string
		wantValue    string
		wantExpected string
	}{
		{
			name:         "style",
			options:      Options{Style: stringPtr("bad")},
			wantName:     "style",
			wantValue:    "bad",
			wantExpected: `one of "long", "short", "narrow"`,
		},
		{
			name:         "style empty",
			options:      Options{Style: stringPtr("")},
			wantName:     "style",
			wantValue:    "",
			wantExpected: `one of "long", "short", "narrow"`,
		},
		{
			name:         "locale matcher",
			options:      Options{LocaleMatcher: stringPtr("bad")},
			wantName:     "localeMatcher",
			wantValue:    "bad",
			wantExpected: `one of "lookup", "best fit"`,
		},
		{
			name:         "locale matcher empty",
			options:      Options{LocaleMatcher: stringPtr("")},
			wantName:     "localeMatcher",
			wantValue:    "",
			wantExpected: `one of "lookup", "best fit"`,
		},
		{
			name:         "numeric",
			options:      Options{Numeric: stringPtr("bad")},
			wantName:     "numeric",
			wantValue:    "bad",
			wantExpected: `one of "always", "auto"`,
		},
		{
			name:         "numeric empty",
			options:      Options{Numeric: stringPtr("")},
			wantName:     "numeric",
			wantValue:    "",
			wantExpected: `one of "always", "auto"`,
		},
		{
			name:         "numbering system",
			options:      Options{NumberingSystem: stringPtr("bad!")},
			wantName:     "numberingSystem",
			wantValue:    "bad!",
			wantExpected: "a Unicode locale extension type",
		},
		{
			name:         "numbering system empty",
			options:      Options{NumberingSystem: stringPtr("")},
			wantName:     "numberingSystem",
			wantValue:    "",
			wantExpected: "a Unicode locale extension type",
		},
	}
	for _, tc := range constructorErrors {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(locale.List{intltest.Locale(t, "en-US")}, tc.options)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s) error = %v, want intlerr.ErrInvalidOption", tc.name, err)
			}
			testcontract.AssertOptionError(t, err, "relativetimeformat", intlerr.InvalidOption, tc.wantName, tc.wantValue, "en-US")
			testcontract.AssertOptionExpected(t, err, tc.wantExpected)
		})
	}

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}
	if _, err := format.Format(Int(int64(1)), Unit("bad")); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("Format(invalid unit) error = %v, want intlerr.ErrInvalidValue", err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "relativetimeformat", "unit", "bad", format.ResolvedOptions().Locale.String())
		testcontract.AssertErrorExpected(t, err, `one of "second", "minute", "hour", "day", "week", "month", "quarter", "year", or their plural forms`)
	}
	if _, err := format.FormatToParts(Int(math.MinInt64), Day); err != nil {
		t.Fatalf("FormatToParts(MinInt64, Day) error = %v", err)
	}
	if got, err := format.Format(Int(math.MinInt64), Day); err != nil {
		t.Fatalf("Format(MinInt64, Day) error = %v", err)
	} else if !strings.HasPrefix(got, "9,223,372,036,854,775,808 days ago") {
		t.Fatalf("Format(MinInt64, Day) = %q, want exact absolute magnitude", got)
	}
	if _, err := format.Format(Uint(1), Unit("bad")); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("Format(invalid unit) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := format.FormatToParts(Uint(1), Unit("bad")); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("FormatToParts(invalid unit) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := format.FormatToParts(Float(math.Inf(1)), Day); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("FormatToParts(+Inf) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "relativetimeformat", "value", "+Inf", format.ResolvedOptions().Locale.String())
		testcontract.AssertErrorExpected(t, err, "a finite numeric value")
	}
	if _, err := format.Format(Float(math.NaN()), Day); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("Format(NaN) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "relativetimeformat", "value", "NaN", format.ResolvedOptions().Locale.String())
		testcontract.AssertErrorExpected(t, err, "a finite numeric value")
	}
	if _, err := Decimal("not-a-number"); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("Decimal(invalid) error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "relativetimeformat", "value", "not-a-number", "")
		testcontract.AssertErrorExpected(t, err, "a finite numeric value")
	}
	for _, matcher := range []string{"bad", ""} {
		t.Run(matcher, func(t *testing.T) {
			t.Parallel()
			_, err := SupportedLocalesOf(nil, Options{LocaleMatcher: stringPtr(matcher)})
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "relativetimeformat", intlerr.InvalidOption, "localeMatcher", matcher, "en")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}
}

func mustDecimalValue(t *testing.T, value string) Value {
	t.Helper()

	v, err := Decimal(value)
	if err != nil {
		t.Fatalf("Decimal(%q) error = %v", value, err)
	}
	return v
}
