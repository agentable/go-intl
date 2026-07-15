package decimal

import "math/big"

// Abs returns d without a sign.
func Abs(d Decimal) Decimal {
	d.negative = false
	d.inner.Negative = false
	return d
}

// Neg returns d with its sign flipped. Zero is returned unchanged, so no
// negative zero is produced.
func Neg(d Decimal) Decimal {
	if d.IsZero() {
		return d
	}
	d.negative = !d.negative
	d.inner.Negative = !d.inner.Negative
	return d
}

// WithSign returns d with the requested sign, including for zero.
func WithSign(d Decimal, negative bool) Decimal {
	d.negative = negative
	d.inner.Negative = negative && !d.IsZero()
	return d
}

// MulInt returns d multiplied by n.
func MulInt(d Decimal, n int64) Decimal {
	if !d.IsFinite() {
		return d
	}
	coefficient := d.inner.Coeff.MathBigInt()
	factor := new(big.Int).Abs(big.NewInt(n))
	coefficient.Mul(coefficient, factor)
	return New(d.Negative() != (n < 0), coefficient, d.inner.Exponent)
}

// Scale10 returns d multiplied by 10^exp.
func Scale10(d Decimal, exp int32) Decimal {
	if !d.IsFinite() {
		return d
	}
	d.inner.Exponent += exp
	return d
}
