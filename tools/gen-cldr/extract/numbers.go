package extract

import (
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

type Numbers = map[string]cldr.Numbers

type CurrencyData struct {
	Fractions  map[string]cldr.CurrencyFraction
	Currencies map[string]cldr.Currencies
}

// displayCurrencyAllowlist limits which currencies retain locale-specific
// display names and symbols. Fractions stay broad because supportedValuesOf
// requires the canonical ISO 4217 universe; display payload is the
// bandwidth-heavy part and remains profile-scoped.
//
// The set covers the major world currencies typically encountered by
// MessageFormat 2.0 / Intl.NumberFormat consumers. Adding a code here costs
// roughly the size of its locale-specific names across the profile (a few
// hundred bytes per locale per currency); we err on the side of inclusion
// because users formatting :currency expect a symbol or full name, not a
// fallback to the ISO code.
var displayCurrencyAllowlist = [...]string{
	"AUD",
	"BRL",
	"CAD",
	"CHF",
	"CNY",
	"CZK",
	"DKK",
	"EUR",
	"GBP",
	"HKD",
	"HUF",
	"IDR",
	"ILS",
	"INR",
	"JPY",
	"KRW",
	"MXN",
	"MYR",
	"NOK",
	"NZD",
	"PHP",
	"PLN",
	"RUB",
	"SEK",
	"SGD",
	"THB",
	"TRY",
	"TWD",
	"USD",
	"VND",
	"ZAR",
}

func ExtractNumbers(raw map[string]cldr.Numbers, locales []string) Numbers {
	selected := localeSet(locales)
	out := make(Numbers, len(selected))
	for locale, numbers := range raw {
		if selected[locale] {
			out[locale] = numbers
		}
	}
	return out
}

func ExtractCurrencies(fractions map[string]cldr.CurrencyFraction, currencies map[string]cldr.Currencies, locales []string) CurrencyData {
	selected := localeSet(locales)
	filteredFractions := maps.Clone(fractions)
	filteredCurrencies := make(map[string]cldr.Currencies, len(selected))
	for locale, byCurrency := range currencies {
		if !selected[locale] {
			continue
		}
		filtered := make(cldr.Currencies, len(displayCurrencyAllowlist))
		for code, names := range byCurrency {
			if isDisplayCurrencyAllowed(code) {
				filtered[code] = names
			}
		}
		filteredCurrencies[locale] = filtered
	}
	return CurrencyData{Fractions: filteredFractions, Currencies: filteredCurrencies}
}

func isDisplayCurrencyAllowed(code string) bool {
	_, ok := slices.BinarySearch(displayCurrencyAllowlist[:], code)
	return ok
}

func localeSet(locales []string) map[string]bool {
	out := make(map[string]bool, len(locales))
	for _, locale := range locales {
		out[locale] = true
	}
	return out
}
