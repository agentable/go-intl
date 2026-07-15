package relativetimeformat

import (
	"fmt"
	"math"
	"strconv"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

// Value is an opaque ECMAScript Number input for RelativeTimeFormat methods.
type Value struct {
	number     float64
	errValue   string
	invalidErr error
}

// Int converts value to the same ECMAScript Number observed by native Intl.
func Int(value int64) Value {
	return Float(float64(value))
}

// Uint converts value to the same ECMAScript Number observed by native Intl.
func Uint(value uint64) Value {
	return Float(float64(value))
}

// Float returns a float64 relative-time value.
func Float(value float64) Value {
	text := fmt.Sprint(value)
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return Value{number: value, errValue: text, invalidErr: decimal.ErrInvalidDecimal}
	}
	return Value{number: value, errValue: text}
}

func (v Value) literalKey() string {
	// ECMAScript Number::toString collapses both signed zeros to "0".
	if v.number == 0 {
		return "0"
	}
	return strconv.FormatFloat(v.number, 'f', -1, 64)
}

func (v Value) isPast() bool {
	return math.Signbit(v.number)
}

func (v Value) numberFormatValue() numberformat.Value {
	return numberformat.Float(math.Abs(v.number))
}

func (v Value) pluralRulesValue() pluralrules.Value {
	return pluralrules.Float(v.number)
}
