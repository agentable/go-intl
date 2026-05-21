package decimal

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

func TestNew(t *testing.T) {
	t.Parallel()

	d := New(false, big.NewInt(12345), -2)
	if d.Form() != Finite {
		t.Fatalf("Form() = %v, want %v", d.Form(), Finite)
	}
	if d.Coeff() != "12345" {
		t.Fatalf("Coeff() = %q, want %q", d.Coeff(), "12345")
	}
	if d.Exponent() != -2 {
		t.Fatalf("Exponent() = %d, want -2", d.Exponent())
	}
	if d.String() != "123.45" {
		t.Fatalf("String() = %q, want %q", d.String(), "123.45")
	}
	if d.Sign() != 1 {
		t.Fatalf("Sign() = %d, want 1", d.Sign())
	}
}

func TestNewPreservesNegativeZero(t *testing.T) {
	t.Parallel()

	d := New(true, big.NewInt(0), 0)
	if !d.IsZero() {
		t.Fatal("IsZero() = false, want true")
	}
	if !d.Negative() {
		t.Fatal("Negative() = false, want true")
	}
	if d.Sign() != 0 {
		t.Fatalf("Sign() = %d, want 0", d.Sign())
	}
}

func TestFromInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   int64
		want string
		sign int
	}{
		{name: "zero", in: 0, want: "0", sign: 0},
		{name: "positive", in: 42, want: "42", sign: 1},
		{name: "negative", in: -42, want: "-42", sign: -1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := FromInt64(tc.in)
			if got.String() != tc.want {
				t.Fatalf("String() = %q, want %q", got.String(), tc.want)
			}
			if got.Sign() != tc.sign {
				t.Fatalf("Sign() = %d, want %d", got.Sign(), tc.sign)
			}
		})
	}
}

func TestFromFloat64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		in       float64
		want     string
		form     Form
		negative bool
	}{
		{name: "zero", in: 0, want: "0", form: Finite},
		{name: "negative zero", in: math.Copysign(0, -1), want: "0", form: Finite, negative: true},
		{name: "decimal", in: 3.14, want: "3.14", form: Finite},
		{name: "large", in: 1e41, want: "100000000000000000000000000000000000000000", form: Finite},
		{name: "small", in: 1e-10, want: "0.0000000001", form: Finite},
		{name: "NaN", in: math.NaN(), want: "NaN", form: NaN},
		{name: "positive infinity", in: math.Inf(1), want: "Infinity", form: Infinite},
		{name: "negative infinity", in: math.Inf(-1), want: "-Infinity", form: Infinite, negative: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := FromFloat64(tc.in)
			if got.String() != tc.want {
				t.Fatalf("String() = %q, want %q", got.String(), tc.want)
			}
			if got.Form() != tc.form {
				t.Fatalf("Form() = %v, want %v", got.Form(), tc.form)
			}
			if got.Negative() != tc.negative {
				t.Fatalf("Negative() = %v, want %v", got.Negative(), tc.negative)
			}
		})
	}
}

func TestParseString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		in       string
		want     string
		form     Form
		negative bool
	}{
		{name: "decimal", in: "123.456", want: "123.456", form: Finite},
		{name: "scientific positive", in: "1.5e3", want: "1500", form: Finite},
		{name: "scientific negative", in: "1.5e-3", want: "0.0015", form: Finite},
		{name: "high precision", in: "1234567891234567.35", want: "1234567891234567.35", form: Finite},
		{name: "negative zero", in: "-0", want: "0", form: Finite, negative: true},
		{name: "NaN", in: "NaN", want: "NaN", form: NaN},
		{name: "positive infinity", in: "Infinity", want: "Infinity", form: Infinite},
		{name: "negative infinity", in: "-Infinity", want: "-Infinity", form: Infinite, negative: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ParseString(tc.in)
			if err != nil {
				t.Fatalf("ParseString(%q) err = %v", tc.in, err)
			}
			if got.String() != tc.want {
				t.Fatalf("String() = %q, want %q", got.String(), tc.want)
			}
			if got.Form() != tc.form {
				t.Fatalf("Form() = %v, want %v", got.Form(), tc.form)
			}
			if got.Negative() != tc.negative {
				t.Fatalf("Negative() = %v, want %v", got.Negative(), tc.negative)
			}
		})
	}
}

func TestParseStringInvalid(t *testing.T) {
	t.Parallel()

	_, err := ParseString("not a number")
	if !errors.Is(err, ErrInvalidDecimal) {
		t.Fatalf("err = %v, want errors.Is(ErrInvalidDecimal)", err)
	}
}
