package decimal

import "testing"

func BenchmarkParseString(b *testing.B) {
	for b.Loop() {
		if _, err := ParseString("1234567890.123456789"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLog10Floor(b *testing.B) {
	x, err := ParseString("123456789012345678901234567890.12345")
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		if _, err := Log10Floor(x); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkQuantizeToIncrement(b *testing.B) {
	x, err := ParseString("12345.6789")
	if err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		_ = QuantizeToIncrement(x, 25, -2, RoundHalfExpand)
	}
}
