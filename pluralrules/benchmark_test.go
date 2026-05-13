package pluralrules

import (
	"testing"

	"github.com/agentable/go-intl/locale"
)

func BenchmarkPluralRules_Cardinal_Cached(b *testing.B) {
	rules := benchmarkPluralRules(b)

	b.ReportAllocs()
	for b.Loop() {
		_ = rules.SelectInt(2)
	}
}

func BenchmarkPluralRules_Ordinal_Cached(b *testing.B) {
	rules := benchmarkPluralRules(b, Options{Type: Ordinal})

	b.ReportAllocs()
	for b.Loop() {
		_ = rules.SelectInt(2)
	}
}

func BenchmarkPluralRules_SelectRange_Cached(b *testing.B) {
	rules := benchmarkPluralRules(b)

	b.ReportAllocs()
	for b.Loop() {
		_ = rules.SelectRangeInt(1, 2)
	}
}

func benchmarkPluralRules(b *testing.B, opts ...Options) *PluralRules {
	b.Helper()

	rules, err := New(locale.MustParse("en-US"), opts...)
	if err != nil {
		b.Fatal(err)
	}
	return rules
}
