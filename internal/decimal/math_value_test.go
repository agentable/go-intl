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
		{name: "int8", in: int8(-8), want: "-8", form: Finite, negative: true},
		{name: "int16", in: int16(-16), want: "-16", form: Finite, negative: true},
		{name: "int32", in: int32(-32), want: "-32", form: Finite, negative: true},
		{name: "int64", in: int64(-64), want: "-64", form: Finite, negative: true},
		{name: "uint", in: uint(7), want: "7", form: Finite},
		{name: "uint8", in: uint8(8), want: "8", form: Finite},
		{name: "uint16", in: uint16(16), want: "16", form: Finite},
		{name: "uint32", in: uint32(32), want: "32", form: Finite},
		{name: "uint64", in: uint64(42), want: "42", form: Finite},
		{name: "uintptr", in: uintptr(9), want: "9", form: Finite},
		{name: "big int", in: big.NewInt(123456789123456789), want: "123456789123456789", form: Finite},
		{name: "negative big int value", in: *big.NewInt(-123), want: "-123", form: Finite, negative: true},
		{name: "nil big int pointer", in: (*big.Int)(nil), want: "0", form: Finite},
		{name: "float32", in: float32(1.25), want: "1.25", form: Finite},
		{name: "float", in: 3.14, want: "3.14", form: Finite},
		{name: "negative zero", in: math.Copysign(0, -1), want: "0", form: Finite, negative: true},
		{name: "NaN", in: math.NaN(), want: "NaN", form: NaN},
		{name: "infinity", in: math.Inf(1), want: "Infinity", form: Infinite},
		{name: "string", in: "1.5e3", want: "1500", form: Finite},
		{name: "invalid string", in: "not a number", want: "NaN", form: NaN},
		{name: "decimal", in: FromInt64(-5), want: "-5", form: Finite, negative: true},
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
