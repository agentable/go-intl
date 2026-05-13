package gointl

import (
	"errors"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

// ErrInvalidKey classifies invalid Intl.supportedValuesOf keys.
var ErrInvalidKey = errors.New("intl: invalid key")

// ErrInvalidOption classifies formatter option validation failures.
//
// It aliases the shared ECMA-402 option sentinel exposed by formatter packages.
// Caller pattern: errors.Is(err, gointl.ErrInvalidOption).
var ErrInvalidOption = numberformat.ErrInvalidOption

// ErrUnsupportedLocale classifies invalid locale input at the root boundary.
//
// It aliases locale.ErrInvalidLocale. Caller pattern:
// errors.Is(err, gointl.ErrUnsupportedLocale).
var ErrUnsupportedLocale = locale.ErrInvalidLocale

// ErrUnsupportedTimeZone classifies invalid or unsupported DateTimeFormat time zones.
//
// It aliases datetimeformat.ErrUnsupportedTimeZone. Caller pattern:
// errors.Is(err, gointl.ErrUnsupportedTimeZone).
var ErrUnsupportedTimeZone = datetimeformat.ErrUnsupportedTimeZone
