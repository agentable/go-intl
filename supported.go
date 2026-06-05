package gointl

import (
	"slices"

	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrtimezone "github.com/agentable/go-intl/internal/cldr/timezone"
	"github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
)

func SupportedCalendars() []string {
	return slices.Clone(cldrdate.SupportedCalendars())
}

func SupportedCollations() []string {
	return slices.Clone(collation.SupportedCollations())
}

func SupportedCurrencies() []string {
	return slices.Clone(cldrcurrency.SupportedCodes())
}

func SupportedNumberingSystems() []string {
	return slices.Clone(cldrnumber.SupportedNumberingSystems())
}

func SupportedTimeZones() []string {
	return slices.Clone(cldrtimezone.SupportedTimeZones())
}

func SupportedUnits() []string {
	return slices.Clone(ecma402.SanctionedSimpleUnitIdentifiers())
}
