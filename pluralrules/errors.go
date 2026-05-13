package pluralrules

import (
	"errors"

	"github.com/agentable/go-intl/internal/ecma402"
)

// ErrInvalidOption classifies PluralRules option validation failures.
//
// New wraps this sentinel when an option value or option combination is outside
// the ECMA-402 contract. Caller pattern:
// errors.Is(err, pluralrules.ErrInvalidOption).
var ErrInvalidOption = ecma402.ErrInvalidOption

// ErrInvalidValue classifies PluralRules numeric input failures.
//
// SelectFloat64 and SelectDecimal wrap this sentinel when the input is
// malformed or non-finite. Caller pattern:
// errors.Is(err, pluralrules.ErrInvalidValue).
var ErrInvalidValue = errors.New("pluralrules: invalid value")
