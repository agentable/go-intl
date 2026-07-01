package ecma402

import (
	"math/big"

	"github.com/agentable/go-intl/internal/decimal"
)

// NumericValueKind identifies the exact Go bridge used to create a numeric
// method input when a formatter can use integer-specific behavior.
type NumericValueKind uint8

const (
	// NumericValueDecimal is the general ECMA-402 mathematical-value path.
	NumericValueDecimal NumericValueKind = iota
	// NumericValueInt64 preserves a signed integer bridge.
	NumericValueInt64
	// NumericValueUint64 preserves an unsigned integer bridge.
	NumericValueUint64
)

// NumericValue is the shared internal record behind public formatter numeric
// bridge types. Public packages keep their own Value wrappers.
type NumericValue struct {
	Decimal decimal.Decimal
	Kind    NumericValueKind
	Int64   int64
	Uint64  uint64
}

// DecimalNumericValue returns a general mathematical numeric value.
func DecimalNumericValue(value decimal.Decimal) NumericValue {
	return NumericValue{Decimal: value}
}

// Int64NumericValue returns a numeric value preserving its signed integer form.
func Int64NumericValue(value int64) NumericValue {
	return NumericValue{Decimal: decimal.FromInt64(value), Kind: NumericValueInt64, Int64: value}
}

// Uint64NumericValue returns a numeric value preserving its unsigned integer form.
func Uint64NumericValue(value uint64) NumericValue {
	return NumericValue{Decimal: decimal.FromUint64(value), Kind: NumericValueUint64, Uint64: value}
}

// Float64NumericValue returns a numeric value for a float64 bridge.
func Float64NumericValue(value float64) NumericValue {
	return DecimalNumericValue(decimal.FromFloat64(value))
}

// BigIntNumericValue returns a numeric value for an arbitrary-precision integer
// bridge. Nil means zero.
func BigIntNumericValue(value *big.Int) NumericValue {
	return DecimalNumericValue(decimal.FromBigInt(value))
}

// Int64Magnitude returns the unsigned absolute magnitude of value, including
// math.MinInt64.
func Int64Magnitude(value int64) uint64 {
	const minInt64 = -1 << 63
	if value >= 0 {
		return uint64(value)
	}
	if value == minInt64 {
		return 1 << 63
	}
	return uint64(-value)
}
