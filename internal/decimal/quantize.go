package decimal

import (
	"math/big"
	"slices"

	"github.com/cockroachdb/apd/v3"
)

var ValidRoundingIncrements = []int{1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000, 2000, 2500, 5000}

func IsValidRoundingIncrement(inc int) bool {
	return slices.Contains(ValidRoundingIncrements, inc)
}

func QuantizeToIncrement(x Decimal, increment int, exp int32, mode RoundingMode) Decimal {
	if !x.IsFinite() || !IsValidRoundingIncrement(increment) {
		return x
	}
	step := New(false, big.NewInt(int64(increment)), exp)
	if step.IsZero() {
		return x
	}
	quotient := div(x, step)
	lowerCount := floor(quotient)
	upperCount := ceil(quotient)
	r1 := mul(lowerCount, step)
	r2 := mul(upperCount, step)
	return ApplyUnsignedRoundingMode(x, r1, r2, mode)
}

func div(a, b Decimal) Decimal {
	var out apd.Decimal
	_, _ = decimalContext.Quo(&out, &a.inner, &b.inner)
	return Decimal{inner: out, negative: out.Negative}
}

func mul(a, b Decimal) Decimal {
	var out apd.Decimal
	_, _ = decimalContext.Mul(&out, &a.inner, &b.inner)
	return Decimal{inner: out, negative: out.Negative}
}

func floor(d Decimal) Decimal {
	var out apd.Decimal
	_, _ = decimalContext.Floor(&out, &d.inner)
	return Decimal{inner: out, negative: out.Negative}
}

func ceil(d Decimal) Decimal {
	var out apd.Decimal
	_, _ = decimalContext.Ceil(&out, &d.inner)
	return Decimal{inner: out, negative: out.Negative}
}
