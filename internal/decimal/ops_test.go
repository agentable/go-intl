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

func TestNeg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "1.5", want: "-1.5"},
		{in: "-2.25", want: "2.25"},
		{in: "0", want: "0"},
		{in: "100", want: "-100"},
		{in: "-0.001", want: "0.001"},
	}
	for _, tc := range tests {
		got := Neg(mustParseDecimal(t, tc.in)).String()
		if got != tc.want {
			t.Errorf("Neg(%s) = %s, want %s", tc.in, got, tc.want)
		}
	}
}
