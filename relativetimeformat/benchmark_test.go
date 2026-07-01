package relativetimeformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func BenchmarkRelativeTimeFormat_Long_Cached(b *testing.B) {
	format := benchmarkRelativeTimeFormat(b, Options{})
	value := Int(-3)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := format.Format(value, Day); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRelativeTimeFormat_NumericAuto_Cached(b *testing.B) {
	format := benchmarkRelativeTimeFormat(b, Options{Numeric: stringPtr(NumericAuto)})
	value := Int(-1)

	b.ReportAllocs()
	for b.Loop() {
		if _, err := format.Format(value, Day); err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkRelativeTimeFormat(b *testing.B, opts Options) *RelativeTimeFormat {
	b.Helper()

	format, err := New(locale.List{intltest.Locale(b, "en-US")}, opts)
	if err != nil {
		b.Fatal(err)
	}
	return format
}
