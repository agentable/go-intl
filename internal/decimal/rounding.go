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
	d1 := sub(x, r1)
	d2 := sub(r2, x)
	switch d1.Cmp(d2) {
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
