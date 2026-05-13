package decimal

import (
	"errors"
	"testing"

	ecma402types "github.com/agentable/go-intl/internal/ecma402/types"
)

var _ ecma402types.MathematicalValue = Decimal{}

func TestZero(t *testing.T) {
	t.Parallel()

	if Zero.Form() != Finite {
		t.Fatalf("Zero.Form() = %v, want %v", Zero.Form(), Finite)
	}
	if Zero.Sign() != 0 {
		t.Fatalf("Zero.Sign() = %d, want 0", Zero.Sign())
	}
	if Zero.Negative() {
		t.Fatal("Zero.Negative() = true, want false")
	}
	if !Zero.IsZero() {
		t.Fatal("Zero.IsZero() = false, want true")
	}
	if Zero.Coeff() != "0" {
		t.Fatalf("Zero.Coeff() = %q, want %q", Zero.Coeff(), "0")
	}
	if Zero.Exponent() != 0 {
		t.Fatalf("Zero.Exponent() = %d, want 0", Zero.Exponent())
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
			if tc.value.Form() != tc.form {
				t.Fatalf("Form() = %v, want %v", tc.value.Form(), tc.form)
			}
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
			if tc.value.IsInfinity() != tc.isInf {
				t.Fatalf("IsInfinity() = %v, want %v", tc.value.IsInfinity(), tc.isInf)
			}
			if tc.value.IsNegative() != tc.negative {
				t.Fatalf("IsNegative() = %v, want %v", tc.value.IsNegative(), tc.negative)
			}
		})
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

func TestAbsDiffCmp(t *testing.T) {
	t.Parallel()

	base := mustParseDecimal(t, "1.2345")
	closer := mustParseDecimal(t, "1.235")
	farther := mustParseDecimal(t, "1.23")
	if got := AbsDiffCmp(base, closer, farther); got >= 0 {
		t.Fatalf("AbsDiffCmp(base, closer, farther) = %d, want closer first", got)
	}

	base = mustParseDecimal(t, "-1.2345")
	closer = mustParseDecimal(t, "-1.235")
	farther = mustParseDecimal(t, "-1.23")
	if got := AbsDiffCmp(base, closer, farther); got >= 0 {
		t.Fatalf("AbsDiffCmp(negative base, closer, farther) = %d, want closer first", got)
	}
}
