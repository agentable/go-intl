package localematcher

import "github.com/agentable/go-intl/locale"

// ErrInvalidLocale classifies invalid locale inputs passed to locale matching.
//
// It aliases locale.ErrInvalidLocale so callers can match either layer with
// errors.Is.
var ErrInvalidLocale = locale.ErrInvalidLocale
