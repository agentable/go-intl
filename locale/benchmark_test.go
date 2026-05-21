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
