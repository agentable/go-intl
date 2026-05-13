package numberformat

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/locale"
)

func TestNumberFormatFormatDecimal(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1234.5); got != "1,234.5" {
		t.Fatalf("Format(1234.5) = %q, want 1,234.5", got)
	}
}

func TestNumberFormatFractionDigitOptions(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{FractionDigits: FractionDigits(2, 2)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1.234); got != "1.23" {
		t.Fatalf("Format(1.234) = %q, want 1.23", got)
	}
	if got := format.formatValue(1); got != "1.00" {
		t.Fatalf("Format(1) = %q, want 1.00", got)
	}
}

func TestNumberFormatUseGroupingFalse(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{UseGrouping: UseGroupingFalse})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1234.5); got != "1234.5" {
		t.Fatalf("Format(1234.5) = %q, want 1234.5", got)
	}
}

func TestNumberFormatUseGroupingMin2(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{UseGrouping: UseGroupingMin2})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		in   any
		want string
	}{
		{in: 9999, want: "9999"},
		{in: 10000, want: "10,000"},
		{in: 1000000, want: "1,000,000"},
	}
	for _, tc := range tests {
		if got := format.formatValue(tc.in); got != tc.want {
			t.Fatalf("Format(%v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNumberFormatCLDRPrimarySecondaryGrouping(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("hi"))
	if err != nil {
		t.Fatal(err)
	}
	if got := format.FormatInt64(123456789); got != "12,34,56,789" {
		t.Fatalf("FormatInt64(123456789) = %q, want Indian grouping", got)
	}
	parts, err := format.FormatDecimalToParts("123456.456")
	if err != nil {
		t.Fatal(err)
	}
	wantParts := []Part{
		{Type: PartInteger, Value: "1"},
		{Type: PartGroup, Value: ","},
		{Type: PartInteger, Value: "23"},
		{Type: PartGroup, Value: ","},
		{Type: PartInteger, Value: "456"},
		{Type: PartDecimal, Value: "."},
		{Type: PartFraction, Value: "456"},
	}
	if !reflect.DeepEqual(parts, wantParts) {
		t.Fatalf("FormatDecimalToParts(123456.456) = %#v, want %#v", parts, wantParts)
	}

	currency, err := New(locale.MustParse("hi"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD")})
	if err != nil {
		t.Fatal(err)
	}
	if got := currency.FormatInt64(123456); got != "$1,23,456.00" {
		t.Fatalf("Currency FormatInt64(123456) = %q, want Indian currency grouping", got)
	}
}

func TestNumberFormatFormatToPartsDecimal(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	want := []Part{
		{Type: PartInteger, Value: "1"},
		{Type: PartGroup, Value: ","},
		{Type: PartInteger, Value: "234"},
		{Type: PartDecimal, Value: "."},
		{Type: PartFraction, Value: "5"},
	}
	if got := format.formatToPartsValue(1234.5); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(1234.5) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatToPartsNegative(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	want := []Part{
		{Type: PartMinusSign, Value: "-"},
		{Type: PartInteger, Value: "5"},
	}
	if got := format.formatToPartsValue(-5); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(-5) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatInvalidInputs(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []any{nil, "not a number", struct{}{}} {
		if got := format.formatValue(in); got != "NaN" {
			t.Fatalf("Format(%#v) = %q, want NaN", in, got)
		}
		want := []Part{{Type: PartNaN, Value: "NaN"}}
		if got := format.formatToPartsValue(in); !reflect.DeepEqual(got, want) {
			t.Fatalf("FormatToParts(%#v) = %#v, want %#v", in, got, want)
		}
	}
}

func TestNumberFormatFormatDecimalInvalidValue(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := format.FormatDecimal("not a number"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("FormatDecimal() error = %v, want ErrInvalidValue", err)
	}
	if _, err := format.FormatDecimalToParts("not a number"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("FormatDecimalToParts() error = %v, want ErrInvalidValue", err)
	}
	if _, err := format.FormatRangeDecimal("1", "not a number"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("FormatRangeDecimal() error = %v, want ErrInvalidValue", err)
	}
}

func TestNumberFormatSignDisplayAlways(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{SignDisplay: AlwaysSignDisplay})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(5); got != "+5" {
		t.Fatalf("Format(5) = %q, want +5", got)
	}
	want := []Part{{Type: PartPlusSign, Value: "+"}, {Type: PartInteger, Value: "5"}}
	if got := format.formatToPartsValue(5); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(5) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatSignDisplayModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode string
		in   any
		want string
	}{
		{name: "always zero", mode: "always", in: 0, want: "+0"},
		{name: "always negative zero", mode: "always", in: math.Copysign(0, -1), want: "-0"},
		{name: "except zero positive", mode: "exceptZero", in: 5, want: "+5"},
		{name: "except zero zero", mode: "exceptZero", in: 0, want: "0"},
		{name: "except zero negative zero", mode: "exceptZero", in: math.Copysign(0, -1), want: "0"},
		{name: "negative positive", mode: "negative", in: 5, want: "5"},
		{name: "negative negative", mode: "negative", in: -5, want: "-5"},
		{name: "negative negative zero", mode: "negative", in: math.Copysign(0, -1), want: "0"},
		{name: "never negative", mode: "never", in: -5, want: "5"},
		{name: "never negative zero", mode: "never", in: math.Copysign(0, -1), want: "0"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.MustParse("en"), Options{SignDisplay: SignDisplay(tc.mode)})
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(tc.in); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNumberFormatFormatInfinity(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(math.Inf(1)); got != "∞" {
		t.Fatalf("Format(+Inf) = %q, want ∞", got)
	}
	if got := format.formatValue(math.Inf(-1)); got != "-∞" {
		t.Fatalf("Format(-Inf) = %q, want -∞", got)
	}
	want := []Part{{Type: PartMinusSign, Value: "-"}, {Type: PartInfinity, Value: "∞"}}
	if got := format.formatToPartsValue(math.Inf(-1)); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(-Inf) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatSignDisplaySpecialValues(t *testing.T) {
	t.Parallel()

	always, err := New(locale.MustParse("en"), Options{SignDisplay: AlwaysSignDisplay})
	if err != nil {
		t.Fatal(err)
	}
	if got := always.formatValue(math.NaN()); got != "+NaN" {
		t.Fatalf("always Format(NaN) = %q, want +NaN", got)
	}
	if got := always.formatValue(math.Inf(1)); got != "+∞" {
		t.Fatalf("always Format(+Inf) = %q, want +∞", got)
	}

	never, err := New(locale.MustParse("en"), Options{SignDisplay: NeverSignDisplay})
	if err != nil {
		t.Fatal(err)
	}
	if got := never.formatValue(math.Inf(-1)); got != "∞" {
		t.Fatalf("never Format(-Inf) = %q, want ∞", got)
	}
}

func TestNumberFormatSignDisplayNotation(t *testing.T) {
	t.Parallel()

	scientific, err := New(locale.MustParse("en"), Options{Notation: ScientificNotation, SignDisplay: AlwaysSignDisplay, FractionDigits: MaximumFractionDigits(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got := scientific.formatValue(1200); got != "+1.2E3" {
		t.Fatalf("scientific Format(1200) = %q, want +1.2E3", got)
	}

	compact, err := New(locale.MustParse("en"), Options{Notation: CompactNotation, SignDisplay: NeverSignDisplay, FractionDigits: MaximumFractionDigits(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got := compact.formatValue(-1500); got != "1.5K" {
		t.Fatalf("compact Format(-1500) = %q, want 1.5K", got)
	}
}

func TestNumberFormatFormatPercent(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Style: PercentStyle})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(0.123); got != "12%" {
		t.Fatalf("Format(0.123) = %q, want 12%%", got)
	}
	want := []Part{{Type: PartInteger, Value: "12"}, {Type: PartPercentSign, Value: "%"}}
	if got := format.formatToPartsValue(0.123); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(0.123) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatPercentPreservesDecimalMagnitude(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Style: PercentStyle, UseGrouping: UseGroupingFalse})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := format.formatValue("9007199254740993"), "900719925474099300%"; got != want {
		t.Fatalf("Format(large percent) = %q, want %q", got, want)
	}
}

func TestNumberFormatScientificPreservesDecimalMagnitude(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Notation: ScientificNotation, SignificantDigits: MaximumSignificantDigits(16), UseGrouping: UseGroupingFalse})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := format.formatValue("9007199254740993"), "9.007199254740993E15"; got != want {
		t.Fatalf("Format(large scientific) = %q, want %q", got, want)
	}
}

func TestNumberFormatFormatCurrencyUSD(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD")})
	if err != nil {
		t.Fatal(err)
	}
	resolved := format.ResolvedOptions()
	if resolved.MinimumFractionDigits != 2 || resolved.MaximumFractionDigits != 2 {
		t.Fatalf("currency fraction digits = %d/%d, want 2/2", resolved.MinimumFractionDigits, resolved.MaximumFractionDigits)
	}
	if got := format.formatValue(1234.5); got != "$1,234.50" {
		t.Fatalf("Format(1234.5) = %q, want $1,234.50", got)
	}
	want := []Part{{Type: PartCurrency, Value: "$"}, {Type: PartInteger, Value: "1"}, {Type: PartGroup, Value: ","}, {Type: PartInteger, Value: "234"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "50"}}
	if got := format.formatToPartsValue(1234.5); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(1234.5) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatCurrencyJPY(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("JPY")})
	if err != nil {
		t.Fatal(err)
	}
	resolved := format.ResolvedOptions()
	if resolved.MinimumFractionDigits != 0 || resolved.MaximumFractionDigits != 0 {
		t.Fatalf("currency fraction digits = %d/%d, want 0/0", resolved.MinimumFractionDigits, resolved.MaximumFractionDigits)
	}
	if got := format.formatValue(1234.5); got != "¥1,235" {
		t.Fatalf("Format(1234.5) = %q, want ¥1,235", got)
	}
}

func TestNumberFormatFormatCurrencyCode(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencyDisplay: CurrencyDisplayCode})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(12); got != "USD12.00" {
		t.Fatalf("Format(12) = %q, want USD12.00", got)
	}
	want := []Part{{Type: PartCurrency, Value: "USD"}, {Type: PartInteger, Value: "12"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}}
	if got := format.formatToPartsValue(12); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(12) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatCurrencyPatternPlacement(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencyDisplay: CurrencyDisplayCode})
	if err != nil {
		t.Fatal(err)
	}
	pattern := format.cldrLoc.CurrencyPattern(format.ResolvedOptions().NumberingSystem, "standard")
	if got := joinParts(format.formatToPartsValue(12)); got != strings.Replace(pattern, "¤#,##0.00", "USD12.00", 1) {
		t.Fatalf("FormatToParts(12) joined = %q, want CLDR currency pattern %q", got, pattern)
	}
}

func TestNumberFormatCurrencySignPlacement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		in   any
		want string
	}{
		{name: "standard negative", opts: Options{Style: CurrencyStyle, Currency: CurrencyCode("USD")}, in: -12, want: "-$12.00"},
		{name: "standard positive sign", opts: Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), SignDisplay: AlwaysSignDisplay}, in: 12, want: "+$12.00"},
		{name: "accounting negative", opts: Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencySign: AccountingCurrencySign}, in: -12, want: "($12.00)"},
		{name: "accounting hidden sign", opts: Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencySign: AccountingCurrencySign, SignDisplay: NeverSignDisplay}, in: -12, want: "$12.00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.MustParse("en-US"), tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(tc.in); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNumberFormatCurrencyAccountingParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencySign: AccountingCurrencySign})
	if err != nil {
		t.Fatal(err)
	}
	want := []Part{
		{Type: PartLiteral, Value: "("},
		{Type: PartCurrency, Value: "$"},
		{Type: PartInteger, Value: "12"},
		{Type: PartDecimal, Value: "."},
		{Type: PartFraction, Value: "00"},
		{Type: PartLiteral, Value: ")"},
	}
	if got := format.formatToPartsValue(-12); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(-12) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatCurrencyName(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencyDisplay: CurrencyDisplayName})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(12); got != "12.00 US dollars" {
		t.Fatalf("Format(12) = %q, want 12.00 US dollars", got)
	}
	want := []Part{{Type: PartInteger, Value: "12"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}, {Type: PartLiteral, Value: " "}, {Type: PartCurrency, Value: "US dollars"}}
	if got := format.formatToPartsValue(12); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(12) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatCurrencyNamePlural(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencyDisplay: CurrencyDisplayName})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1); got != "1.00 US dollar" {
		t.Fatalf("Format(1) = %q, want 1.00 US dollar", got)
	}
	if got := format.formatValue(2); got != "2.00 US dollars" {
		t.Fatalf("Format(2) = %q, want 2.00 US dollars", got)
	}
}

func TestNumberFormatCurrencyDisplayNameUsesRoundedPlural(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{
		Style:           CurrencyStyle,
		Currency:        CurrencyCode("USD"),
		CurrencyDisplay: CurrencyDisplayName,
		FractionDigits:  FractionDigits(0, 0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := format.formatValue(1.2), "1 US dollar"; got != want {
		t.Fatalf("Format(1.2) = %q, want %q", got, want)
	}
}

func TestNumberFormatFormatCurrencyNarrowSymbol(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencyDisplay: CurrencyDisplayNarrowSymbol})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(12); got != "$12.00" {
		t.Fatalf("Format(12) = %q, want $12.00", got)
	}
	want := []Part{{Type: PartCurrency, Value: "$"}, {Type: PartInteger, Value: "12"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "00"}}
	if got := format.formatToPartsValue(12); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(12) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatUnitMeter(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Style: UnitStyle, Unit: UnitIdentifier("meter")})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1); got != "1 m" {
		t.Fatalf("Format(1) = %q, want 1 m", got)
	}
	want := []Part{{Type: PartInteger, Value: "1"}, {Type: PartLiteral, Value: " "}, {Type: PartUnit, Value: "m"}}
	if got := format.formatToPartsValue(1); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(1) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatScientific(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Notation: ScientificNotation, FractionDigits: MaximumFractionDigits(2)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(12345); got != "1.23E4" {
		t.Fatalf("Format(12345) = %q, want 1.23E4", got)
	}
	want := []Part{{Type: PartInteger, Value: "1"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "23"}, {Type: PartExponentSeparator, Value: "E"}, {Type: PartExponentInteger, Value: "4"}}
	if got := format.formatToPartsValue(12345); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(12345) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatEngineering(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Notation: EngineeringNotation, FractionDigits: MaximumFractionDigits(2)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(12345); got != "12.35E3" {
		t.Fatalf("Format(12345) = %q, want 12.35E3", got)
	}
}

func TestNumberFormatFormatScientificNegativeExponent(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Notation: ScientificNotation, FractionDigits: MaximumFractionDigits(2)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(0.0123); got != "1.23E-2" {
		t.Fatalf("Format(0.0123) = %q, want 1.23E-2", got)
	}
	want := []Part{{Type: PartInteger, Value: "1"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "23"}, {Type: PartExponentSeparator, Value: "E"}, {Type: PartExponentMinusSign, Value: "-"}, {Type: PartExponentInteger, Value: "2"}}
	if got := format.formatToPartsValue(0.0123); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(0.0123) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatCompact(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Notation: CompactNotation, FractionDigits: MaximumFractionDigits(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1500); got != "1.5K" {
		t.Fatalf("FormatInt(1500) = %q, want 1.5K", got)
	}
	want := []Part{{Type: PartInteger, Value: "1"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "5"}, {Type: PartCompact, Value: "K"}}
	if got := format.formatToPartsValue(1500); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(1500) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatCompactLong(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{Notation: CompactNotation, CompactDisplay: LongCompactDisplay, FractionDigits: MaximumFractionDigits(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1500); got != "1.5 thousand" {
		t.Fatalf("FormatInt(1500) = %q, want 1.5 thousand", got)
	}
	want := []Part{{Type: PartInteger, Value: "1"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "5"}, {Type: PartLiteral, Value: " "}, {Type: PartCompact, Value: "thousand"}}
	if got := format.formatToPartsValue(1500); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(1500) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatRoundingOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		in   any
		want string
	}{
		{name: "half expand", opts: Options{FractionDigits: MaximumFractionDigits(0)}, in: 1.5, want: "2"},
		{name: "half trunc", opts: Options{FractionDigits: MaximumFractionDigits(0), RoundingMode: HalfTruncRoundingMode}, in: 1.5, want: "1"},
		{name: "half even lower", opts: Options{FractionDigits: MaximumFractionDigits(0), RoundingMode: HalfEvenRoundingMode}, in: 2.5, want: "2"},
		{name: "half even upper", opts: Options{FractionDigits: MaximumFractionDigits(0), RoundingMode: HalfEvenRoundingMode}, in: 3.5, want: "4"},
		{name: "ceil negative", opts: Options{FractionDigits: MaximumFractionDigits(0), RoundingMode: CeilRoundingMode}, in: -1.9, want: "-1"},
		{name: "floor negative", opts: Options{FractionDigits: MaximumFractionDigits(0), RoundingMode: FloorRoundingMode}, in: -1.1, want: "-2"},
		{name: "rounding increment", opts: Options{FractionDigits: FractionDigits(2, 2), RoundingIncrement: 5}, in: 1.23, want: "1.25"},
		{name: "strip integer zeros", opts: Options{FractionDigits: FractionDigits(2, 2), TrailingZeroDisplay: StripIfIntegerTrailingZeroDisplay}, in: 1, want: "1"},
		{name: "minimum integer digits", opts: Options{MinimumIntegerDigits: 3, FractionDigits: MaximumFractionDigits(0)}, in: 5, want: "005"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.MustParse("en"), tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(tc.in); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNumberFormatSignificantDigitOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		in   any
		want string
	}{
		{name: "minimum significant zero", opts: Options{SignificantDigits: MinimumSignificantDigits(3)}, in: 0, want: "0.00"},
		{name: "minimum significant decimal", opts: Options{SignificantDigits: MinimumSignificantDigits(3)}, in: 1.2, want: "1.20"},
		{name: "maximum significant rounds integer", opts: Options{SignificantDigits: MaximumSignificantDigits(3)}, in: 1234.5, want: "1,230"},
		{name: "minimum and maximum significant small decimal", opts: Options{SignificantDigits: SignificantDigits(2, 4)}, in: 0.0012345, want: "0.001235"},
		{name: "minimum integer digits", opts: Options{MinimumIntegerDigits: 3, SignificantDigits: MinimumSignificantDigits(3)}, in: 1.2, want: "001.20"},
		{name: "strip integer zeros", opts: Options{SignificantDigits: MinimumSignificantDigits(3), TrailingZeroDisplay: StripIfIntegerTrailingZeroDisplay}, in: 1, want: "1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.MustParse("en"), tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(tc.in); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNumberFormatRoundingPriority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts Options
		in   any
		want string
	}{
		{name: "more precision chooses significant", opts: Options{FractionDigits: FractionDigits(2, 2), SignificantDigits: SignificantDigits(2, 4), RoundingPriority: MorePrecisionRoundingPriority}, in: 1.2345, want: "1.235"},
		{name: "less precision chooses fraction", opts: Options{FractionDigits: FractionDigits(2, 2), SignificantDigits: SignificantDigits(2, 4), RoundingPriority: LessPrecisionRoundingPriority}, in: 1.2345, want: "1.23"},
		{name: "more precision keeps integer fraction", opts: Options{FractionDigits: FractionDigits(2, 2), SignificantDigits: SignificantDigits(2, 4), RoundingPriority: MorePrecisionRoundingPriority}, in: 12345, want: "12,345.00"},
		{name: "less precision rounds integer significant", opts: Options{FractionDigits: FractionDigits(2, 2), SignificantDigits: SignificantDigits(2, 4), RoundingPriority: LessPrecisionRoundingPriority}, in: 12345, want: "12,350"},
		{name: "more precision default significant", opts: Options{FractionDigits: MaximumFractionDigits(2), RoundingPriority: MorePrecisionRoundingPriority}, in: 1.2345, want: "1.2345"},
		{name: "less precision default significant", opts: Options{FractionDigits: MaximumFractionDigits(2), RoundingPriority: LessPrecisionRoundingPriority}, in: 1.2345, want: "1.23"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.MustParse("en"), tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(tc.in); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNumberFormatRoundingIncrementRequiresFixedFractionDigits(t *testing.T) {
	t.Parallel()

	if _, err := New(locale.MustParse("en"), Options{RoundingIncrement: 5}); !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
	if _, err := New(locale.MustParse("en"), Options{FractionDigits: FractionDigits(2, 2), RoundingIncrement: 5}); err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if _, err := New(locale.MustParse("en"), Options{FractionDigits: FractionDigits(2, 2), RoundingIncrement: 5, RoundingPriority: MorePrecisionRoundingPriority}); !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New() error = %v, want ErrInvalidOption", err)
	}
}

func TestPartitionDecimalUsesNumberSymbols(t *testing.T) {
	t.Parallel()

	symbols := cldr.NumberSymbols{Decimal: "·", Group: "_", Minus: "−"}
	want := []Part{
		{Type: PartMinusSign, Value: "−"},
		{Type: PartInteger, Value: "1"},
		{Type: PartGroup, Value: "_"},
		{Type: PartInteger, Value: "234"},
		{Type: PartDecimal, Value: "·"},
		{Type: PartFraction, Value: "5"},
	}
	if got := partitionDecimal("-1,234.5", symbols); !reflect.DeepEqual(got, want) {
		t.Fatalf("partitionDecimal() = %#v, want %#v", got, want)
	}
}

func TestLocalizeNumberStringUsesNumberSymbols(t *testing.T) {
	t.Parallel()

	symbols := cldr.NumberSymbols{Decimal: "·", Group: "_", Minus: "−"}
	if got := localizeNumberString("-1,234.5", symbols); got != "−1_234·5" {
		t.Fatalf("localizeNumberString() = %q, want %q", got, "−1_234·5")
	}
}

func TestNumberFormatUsesCLDRNumberSymbols(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{SignDisplay: AlwaysSignDisplay})
	if err != nil {
		t.Fatal(err)
	}
	symbols := format.symbols()
	if got := format.formatValue(nil); got != symbols.Plus+symbols.NaN {
		t.Fatalf("Format(nil) = %q, want CLDR signed NaN %q", got, symbols.Plus+symbols.NaN)
	}
	if got := format.formatValue(math.Inf(-1)); got != symbols.Minus+symbols.Infinity {
		t.Fatalf("Format(-Inf) = %q, want CLDR infinity %q", got, symbols.Minus+symbols.Infinity)
	}
	want := []Part{{Type: PartPlusSign, Value: symbols.Plus}, {Type: PartInteger, Value: "5"}}
	if got := format.formatToPartsValue(5); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(5) = %#v, want %#v", got, want)
	}

	percent, err := New(locale.MustParse("en"), Options{Style: PercentStyle})
	if err != nil {
		t.Fatal(err)
	}
	want = []Part{{Type: PartInteger, Value: "12"}, {Type: PartPercentSign, Value: symbols.Percent}}
	if got := percent.formatToPartsValue(0.123); !reflect.DeepEqual(got, want) {
		t.Fatalf("percent FormatToParts() = %#v, want %#v", got, want)
	}

	scientific, err := New(locale.MustParse("en"), Options{Notation: ScientificNotation, FractionDigits: MaximumFractionDigits(2)})
	if err != nil {
		t.Fatal(err)
	}
	want = []Part{{Type: PartInteger, Value: "1"}, {Type: PartDecimal, Value: symbols.Decimal}, {Type: PartFraction, Value: "23"}, {Type: PartExponentSeparator, Value: symbols.Exponential}, {Type: PartExponentMinusSign, Value: symbols.Minus}, {Type: PartExponentInteger, Value: "2"}}
	if got := scientific.formatToPartsValue(0.0123); !reflect.DeepEqual(got, want) {
		t.Fatalf("scientific FormatToParts() = %#v, want %#v", got, want)
	}
}

func TestNumberFormatEqualRangeUsesCLDRApproximateSign(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	symbols := format.symbols()
	if got := format.formatRangeValue(2, 2); got != symbols.ApproxSign+"2" {
		t.Fatalf("FormatRange(2, 2) = %q, want %q", got, symbols.ApproxSign+"2")
	}
	parts := format.formatRangeToPartsValue(2, 2)
	if len(parts) == 0 || parts[0] != (RangePart{Type: PartApproximatelySign, Value: symbols.ApproxSign, Source: SourceShared}) {
		t.Fatalf("FormatRangeToParts(2, 2) = %#v, want CLDR approximate sign first", parts)
	}
}

func TestNumberFormatConcurrentFormat(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{FractionDigits: FractionDigits(2, 2)})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := range 100 {
		wg.Go(func() {
			if got := format.formatValue(1234 + float64(i)/100); got == "" {
				t.Error("Format() returned empty string")
			}
		})
	}
	wg.Wait()
}
