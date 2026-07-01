package collator_test

import (
	"testing"

	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func BenchmarkCollator_Compare_Cached(b *testing.B) {
	compare := benchmarkCollator(b, collator.Options{})

	b.ReportAllocs()
	for b.Loop() {
		if got := compare.Compare("alpha", "beta"); got >= 0 {
			b.Fatalf("Compare(alpha, beta) = %d", got)
		}
	}
}

func BenchmarkCollator_CompareNumeric_Cached(b *testing.B) {
	compare := benchmarkCollator(b, collator.Options{Numeric: boolPtr(true)})

	b.ReportAllocs()
	for b.Loop() {
		if got := compare.Compare("2", "10"); got >= 0 {
			b.Fatalf("Compare(2, 10) = %d", got)
		}
	}
}

func benchmarkCollator(b *testing.B, opts collator.Options) *collator.Collator {
	b.Helper()

	compare, err := collator.New(locale.List{intltest.Locale(b, "en-US")}, opts)
	if err != nil {
		b.Fatal(err)
	}
	return compare
}
