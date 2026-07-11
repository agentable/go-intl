package decimal

import (
	"errors"
	"testing"
)

func TestZero(t *testing.T) {
	t.Parallel()

	assertDecimalForm(t, Zero, Finite)
	if Zero.Sign() != 0 {
		t.Fatalf("Zero.Sign() = %d, want 0", Zero.Sign())
	}
	if Zero.Negative() {
		t.Fatal("Zero.Negative() = true, want false")
	}
	if !Zero.IsZero() {
		t.Fatal("Zero.IsZero() = false, want true")
	}
}

func TestSpecialValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    Decimal
		form     Form
		sign     int
		negative bool
		isNaN    bool
		isInf    bool
	}{
		{name: "NaN", value: NaNValue, form: NaN, sign: 0, isNaN: true},
		{name: "positive infinity", value: PosInfinity, form: Infinite, sign: 1, isInf: true},
		{name: "negative infinity", value: NegInfinity, form: Infinite, sign: -1, negative: true, isInf: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertDecimalForm(t, tc.value, tc.form)
			if tc.value.Sign() != tc.sign {
				t.Fatalf("Sign() = %d, want %d", tc.value.Sign(), tc.sign)
			}
			if tc.value.Negative() != tc.negative {
				t.Fatalf("Negative() = %v, want %v", tc.value.Negative(), tc.negative)
			}
			if tc.value.IsNaN() != tc.isNaN {
				t.Fatalf("IsNaN() = %v, want %v", tc.value.IsNaN(), tc.isNaN)
			}
			if tc.value.IsInf() != tc.isInf {
				t.Fatalf("IsInf() = %v, want %v", tc.value.IsInf(), tc.isInf)
			}
		})
	}
}

func assertDecimalForm(t *testing.T, d Decimal, want Form) {
	t.Helper()

	switch want {
	case Finite:
		if !d.IsFinite() {
			t.Fatal("IsFinite() = false, want true")
		}
	case Infinite:
		if !d.IsInf() {
			t.Fatal("IsInf() = false, want true")
		}
	case NaN, NaNSignaling:
		if !d.IsNaN() {
			t.Fatal("IsNaN() = false, want true")
		}
	default:
		t.Fatalf("unknown decimal form %v", want)
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()

	for _, err := range []error{
		ErrInvalidDecimal,
		ErrInvalidRoundingIncrement,
		ErrNaNComparison,
		ErrLog10Domain,
	} {
		if !errors.Is(err, err) {
			t.Fatalf("errors.Is(%v, itself) = false", err)
		}
	}
}
