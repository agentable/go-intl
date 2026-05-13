package datetimeformat

import (
	"errors"

	ecma402 "github.com/agentable/go-intl/internal/ecma402"
)

// ErrInvalidOption classifies DateTimeFormat option validation failures.
//
// New wraps this sentinel when a field option or option combination is outside
// the ECMA-402 contract. Caller pattern:
// errors.Is(err, datetimeformat.ErrInvalidOption).
var ErrInvalidOption = ecma402.ErrInvalidOption

// ErrUnsupportedTimeZone classifies invalid or unsupported time-zone options.
//
// New wraps this sentinel when Options.TimeZone cannot be resolved to a supported
// IANA location or fixed offset. Caller pattern:
// errors.Is(err, datetimeformat.ErrUnsupportedTimeZone).
var ErrUnsupportedTimeZone = errors.New("unsupported time zone")
