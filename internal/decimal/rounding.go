package decimal

import (
	"fmt"
	"math/big"
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

var roundingModeNames = [...]string{
	RoundCeil:       "ceil",
	RoundFloor:      "floor",
	RoundExpand:     "expand",
	RoundTrunc:      "trunc",
	RoundHalfCeil:   "halfCeil",
	RoundHalfFloor:  "halfFloor",
	RoundHalfExpand: "halfExpand",
	RoundHalfTrunc:  "halfTrunc",
	RoundHalfEven:   "halfEven",
}

func (m RoundingMode) String() string {
	if m < 0 || int(m) >= len(roundingModeNames) {
		return "unknown"
	}
	return roundingModeNames[m]
}

func ParseRoundingMode(s string) (RoundingMode, error) {
	for mode, name := range roundingModeNames {
		if s == name {
			return RoundingMode(mode), nil
		}
	}
	return 0, fmt.Errorf("decimal: invalid rounding mode %q: %w", s, ErrInvalidDecimal)
}

func UnsignedRoundingMode(mode RoundingMode, isNegative bool) RoundingMode {
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
	if x.Cmp(r1) == 0 || r1.Cmp(r2) == 0 {
		return r1
	}
	if x.Cmp(r2) == 0 {
		return r2
	}
	switch mode {
	case RoundTrunc:
		return r1
	case RoundExpand:
		return r2
	case RoundCeil, RoundFloor, RoundHalfCeil, RoundHalfFloor, RoundHalfExpand, RoundHalfTrunc, RoundHalfEven:
	}
	switch compareRoundingDistances(x, r1, r2) {
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

func compareRoundingDistances(x, r1, r2 Decimal) int {
	return MulInt(x, 2).Cmp(addRoundingBoundaries(r1, r2))
}

func halfEvenLower(r1, r2 Decimal) bool {
	lower := r1.inner.Coeff.MathBigInt()
	step := r2.inner.Coeff.MathBigInt()
	step.Sub(step, lower)
	if step.Sign() == 0 {
		return true
	}
	cardinality := new(big.Int)
	remainder := new(big.Int)
	cardinality.QuoRem(lower, step, remainder)
	return remainder.Sign() == 0 && cardinality.Bit(0) == 0
}

// addRoundingBoundaries relies on the ApplyUnsignedRoundingMode contract: r1
// and r2 are adjacent non-negative multiples of one rounding increment and
// therefore share an exponent.
func addRoundingBoundaries(r1, r2 Decimal) Decimal {
	coeff := r1.inner.Coeff.MathBigInt()
	coeff.Add(coeff, r2.inner.Coeff.MathBigInt())
	return New(false, coeff, r1.inner.Exponent)
}
