package numberformat

import (
	"testing"

	"github.com/agentable/go-intl/locale"
)

func BenchmarkNumberFormat_Decimal_PerCall(b *testing.B) {
	loc := locale.MustParse("en-US")

	b.ReportAllocs()
	for b.Loop() {
		format, err := New(loc)
		if err != nil {
			b.Fatal(err)
		}
		_ = format.FormatInt64(1234567)
	}
}

func BenchmarkNumberFormat_Decimal_Cached(b *testing.B) {
	format, err := New(locale.MustParse("en-US"))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.FormatInt64(1234567)
	}
}

func BenchmarkNumberFormat_Percent_Cached(b *testing.B) {
	format, err := New(locale.MustParse("en-US"), Options{Style: PercentStyle})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.FormatFloat64(0.5)
	}
}

func BenchmarkNumberFormat_Currency_Cached(b *testing.B) {
	format, err := New(locale.MustParse("en-US"), Options{Style: CurrencyStyle, Currency: CurrencyCode("USD")})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.FormatFloat64(1234.5)
	}
}

func BenchmarkNumberFormat_Compact_Cached(b *testing.B) {
	format, err := New(locale.MustParse("en-US"), Options{Notation: CompactNotation})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.FormatInt(1234567)
	}
}

func BenchmarkNumberFormat_FormatToParts_Cached(b *testing.B) {
	format, err := New(locale.MustParse("en-US"))
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.FormatFloat64ToParts(1234.5)
	}
}

func BenchmarkNumberFormat_New(b *testing.B) {
	loc := locale.MustParse("en-US")

	b.ReportAllocs()
	for b.Loop() {
		format, err := New(loc)
		if err != nil {
			b.Fatal(err)
		}
		_ = format
	}
}
