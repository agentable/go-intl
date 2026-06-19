package numberformat

import (
	"math"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestNumberFormatFormatRangeEqual(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatRangeValue(1, 1); got != "~1" {
		t.Fatalf("FormatRange(1, 1) = %q, want ~1", got)
	}
}

func TestNumberFormatRangeRejectsNaN(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.FormatRange(Float(math.NaN()), Int(1)); got != "" {
		t.Fatalf("FormatRange(NaN, 1) = %q, want empty invalid range output", got)
	}
	if got := format.FormatRangeToParts(Int(1), Float(math.Inf(1))); got != nil {
		t.Fatalf("FormatRangeToParts(1, +Inf) = %#v, want nil invalid range output", got)
	}
	start, err := Decimal("NaN")
	if err != nil {
		t.Fatal(err)
	}
	end, err := Decimal("Infinity")
	if err != nil {
		t.Fatal(err)
	}
	if got := format.FormatRange(start, Int(1)); got != "" {
		t.Fatalf("FormatRange(NaN, 1) = %q, want empty invalid range output", got)
	}
	if got := format.FormatRangeToParts(Int(1), end); got != nil {
		t.Fatalf("FormatRangeToParts(1, Infinity) = %#v, want nil invalid range output", got)
	}
}

func TestNumberFormatFormatRangeDistinct(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatRangeValue(1, 2); got != "1–2" {
		t.Fatalf("FormatRange(1, 2) = %q, want 1–2", got)
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
	if got := format.formatRangeToPartsValue(1, 2); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(1, 2) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatPublicRangeIntegerBridges(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: UseGroupingFalse})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name   string
		format func() string
		want   string
	}{
		{name: "int", format: func() string { return format.FormatRange(Int(int64(-1)), Int(int64(2))) }, want: "-1–2"},
		{name: "int64", format: func() string { return format.FormatRange(Int(-1), Int(2)) }, want: "-1–2"},
		{name: "uint", format: func() string { return format.FormatRange(Uint(uint64(1)), Uint(uint64(2))) }, want: "1–2"},
		{name: "uint64 max", format: func() string { return format.FormatRange(Uint(^uint64(0)-1), Uint(^uint64(0))) }, want: "18446744073709551614–18446744073709551615"},
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

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: UseGroupingFalse})
	if err != nil {
		t.Fatal(err)
	}

	gotFloat := format.FormatRange(Float(1.5), Float(2.5))
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
	gotDecimal := format.FormatRange(start, end)
	if gotDecimal != "1.25–1.50" {
		t.Fatalf("FormatRangeDecimal(1.25, 1.50) = %q, want 1.25–1.50", gotDecimal)
	}
}

func TestNumberFormatPublicRangeToPartsIntegerBridges(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: UseGroupingFalse})
	if err != nil {
		t.Fatal(err)
	}

	intParts := format.FormatRangeToParts(Int(int64(-1)), Int(int64(2)))
	if len(intParts) != 4 || intParts[0].Type != PartMinusSign || intParts[0].Source != SourceStartRange || intParts[2].Value != "–" || intParts[3].Source != SourceEndRange {
		t.Fatalf("FormatRangeIntToParts(-1, 2) = %#v, want start minus, shared separator, end integer", intParts)
	}

	uintParts := format.FormatRangeToParts(Uint(5), Uint(5))
	if len(uintParts) != 2 || uintParts[0].Type != PartApproximatelySign || uintParts[0].Source != SourceShared || uintParts[1].Value != "5" || uintParts[1].Source != SourceShared {
		t.Fatalf("FormatRangeUint64ToParts(5, 5) = %#v, want shared approximate 5", uintParts)
	}

	uintBridgeParts := format.FormatRangeToParts(Uint(uint64(1)), Uint(uint64(2)))
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
	if got := format.formatRangeValue(2, 1); got != "2–1" {
		t.Fatalf("FormatRange(2, 1) = %q, want 2–1", got)
	}
	want := []RangePart{
		{Type: PartInteger, Value: "2", Source: SourceStartRange},
		{Type: PartLiteral, Value: "–", Source: SourceShared},
		{Type: PartInteger, Value: "1", Source: SourceEndRange},
	}
	if got := format.formatRangeToPartsValue(2, 1); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(2, 1) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatRangeReversedUsesExactDecimalComparison(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{UseGrouping: UseGroupingFalse})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatRangeValue("9007199254740993", "9007199254740992"); got != "9007199254740993–9007199254740992" {
		t.Fatalf("FormatRange() = %q, want input-order collapsed range", got)
	}
}

func TestNumberFormatRangeEqualAfterRoundingUsesApproximateSign(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{MaximumFractionDigits: intPtr(0)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.FormatRange(Float(1.1), Float(1.2)); got != "~1" {
		t.Fatalf("FormatRange(1.1, 1.2) = %q; want ~1", got)
	}
	want := []RangePart{
		{Type: PartApproximatelySign, Value: "~", Source: SourceShared},
		{Type: PartInteger, Value: "1", Source: SourceShared},
	}
	if got := format.formatRangeToPartsValue(1.1, 1.2); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(1.1, 1.2) = %#v, want %#v", got, want)
	}
}

func TestNumberFormatFormatRangeToPartsCollapsesCurrency(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Style: CurrencyStyle, Currency: Currency("USD"), CurrencyDisplay: CurrencyDisplayCode})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatRangeValue(1, 2); got != "USD1.00–2.00" {
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
	if got := format.formatRangeToPartsValue(1, 2); !reflect.DeepEqual(got, want) {
		t.Fatalf("FormatRangeToParts(1, 2) = %#v, want %#v", got, want)
	}
}
