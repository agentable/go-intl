package pluralrules

import (
	"math/big"

	"github.com/agentable/go-intl/internal/decimal"
)

type valueKind uint8

const (
	valueDecimal valueKind = iota
	valueInt64
	valueUint64
)

// Value is an opaque ECMA-402 numeric input for PluralRules methods.
type Value struct {
	decimal decimal.Decimal
	kind    valueKind
	int64   int64
	uint64  uint64
}

// Int returns a signed integer numeric value.
func Int(v int64) Value {
	return Value{decimal: decimal.FromInt64(v), kind: valueInt64, int64: v}
}

// Uint returns an unsigned integer numeric value.
func Uint(v uint64) Value {
	return Value{decimal: decimal.New(false, new(big.Int).SetUint64(v), 0), kind: valueUint64, uint64: v}
}

// Float returns a float64 numeric value.
func Float(v float64) Value {
	return Value{decimal: decimal.FromFloat64(v)}
}

// BigInt returns an arbitrary-precision integer numeric value. A nil value is
// treated as zero.
func BigInt(v *big.Int) Value {
	return Value{decimal: decimal.New(v != nil && v.Sign() < 0, v, 0)}
}

// BigFloat returns an arbitrary-precision floating-point numeric value. A nil
// value is treated as zero.
func BigFloat(v *big.Float) Value {
	if v == nil {
		return Value{}
	}
	if v.IsInf() {
		if v.Signbit() {
			return Value{decimal: decimal.NegInfinity}
		}
		return Value{decimal: decimal.PosInfinity}
	}
	d, err := decimal.ParseString(v.Text('g', -1))
	if err != nil {
		return Value{decimal: decimal.NaNValue}
	}
	return Value{decimal: d}
}

// Decimal parses a finite ECMA-402 decimal-string bridge value.
func Decimal(s string) (Value, error) {
	d, err := parseFiniteDecimalValue("decimal", s)
	if err != nil {
		return Value{}, err
	}
	return Value{decimal: d}, nil
}
