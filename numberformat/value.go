package numberformat

import (
	"math/big"

	"github.com/agentable/go-intl/internal/ecma402"
)

// Value is an opaque ECMA-402 numeric input for NumberFormat methods.
type Value struct {
	numeric ecma402.NumericValue
}

// Int returns a signed integer numeric value.
func Int(v int64) Value {
	return Value{numeric: ecma402.Int64NumericValue(v)}
}

// Uint returns an unsigned integer numeric value.
func Uint(v uint64) Value {
	return Value{numeric: ecma402.Uint64NumericValue(v)}
}

// Float returns a float64 numeric value, including NaN and infinities.
func Float(v float64) Value {
	return Value{numeric: ecma402.Float64NumericValue(v)}
}

// BigInt returns an arbitrary-precision integer numeric value. A nil value is
// treated as zero.
func BigInt(v *big.Int) Value {
	return Value{numeric: ecma402.BigIntNumericValue(v)}
}

// ParseValue parses an Intl numeric string, retaining exact finite precision.
// Empty or whitespace-only input is zero; unsigned radix integers are accepted.
// String range extremes become signed infinity or zero. Malformed input returns an error.
func ParseValue(s string) (Value, error) {
	d, err := ecma402.ParseDecimalInput(s)
	if err != nil {
		return Value{}, ecma402.InvalidDecimalValueError(numberFormatOwner, "decimal", s, err)
	}
	return Value{numeric: ecma402.DecimalNumericValue(d)}, nil
}
