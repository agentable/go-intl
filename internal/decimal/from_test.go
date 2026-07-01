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
	assertDecimalForm(t, d, Finite)
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

func TestFromUint64(t *testing.T) {
	t.Parallel()

	got := FromUint64(math.MaxUint64)
	if got.String() != "18446744073709551615" {
		t.Fatalf("String() = %q, want %q", got.String(), "18446744073709551615")
	}
	if got.Sign() != 1 {
		t.Fatalf("Sign() = %d, want 1", got.Sign())
	}
}

func TestFromBigInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		in       *big.Int
		want     string
		negative bool
	}{
		{name: "nil", want: "0"},
		{name: "positive", in: big.NewInt(42), want: "42"},
		{name: "negative", in: big.NewInt(-42), want: "-42", negative: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := FromBigInt(tc.in)
			if got.String() != tc.want {
				t.Fatalf("String() = %q, want %q", got.String(), tc.want)
			}
			if got.Negative() != tc.negative {
				t.Fatalf("Negative() = %v, want %v", got.Negative(), tc.negative)
			}
		})
	}
}

func TestFromBigIntCopiesInput(t *testing.T) {
	t.Parallel()

	in := big.NewInt(42)
	got := FromBigInt(in)
	in.SetInt64(7)

	if got.String() != "42" {
		t.Fatalf("String() after caller mutation = %q, want 42", got.String())
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
			assertDecimalForm(t, got, tc.form)
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
			assertDecimalForm(t, got, tc.form)
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
