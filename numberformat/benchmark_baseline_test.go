package numberformat

import (
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

func BenchmarkBaseline_XText_Decimal(b *testing.B) {
	printer := message.NewPrinter(language.AmericanEnglish)

	b.ReportAllocs()
	for b.Loop() {
		_ = printer.Sprintf("%v", 1234567)
	}
}

func BenchmarkBaseline_XText_Percent(b *testing.B) {
	printer := message.NewPrinter(language.AmericanEnglish)

	b.ReportAllocs()
	for b.Loop() {
		_ = printer.Sprintf("%v", number.Percent(0.5))
	}
}
