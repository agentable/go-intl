package decimal

import (
	"fmt"
	"math/big"
)

func ToIntlMathematicalValue(value any) (Decimal, error) {
	switch v := value.(type) {
	case nil:
		return Zero, nil
	case bool:
		if v {
			return FromInt64(1), nil
		}
		return Zero, nil
	case int:
		return FromInt64(int64(v)), nil
	case int8:
		return FromInt64(int64(v)), nil
	case int16:
		return FromInt64(int64(v)), nil
	case int32:
		return FromInt64(int64(v)), nil
	case int64:
		return FromInt64(v), nil
	case uint:
		return New(false, new(big.Int).SetUint64(uint64(v)), 0), nil
	case uint8:
		return FromInt64(int64(v)), nil
	case uint16:
		return FromInt64(int64(v)), nil
	case uint32:
		return FromInt64(int64(v)), nil
	case uint64:
		return New(false, new(big.Int).SetUint64(v), 0), nil
	case uintptr:
		return New(false, new(big.Int).SetUint64(uint64(v)), 0), nil
	case *big.Int:
		if v == nil {
			return Zero, nil
		}
		return New(v.Sign() < 0, v, 0), nil
	case big.Int:
		return New(v.Sign() < 0, &v, 0), nil
	case float32:
		return FromFloat64(float64(v)), nil
	case float64:
		return FromFloat64(v), nil
	case string:
		d, ok := parseStringValue(v)
		if !ok {
			return NaNValue, nil
		}
		return d, nil
	case Decimal:
		return v, nil
	}
	return Decimal{}, fmt.Errorf("decimal: unsupported value %T: %w", value, ErrInvalidDecimal)
}

func parseStringValue(s string) (Decimal, bool) {
	d, err := ParseString(s)
	if err != nil {
		return Decimal{}, false
	}
	return d, true
}
