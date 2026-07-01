package ecma402

import (
	"github.com/agentable/go-intl/internal/intlerr"
)

// ErrInvalidOption is the sentinel for ECMA-402 RangeError equivalents — a
// value outside its allowed enumeration, an out-of-range numeric option, or a
// malformed identifier (currency code, unit identifier, time zone name).
var ErrInvalidOption = intlerr.ErrInvalidOption

// ErrUnsupportedOption is the sentinel for valid ECMA-402 options that the
// active implementation does not support.
var ErrUnsupportedOption = intlerr.ErrUnsupportedOption

// ErrInvalidValue is the sentinel for ECMA-402 runtime value failures.
var ErrInvalidValue = intlerr.ErrInvalidValue

// ErrInvalidCode is the sentinel for invalid DisplayNames code inputs.
var ErrInvalidCode = intlerr.ErrInvalidCode

// Error records structured Intl error context while preserving sentinel
// matching through Unwrap. It aliases the root Intl structured error type.
type Error = intlerr.Error

// OptionError records option validation context while preserving sentinel
// matching through Unwrap. It aliases the root Intl structured error type.
type OptionError = Error

// InvalidOptionErrorExpected records option validation context with
// caller-provided expected guidance.
func InvalidOptionErrorExpected(owner, name, value, loc, expected string, err error) error {
	return intlerr.NewInvalidOptionExpected(owner, name, value, loc, expected, err)
}

// UnsupportedOptionErrorExpected records unsupported-option context with
// caller-provided expected guidance.
func UnsupportedOptionErrorExpected(owner, name, value, loc, expected string, err error) error {
	return intlerr.NewUnsupportedOptionExpected(owner, name, value, loc, expected, err)
}

// InvalidValueErrorExpected records runtime value validation context with
// caller-provided expected guidance.
func InvalidValueErrorExpected(owner, name, value, loc, expected string, err error) error {
	return intlerr.NewInvalidValueExpected(owner, name, value, loc, expected, err)
}

// InvalidCodeErrorExpected records DisplayNames code validation context with
// caller-provided expected guidance.
func InvalidCodeErrorExpected(owner, name, value, loc, expected string, err error) error {
	return intlerr.NewInvalidCodeExpected(owner, name, value, loc, expected, err)
}
