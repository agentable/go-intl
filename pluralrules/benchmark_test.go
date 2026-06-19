package pluralrules

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func BenchmarkPluralRules_Cardinal_Cached(b *testing.B) {
	rules := benchmarkPluralRules(b, Options{})

	b.ReportAllocs()
	for b.Loop() {
		_ = mustCategory(rules.Select(Int(int64(2))))
	}
}

func BenchmarkPluralRules_Ordinal_Cached(b *testing.B) {
	rules := benchmarkPluralRules(b, Options{Type: Ordinal})

	b.ReportAllocs()
	for b.Loop() {
		_ = mustCategory(rules.Select(Int(int64(2))))
	}
}

func BenchmarkPluralRules_SelectRange_Cached(b *testing.B) {
	rules := benchmarkPluralRules(b, Options{})

	b.ReportAllocs()
	for b.Loop() {
		_ = mustCategory(rules.SelectRange(Int(int64(1)), Int(int64(2))))
	}
}

func benchmarkPluralRules(b *testing.B, opts Options) *PluralRules {
	b.Helper()

	rules, err := New(locale.List{intltest.Locale(b, "en-US")}, opts)
	if err != nil {
		b.Fatal(err)
	}
	return rules
}
