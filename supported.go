package gointl

import (
	"slices"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
)

func SupportedCalendars() []string {
	return slices.Clone(cldr.SupportedCalendars())
}

func SupportedCollations() []string {
	return slices.Clone(collation.SupportedCollations())
}

func SupportedCurrencies() []string {
	return slices.Clone(cldr.SupportedCurrencies())
}

func SupportedNumberingSystems() []string {
	return slices.Clone(cldr.SupportedNumberingSystems())
}

func SupportedTimeZones() []string {
	return slices.Clone(cldr.SupportedTimeZones())
}

func SupportedUnits() []string {
	return slices.Clone(ecma402.SanctionedSimpleUnitIdentifiers())
}
