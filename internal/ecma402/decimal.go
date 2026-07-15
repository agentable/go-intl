package ecma402

import (
	"fmt"

	"github.com/agentable/go-intl/internal/decimal"
)

const (
	// ExpectedDecimalValue describes ECMA-402 decimal-string bridge inputs that
	// allow NumberFormat's special numeric values.
	ExpectedDecimalValue = "a well-formed decimal string, NaN, Infinity, or -Infinity"
	// ExpectedFiniteNumericValue describes ECMA-402 method inputs that reject
	// NaN and infinities.
	ExpectedFiniteNumericValue = "a finite numeric value"
)

// ParseDecimalInput parses an ECMA-402 decimal-string bridge value.
func ParseDecimalInput(value string) (decimal.Decimal, error) {
	return decimal.ParseString(value)
}

// ParseFiniteDecimalInput parses a decimal-string bridge value that must be
// finite at the owning operation boundary.
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

// InvalidDecimalValueError records a decimal-string bridge value failure with
// the shared ECMA-402 decimal-input guidance.
func InvalidDecimalValueError(owner, name, value string, err error) error {
	return InvalidValueErrorExpected(owner, name, value, "", ExpectedDecimalValue, err)
}

// InvalidFiniteNumericValueError records a finite numeric value failure with
// the shared ECMA-402 finite-input guidance.
func InvalidFiniteNumericValueError(owner, name, value, loc string, err error) error {
	return InvalidValueErrorExpected(owner, name, value, loc, ExpectedFiniteNumericValue, err)
}
