package pluralrules

import (
	"testing"

	"golang.org/x/text/feature/plural"
	"golang.org/x/text/language"
)

func BenchmarkBaseline_XText_Plural_Cardinal(b *testing.B) {
	tag := language.AmericanEnglish

	b.ReportAllocs()
	for b.Loop() {
		_ = plural.Cardinal.MatchPlural(tag, 2, 0, 0, 0, 0)
	}
}

func BenchmarkBaseline_XText_Plural_Ordinal(b *testing.B) {
	tag := language.AmericanEnglish

	b.ReportAllocs()
	for b.Loop() {
		_ = plural.Ordinal.MatchPlural(tag, 2, 0, 0, 0, 0)
	}
}
