package gointl

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

type Locale = locale.Locale
type NumberFormat = numberformat.NumberFormat
type DateTimeFormat = datetimeformat.DateTimeFormat
type PluralRules = pluralrules.PluralRules
type PluralCategory = pluralrules.Category

type SupportedValueKey string

const (
	SupportedValueCalendar        SupportedValueKey = "calendar"
	SupportedValueCollation       SupportedValueKey = "collation"
	SupportedValueCurrency        SupportedValueKey = "currency"
	SupportedValueNumberingSystem SupportedValueKey = "numberingSystem"
	SupportedValueTimeZone        SupportedValueKey = "timeZone"
	SupportedValueUnit            SupportedValueKey = "unit"
)

func GetCanonicalLocales(locales ...locale.Locale) []locale.Locale {
	seen := map[string]bool{}
	out := make([]locale.Locale, 0, len(locales))
	for _, loc := range locales {
		key := loc.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, loc)
	}
	return out
}

func SupportedValuesOf(key SupportedValueKey) ([]string, error) {
	var values []string
	switch key {
	case SupportedValueCalendar:
		values = cldr.SupportedCalendars()
	case SupportedValueCollation:
		values = cldr.SupportedCollations()
	case SupportedValueCurrency:
		values = cldr.SupportedCurrencies()
	case SupportedValueNumberingSystem:
		values = cldr.SupportedNumberingSystems()
	case SupportedValueTimeZone:
		values = cldr.SupportedTimeZones()
	case SupportedValueUnit:
		values = supportedUnitValues()
	default:
		return nil, fmt.Errorf("intl: unsupported value key %q: %w", key, ErrInvalidKey)
	}
	return slices.Clone(values), nil
}

func supportedUnitValues() []string {
	seen := map[string]bool{}
	for _, unit := range ecma402.SanctionedUnits {
		if _, value, ok := strings.Cut(unit, "-"); ok {
			unit = value
		}
		seen[unit] = true
	}
	return slices.Sorted(maps.Keys(seen))
}
