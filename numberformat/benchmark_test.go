package numberformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func BenchmarkNumberFormat_Decimal_PerCall(b *testing.B) {
	loc := intltest.Locale(b, "en-US")

	b.ReportAllocs()
	for b.Loop() {
		format, err := New(locale.List{loc}, Options{})
		if err != nil {
			b.Fatal(err)
		}
		_ = format.Format(Int(1234567))
	}
}

func BenchmarkNumberFormat_Decimal_Cached(b *testing.B) {
	format, err := New(locale.List{intltest.Locale(b, "en-US")}, Options{})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.Format(Int(1234567))
	}
}

func BenchmarkNumberFormat_Percent_Cached(b *testing.B) {
	format, err := New(locale.List{intltest.Locale(b, "en-US")}, Options{Style: stringPtr(PercentStyle)})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.Format(Float(0.5))
	}
}

func BenchmarkNumberFormat_Currency_Cached(b *testing.B) {
	format, err := New(locale.List{intltest.Locale(b, "en-US")}, Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD")})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.Format(Float(1234.5))
	}
}

func BenchmarkNumberFormat_Compact_Cached(b *testing.B) {
	format, err := New(locale.List{intltest.Locale(b, "en-US")}, Options{Notation: stringPtr(CompactNotation)})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.Format(Int(int64(1234567)))
	}
}

func BenchmarkNumberFormat_FormatToParts_Cached(b *testing.B) {
	format, err := New(locale.List{intltest.Locale(b, "en-US")}, Options{})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		_ = format.FormatToParts(Float(1234.5))
	}
}

func BenchmarkNumberFormat_FormatRange_Cached(b *testing.B) {
	format, err := New(locale.List{intltest.Locale(b, "en-US")}, Options{})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := format.FormatRange(Int(1234), Int(5678)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkNumberFormat_New(b *testing.B) {
	loc := intltest.Locale(b, "en-US")

	b.ReportAllocs()
	for b.Loop() {
		format, err := New(locale.List{loc}, Options{})
		if err != nil {
			b.Fatal(err)
		}
		_ = format
	}
}
