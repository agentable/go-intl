package relativetimeformat

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"

	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/locale"
)

func TestRelativeTimeFormatResolvedOptionsDefaults(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
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

	format, err := New(locale.List{locale.MustParse("ban")}, Options{})
	if err != nil {
		t.Fatalf("New(ban) error = %v", err)
	}
	if got := format.ResolvedOptions().Locale.String(); got == "ban" {
		t.Fatalf("ResolvedOptions().Locale = %q, want supported fallback", got)
	}
	if got, err := format.FormatInt(1, Day); err != nil {
		t.Fatalf("FormatInt(1, Day) error = %v", err)
	} else if got == "" {
		t.Fatal("FormatInt(1, Day) returned empty fallback output")
	}
}

func TestRelativeTimeFormatFormatIntPastSecond(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatInt(-1, Second)
	if err != nil {
		t.Fatalf("FormatInt(-1, Second) error = %v", err)
	}
	if got != "1 second ago" {
		t.Fatalf("FormatInt(-1, Second) = %q, want %q", got, "1 second ago")
	}
}

func TestRelativeTimeFormatFormatFloat64FutureDay(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatFloat64(1.5, Days)
	if err != nil {
		t.Fatalf("FormatFloat64(1.5, Days) error = %v", err)
	}
	if got != "in 1.5 days" {
		t.Fatalf("FormatFloat64(1.5, Days) = %q, want %q", got, "in 1.5 days")
	}
}

func TestRelativeTimeFormatFormatDecimalPastMinute(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatDecimal("-2", Minutes)
	if err != nil {
		t.Fatalf("FormatDecimal(-2, Minutes) error = %v", err)
	}
	if got != "2 minutes ago" {
		t.Fatalf("FormatDecimal(-2, Minutes) = %q, want %q", got, "2 minutes ago")
	}
}

func TestRelativeTimeFormatNumberingSystemOptionLocalizesDigits(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{NumberingSystem: "arab"})
	if err != nil {
		t.Fatalf("New(en-US, numberingSystem=arab) error = %v", err)
	}
	if got := format.ResolvedOptions().NumberingSystem; got != "arab" {
		t.Fatalf("ResolvedOptions().NumberingSystem = %q, want arab", got)
	}

	got, err := format.FormatInt(2, Weeks)
	if err != nil {
		t.Fatalf("FormatInt(2, Weeks) error = %v", err)
	}
	if got != "in ٢ weeks" {
		t.Fatalf("FormatInt(2, Weeks) = %q, want Arabic-Indic digit", got)
	}
}

func TestRelativeTimeFormatFallsBackToLongStyleData(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Style: ShortStyle})
	if err != nil {
		t.Fatalf("New(en-US, short) error = %v", err)
	}
	format.fields = map[string]map[string]cldrrelativetime.RelativeTimeField{
		string(Day): {
			string(LongStyle): {
				Future: map[string]string{"one": "in {0} day", "other": "in {0} days"},
				Past:   map[string]string{"one": "{0} day ago", "other": "{0} days ago"},
			},
		},
	}

	got, err := format.FormatInt(1, Day)
	if err != nil {
		t.Fatalf("FormatInt(1, Day) with long fallback data error = %v", err)
	}
	if got != "in 1 day" {
		t.Fatalf("FormatInt(1, Day) with long fallback data = %q, want in 1 day", got)
	}
}

func TestRelativeTimeFormatNumericAutoLiteral(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Numeric: NumericAuto})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatInt(-1, Day)
	if err != nil {
		t.Fatalf("FormatInt(-1, Day) error = %v", err)
	}
	if got != "yesterday" {
		t.Fatalf("FormatInt(-1, Day) = %q, want %q", got, "yesterday")
	}
}

func TestRelativeTimeFormatNumericAutoLiteralToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Numeric: NumericAuto})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	parts, err := format.FormatIntToParts(-1, Day)
	if err != nil {
		t.Fatalf("FormatIntToParts(-1, Day) error = %v", err)
	}
	if len(parts) != 1 || parts[0].Type != PartLiteral || parts[0].Value != "yesterday" || parts[0].Unit != "" {
		t.Fatalf("FormatIntToParts(-1, Day) = %#v, want literal yesterday without unit", parts)
	}
}

func TestRelativeTimeFormatNumericAutoDecimalNegativeZeroLiteral(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{Numeric: NumericAuto})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatDecimal("-0", Day)
	if err != nil {
		t.Fatalf("FormatDecimal(-0, Day) error = %v", err)
	}
	if got != "today" {
		t.Fatalf("FormatDecimal(-0, Day) = %q, want %q", got, "today")
	}

	parts, err := format.FormatDecimalToParts("-0", Day)
	if err != nil {
		t.Fatalf("FormatDecimalToParts(-0, Day) error = %v", err)
	}
	if len(parts) != 1 || parts[0].Type != PartLiteral || parts[0].Value != "today" || parts[0].Unit != "" {
		t.Fatalf("FormatDecimalToParts(-0, Day) = %#v, want literal today without unit", parts)
	}
}

func TestRelativeTimeFormatPluralUnitAliases(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name  string
		value int
		unit  Unit
		want  string
	}{
		{name: "seconds", value: -2, unit: Seconds, want: "2 seconds ago"},
		{name: "minutes", value: 2, unit: Minutes, want: "in 2 minutes"},
		{name: "hours", value: -2, unit: Hours, want: "2 hours ago"},
		{name: "days", value: 2, unit: Days, want: "in 2 days"},
		{name: "weeks", value: -2, unit: Weeks, want: "2 weeks ago"},
		{name: "months", value: 2, unit: Months, want: "in 2 months"},
		{name: "quarters", value: 1, unit: Quarters, want: "in 1 quarter"},
		{name: "years", value: -2, unit: Years, want: "2 years ago"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := format.FormatInt(tc.value, tc.unit)
			if err != nil {
				t.Fatalf("FormatInt(%d, %s) error = %v", tc.value, tc.unit, err)
			}
			if got != tc.want {
				t.Fatalf("FormatInt(%d, %s) = %q, want %q", tc.value, tc.unit, got, tc.want)
			}
		})
	}
}

func TestRelativeTimeFormatFormatIntToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatIntToParts(-1, Second)
	if err != nil {
		t.Fatalf("FormatIntToParts(-1, Second) error = %v", err)
	}
	want := []Part{
		{Type: PartInteger, Value: "1", Unit: Second},
		{Type: PartLiteral, Value: " second ago"},
	}
	intltest.DiffParts(t, got, want)
}

func TestRelativeTimeFormatFormatUint(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatUint(2, Week)
	if err != nil {
		t.Fatalf("FormatUint(2, Week) error = %v", err)
	}
	if got != "in 2 weeks" {
		t.Fatalf("FormatUint(2, Week) = %q, want %q", got, "in 2 weeks")
	}

	got, err = format.FormatUint64(^uint64(0), Year)
	if err != nil {
		t.Fatalf("FormatUint64(max, Year) error = %v", err)
	}
	if !strings.HasPrefix(got, "in 18,446,744,073,709,551,615 years") {
		t.Fatalf("FormatUint64(max, Year) = %q, want exact uint64 magnitude", got)
	}
}

func TestRelativeTimeFormatFormatUnsignedToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	parts, err := format.FormatUintToParts(2, Weeks)
	if err != nil {
		t.Fatalf("FormatUintToParts(2, Weeks) error = %v", err)
	}
	if got := relativePartsText(parts); got != "in 2 weeks" {
		t.Fatalf("FormatUintToParts(2, Weeks) text = %q, want %q", got, "in 2 weeks")
	}
	if len(parts) != 3 || parts[1].Type != PartInteger || parts[1].Unit != Week {
		t.Fatalf("FormatUintToParts(2, Weeks) = %#v, want integer part tagged with singular week", parts)
	}

	maxParts, err := format.FormatUint64ToParts(^uint64(0), Years)
	if err != nil {
		t.Fatalf("FormatUint64ToParts(max, Years) error = %v", err)
	}
	if got := relativePartsText(maxParts); !strings.HasPrefix(got, "in 18,446,744,073,709,551,615 years") {
		t.Fatalf("FormatUint64ToParts(max, Years) text = %q, want exact uint64 magnitude", got)
	}
}

func TestRelativeTimeFormatFormatFloat64ToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	got, err := format.FormatFloat64ToParts(1.5, Days)
	if err != nil {
		t.Fatalf("FormatFloat64ToParts(1.5, Days) error = %v", err)
	}
	want := []Part{
		{Type: PartLiteral, Value: "in "},
		{Type: PartInteger, Value: "1", Unit: Day},
		{Type: PartDecimal, Value: ".", Unit: Day},
		{Type: PartFraction, Value: "5", Unit: Day},
		{Type: PartLiteral, Value: " days"},
	}
	intltest.DiffParts(t, got, want)
}

func TestRelativeTimeFormatFormatDecimalToPartsNegativeZero(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	parts, err := format.FormatDecimalToParts("-0", Days)
	if err != nil {
		t.Fatalf("FormatDecimalToParts(-0, Days) error = %v", err)
	}
	if got := relativePartsText(parts); got != "0 days ago" {
		t.Fatalf("FormatDecimalToParts(-0, Days) text = %q, want %q", got, "0 days ago")
	}
	if len(parts) < 1 || parts[0].Type != PartInteger || parts[0].Unit != Day {
		t.Fatalf("FormatDecimalToParts(-0, Days) = %#v, want integer part tagged with singular day", parts)
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

	requested := locale.List{
		locale.MustParse("fr-FR"),
		locale.MustParse("en-US"),
		locale.MustParse("zh-Hans-CN"),
	}
	got, err := SupportedLocalesOf(requested, Options{LocaleMatcher: LookupLocaleMatcher})
	if err != nil {
		t.Fatalf("SupportedLocalesOf() error = %v", err)
	}
	want := []string{"fr-FR", "en-US", "zh-Hans-CN"}
	if len(got) != len(want) {
		t.Fatalf("SupportedLocalesOf() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i].String() != want[i] {
			t.Fatalf("SupportedLocalesOf()[%d] = %q, want %q", i, got[i].String(), want[i])
		}
	}
}

func TestRelativeTimeFormatErrors(t *testing.T) {
	t.Parallel()

	if _, err := New(locale.List{locale.MustParse("en-US")}, Options{Style: Style("bad")}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	}
	if _, err := New(locale.List{locale.MustParse("en-US")}, Options{LocaleMatcher: LocaleMatcher("bad")}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(invalid localeMatcher) error = %v, want intlerr.ErrInvalidOption", err)
	}
	if _, err := New(locale.List{locale.MustParse("en-US")}, Options{Numeric: Numeric("bad")}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(invalid numeric) error = %v, want intlerr.ErrInvalidOption", err)
	}
	if _, err := New(locale.List{locale.MustParse("en-US")}, Options{NumberingSystem: "bad!"}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(invalid numberingSystem) error = %v, want intlerr.ErrInvalidOption", err)
	}

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}
	if _, err := format.FormatInt(1, Unit("bad")); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("FormatInt(invalid unit) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := format.FormatInt64ToParts(math.MinInt64, Day); err != nil {
		t.Fatalf("FormatInt64ToParts(MinInt64, Day) error = %v", err)
	}
	if got, err := format.FormatInt64(math.MinInt64, Day); err != nil {
		t.Fatalf("FormatInt64(MinInt64, Day) error = %v", err)
	} else if !strings.HasPrefix(got, "9,223,372,036,854,775,808 days ago") {
		t.Fatalf("FormatInt64(MinInt64, Day) = %q, want exact absolute magnitude", got)
	}
	if _, err := format.FormatUint64(1, Unit("bad")); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("FormatUint64(invalid unit) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := format.FormatUint64ToParts(1, Unit("bad")); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("FormatUint64ToParts(invalid unit) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := format.FormatFloat64ToParts(math.Inf(1), Day); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("FormatFloat64ToParts(+Inf) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := format.FormatFloat64(math.NaN(), Day); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("FormatFloat64(NaN) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := format.FormatDecimal("not-a-number", Day); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("FormatDecimal(invalid) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := format.FormatDecimalToParts("not-a-number", Day); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("FormatDecimalToParts(invalid) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := SupportedLocalesOf(nil, Options{LocaleMatcher: LocaleMatcher("bad")}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("SupportedLocalesOf(invalid matcher) error = %v, want intlerr.ErrInvalidOption", err)
	}
}
