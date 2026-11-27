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
	lowerCount, exact := lowerIncrementCount(x, stepCoeff, exp)
	if exact {
		return x
	}
	r1 := New(false, new(big.Int).Mul(lowerCount, stepCoeff), exp)
	upperCount := new(big.Int).Add(lowerCount, big.NewInt(1))
	r2 := New(false, new(big.Int).Mul(upperCount, stepCoeff), exp)
	return ApplyUnsignedRoundingMode(x, r1, r2, mode)
}

func lowerIncrementCount(x Decimal, stepCoeff *big.Int, stepExp int32) (count *big.Int, exact bool) {
	dividend := x.inner.Coeff.MathBigInt()
	divisor := stepCoeff
	shift := int64(x.inner.Exponent) - int64(stepExp)
	if shift >= 0 {
		// Test divisibility with modular exponentiation first. A short input such
		// as 1e1000000000 must not allocate a billion-digit 10^shift merely to
		// discover that it already lies on the rounding increment.
		factor := new(big.Int).Exp(big.NewInt(10), big.NewInt(shift), divisor)
		remainder := new(big.Int).Mod(dividend, divisor)
		remainder.Mul(remainder, factor).Mod(remainder, divisor)
		if remainder.Sign() == 0 {
			return nil, true
		}
		dividend.Mul(dividend, powerOfTen(shift))
	} else {
		gap := -shift
		if gap >= x.inner.NumDigits() {
			return new(big.Int), false
		}
		divisor = powerOfTen(gap)
		divisor.Mul(divisor, stepCoeff)
	}
	return dividend.Quo(dividend, divisor), false
}

func powerOfTen(exp int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(exp), nil)
}
