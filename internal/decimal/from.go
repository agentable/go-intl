package decimal

import (
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/cockroachdb/apd/v3"
)

func New(negative bool, coeff *big.Int, exp int32) Decimal {
	if coeff == nil {
		coeff = new(big.Int)
	}
	var apdCoeff apd.BigInt
	apdCoeff.SetMathBigInt(new(big.Int).Abs(coeff))
	inner := *apd.NewWithBigInt(&apdCoeff, exp)
	if negative || coeff.Sign() < 0 {
		inner.Negative = coeff.Sign() != 0
	}
	return Decimal{inner: inner, negative: negative || coeff.Sign() < 0}
}

func FromInt64(n int64) Decimal {
	var inner apd.Decimal
	inner.SetFinite(n, 0)
	return Decimal{inner: inner, negative: n < 0}
}

func FromFloat64(f float64) Decimal {
	switch {
	case math.IsNaN(f):
		return NaNValue
	case math.IsInf(f, 1):
		return PosInfinity
	case math.IsInf(f, -1):
		return NegInfinity
	case f == 0:
		return New(math.Signbit(f), big.NewInt(0), 0)
	}
	d, err := ParseString(strconv.FormatFloat(f, 'g', -1, 64))
	if err != nil {
		return NaNValue
	}
	return d
}

func ParseString(s string) (Decimal, error) {
	switch s {
	case "NaN":
		return NaNValue, nil
	case "Infinity", "+Infinity":
		return PosInfinity, nil
	case "-Infinity":
		return NegInfinity, nil
	}
	var inner apd.Decimal
	if _, _, err := inner.SetString(s); err != nil {
		return Decimal{}, fmt.Errorf("decimal: parse %q: %w", s, ErrInvalidDecimal)
	}
	return Decimal{inner: inner, negative: inner.Negative}, nil
}
