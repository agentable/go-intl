package decimal

import (
	"math/big"
	"slices"
)

var roundingIncrements = [...]int{1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000, 2000, 2500, 5000}

func RoundingIncrements() []int {
	return slices.Clone(roundingIncrements[:])
}

func IsValidRoundingIncrement(inc int) bool {
	return slices.Contains(roundingIncrements[:], inc)
}

func QuantizeToIncrement(x Decimal, increment int, exp int32, mode RoundingMode) Decimal {
	if !x.IsFinite() || !IsValidRoundingIncrement(increment) {
		return x
	}
	stepCoeff := big.NewInt(int64(increment))
	lowerCount, _, exact := quotientRemainder(x, stepCoeff, exp)
	if exact {
		return x
	}
	r1 := New(false, new(big.Int).Mul(lowerCount, stepCoeff), exp)
	upperCount := new(big.Int).Add(lowerCount, big.NewInt(1))
	r2 := New(false, new(big.Int).Mul(upperCount, stepCoeff), exp)
	return ApplyUnsignedRoundingMode(x, r1, r2, mode)
}

func quotientRemainder(x Decimal, divisorCoeff *big.Int, divisorExp int32) (quotient, remainder *big.Int, exact bool) {
	dividend := x.inner.Coeff.MathBigInt()
	divisor := new(big.Int).Set(divisorCoeff)
	shift := int64(x.inner.Exponent) - int64(divisorExp)
	if shift >= 0 {
		// Test divisibility with modular exponentiation first. A short input such
		// as 1e1000000000 must not allocate a billion-digit 10^shift merely to
		// discover that it already lies on the rounding increment.
		factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(shift), divisor)
		remainder := new(big.Int).Mod(new(big.Int).Set(dividend), divisor)
		remainder.Mul(remainder, factor).Mod(remainder, divisor)
		if remainder.Sign() == 0 {
			return nil, nil, true
		}
		dividend.Mul(dividend, powerOfTen(shift))
	} else {
		gap := -shift
		if gap >= x.inner.NumDigits() {
			return new(big.Int), dividend, false
		}
		divisor.Mul(divisor, powerOfTen(gap))
	}
	quotient = new(big.Int)
	remainder = new(big.Int)
	quotient.QuoRem(dividend, divisor, remainder)
	return quotient, remainder, false
}

func powerOfTen(exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
}
