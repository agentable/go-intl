package numberformat

import (
	"errors"
	"math"
	"math/big"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func mustNumberSymbolsForLocale(t *testing.T, tag string) cldrnumber.NumberSymbols {
	t.Helper()

	loc := mustNumberLocale(t, tag)
	return loc.NumberSymbols("")
}

func mustNumberLocale(t *testing.T, tag string) cldrnumber.Locale {
	t.Helper()

	loc, ok := cldrnumber.ResolveLocale(tag)
	if !ok {
		t.Fatalf("cldrnumber.ResolveLocale(%q) = false", tag)
	}
	return loc
}

func TestNumberFormatFormat(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.Format(Float(1234.5)); got != "1,234.5" {
		t.Fatalf("Format(1234.5) = %q, want 1,234.5", got)
	}
}

func TestNumberFormatFormatEqualsFormatToPartsJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		opts   Options
		value  float64
	}{
		{name: "decimal", locale: "en", opts: Options{}, value: 1234.5},
		{name: "german grouping", locale: "de-DE", opts: Options{}, value: 1234.5},
		{name: "non latin digits", locale: "en", opts: Options{NumberingSystem: stringPtr("arab")}, value: 1234.5},
		{name: "percent", locale: "en", opts: Options{Style: stringPtr(PercentStyle)}, value: 0.5},
		{name: "currency", locale: "en", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD")}, value: 1234.5},
		{name: "currency accounting", locale: "en-US", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencySign: stringPtr(AccountingCurrencySign)}, value: -12},
		{name: "currency name", locale: "en-US", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName)}, value: 1},
		{name: "currency name accounting", locale: "en-US", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName), CurrencySign: stringPtr(AccountingCurrencySign)}, value: -12},
		{name: "unit", locale: "en", opts: Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter")}, value: 2},
		{name: "scientific", locale: "en", opts: Options{Notation: stringPtr(ScientificNotation), MaximumFractionDigits: intPtr(2)}, value: 0.0123},
		{name: "engineering", locale: "en", opts: Options{Notation: stringPtr(EngineeringNotation), MaximumFractionDigits: intPtr(2)}, value: 12345},
		{name: "compact", locale: "en", opts: Options{Notation: stringPtr(CompactNotation), MaximumFractionDigits: intPtr(1)}, value: 1500},
		{name: "compact long", locale: "en", opts: Options{Notation: stringPtr(CompactNotation), CompactDisplay: stringPtr(LongCompactDisplay), MaximumFractionDigits: intPtr(1)}, value: 1500},
		{name: "nan", locale: "en", opts: Options{}, value: math.NaN()},
		{name: "negative infinity unit", locale: "en", opts: Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter")}, value: math.Inf(-1)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, tc.locale)}, tc.opts)
			if err != nil {
				t.Fatalf("New(%s) error = %v", tc.name, err)
			}
			if got, want := format.Format(Float(tc.value)), joinNumberParts(format.FormatToParts(Float(tc.value))); got != want {
				t.Fatalf("Format(%v) = %q, want joined FormatToParts %q", tc.value, got, want)
			}
		})
	}
}

func joinNumberParts(parts []Part) string {
	size := 0
	for _, part := range parts {
		size += len(part.Value)
	}
	var b strings.Builder
	b.Grow(size)
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func TestNumberFormatFractionDigitOptions(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2)})
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

func TestNumberFormatMinimumFractionDigitsOnly(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumFractionDigits: intPtr(3)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.Format(Int(1)); got != "1.000" {
		t.Fatalf("Format(1) = %q, want 1.000", got)
	}
}

func TestNumberFormatUseGroupingFalse(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1234.5); got != "1234.5" {
		t.Fatalf("Format(1234.5) = %q, want 1234.5", got)
	}
}

func TestNumberFormatUseGroupingMin2(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingMin2)})
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

	format, err := New(locale.List{intltest.Locale(t, "hi")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.Format(Int(123456789)); got != "12,34,56,789" {
		t.Fatalf("Format(123456789) = %q, want Indian grouping", got)
	}
	value, err := Decimal("123456.456")
	if err != nil {
		t.Fatal(err)
	}
	parts := format.FormatToParts(value)
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
		t.Fatalf("FormatToParts(123456.456) = %#v, want %#v", parts, wantParts)
	}

	currency, err := New(locale.List{intltest.Locale(t, "hi")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD")})
	if err != nil {
		t.Fatal(err)
	}
	if got := currency.Format(Int(123456)); got != "$1,23,456.00" {
		t.Fatalf("Currency Format(123456) = %q, want Indian currency grouping", got)
	}
}

func TestNumberFormatPublicIntegerBridges(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}
	maxUint64 := ^uint64(0)

	tests := []struct {
		name   string
		format func() string
		want   string
	}{
		{name: "int", format: func() string { return format.Format(Int(int64(-12))) }, want: "-12"},
		{name: "int64", format: func() string { return format.Format(Int(-12)) }, want: "-12"},
		{name: "uint", format: func() string { return format.Format(Uint(uint64(12))) }, want: "12"},
		{name: "uint64 max", format: func() string { return format.Format(Uint(maxUint64)) }, want: "18446744073709551615"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.format(); got != tc.want {
				t.Fatalf("%s format = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestNumberFormatDecimalInputUsesMathematicalValue(t *testing.T) {
	t.Parallel()

	locales := locale.List{intltest.Locale(t, "en")}
	defaultFormat, err := New(locales, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}
	fixedFormat, err := New(locales, Options{
		UseGrouping:           stringPtr(UseGroupingFalse),
		MinimumFractionDigits: intPtr(2),
		MaximumFractionDigits: intPtr(2),
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input       string
		wantDefault string
		wantFixed   string
	}{
		{input: "1", wantDefault: "1", wantFixed: "1.00"},
		{input: "1.0", wantDefault: "1", wantFixed: "1.00"},
		{input: "1.00", wantDefault: "1", wantFixed: "1.00"},
		{input: "1.50", wantDefault: "1.5", wantFixed: "1.50"},
		{input: "-0.00", wantDefault: "-0", wantFixed: "-0.00"},
	}
	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			value, err := Decimal(tc.input)
			if err != nil {
				t.Fatal(err)
			}
			if got := defaultFormat.Format(value); got != tc.wantDefault {
				t.Errorf("default Format(Decimal(%q)) = %q, want %q", tc.input, got, tc.wantDefault)
			}
			if got := fixedFormat.Format(value); got != tc.wantFixed {
				t.Errorf("fixed Format(Decimal(%q)) = %q, want %q", tc.input, got, tc.wantFixed)
			}
		})
	}
}

func TestNumberFormatPublicIntegerBridgesUseGroupedLatnFastPath(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		format func() string
		want   string
	}{
		{name: "int64 min", format: func() string { return format.Format(Int(math.MinInt64)) }, want: "-9,223,372,036,854,775,808"},
		{name: "uint64 max", format: func() string { return format.Format(Uint(^uint64(0))) }, want: "18,446,744,073,709,551,615"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.format(); got != tc.want {
				t.Fatalf("%s format = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestNumberFormatFormatValueCoversUnsignedFastPath(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		in   any
		want string
	}{
		{name: "uint8", in: uint8(8), want: "8"},
		{name: "uint16", in: uint16(16), want: "16"},
		{name: "uint32", in: uint32(32), want: "32"},
		{name: "uint64", in: uint64(64), want: "64"},
		{name: "uintptr", in: uintptr(9), want: "9"},
		{name: "big int", in: big.NewInt(42), want: "42"},
		{name: "big int value", in: *big.NewInt(43), want: "43"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := format.formatValue(tc.in); got != tc.want {
				t.Fatalf("formatValue(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestNumberFormatPublicToPartsBridges(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}

	intParts := format.FormatToParts(Int(int64(-12)))
	if len(intParts) != 2 || intParts[0].Type != PartMinusSign || intParts[1].Type != PartInteger || intParts[1].Value != "12" {
		t.Fatalf("FormatToParts(-12) = %#v, want minus sign and integer 12", intParts)
	}

	uintParts := format.FormatToParts(Uint(uint64(12)))
	if len(uintParts) != 1 || uintParts[0].Type != PartInteger || uintParts[0].Value != "12" {
		t.Fatalf("FormatToParts(12) = %#v, want integer 12", uintParts)
	}

	maxUintParts := format.FormatToParts(Uint(^uint64(0)))
	if len(maxUintParts) != 1 || maxUintParts[0].Type != PartInteger || maxUintParts[0].Value != "18446744073709551615" {
		t.Fatalf("FormatToParts(max) = %#v, want exact uint64 integer", maxUintParts)
	}

	floatParts := format.FormatToParts(Float(1.5))
	if len(floatParts) != 3 || floatParts[0].Value != "1" || floatParts[1].Type != PartDecimal || floatParts[2].Value != "5" {
		t.Fatalf("FormatToParts(1.5) = %#v, want integer/decimal/fraction", floatParts)
	}
}

func TestNumberFormatToPartsLocalizesDigitsWithoutMutatingSource(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{NumberingSystem: stringPtr("arab")})
	if err != nil {
		t.Fatal(err)
	}
	parts := format.FormatToParts(Int(1234))
	if len(parts) != 3 || parts[0].Value != "١" || parts[1].Type != PartGroup || parts[2].Value != "٢٣٤" {
		t.Fatalf("FormatToParts(1234) = %#v, want Arabic-Indic grouped integer parts", parts)
	}
	parts[0].Value = "mutated"
	next := format.FormatToParts(Int(1234))
	if next[0].Value != "١" {
		t.Fatalf("FormatToParts reused mutable backing data, got %#v", next)
	}
}

func TestNumberFormatFormatToPartsDecimal(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
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

func TestNumberFormatFormatInvalidValue(t *testing.T) {
	t.Parallel()

	if _, err := Decimal("not a number"); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("Decimal() error = %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "numberformat", "decimal", "not a number", "")
		testcontract.AssertErrorExpected(t, err, "a well-formed decimal string, NaN, Infinity, or -Infinity")
	}
}

func TestNumberFormatSignDisplayAlways(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{SignDisplay: stringPtr(AlwaysSignDisplay)})
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

			format, err := New(locale.List{intltest.Locale(t, "en")}, Options{SignDisplay: stringPtr(SignDisplay(tc.mode))})
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
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

	always, err := New(locale.List{intltest.Locale(t, "en")}, Options{SignDisplay: stringPtr(AlwaysSignDisplay)})
	if err != nil {
		t.Fatal(err)
	}
	if got := always.formatValue(math.NaN()); got != "+NaN" {
		t.Fatalf("always Format(NaN) = %q, want +NaN", got)
	}
	if got := always.formatValue(math.Inf(1)); got != "+∞" {
		t.Fatalf("always Format(+Inf) = %q, want +∞", got)
	}

	never, err := New(locale.List{intltest.Locale(t, "en")}, Options{SignDisplay: stringPtr(NeverSignDisplay)})
	if err != nil {
		t.Fatal(err)
	}
	if got := never.formatValue(math.Inf(-1)); got != "∞" {
		t.Fatalf("never Format(-Inf) = %q, want ∞", got)
	}
}

func TestNumberFormatSignDisplayNotation(t *testing.T) {
	t.Parallel()

	scientific, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(ScientificNotation), SignDisplay: stringPtr(AlwaysSignDisplay), MaximumFractionDigits: intPtr(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got := scientific.formatValue(1200); got != "+1.2E3" {
		t.Fatalf("scientific Format(1200) = %q, want +1.2E3", got)
	}

	compact, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(CompactNotation), SignDisplay: stringPtr(NeverSignDisplay), MaximumFractionDigits: intPtr(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got := compact.formatValue(-1500); got != "1.5K" {
		t.Fatalf("compact Format(-1500) = %q, want 1.5K", got)
	}
}

func TestNumberFormatFormatPercent(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(PercentStyle)})
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(PercentStyle), UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := format.formatValue("9007199254740993"), "900719925474099300%"; got != want {
		t.Fatalf("Format(large percent) = %q, want %q", got, want)
	}
}

func TestNumberFormatPercentPreservesArbitraryPrecisionBigInt(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(PercentStyle), UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}
	inputText := "1" + strings.Repeat("0", 99) + "5"
	input, ok := new(big.Int).SetString(inputText, 10)
	if !ok {
		t.Fatalf("SetString(%q) failed", inputText)
	}
	if got, want := format.Format(BigInt(input)), inputText+"00%"; got != want {
		t.Fatalf("Format(%s) = %q, want %q", inputText, got, want)
	}
}

func TestNumberFormatScientificPreservesDecimalMagnitude(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(ScientificNotation), MaximumSignificantDigits: intPtr(16), UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := format.formatValue("9007199254740993"), "9.007199254740993E15"; got != want {
		t.Fatalf("Format(large scientific) = %q, want %q", got, want)
	}
}

func TestNumberFormatFormatCurrencyUSD(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD")})
	if err != nil {
		t.Fatal(err)
	}
	resolved := format.ResolvedOptions()
	if resolved.MinimumFractionDigits == nil || resolved.MaximumFractionDigits == nil ||
		*resolved.MinimumFractionDigits != 2 || *resolved.MaximumFractionDigits != 2 {
		t.Fatalf("currency fraction digits = %v/%v, want 2/2", resolved.MinimumFractionDigits, resolved.MaximumFractionDigits)
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("JPY")})
	if err != nil {
		t.Fatal(err)
	}
	resolved := format.ResolvedOptions()
	if resolved.MinimumFractionDigits == nil || resolved.MaximumFractionDigits == nil ||
		*resolved.MinimumFractionDigits != 0 || *resolved.MaximumFractionDigits != 0 {
		t.Fatalf("currency fraction digits = %v/%v, want 0/0", resolved.MinimumFractionDigits, resolved.MaximumFractionDigits)
	}
	if got := format.formatValue(1234.5); got != "¥1,235" {
		t.Fatalf("Format(1234.5) = %q, want ¥1,235", got)
	}
}

func TestNumberFormatFormatCurrencyCode(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayCode)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayCode)})
	if err != nil {
		t.Fatal(err)
	}
	numberLocale := mustNumberLocale(t, "en-US")
	pattern := numberLocale.CurrencyPattern(format.ResolvedOptions().NumberingSystem, "standard")
	if got := joinNumberParts(format.formatToPartsValue(12)); got != strings.Replace(pattern, "¤#,##0.00", "USD12.00", 1) {
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
		{name: "standard negative", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD")}, in: -12, want: "-$12.00"},
		{name: "standard positive sign", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), SignDisplay: stringPtr(AlwaysSignDisplay)}, in: 12, want: "+$12.00"},
		{name: "accounting negative", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencySign: stringPtr(AccountingCurrencySign)}, in: -12, want: "($12.00)"},
		{name: "accounting hidden sign", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencySign: stringPtr(AccountingCurrencySign), SignDisplay: stringPtr(NeverSignDisplay)}, in: -12, want: "$12.00"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en-US")}, tc.opts)
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencySign: stringPtr(AccountingCurrencySign)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayName)})
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Style:                 stringPtr(CurrencyStyle),
		Currency:              stringPtr("USD"),
		CurrencyDisplay:       stringPtr(CurrencyDisplayName),
		MinimumFractionDigits: intPtr(0), MaximumFractionDigits: intPtr(0),
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

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayNarrowSymbol)})
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter")})
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

func TestNumberFormatRejectsUnitStyleWithoutUnit(t *testing.T) {
	t.Parallel()

	_, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(UnitStyle)})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(unit style without unit) error = %v, want intlerr.ErrInvalidOption", err)
	}
	testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "unit", "", "en")
	testcontract.AssertOptionExpected(t, err, `a sanctioned unit identifier when style is "unit"`)
}

func TestNumberFormatFormatScientific(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(ScientificNotation), MaximumFractionDigits: intPtr(2)})
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(EngineeringNotation), MaximumFractionDigits: intPtr(2)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(12345); got != "12.35E3" {
		t.Fatalf("Format(12345) = %q, want 12.35E3", got)
	}
}

func TestNumberFormatFormatScientificNegativeExponent(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(ScientificNotation), MaximumFractionDigits: intPtr(2)})
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(CompactNotation), MaximumFractionDigits: intPtr(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1500); got != "1.5K" {
		t.Fatalf("Format(1500) = %q, want 1.5K", got)
	}
	want := []Part{{Type: PartInteger, Value: "1"}, {Type: PartDecimal, Value: "."}, {Type: PartFraction, Value: "5"}, {Type: PartCompact, Value: "K"}}
	if got := format.formatToPartsValue(1500); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(1500) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatCompactDefaultPrecision(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(CompactNotation)})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		in   int
		want string
	}{
		{in: 1500, want: "1.5K"},
		{in: 999950, want: "1M"},
		{in: 1234567, want: "1.2M"},
	}
	for _, tc := range tests {
		if got := format.Format(Int(int64(tc.in))); got != tc.want {
			t.Fatalf("Format(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNumberFormatNotationCombinesWithStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  Options
		value float64
		want  string
	}{
		{
			name:  "compact percent",
			opts:  Options{Style: stringPtr(PercentStyle), Notation: stringPtr(CompactNotation)},
			value: 12345,
			want:  "1.2M%",
		},
		{
			name:  "scientific percent",
			opts:  Options{Style: stringPtr(PercentStyle), Notation: stringPtr(ScientificNotation)},
			value: 0.0123,
			want:  "1E0%",
		},
		{
			name:  "compact currency",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), Notation: stringPtr(CompactNotation)},
			value: 1234567,
			want:  "$1.2M",
		},
		{
			name:  "scientific currency",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), Notation: stringPtr(ScientificNotation)},
			value: 1234,
			want:  "$1.234E3",
		},
		{
			name:  "compact unit",
			opts:  Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter"), Notation: stringPtr(CompactNotation)},
			value: 1234567,
			want:  "1.2M m",
		},
		{
			name:  "scientific unit",
			opts:  Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter"), Notation: stringPtr(ScientificNotation)},
			value: 1234,
			want:  "1.234E3 m",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.Format(Float(tc.value)); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestNumberFormatSpecialValuesUseStylePatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		opts  Options
		value float64
		want  string
	}{
		{
			name:  "percent NaN",
			opts:  Options{Style: stringPtr(PercentStyle)},
			value: math.NaN(),
			want:  "NaN%",
		},
		{
			name:  "scientific percent infinity",
			opts:  Options{Style: stringPtr(PercentStyle), Notation: stringPtr(ScientificNotation)},
			value: math.Inf(1),
			want:  "∞%",
		},
		{
			name:  "accounting currency negative infinity",
			opts:  Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencySign: stringPtr(AccountingCurrencySign)},
			value: math.Inf(-1),
			want:  "($∞)",
		},
		{
			name:  "unit infinity",
			opts:  Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter")},
			value: math.Inf(1),
			want:  "∞ m",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.Format(Float(tc.value)); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.value, got, tc.want)
			}
		})
	}
}

func TestNumberFormatFormatCompactLong(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(CompactNotation), CompactDisplay: stringPtr(LongCompactDisplay), MaximumFractionDigits: intPtr(1)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatValue(1500); got != "1.5 thousand" {
		t.Fatalf("Format(1500) = %q, want 1.5 thousand", got)
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
		{name: "half expand", opts: Options{MaximumFractionDigits: intPtr(0)}, in: 1.5, want: "2"},
		{name: "half trunc", opts: Options{MaximumFractionDigits: intPtr(0), RoundingMode: stringPtr(HalfTruncRoundingMode)}, in: 1.5, want: "1"},
		{name: "half even lower", opts: Options{MaximumFractionDigits: intPtr(0), RoundingMode: stringPtr(HalfEvenRoundingMode)}, in: 2.5, want: "2"},
		{name: "half even upper", opts: Options{MaximumFractionDigits: intPtr(0), RoundingMode: stringPtr(HalfEvenRoundingMode)}, in: 3.5, want: "4"},
		{name: "ceil negative", opts: Options{MaximumFractionDigits: intPtr(0), RoundingMode: stringPtr(CeilRoundingMode)}, in: -1.9, want: "-1"},
		{name: "floor negative", opts: Options{MaximumFractionDigits: intPtr(0), RoundingMode: stringPtr(FloorRoundingMode)}, in: -1.1, want: "-2"},
		{name: "rounding increment", opts: Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2), RoundingIncrement: intPtr(5)}, in: 1.23, want: "1.25"},
		{name: "strip integer zeros", opts: Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2), TrailingZeroDisplay: stringPtr(StripIfIntegerTrailingZeroDisplay)}, in: 1, want: "1"},
		{name: "minimum integer digits", opts: Options{MinimumIntegerDigits: intPtr(3), MaximumFractionDigits: intPtr(0)}, in: 5, want: "005"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			if got := format.formatValue(tc.in); got != tc.want {
				t.Fatalf("Format(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestNumberFormatRoundingPreservesArbitraryPrecisionBigInt(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		MinimumFractionDigits: intPtr(0),
		MaximumFractionDigits: intPtr(0),
		RoundingIncrement:     intPtr(10),
		RoundingMode:          stringPtr(TruncRoundingMode),
		UseGrouping:           stringPtr(UseGroupingFalse),
	})
	if err != nil {
		t.Fatal(err)
	}
	inputText := "1" + strings.Repeat("0", 99) + "5"
	input, ok := new(big.Int).SetString(inputText, 10)
	if !ok {
		t.Fatalf("SetString(%q) failed", inputText)
	}
	want := "1" + strings.Repeat("0", 100)
	if got := format.Format(BigInt(input)); got != want {
		t.Fatalf("Format(%s) = %q, want %q", inputText, got, want)
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
		{name: "minimum significant zero", opts: Options{MinimumSignificantDigits: intPtr(3)}, in: 0, want: "0.00"},
		{name: "minimum significant decimal", opts: Options{MinimumSignificantDigits: intPtr(3)}, in: 1.2, want: "1.20"},
		{name: "maximum significant rounds integer", opts: Options{MaximumSignificantDigits: intPtr(3)}, in: 1234.5, want: "1,230"},
		{name: "minimum and maximum significant small decimal", opts: Options{MinimumSignificantDigits: intPtr(2), MaximumSignificantDigits: intPtr(4)}, in: 0.0012345, want: "0.001235"},
		{name: "minimum integer digits", opts: Options{MinimumIntegerDigits: intPtr(3), MinimumSignificantDigits: intPtr(3)}, in: 1.2, want: "001.20"},
		{name: "strip integer zeros", opts: Options{MinimumSignificantDigits: intPtr(3), TrailingZeroDisplay: stringPtr(StripIfIntegerTrailingZeroDisplay)}, in: 1, want: "1"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
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
		{name: "more precision chooses significant", opts: Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2), MinimumSignificantDigits: intPtr(2), MaximumSignificantDigits: intPtr(4), RoundingPriority: stringPtr(MorePrecisionRoundingPriority)}, in: 1.2345, want: "1.235"},
		{name: "less precision chooses fraction", opts: Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2), MinimumSignificantDigits: intPtr(2), MaximumSignificantDigits: intPtr(4), RoundingPriority: stringPtr(LessPrecisionRoundingPriority)}, in: 1.2345, want: "1.23"},
		{name: "more precision keeps integer fraction", opts: Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2), MinimumSignificantDigits: intPtr(2), MaximumSignificantDigits: intPtr(4), RoundingPriority: stringPtr(MorePrecisionRoundingPriority)}, in: 12345, want: "12,345.00"},
		{name: "less precision rounds integer significant", opts: Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2), MinimumSignificantDigits: intPtr(2), MaximumSignificantDigits: intPtr(4), RoundingPriority: stringPtr(LessPrecisionRoundingPriority)}, in: 12345, want: "12,350"},
		{name: "more precision default significant", opts: Options{MaximumFractionDigits: intPtr(2), RoundingPriority: stringPtr(MorePrecisionRoundingPriority)}, in: 1.2345, want: "1.2345"},
		{name: "less precision default significant", opts: Options{MaximumFractionDigits: intPtr(2), RoundingPriority: stringPtr(LessPrecisionRoundingPriority)}, in: 1.2345, want: "1.23"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
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

	defaults, err := New(locale.List{intltest.Locale(t, "en")}, Options{RoundingIncrement: intPtr(5)})
	if err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	resolved := defaults.ResolvedOptions()
	if resolved.MinimumFractionDigits == nil || resolved.MaximumFractionDigits == nil ||
		*resolved.MinimumFractionDigits != 0 || *resolved.MaximumFractionDigits != 0 {
		t.Fatalf("fraction digits = %v/%v, want 0/0", resolved.MinimumFractionDigits, resolved.MaximumFractionDigits)
	}
	if _, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2), RoundingIncrement: intPtr(5)}); err != nil {
		t.Fatalf("New() error = %v, want nil", err)
	}
	if _, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2), RoundingIncrement: intPtr(5), RoundingPriority: stringPtr(MorePrecisionRoundingPriority)}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
	} else {
		testcontract.AssertOptionError(t, err, "numberformat", intlerr.InvalidOption, "roundingIncrement", "5", "en")
		testcontract.AssertOptionExpected(t, err, "roundingIncrement 1 unless fraction digit rounding uses equal minimumFractionDigits and maximumFractionDigits")
	}
}

func TestNumberFormatToPartsUsesCLDRNumberSymbols(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "de-DE")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	value, err := Decimal("-1234.5")
	if err != nil {
		t.Fatal(err)
	}
	symbols := mustNumberSymbolsForLocale(t, "de-DE")
	want := []Part{
		{Type: PartMinusSign, Value: symbols.Minus},
		{Type: PartInteger, Value: "1"},
		{Type: PartGroup, Value: symbols.Group},
		{Type: PartInteger, Value: "234"},
		{Type: PartDecimal, Value: symbols.Decimal},
		{Type: PartFraction, Value: "5"},
	}
	if got := format.FormatToParts(value); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatToParts(-1234.5) = %#v, want %#v", got, want)
	}
}

func TestLocalizeNumberStringUsesNumberSymbols(t *testing.T) {
	t.Parallel()

	symbols := cldrnumber.NumberSymbols{Decimal: "·", Group: "_", Minus: "−"}
	if got := localizeNumberString("-1,234.5", symbols); got != "−1_234·5" {
		t.Fatalf("localizeNumberString() = %q, want %q", got, "−1_234·5")
	}
	symbols = cldrnumber.NumberSymbols{Decimal: ",", Group: ".", Minus: "-"}
	if got := localizeNumberString("-1,234.5", symbols); got != "-1.234,5" {
		t.Fatalf("localizeNumberString() = %q, want %q", got, "-1.234,5")
	}
}

func TestNumberFormatUsesCLDRNumberSymbols(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{SignDisplay: stringPtr(AlwaysSignDisplay)})
	if err != nil {
		t.Fatal(err)
	}
	symbols := mustNumberSymbolsForLocale(t, "en")
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

	percent, err := New(locale.List{intltest.Locale(t, "en")}, Options{Style: stringPtr(PercentStyle)})
	if err != nil {
		t.Fatal(err)
	}
	want = []Part{{Type: PartInteger, Value: "12"}, {Type: PartPercentSign, Value: symbols.Percent}}
	if got := percent.formatToPartsValue(0.123); !reflect.DeepEqual(got, want) {
		t.Fatalf("percent FormatToParts() = %#v, want %#v", got, want)
	}

	scientific, err := New(locale.List{intltest.Locale(t, "en")}, Options{Notation: stringPtr(ScientificNotation), MaximumFractionDigits: intPtr(2)})
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	symbols := mustNumberSymbolsForLocale(t, "en")
	if got := mustFormatRangeValue(t, format, 2, 2); got != symbols.ApproxSign+"2" {
		t.Fatalf("FormatRange(2, 2) = %q, want %q", got, symbols.ApproxSign+"2")
	}
	parts := mustFormatRangeToPartsValue(t, format, 2, 2)
	if len(parts) == 0 || parts[0] != (RangePart{Type: PartApproximatelySign, Value: symbols.ApproxSign, Source: SourceShared}) {
		t.Fatalf("FormatRangeToParts(2, 2) = %#v, want CLDR approximate sign first", parts)
	}
}

func TestNumberFormatConcurrentFormat(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{MinimumFractionDigits: intPtr(2), MaximumFractionDigits: intPtr(2)})
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
