package decimal

// Abs returns d without a sign.
func Abs(d Decimal) Decimal {
	d.negative = false
	d.inner.Negative = false
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
