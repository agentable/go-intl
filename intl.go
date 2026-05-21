package gointl

import (
	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
	"github.com/agentable/go-intl/segmenter"
)

type Locale = locale.Locale
type NumberFormat = numberformat.NumberFormat
type DateTimeFormat = datetimeformat.DateTimeFormat
type PluralRules = pluralrules.PluralRules
type ListFormat = listformat.ListFormat
type RelativeTimeFormat = relativetimeformat.RelativeTimeFormat
type DurationFormat = durationformat.DurationFormat
type DisplayNames = displaynames.DisplayNames
type Collator = collator.Collator
type Segmenter = segmenter.Segmenter

func GetCanonicalLocales(locales locale.List) locale.List {
	return locale.CanonicalizeList(locales)
}
