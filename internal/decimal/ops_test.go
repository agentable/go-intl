package decimal

import "testing"

func TestDecimalOps(t *testing.T) {
	t.Parallel()

	d := mustParseDecimal(t, "-123.45")
	if got, want := Abs(d).String(), "123.45"; got != want {
		t.Fatalf("Abs(%s) = %s, want %s", d.String(), got, want)
	}
	if got, want := MulInt(d, 100).String(), "-12345.00"; got != want {
		t.Fatalf("MulInt(%s, 100) = %s, want %s", d.String(), got, want)
	}
	if got, want := Scale10(d, -2).String(), "-1.2345"; got != want {
		t.Fatalf("Scale10(%s, -2) = %s, want %s", d.String(), got, want)
	}
}
