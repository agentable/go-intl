package gointl

import (
	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrtimezone "github.com/agentable/go-intl/internal/cldr/timezone"
	"github.com/agentable/go-intl/internal/collation"
	"github.com/agentable/go-intl/internal/ecma402"
)

// SupportedCalendars returns the calendar identifiers for
// Intl.supportedValuesOf("calendar").
//
// The returned slice is sorted, canonical, and independent.
func SupportedCalendars() []string {
	return cldrdate.SupportedCalendars()
}

// SupportedCollations returns the active Collator backend collation identifiers
// for Intl.supportedValuesOf("collation").
//
// The returned slice is sorted, canonical, and independent.
func SupportedCollations() []string {
	return collation.SupportedCollations()
}

// SupportedCurrencies returns the currency identifiers for
// Intl.supportedValuesOf("currency").
//
// The returned slice is sorted, canonical, and independent.
func SupportedCurrencies() []string {
	return cldrcurrency.SupportedCodes()
}

// SupportedNumberingSystems returns the numbering-system identifiers for
// Intl.supportedValuesOf("numberingSystem").
//
// The returned slice is sorted, canonical, and independent.
func SupportedNumberingSystems() []string {
	return cldrnumber.SupportedNumberingSystems()
}

// SupportedTimeZones returns the primary IANA time-zone identifiers for
// Intl.supportedValuesOf("timeZone").
//
// The returned slice is sorted, canonical, and independent.
func SupportedTimeZones() []string {
	return cldrtimezone.SupportedTimeZones()
}

// SupportedUnits returns the sanctioned unit identifiers for
// Intl.supportedValuesOf("unit").
//
// The returned slice is sorted, canonical, and independent.
func SupportedUnits() []string {
	return ecma402.SanctionedSimpleUnitIdentifiers()
}
