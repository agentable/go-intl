package numberformat

import (
	"errors"

	ecma402 "github.com/agentable/go-intl/internal/ecma402"
)

// ErrInvalidOption classifies NumberFormat option validation failures.
//
// New wraps this sentinel when an option value or option combination is outside
// the ECMA-402 contract. Caller pattern:
// errors.Is(err, numberformat.ErrInvalidOption).
var ErrInvalidOption = ecma402.ErrInvalidOption

// ErrInvalidValue classifies NumberFormat decimal input failures.
//
// FormatDecimal, FormatDecimalToParts, and decimal range methods wrap this
// sentinel when the input is malformed. Caller pattern:
// errors.Is(err, numberformat.ErrInvalidValue).
var ErrInvalidValue = errors.New("numberformat: invalid value")
