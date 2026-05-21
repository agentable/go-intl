package ecma402

import (
	"github.com/agentable/go-intl/internal/intlerr"
)

// ErrInvalidOption is the sentinel for ECMA-402 RangeError equivalents — a
// value outside its allowed enumeration, an out-of-range numeric option, or a
// malformed identifier (currency code, unit identifier, time zone name).
var ErrInvalidOption = intlerr.ErrInvalidOption

// OptionError records option validation context while preserving sentinel
// matching through Unwrap. It aliases the root Intl structured error type.
type OptionError = intlerr.Error

// InvalidOptionError returns an OptionError wrapping an invalid-option
// sentinel.
func InvalidOptionError(owner, name, value, loc string, err error) error {
	return intlerr.New(intlerr.InvalidOption, owner, name, value, loc, err)
}

// UnsupportedOptionError returns an OptionError wrapping an unsupported-option
// sentinel.
func UnsupportedOptionError(owner, name, value, loc string, err error) error {
	return intlerr.New(intlerr.UnsupportedOption, owner, name, value, loc, err)
}
