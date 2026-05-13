package locale

import (
	"errors"

	"github.com/agentable/go-intl/internal/ecma402"
)

// ErrInvalidLocale classifies invalid BCP 47 tags and invalid Locale fields.
//
// Parse and New wrap this sentinel for caller-provided locale syntax and
// Unicode extension failures. Caller pattern:
// errors.Is(err, locale.ErrInvalidLocale).
var ErrInvalidLocale = errors.New("locale: invalid locale")

var (
	// ErrInvalidOption classifies ECMA-402 RangeError-equivalent option failures.
	//
	// It aliases internal/ecma402.ErrInvalidOption for package-local caller
	// matching. Caller pattern: errors.Is(err, locale.ErrInvalidOption).
	ErrInvalidOption = ecma402.ErrInvalidOption
)
