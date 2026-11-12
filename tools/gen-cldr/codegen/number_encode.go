package codegen

import (
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// numberSymbolFieldOrder fixes the wire order of the NumberSymbols
// fields. The encoder writes each field as a StringRef in this order and the
// decoder reads them back in the same order, so the table is the single source
// of truth for both sides of the mirror.
var numberSymbolFieldOrder = [...]func(cldr.NumberSymbols) string{
	func(s cldr.NumberSymbols) string { return s.Decimal },
	func(s cldr.NumberSymbols) string { return s.Group },
	func(s cldr.NumberSymbols) string { return s.Percent },
	func(s cldr.NumberSymbols) string { return s.Plus },
	func(s cldr.NumberSymbols) string { return s.Minus },
	func(s cldr.NumberSymbols) string { return s.NaN },
	func(s cldr.NumberSymbols) string { return s.Infinity },
	func(s cldr.NumberSymbols) string { return s.ApproxSign },
	func(s cldr.NumberSymbols) string { return s.RangeSign },
	func(s cldr.NumberSymbols) string { return s.PerMille },
	func(s cldr.NumberSymbols) string { return s.Exponential },
	func(s cldr.NumberSymbols) string { return s.SuperscriptingExponent },
	func(s cldr.NumberSymbols) string { return s.TimeSeparator },
	func(s cldr.NumberSymbols) string { return s.CurrencyDecimal },
	func(s cldr.NumberSymbols) string { return s.CurrencyGroup },
}

// encodeNumbers emits per-locale number records with a locale-delta index of
// offset/length references. Supported locales and numbering systems keep their
// independent narrow indexes.
func encodeNumbers(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.Numbers
	localeIndex := localeIndexMap(input.Locales)

	var main, index blobEncoder
	locales := sortedLocaleKeys(data)
	if err := index.appendLocaleDeltaRecords(locales, localeIndex, func(locale string) {
		start := len(main.bytes())
		encodeNumberLocale(&main, data[locale], table)
		index.appendStringRef(StringRef{start: uint32(start), length: uint32(len(main.bytes()) - start)})
	}); err != nil {
		return nil, err
	}

	var supported blobEncoder
	supported.appendStringRefSlice(locales, table)

	var numberingSystems blobEncoder
	nsExtras := numberingSystemExtras(data)
	numberingSystems.appendStringRefSlice(nsExtras, table)

	return renderPayloadFile("number", table,
		payloadBlob{"_numberIndexBlob", index.bytes()},
		payloadBlob{"_numberBlob", main.bytes()},
		payloadBlob{"_numberSupportedBlob", supported.bytes()},
		payloadBlob{"_numberingSystemBlob", numberingSystems.bytes()},
	)
}

// encodeNumberLocale serializes one locale's numberData in the fixed order the
// decoder reads it: default numbering system, symbols-by-NS, decimal/percent/
// scientific pattern maps, currency style map, currency spacing text,
// currency-name placement map, and decimal/currency compact pattern trees.
func encodeNumberLocale(e *blobEncoder, n cldr.Numbers, table *StringTable) {
	e.appendStringRef(table.Add(n.DefaultNumberingSystem))
	e.appendUvarint(uint64(n.MinimumGroupingDigits))
	encodeNumberSymbols(e, n.Symbols, table)
	e.appendStringRefMap(n.DecimalPatterns, table)
	e.appendStringRefMap(n.PercentPatterns, table)
	e.appendStringRefMap(n.ScientificPatterns, table)
	encodeCurrencyPatterns(e, n.CurrencyPatterns, table)
	encodeCurrencySpacing(e, n.CurrencySpacing, table)
	encodeCurrencyNamePatterns(e, n.CurrencyNamePatterns, table)
	encodeCompactPatterns(e, n.CompactPatterns, table)
	encodeCompactPatterns(e, n.CurrencyCompactPatterns, table)
}

// encodeNumberSymbols serializes the per-numbering-system symbols map. Each
// numbering system key is followed by its symbol StringRefs in
// numberSymbolFieldOrder.
func encodeNumberSymbols(e *blobEncoder, symbols map[string]cldr.NumberSymbols, table *StringTable) {
	appendStringRefKeyMap(e, symbols, table, func(s cldr.NumberSymbols) {
		for _, field := range numberSymbolFieldOrder {
			e.appendStringRef(table.Add(field(s)))
		}
	})
}

// encodeCurrencyPatterns serializes the ns -> sign -> pattern map as a sorted
// numbering-system key map whose values are sorted sign string-maps.
func encodeCurrencyPatterns(e *blobEncoder, values map[string]map[string]string, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(signPatterns map[string]string) {
		e.appendStringRefMap(signPatterns, table)
	})
}

func encodeCurrencySpacing(e *blobEncoder, values map[string]cldr.CurrencySpacing, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(spacing cldr.CurrencySpacing) {
		e.appendStringRef(table.Add(spacing.BeforeCurrency))
		e.appendStringRef(table.Add(spacing.AfterCurrency))
	})
}

// encodeCurrencyNamePatterns serializes the ns -> plural -> placement pattern
// tree. The same plural category selects the localized name and its position.
func encodeCurrencyNamePatterns(e *blobEncoder, values map[string]map[string]string, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(pluralPatterns map[string]string) {
		e.appendStringRefMap(pluralPatterns, table)
	})
}

// encodeCompactPatterns serializes the ns -> display -> exp -> plural -> pattern
// tree. Numbering-system and display keys are sorted string-maps; the exp level
// is a sorted-uvarint key map (exponents are small non-negative ints) whose
// values are sorted plural string-maps.
func encodeCompactPatterns(e *blobEncoder, values map[string]map[string]map[int]map[string]string, table *StringTable) {
	appendStringRefKeyMap(e, values, table, func(displays map[string]map[int]map[string]string) {
		appendStringRefKeyMap(e, displays, table, func(exps map[int]map[string]string) {
			encodeCompactExponentPatterns(e, exps, table)
		})
	})
}

func encodeCompactExponentPatterns(e *blobEncoder, exps map[int]map[string]string, table *StringTable) {
	expKeys := slices.Sorted(maps.Keys(exps))
	e.appendUvarint(uint64(len(expKeys)))
	for _, exp := range expKeys {
		e.appendUvarint(uint64(exp))
		e.appendStringRefMap(exps[exp], table)
	}
}

// numberingSystemExtras returns the numbering systems that appear in the
// generated number data (each locale's default numbering system plus any
// numbering system carrying symbols), in sorted order. The number domain's
// SupportedNumberingSystems accessor merges these with the ECMA-402 simple
// numbering-system set.
func numberingSystemExtras(numbers extract.Numbers) []string {
	seen := map[string]bool{}
	for _, data := range numbers {
		if data.DefaultNumberingSystem != "" {
			seen[data.DefaultNumberingSystem] = true
		}
		for numberingSystem := range data.Symbols {
			seen[numberingSystem] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}
