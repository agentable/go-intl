package decimal

import (
	"fmt"

	"github.com/cockroachdb/apd/v3"
)

type RoundingMode int

const (
	RoundCeil RoundingMode = iota
	RoundFloor
	RoundExpand
	RoundTrunc
	RoundHalfCeil
	RoundHalfFloor
	RoundHalfExpand
	RoundHalfTrunc
	RoundHalfEven
)

func (m RoundingMode) String() string {
	switch m {
	case RoundCeil:
		return "ceil"
	case RoundFloor:
		return "floor"
	case RoundExpand:
		return "expand"
	case RoundTrunc:
		return "trunc"
	case RoundHalfCeil:
		return "halfCeil"
	case RoundHalfFloor:
		return "halfFloor"
	case RoundHalfExpand:
		return "halfExpand"
	case RoundHalfTrunc:
		return "halfTrunc"
	case RoundHalfEven:
		return "halfEven"
	default:
		return "unknown"
	}
}

func ParseRoundingMode(s string) (RoundingMode, error) {
	switch s {
	case "ceil":
		return RoundCeil, nil
	case "floor":
		return RoundFloor, nil
	case "expand":
		return RoundExpand, nil
	case "trunc":
		return RoundTrunc, nil
	case "halfCeil":
		return RoundHalfCeil, nil
	case "halfFloor":
		return RoundHalfFloor, nil
	case "halfExpand":
		return RoundHalfExpand, nil
	case "halfTrunc":
		return RoundHalfTrunc, nil
	case "halfEven":
		return RoundHalfEven, nil
	default:
		return 0, fmt.Errorf("decimal: invalid rounding mode %q: %w", s, ErrInvalidDecimal)
	}
}

func GetUnsignedRoundingMode(mode RoundingMode, isNegative bool) RoundingMode {
	if isNegative {
		switch mode {
		case RoundCeil, RoundTrunc:
			return RoundTrunc
		case RoundFloor, RoundExpand:
			return RoundExpand
		case RoundHalfCeil, RoundHalfTrunc:
			return RoundHalfTrunc
		case RoundHalfFloor, RoundHalfExpand:
			return RoundHalfExpand
		case RoundHalfEven:
			return RoundHalfEven
		}
	}
	switch mode {
	case RoundCeil, RoundExpand:
		return RoundExpand
	case RoundFloor, RoundTrunc:
		return RoundTrunc
	case RoundHalfCeil, RoundHalfExpand:
		return RoundHalfExpand
	case RoundHalfFloor, RoundHalfTrunc:
		return RoundHalfTrunc
	case RoundHalfEven:
		return RoundHalfEven
	}
	return RoundHalfEven
}

func ApplyUnsignedRoundingMode(x, r1, r2 Decimal, mode RoundingMode) Decimal {
	if compare(x, r1) == 0 || compare(r1, r2) == 0 {
		return r1
	}
	if compare(x, r2) == 0 {
		return r2
	}
	switch mode {
	case RoundTrunc:
		return r1
	case RoundExpand:
		return r2
	case RoundCeil, RoundFloor, RoundHalfCeil, RoundHalfFloor, RoundHalfExpand, RoundHalfTrunc, RoundHalfEven:
	}
	d1 := sub(x, r1)
	d2 := sub(r2, x)
	switch compare(d1, d2) {
	case -1:
		return r1
	case 1:
		return r2
	}
	switch mode {
	case RoundHalfTrunc:
		return r1
	case RoundHalfExpand:
		return r2
	case RoundCeil, RoundFloor, RoundExpand, RoundTrunc, RoundHalfCeil, RoundHalfFloor, RoundHalfEven:
	}
	if halfEvenLower(r1, r2) {
		return r1
	}
	return r2
}

func compare(a, b Decimal) int {
	return a.inner.Cmp(&b.inner)
}

func sub(a, b Decimal) Decimal {
	var out apd.Decimal
	_, _ = decimalContext.Sub(&out, &a.inner, &b.inner)
	return Decimal{inner: out, negative: out.Negative}
}

func halfEvenLower(r1, r2 Decimal) bool {
	step := sub(r2, r1)
	if step.IsZero() {
		return true
	}
	var quotient apd.Decimal
	_, _ = decimalContext.Quo(&quotient, &r1.inner, &step.inner)
	var remainder apd.Decimal
	_, _ = decimalContext.Rem(&remainder, &quotient, apd.New(2, 0))
	return remainder.Coeff.Sign() == 0
}

var decimalContext = apd.Context{Precision: 100, MaxExponent: 1000000, MinExponent: -1000000}
