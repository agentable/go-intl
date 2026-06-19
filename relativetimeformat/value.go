package relativetimeformat

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

type valueKind uint8

const (
	valueInt64 valueKind = iota
	valueUint64
	valueFloat64
	valueDecimal
	valueInvalid
)

// Value is an opaque ECMA-402 numeric input for RelativeTimeFormat methods.
type Value struct {
	kind        valueKind
	literalKey  string
	errValue    string
	past        bool
	numberValue numberformat.Value
	pluralValue pluralrules.Value
}

// Int returns a signed integer relative-time value.
func Int(value int64) Value {
	text := strconv.FormatInt(value, 10)
	numberValue := numberformat.Int(value)
	if value < 0 {
		absValue, err := numberformat.Decimal(strings.TrimPrefix(text, "-"))
		if err == nil {
			numberValue = absValue
		}
	}
	return Value{
		kind:        valueInt64,
		literalKey:  text,
		errValue:    text,
		past:        value < 0,
		numberValue: numberValue,
		pluralValue: pluralrules.Int(value),
	}
}

// Uint returns an unsigned integer relative-time value.
func Uint(value uint64) Value {
	text := strconv.FormatUint(value, 10)
	return Value{
		kind:        valueUint64,
		literalKey:  text,
		errValue:    text,
		numberValue: numberformat.Uint(value),
		pluralValue: pluralrules.Uint(value),
	}
}

// Float returns a float64 relative-time value.
func Float(value float64) Value {
	text := fmt.Sprint(value)
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return Value{kind: valueInvalid, errValue: text}
	}
	return Value{
		kind:        valueFloat64,
		literalKey:  strconv.FormatFloat(value, 'f', -1, 64),
		errValue:    text,
		past:        math.Signbit(value),
		numberValue: numberformat.Float(math.Abs(value)),
		pluralValue: pluralrules.Float(value),
	}
}

// Decimal parses a finite ECMA-402 decimal-string bridge value.
func Decimal(value string) (Value, error) {
	pluralValue, err := pluralrules.Decimal(value)
	if err != nil {
		return Value{}, invalidValue("value", value)
	}
	absValue := strings.TrimPrefix(value, "-")
	numberValue, err := numberformat.Decimal(absValue)
	if err != nil {
		return Value{}, invalidValue("value", value)
	}
	return Value{
		kind:        valueDecimal,
		literalKey:  decimalRelativeLiteralKey(value),
		errValue:    value,
		past:        strings.HasPrefix(value, "-"),
		numberValue: numberValue,
		pluralValue: pluralValue,
	}, nil
}
