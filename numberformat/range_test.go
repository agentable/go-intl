package numberformat

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/locale"
)

func TestNumberFormatFormatRangeEqual(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatRangeValue(1, 1); got != "~1" {
		t.Fatalf("FormatRange(1, 1) = %q, want ~1", got)
	}
}

func TestNumberFormatRangeRejectsNaN(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := format.FormatRangeFloat64(math.NaN(), 1); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("FormatRangeFloat64(NaN, 1) error = %v, want ErrInvalidValue", err)
	}
	if _, err := format.FormatRangeDecimal("NaN", "1"); !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("FormatRangeDecimal(NaN, 1) error = %v, want ErrInvalidValue", err)
	}
}

func TestNumberFormatFormatRangeDistinct(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatRangeValue(1, 2); got != "1–2" {
		t.Fatalf("FormatRange(1, 2) = %q, want 1–2", got)
	}
}

func TestNumberFormatFormatRangeToParts(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
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

func TestNumberFormatFormatRangeReversed(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"))
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

	format, err := New(locale.MustParse("en"), Options{UseGrouping: UseGroupingFalse})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatRangeValue("9007199254740993", "9007199254740992"); got != "9007199254740993–9007199254740992" {
		t.Fatalf("FormatRange() = %q, want input-order collapsed range", got)
	}
}

func TestNumberFormatRangeEqualAfterRoundingUsesApproximateSign(t *testing.T) {
	t.Parallel()

	format, err := New(locale.MustParse("en"), Options{FractionDigits: MaximumFractionDigits(0)})
	if err != nil {
		t.Fatal(err)
	}
	if got := format.formatRangeValue(1.1, 1.2); got != "~1" {
		t.Fatalf("FormatRange(1.1, 1.2) = %q, want ~1", got)
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

	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD"), CurrencyDisplay: CurrencyDisplayCode})
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
