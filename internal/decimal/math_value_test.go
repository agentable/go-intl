package decimal

import (
	"errors"
	"math"
	"math/big"
	"testing"
)

func TestToIntlMathematicalValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		in       any
		want     string
		form     Form
		negative bool
	}{
		{name: "nil", in: nil, want: "0", form: Finite},
		{name: "true", in: true, want: "1", form: Finite},
		{name: "false", in: false, want: "0", form: Finite},
		{name: "int", in: int(-42), want: "-42", form: Finite, negative: true},
		{name: "uint64", in: uint64(42), want: "42", form: Finite},
		{name: "big int", in: big.NewInt(123456789123456789), want: "123456789123456789", form: Finite},
		{name: "float", in: 3.14, want: "3.14", form: Finite},
		{name: "negative zero", in: math.Copysign(0, -1), want: "0", form: Finite, negative: true},
		{name: "NaN", in: math.NaN(), want: "NaN", form: NaN},
		{name: "infinity", in: math.Inf(1), want: "Infinity", form: Infinite},
		{name: "string", in: "1.5e3", want: "1500", form: Finite},
		{name: "invalid string", in: "not a number", want: "NaN", form: NaN},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := ToIntlMathematicalValue(tc.in)
			if err != nil {
				t.Fatalf("ToIntlMathematicalValue(%v) err = %v", tc.in, err)
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

func TestToIntlMathematicalValueUnsupported(t *testing.T) {
	t.Parallel()

	_, err := ToIntlMathematicalValue(struct{}{})
	if !errors.Is(err, ErrInvalidDecimal) {
		t.Fatalf("err = %v, want errors.Is(ErrInvalidDecimal)", err)
	}
}
