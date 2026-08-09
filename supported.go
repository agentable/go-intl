package gointl

import (
	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/tz"
)

// SupportedCalendars returns the calendar identifiers for
// Intl.supportedValuesOf("calendar").
//
// The returned slice is sorted, canonical, and independent.
func SupportedCalendars() []string {
	return cldrdate.SupportedCalendars()
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
	return tz.SupportedTimeZones()
}

// SupportedUnits returns the sanctioned unit identifiers for
// Intl.supportedValuesOf("unit").
//
// The returned slice is sorted, canonical, and independent.
func SupportedUnits() []string {
	return ecma402.SanctionedSimpleUnitIdentifiers()
}
