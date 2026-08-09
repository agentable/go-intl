package gointl

import (
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
)

// Locale is the root Intl.Locale constructor-property alias for locale.Locale.
type Locale = locale.Locale

// NumberFormat is the root Intl.NumberFormat constructor-property alias for
// numberformat.NumberFormat.
type NumberFormat = numberformat.NumberFormat

// DateTimeFormat is the root Intl.DateTimeFormat constructor-property alias for
// datetimeformat.DateTimeFormat.
type DateTimeFormat = datetimeformat.DateTimeFormat

// PluralRules is the root Intl.PluralRules constructor-property alias for
// pluralrules.PluralRules.
type PluralRules = pluralrules.PluralRules

// ListFormat is the root Intl.ListFormat constructor-property alias for
// listformat.ListFormat.
type ListFormat = listformat.ListFormat

// RelativeTimeFormat is the root Intl.RelativeTimeFormat constructor-property alias
// for relativetimeformat.RelativeTimeFormat.
type RelativeTimeFormat = relativetimeformat.RelativeTimeFormat

// DurationFormat is the root Intl.DurationFormat constructor-property alias for
// durationformat.DurationFormat.
type DurationFormat = durationformat.DurationFormat

// DisplayNames is the root Intl.DisplayNames constructor-property alias for
// displaynames.DisplayNames.
type DisplayNames = displaynames.DisplayNames

// GetCanonicalLocales returns the Intl.getCanonicalLocales canonical locale list.
//
// It keeps the first occurrence of each canonical locale while preserving
// request order. Raw locale string parsing stays with the locale package.
func GetCanonicalLocales(locales locale.List) locale.List {
	return ecma402.CanonicalLocaleList(locales)
}
