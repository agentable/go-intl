package codegen

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

func renderNumbers(numbers extract.Numbers, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("type NumberSymbols struct { Decimal, Group, Percent, Plus, Minus, NaN, Infinity, ApproxSign, PerMille, Exponential, SuperscriptingExponent string }\n\n")
	b.WriteString("type numberData struct { defaultNS string; symbols map[string]NumberSymbols; decimal, percent, scientific map[string]string; currency map[string]map[string]string; compact map[string]map[string]map[int]map[string]string }\n\n")
	b.WriteString("var numbersByLocale = map[Locale]numberData{\n")
	for _, locale := range sortedLocaleKeys(numbers) {
		n := numbers[locale]
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: {defaultNS: ")
		b.WriteString(strconv.Quote(n.DefaultNumberingSystem))
		b.WriteString(", symbols: map[string]NumberSymbols{\n")
		for _, ns := range slices.Sorted(maps.Keys(n.Symbols)) {
			b.WriteString("\t\t")
			b.WriteString(strconv.Quote(ns))
			b.WriteString(": ")
			b.WriteString(numberSymbolsLiteral(n.Symbols[ns], table))
			b.WriteString(",\n")
		}
		b.WriteString("\t}, decimal: ")
		b.WriteString(stringMapLiteral(n.DecimalPatterns, table))
		b.WriteString(", percent: ")
		b.WriteString(stringMapLiteral(n.PercentPatterns, table))
		b.WriteString(", scientific: ")
		b.WriteString(stringMapLiteral(n.ScientificPatterns, table))
		b.WriteString(", currency: ")
		b.WriteString(patternStyleMapLiteral(n.CurrencyPatterns, table))
		b.WriteString(", compact: ")
		b.WriteString(compactPatternsLiteral(n.CompactPatterns, table))
		b.WriteString("},\n")
	}
	b.WriteString("}\n\n")
	b.WriteString(`func (l Locale) NumberSymbols(numberingSystem string) NumberSymbols {
	data := numbersByLocale[l]
	if numberingSystem == "" { numberingSystem = data.defaultNS }
	return data.symbols[numberingSystem]
}

func (l Locale) DecimalPattern(numberingSystem string) string {
	data := numbersByLocale[l]
	if numberingSystem == "" { numberingSystem = data.defaultNS }
	return data.decimal[numberingSystem]
}

func (l Locale) PercentPattern(numberingSystem string) string {
	data := numbersByLocale[l]
	if numberingSystem == "" { numberingSystem = data.defaultNS }
	return data.percent[numberingSystem]
}

func (l Locale) CurrencyPattern(numberingSystem, sign string) string {
	data := numbersByLocale[l]
	if numberingSystem == "" { numberingSystem = data.defaultNS }
	patterns := data.currency[numberingSystem]
	if pattern := patterns[sign]; pattern != "" { return pattern }
	return patterns["standard"]
}

func (l Locale) ScientificPattern(numberingSystem string) string {
	data := numbersByLocale[l]
	if numberingSystem == "" { numberingSystem = data.defaultNS }
	return data.scientific[numberingSystem]
}

func (l Locale) CompactPattern(numberingSystem, display string, exp int, plural string) string {
	data := numbersByLocale[l]
	if numberingSystem == "" { numberingSystem = data.defaultNS }
	patterns := data.compact[numberingSystem][display][exp]
	if pattern := patterns[plural]; pattern != "" { return pattern }
	return patterns["other"]
}

func (l Locale) DefaultNumberingSystem() string { return numbersByLocale[l].defaultNS }
`)
	return FormatFile([]byte(b.String()))
}

func renderCurrencies(data extract.CurrencyData, table *StringTable) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package cldr\n\n")
	b.WriteString("type CurrencyData struct{ DefaultDigits, CashDigits, Rounding int }\n\n")
	b.WriteString("type currencyNames struct{ display map[string]string; symbol, narrow string }\n\n")
	b.WriteString("var currencyFractions = map[string]CurrencyData{\n")
	for _, code := range slices.Sorted(maps.Keys(data.Fractions)) {
		f := data.Fractions[code]
		fmt.Fprintf(&b, "\t%q: {DefaultDigits: %d, CashDigits: %d, Rounding: %d},\n", code, f.Digits, f.CashDigits, f.Rounding)
	}
	b.WriteString("}\n\n")
	b.WriteString("var currenciesByLocale = map[Locale]map[string]currencyNames{\n")
	for _, locale := range sortedLocaleKeys(data.Currencies) {
		b.WriteString("\tlocaleIndex[")
		b.WriteString(strconv.Quote(locale))
		b.WriteString("]: {\n")
		for _, code := range slices.Sorted(maps.Keys(data.Currencies[locale])) {
			names := data.Currencies[locale][code]
			b.WriteString("\t\t")
			b.WriteString(strconv.Quote(code))
			b.WriteString(": {display: ")
			b.WriteString(stringMapLiteral(names.Display, table))
			b.WriteString(", symbol: ")
			b.WriteString(refStringLiteral(names.Symbol, table))
			b.WriteString(", narrow: ")
			b.WriteString(refStringLiteral(names.Narrow, table))
			b.WriteString("},\n")
		}
		b.WriteString("\t},\n")
	}
	b.WriteString("}\n\n")
	b.WriteString(`func CurrencyDigits(code string) CurrencyData {
	if data, ok := currencyFractions[code]; ok { return data }
	return currencyFractions["DEFAULT"]
}

func (l Locale) CurrencyDisplayName(code, plural string) string {
	if plural == "" { plural = "other" }
	data := currenciesByLocale[l][code]
	if name := data.display[plural]; name != "" { return name }
	return data.display["other"]
}

func (l Locale) CurrencySymbol(code, width string) string {
	data := currenciesByLocale[l][code]
	if width == "narrow" && data.narrow != "" { return data.narrow }
	return data.symbol
}
`)
	return FormatFile([]byte(b.String()))
}

func numberSymbolsLiteral(s cldr.NumberSymbols, table *StringTable) string {
	return "NumberSymbols{" +
		"Decimal: " + refStringLiteral(s.Decimal, table) + ", " +
		"Group: " + refStringLiteral(s.Group, table) + ", " +
		"Percent: " + refStringLiteral(s.Percent, table) + ", " +
		"Plus: " + refStringLiteral(s.Plus, table) + ", " +
		"Minus: " + refStringLiteral(s.Minus, table) + ", " +
		"NaN: " + refStringLiteral(s.NaN, table) + ", " +
		"Infinity: " + refStringLiteral(s.Infinity, table) + ", " +
		"ApproxSign: " + refStringLiteral(s.ApproxSign, table) + ", " +
		"PerMille: " + refStringLiteral(s.PerMille, table) + ", " +
		"Exponential: " + refStringLiteral(s.Exponential, table) + ", " +
		"SuperscriptingExponent: " + refStringLiteral(s.SuperscriptingExponent, table) + "}"
}

func stringMapLiteral(values map[string]string, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]string", values, func(value string) string {
		return refStringLiteral(value, table)
	})
}

func patternStyleMapLiteral(values map[string]map[string]string, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]map[string]string", values, func(patterns map[string]string) string {
		return stringMapLiteral(patterns, table)
	})
}

func compactPatternsLiteral(values map[string]map[string]map[int]map[string]string, table *StringTable) string {
	return stringKeyedMapLiteral("map[string]map[string]map[int]map[string]string", values, func(displays map[string]map[int]map[string]string) string {
		return stringKeyedMapLiteral("map[string]map[int]map[string]string", displays, func(patterns map[int]map[string]string) string {
			var b strings.Builder
			b.WriteString("{")
			for _, exp := range slices.Sorted(maps.Keys(patterns)) {
				b.WriteString(strconv.Itoa(exp))
				b.WriteString(": ")
				b.WriteString(stringMapLiteral(patterns[exp], table))
				b.WriteString(", ")
			}
			b.WriteString("}")
			return b.String()
		})
	})
}

func refStringLiteral(value string, table *StringTable) string {
	if value == "" {
		return strconv.Quote("")
	}
	return table.Add(value).GoLiteral() + ".string()"
}

func sortedLocaleKeys[T any](values map[string]T) []string {
	return slices.Sorted(maps.Keys(values))
}
