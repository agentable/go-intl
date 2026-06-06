package locale

import "testing"

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
	b.ReportAllocs()
	for b.Loop() {
		loc, err := New("en-US", Options{Calendar: "gregory", HourCycle: "h23", NumberingSystem: "latn"})
		if err != nil {
			b.Fatal(err)
		}
		_ = loc
	}
}

func BenchmarkLocale_Equal_Cached(b *testing.B) {
	loc := MustParse("en-US-u-ca-gregory-hc-h23-nu-latn")
	other := MustParse("en-US-u-nu-latn-hc-h23-ca-gregory")
	b.ReportAllocs()
	for b.Loop() {
		_ = loc.Equal(other)
	}
}
