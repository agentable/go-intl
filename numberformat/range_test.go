package numberformat

import (
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
)

func mustFormatRange(t *testing.T, f *NumberFormat, start, end Value) string {
	t.Helper()

	out, err := f.FormatRange(start, end)
	if err != nil {
		t.Fatalf("FormatRange(%v, %v) error = %v", start.numeric.Decimal, end.numeric.Decimal, err)
	}
	return out
}

func mustFormatRangeToParts(t *testing.T, f *NumberFormat, start, end Value) []RangePart {
	t.Helper()

	parts, err := f.FormatRangeToParts(start, end)
	if err != nil {
		t.Fatalf("FormatRangeToParts(%v, %v) error = %v", start.numeric.Decimal, end.numeric.Decimal, err)
	}
	return parts
}

func TestNumberFormatFormatRangeEqual(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, Float(1), Float(1)); got != "~1" {
		t.Fatalf("FormatRange(1, 1) = %q, want ~1", got)
	}
}

func TestNumberFormatRangeRejectsNaN(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got, err := format.FormatRange(Float(math.NaN()), Int(1)); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("FormatRange(NaN, 1) = %q, error %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", got, err)
	} else if got != "" {
		t.Fatalf("FormatRange(NaN, 1) = %q with error %v, want empty output", got, err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "numberformat", "range", "start=NaN end=1", "en")
		testcontract.AssertErrorExpected(t, err, "numeric range values that are not NaN")
	}
	if parts, err := format.FormatRangeToParts(Float(math.NaN()), Int(1)); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("FormatRangeToParts(NaN, 1) = %#v, error %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", parts, err)
	} else if parts != nil {
		t.Fatalf("FormatRangeToParts(NaN, 1) = %#v with error %v, want nil parts", parts, err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "numberformat", "range", "start=NaN end=1", "en")
		testcontract.AssertErrorExpected(t, err, "numeric range values that are not NaN")
	}
	start, err := Decimal("NaN")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := format.FormatRange(start, Int(1)); !errors.Is(err, intlerr.ErrInvalidValue) || !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("FormatRange(decimal NaN, 1) = %q, error %v, want intlerr.ErrInvalidValue and ErrInvalidDecimal", got, err)
	} else if got != "" {
		t.Fatalf("FormatRange(decimal NaN, 1) = %q with error %v, want empty output", got, err)
	} else {
		testcontract.AssertIntlError(t, err, intlerr.InvalidValue, "numberformat", "range", "start=NaN end=1", "en")
		testcontract.AssertErrorExpected(t, err, "numeric range values that are not NaN")
	}
}

func TestNumberFormatRangeAcceptsInfinity(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	end, err := Decimal("Infinity")
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, Int(1), end); got != "1–∞" {
		t.Fatalf("FormatRange(1, Infinity) = %q, want 1–∞", got)
	}
	parts := mustFormatRangeToParts(t, format, Int(1), Float(math.Inf(1)))
	if len(parts) != 3 || parts[2] != (RangePart{Type: PartInfinity, Value: "∞", Source: SourceEndRange}) {
		t.Fatalf("FormatRangeToParts(1, +Inf) = %#v, want end infinity part", parts)
	}
}

func TestNumberFormatFormatRangeDistinct(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, Float(1), Float(2)); got != "1–2" {
		t.Fatalf("FormatRange(1, 2) = %q, want 1–2", got)
	}
}

func TestNumberFormatRangeUsesLocaleRangeSign(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		locale string
		want   string
	}{
		{name: "hyphen", locale: "zh", want: "1-2"},
		{name: "wave dash", locale: "ja", want: "1～2"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, tc.locale)}, Options{})
			if err != nil {
				t.Fatal(err)
			}
			if got := mustFormatRange(t, format, Int(1), Int(2)); got != tc.want {
				t.Fatalf("FormatRange(1, 2) for %s = %q, want %q", tc.locale, got, tc.want)
			}
		})
	}
}

func TestNumberFormatFormatRangeToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	want := []RangePart{
		{Type: PartInteger, Value: "1", Source: SourceStartRange},
		{Type: PartLiteral, Value: "–", Source: SourceShared},
		{Type: PartInteger, Value: "2", Source: SourceEndRange},
	}
	if got := mustFormatRangeToParts(t, format, Float(1), Float(2)); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(1, 2) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatRangeTextEqualsPartsJoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		opts       Options
		start, end Value
	}{
		{name: "distinct", start: Int(1), end: Int(2)},
		{name: "signed", start: Int(-1), end: Int(2)},
		{name: "reversed", start: Int(2), end: Int(1)},
		{name: "infinity", start: Int(1), end: Float(math.Inf(1))},
		{name: "approximate", opts: Options{MaximumFractionDigits: intPtr(0)}, start: Float(1.1), end: Float(1.2)},
		{name: "scientific", opts: Options{Notation: stringPtr(ScientificNotation), MaximumFractionDigits: intPtr(0)}, start: Int(1000), end: Int(10000)},
		{name: "currency collapse", opts: Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayCode)}, start: Int(1), end: Int(2)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			format, err := New(locale.List{intltest.Locale(t, "en")}, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			text := mustFormatRange(t, format, tc.start, tc.end)
			parts := mustFormatRangeToParts(t, format, tc.start, tc.end)
			var joined strings.Builder
			for _, part := range parts {
				joined.WriteString(part.Value)
			}
			if got := joined.String(); got != text {
				t.Errorf("FormatRangeToParts join = %q, want FormatRange %q", got, text)
			}
		})
	}
}

func TestNumberFormatPublicRangeIntegerBridges(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		format func() string
		want   string
	}{
		{name: "int", format: func() string { return mustFormatRange(t, format, Int(int64(-1)), Int(int64(2))) }, want: "-1 – 2"},
		{name: "int64", format: func() string { return mustFormatRange(t, format, Int(-1), Int(2)) }, want: "-1 – 2"},
		{name: "uint", format: func() string { return mustFormatRange(t, format, Uint(uint64(1)), Uint(uint64(2))) }, want: "1–2"},
		{name: "uint64 max", format: func() string { return mustFormatRange(t, format, Uint(^uint64(0)-1), Uint(^uint64(0))) }, want: "18446744073709551614–18446744073709551615"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := tc.format(); got != tc.want {
				t.Fatalf("%s range = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestNumberFormatPublicRangeDecimalAndFloatBridges(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}

	gotFloat := mustFormatRange(t, format, Float(1.5), Float(2.5))
	if gotFloat != "1.5–2.5" {
		t.Fatalf("FormatRangeFloat64(1.5, 2.5) = %q, want 1.5–2.5", gotFloat)
	}

	start, err := Decimal("1.25")
	if err != nil {
		t.Fatal(err)
	}
	end, err := Decimal("1.50")
	if err != nil {
		t.Fatal(err)
	}
	gotDecimal := mustFormatRange(t, format, start, end)
	if gotDecimal != "1.25–1.5" {
		t.Fatalf("FormatRangeDecimal(1.25, 1.50) = %q, want 1.25–1.5", gotDecimal)
	}

	equalStart, err := Decimal("1.5")
	if err != nil {
		t.Fatal(err)
	}
	equalEnd, err := Decimal("1.50")
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, equalStart, equalEnd); got != "~1.5" {
		t.Fatalf("FormatRangeDecimal(1.5, 1.50) = %q, want ~1.5", got)
	}
}

func TestNumberFormatPublicRangeToPartsIntegerBridges(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}

	intParts := mustFormatRangeToParts(t, format, Int(int64(-1)), Int(int64(2)))
	if len(intParts) != 4 || intParts[0].Type != PartMinusSign || intParts[0].Source != SourceStartRange || intParts[2].Value != " – " || intParts[3].Source != SourceEndRange {
		t.Fatalf("FormatRangeIntToParts(-1, 2) = %#v, want start minus, shared separator, end integer", intParts)
	}

	uintParts := mustFormatRangeToParts(t, format, Uint(5), Uint(5))
	if len(uintParts) != 2 || uintParts[0].Type != PartApproximatelySign || uintParts[0].Source != SourceShared || uintParts[1].Value != "5" || uintParts[1].Source != SourceShared {
		t.Fatalf("FormatRangeUint64ToParts(5, 5) = %#v, want shared approximate 5", uintParts)
	}

	uintBridgeParts := mustFormatRangeToParts(t, format, Uint(uint64(1)), Uint(uint64(2)))
	if len(uintBridgeParts) != 3 || uintBridgeParts[0].Value != "1" || uintBridgeParts[1].Value != "–" || uintBridgeParts[2].Value != "2" {
		t.Fatalf("FormatRangeUintToParts(1, 2) = %#v, want public uint bridge range parts", uintBridgeParts)
	}
}

func TestNumberFormatFormatRangeReversed(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, Float(2), Float(1)); got != "2–1" {
		t.Fatalf("FormatRange(2, 1) = %q, want 2–1", got)
	}
	want := []RangePart{
		{Type: PartInteger, Value: "2", Source: SourceStartRange},
		{Type: PartLiteral, Value: "–", Source: SourceShared},
		{Type: PartInteger, Value: "1", Source: SourceEndRange},
	}
	if got := mustFormatRangeToParts(t, format, Float(2), Float(1)); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(2, 1) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatRangeReversedUsesExactDecimalComparison(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: stringPtr(UseGroupingFalse)})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, mustDecimalValue(t, "9007199254740993"), mustDecimalValue(t, "9007199254740992")); got != "9007199254740993–9007199254740992" {
		t.Fatalf("FormatRange() = %q, want input-order collapsed range", got)
	}
}

func TestNumberFormatRangeEqualAfterRoundingUsesApproximateSign(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{MaximumFractionDigits: intPtr(0)})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, Float(1.1), Float(1.2)); got != "~1" {
		t.Fatalf("FormatRange(1.1, 1.2) = %q; want ~1", got)
	}
	want := []RangePart{
		{Type: PartApproximatelySign, Value: "~", Source: SourceShared},
		{Type: PartInteger, Value: "1", Source: SourceShared},
	}
	if got := mustFormatRangeToParts(t, format, Float(1.1), Float(1.2)); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(1.1, 1.2) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatRangeEqualAfterRoundingPreservesNegativeZero(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{MaximumFractionDigits: intPtr(0)})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, Float(-0.1), Float(-0.2)); got != "~-0" {
		t.Fatalf("FormatRange(-0.1, -0.2) = %q; want ~-0", got)
	}
	want := []RangePart{
		{Type: PartApproximatelySign, Value: "~", Source: SourceShared},
		{Type: PartMinusSign, Value: "-", Source: SourceShared},
		{Type: PartInteger, Value: "0", Source: SourceShared},
	}
	if got := mustFormatRangeToParts(t, format, Float(-0.1), Float(-0.2)); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(-0.1, -0.2) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatRangeScientificExponentChangesVisibleOutput(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		Notation:              stringPtr(ScientificNotation),
		MaximumFractionDigits: intPtr(0),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, Int(1000), Int(10000)); got != "1E3–1E4" {
		t.Fatalf("FormatRange(1000, 10000) = %q; want 1E3–1E4", got)
	}
	want := []RangePart{
		{Type: PartInteger, Value: "1", Source: SourceStartRange},
		{Type: PartExponentSeparator, Value: "E", Source: SourceStartRange},
		{Type: PartExponentInteger, Value: "3", Source: SourceStartRange},
		{Type: PartLiteral, Value: "–", Source: SourceShared},
		{Type: PartInteger, Value: "1", Source: SourceEndRange},
		{Type: PartExponentSeparator, Value: "E", Source: SourceEndRange},
		{Type: PartExponentInteger, Value: "4", Source: SourceEndRange},
	}
	if got := mustFormatRangeToParts(t, format, Int(1000), Int(10000)); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(1000, 10000) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatRangeToPartsCollapsesCurrency(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD"), CurrencyDisplay: stringPtr(CurrencyDisplayCode)})
	if err != nil {
		t.Fatal(err)
	}
	if got := mustFormatRange(t, format, Float(1), Float(2)); got != "USD1.00–2.00" {
		t.Fatalf("FormatRange(1, 2) = %q, want USD1.00–2.00", got)
	}
	want := []RangePart{
		{Type: PartCurrency, Value: "USD", Source: SourceStartRange},
		{Type: PartInteger, Value: "1", Source: SourceStartRange},
		{Type: PartDecimal, Value: ".", Source: SourceStartRange},
		{Type: PartFraction, Value: "00", Source: SourceStartRange},
		{Type: PartLiteral, Value: "–", Source: SourceShared},
		{Type: PartInteger, Value: "2", Source: SourceEndRange},
		{Type: PartDecimal, Value: ".", Source: SourceEndRange},
		{Type: PartFraction, Value: "00", Source: SourceEndRange},
	}
	if got := mustFormatRangeToParts(t, format, Float(1), Float(2)); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(1, 2) = %#v, want %#v", got, want)
	}
}
