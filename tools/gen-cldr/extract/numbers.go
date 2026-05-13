package extract

import "github.com/agentable/go-intl/tools/gen-cldr/cldr"

type Numbers = map[string]cldr.Numbers

type CurrencyData struct {
	Fractions  map[string]cldr.CurrencyFraction
	Currencies map[string]cldr.Currencies
}

var generatedCurrencies = map[string]bool{"DEFAULT": true, "JPY": true, "USD": true}

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
	filteredFractions := make(map[string]cldr.CurrencyFraction, len(generatedCurrencies))
	for code, fraction := range fractions {
		if generatedCurrencies[code] {
			filteredFractions[code] = fraction
		}
	}
	filteredCurrencies := make(map[string]cldr.Currencies, len(selected))
	for locale, byCurrency := range currencies {
		if !selected[locale] {
			continue
		}
		filtered := make(cldr.Currencies, len(generatedCurrencies))
		for code, names := range byCurrency {
			if generatedCurrencies[code] {
				filtered[code] = names
			}
		}
		filteredCurrencies[locale] = filtered
	}
	return CurrencyData{Fractions: filteredFractions, Currencies: filteredCurrencies}
}

func localeSet(locales []string) map[string]bool {
	out := make(map[string]bool, len(locales))
	for _, locale := range locales {
		out[locale] = true
	}
	return out
}
