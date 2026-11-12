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

func BenchmarkNumberFormat_SupportedLocalesOf(b *testing.B) {
	for _, tag := range []string{"en", "nn"} {
		for _, algorithm := range []string{"lookup", "best fit"} {
			b.Run(tag+"/"+algorithm, func(b *testing.B) {
				requested := intltest.LocaleList(b, tag)
				opts := Options{LocaleMatcher: &algorithm}
				if _, err := SupportedLocalesOf(requested, opts); err != nil {
					b.Fatal(err)
				}
				b.ReportAllocs()
				for b.Loop() {
					if _, err := SupportedLocalesOf(requested, opts); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

// Each method receives the same prebuilt mathematical values and formatter.
func BenchmarkNumberFormat_Partition_Cached(b *testing.B) {
	value := Float(1234567.89)
	end := Float(2345678.91)
	for _, tt := range []struct {
		name string
		opts Options
	}{
		{"decimal", Options{}},
		{"currency", Options{Style: stringPtr(CurrencyStyle), Currency: stringPtr("USD")}},
		{"unit", Options{Style: stringPtr(UnitStyle), Unit: stringPtr("meter-per-second")}},
		{"compact", Options{Notation: stringPtr(CompactNotation)}},
		{"scientific", Options{Notation: stringPtr(ScientificNotation)}},
		{"arab", Options{NumberingSystem: stringPtr("arab")}},
		{"precision", Options{MinimumFractionDigits: intPtr(20), MaximumFractionDigits: intPtr(20)}},
	} {
		b.Run(tt.name, func(b *testing.B) {
			f, err := New(intltest.LocaleList(b, "en"), tt.opts)
			if err != nil {
				b.Fatal(err)
			}
			b.Run("text", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = f.Format(value)
				}
			})
			b.Run("parts", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					_ = f.FormatToParts(value)
				}
			})
			b.Run("range", func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					if _, err := f.FormatRange(value, end); err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}
