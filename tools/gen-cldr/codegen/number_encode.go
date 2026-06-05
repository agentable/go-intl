package codegen

import (
	"bytes"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// numberSymbolFieldOrder fixes the wire order of the twelve NumberSymbols
// fields. The encoder writes each field as a StringRef in this order and the
// decoder reads them back in the same order, so the slice is the single source
// of truth for both sides of the mirror.
var numberSymbolFieldOrder = []func(cldr.NumberSymbols) string{
	func(s cldr.NumberSymbols) string { return s.Decimal },
	func(s cldr.NumberSymbols) string { return s.Group },
	func(s cldr.NumberSymbols) string { return s.Percent },
	func(s cldr.NumberSymbols) string { return s.Plus },
	func(s cldr.NumberSymbols) string { return s.Minus },
	func(s cldr.NumberSymbols) string { return s.NaN },
	func(s cldr.NumberSymbols) string { return s.Infinity },
	func(s cldr.NumberSymbols) string { return s.ApproxSign },
	func(s cldr.NumberSymbols) string { return s.PerMille },
	func(s cldr.NumberSymbols) string { return s.Exponential },
	func(s cldr.NumberSymbols) string { return s.SuperscriptingExponent },
	func(s cldr.NumberSymbols) string { return s.TimeSeparator },
}

// encodeNumbers renders the const-only payload for the number domain. It emits a
// private _data string table plus two independent blobs, each prefixed by its
// record count:
//
//   - _numberBlob:          per-locale number-formatting data. Locale indices are
//     written as a sorted delta stream; each locale carries its default
//     numbering system StringRef followed by the symbols, decimal, percent,
//     scientific, currency, and compact sub-blocks. The decoder rebuilds the
//     same map[Locale]numberData the legacy loadNumbersByLocale produced.
//   - _numberSupportedBlob: the locales with number data, in sorted-locale
//     order. A narrow index so SupportedLocales never decodes the main blob.
//
// The two blobs ride independent sync.Once gates in decode.go, so a
// SupportedLocales call never forces the (much larger) main decode.
func encodeNumbers(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.Numbers
	localeIndex := localeIndexMap(input.Locales)

	var main blobEncoder
	locales := sortedLocaleKeys(data)
	main.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		main.appendDelta(mustLocaleIndex(localeIndex, locale))
		encodeNumberLocale(&main, data[locale], table)
	}

	var supported blobEncoder
	supportedTags := numberSupportedLocaleTags(data)
	supported.appendUvarint(uint64(len(supportedTags)))
	for _, tag := range supportedTags {
		supported.appendStringRef(table.Add(tag))
	}

	var numberingSystems blobEncoder
	nsExtras := numberingSystemExtras(data)
	numberingSystems.appendUvarint(uint64(len(nsExtras)))
	for _, ns := range nsExtras {
		numberingSystems.appendStringRef(table.Add(ns))
	}

	var b bytes.Buffer
	b.WriteString("package number\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range []struct {
		name  string
		bytes []byte
	}{
		{"_numberBlob", main.bytes()},
		{"_numberSupportedBlob", supported.bytes()},
		{"_numberingSystemBlob", numberingSystems.bytes()},
	} {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}

// encodeNumberLocale serializes one locale's numberData in the fixed order the
// decoder reads it: default numbering system, symbols-by-NS, decimal/percent/
// scientific pattern maps, currency style map, and compact pattern tree.
func encodeNumberLocale(e *blobEncoder, n cldr.Numbers, table *StringTable) {
	e.appendStringRef(table.Add(n.DefaultNumberingSystem))
	encodeNumberSymbols(e, n.Symbols, table)
	encodeStringMap(e, n.DecimalPatterns, table)
	encodeStringMap(e, n.PercentPatterns, table)
	encodeStringMap(e, n.ScientificPatterns, table)
	encodeCurrencyPatterns(e, n.CurrencyPatterns, table)
	encodeCompactPatterns(e, n.CompactPatterns, table)
}

// encodeNumberSymbols serializes the per-numbering-system symbols map. Each
// numbering system key is followed by its twelve symbol StringRefs in
// numberSymbolFieldOrder.
func encodeNumberSymbols(e *blobEncoder, symbols map[string]cldr.NumberSymbols, table *StringTable) {
	keys := slices.Sorted(maps.Keys(symbols))
	e.appendUvarint(uint64(len(keys)))
	for _, ns := range keys {
		e.appendStringRef(table.Add(ns))
		s := symbols[ns]
		for _, field := range numberSymbolFieldOrder {
			e.appendStringRef(table.Add(field(s)))
		}
	}
}

// encodeCurrencyPatterns serializes the ns -> sign -> pattern map as a sorted
// numbering-system key map whose values are sorted sign string-maps.
func encodeCurrencyPatterns(e *blobEncoder, values map[string]map[string]string, table *StringTable) {
	keys := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(keys)))
	for _, ns := range keys {
		e.appendStringRef(table.Add(ns))
		encodeStringMap(e, values[ns], table)
	}
}

// encodeCompactPatterns serializes the ns -> display -> exp -> plural -> pattern
// tree. Numbering-system and display keys are sorted string-maps; the exp level
// is a sorted-uvarint key map (exponents are small non-negative ints) whose
// values are sorted plural string-maps.
func encodeCompactPatterns(e *blobEncoder, values map[string]map[string]map[int]map[string]string, table *StringTable) {
	nsKeys := slices.Sorted(maps.Keys(values))
	e.appendUvarint(uint64(len(nsKeys)))
	for _, ns := range nsKeys {
		e.appendStringRef(table.Add(ns))
		displays := values[ns]
		displayKeys := slices.Sorted(maps.Keys(displays))
		e.appendUvarint(uint64(len(displayKeys)))
		for _, display := range displayKeys {
			e.appendStringRef(table.Add(display))
			exps := displays[display]
			expKeys := slices.Sorted(maps.Keys(exps))
			e.appendUvarint(uint64(len(expKeys)))
			for _, exp := range expKeys {
				e.appendUvarint(uint64(exp))
				encodeStringMap(e, exps[exp], table)
			}
		}
	}
}

// numberSupportedLocaleTags returns the locales with number data in
// sorted-locale order. The kernel localeRecords table is sorted by tag, so the
// legacy supportedLocaleTags (which scans localeRecords in index order and keeps
// tags with data) yields exactly the sorted key set of the data map.
func numberSupportedLocaleTags(data extract.Numbers) []string {
	return sortedLocaleKeys(data)
}

// numberingSystemExtras returns the numbering systems that appear in the
// generated number data (each locale's default numbering system plus any
// numbering system carrying symbols), in sorted order. The number domain's
// SupportedNumberingSystems accessor merges these with the ECMA-402 simple
// numbering-system set, exactly as the legacy root supported path did.
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
