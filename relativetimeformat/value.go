package relativetimeformat

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

// Value is an opaque ECMA-402 numeric input for RelativeTimeFormat methods.
type Value struct {
	literalKey  string
	errValue    string
	invalidErr  error
	past        bool
	numberValue numberformat.Value
	pluralValue pluralrules.Value
}

// Int returns a signed integer relative-time value.
func Int(value int64) Value {
	canonical := strconv.FormatInt(value, 10)
	past, literalKey := relativeCanonical(canonical)
	numberValue := numberformat.Int(value)
	if past {
		numberValue = numberformat.Uint(ecma402.Int64Magnitude(value))
	}
	return Value{
		literalKey:  literalKey,
		errValue:    canonical,
		past:        past,
		numberValue: numberValue,
		pluralValue: pluralrules.Int(value),
	}
}

// Uint returns an unsigned integer relative-time value.
func Uint(value uint64) Value {
	canonical := strconv.FormatUint(value, 10)
	past, literalKey := relativeCanonical(canonical)
	return Value{
		literalKey:  literalKey,
		errValue:    canonical,
		past:        past,
		numberValue: numberformat.Uint(value),
		pluralValue: pluralrules.Uint(value),
	}
}

// Float returns a float64 relative-time value.
func Float(value float64) Value {
	text := fmt.Sprint(value)
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return Value{errValue: text, invalidErr: decimal.ErrInvalidDecimal}
	}
	past, literalKey := relativeCanonical(strconv.FormatFloat(value, 'f', -1, 64))
	return Value{
		literalKey:  literalKey,
		errValue:    text,
		past:        past,
		numberValue: numberformat.Float(math.Abs(value)),
		pluralValue: pluralrules.Float(value),
	}
}

// Decimal parses a finite ECMA-402 decimal-string bridge value.
func Decimal(value string) (Value, error) {
	pluralValue, err := pluralrules.Decimal(value)
	if err != nil {
		return Value{}, invalidRelativeTimeValue(value, "", err)
	}
	absValue := strings.TrimPrefix(value, "-")
	numberValue, err := numberformat.Decimal(absValue)
	if err != nil {
		return Value{}, invalidRelativeTimeValue(value, "", err)
	}
	past, literalKey := relativeCanonical(value)
	return Value{
		literalKey:  literalKey,
		errValue:    value,
		past:        past,
		numberValue: numberValue,
		pluralValue: pluralValue,
	}, nil
}

// relativeCanonical derives, from a value's sign-preserving canonical string,
// the ECMA-402 tense flag and the numeric=auto literal lookup key — the one
// derivation shared by every numeric bridge. past follows the spec rule "value
// is -0 or value < -0"; literalKey is ToString(value): -0 collapses to "0" and
// trailing fraction zeros drop, so "1.0" matches CLDR's "1" entry.
func relativeCanonical(canonical string) (past bool, literalKey string) {
	past = strings.HasPrefix(canonical, "-")
	d, err := ecma402.ParseFiniteDecimalInput(canonical)
	if err != nil {
		return past, canonical
	}
	return past, trimTrailingFractionZeros(d.String())
}

func trimTrailingFractionZeros(s string) string {
	if !strings.ContainsRune(s, '.') {
		return s
	}
	return strings.TrimSuffix(strings.TrimRight(s, "0"), ".")
}
