package locale

import (
	"testing"

	"golang.org/x/text/language"
)

func BenchmarkLocale_Parse(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		loc, err := Parse("en-US-u-ca-gregory-hc-h23-nu-latn")
		if err != nil {
			b.Fatal(err)
		}
		_ = loc
	}
}

func BenchmarkLocale_New(b *testing.B) {
	tag := language.MustParse("en-US")

	b.ReportAllocs()
	for b.Loop() {
		loc, err := New(tag, Options{Calendar: "gregory", HourCycle: "h23", NumberingSystem: "latn"})
		if err != nil {
			b.Fatal(err)
		}
		_ = loc
	}
}
