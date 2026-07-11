package decimal

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

// MulInt returns d multiplied by n.
func MulInt(d Decimal, n int64) Decimal {
	if !d.IsFinite() {
		return d
	}
	return mul(d, FromInt64(n))
}

// Scale10 returns d multiplied by 10^exp.
func Scale10(d Decimal, exp int32) Decimal {
	if !d.IsFinite() {
		return d
	}
	d.inner.Exponent += exp
	return d
}
