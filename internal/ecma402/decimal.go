package ecma402

import (
	"fmt"

	"github.com/agentable/go-intl/internal/decimal"
)

// ParseDecimalInput parses an ECMA-402 decimal-string bridge value.
func ParseDecimalInput(value string) (decimal.Decimal, error) {
	return decimal.ParseString(value)
}

// ParseFiniteDecimalInput parses a decimal-string bridge value that must be
// finite, such as PluralRules input and NumberFormat range endpoints.
func ParseFiniteDecimalInput(value string) (decimal.Decimal, error) {
	d, err := ParseDecimalInput(value)
	if err != nil {
		return decimal.Decimal{}, err
	}
	if err := RequireFiniteDecimalInput(d); err != nil {
		return decimal.Decimal{}, err
	}
	return d, nil
}

// RequireFiniteDecimalInput rejects NaN and infinities at ECMA-402 method
// boundaries where the native operation would throw on non-finite input.
func RequireFiniteDecimalInput(value decimal.Decimal) error {
	if value.IsFinite() {
		return nil
	}
	return fmt.Errorf("ecma402: non-finite decimal input %q: %w", value.String(), decimal.ErrInvalidDecimal)
}
