package codegen

import (
	"bytes"
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// encodeCurrencies renders the const-only payload for the currency domain. It
// emits a private _data string table plus three independent blobs, each
// prefixed by its record count:
//
//   - _currencyFractionBlob:  the locale-independent fraction table, sorted by
//     ISO code. Each record is a code StringRef followed by three uvarints
//     (default digits, cash digits, rounding). Used by CurrencyDigits.
//   - _currencyNamesBlob:     per-locale display/canonical/symbol/narrow data.
//     Locale indices are written as a sorted delta stream; each locale carries a
//     sorted code map whose values are a display string-map plus the three
//     scalar StringRefs. The decoder rebuilds the same
//     map[Locale]map[string]currencyNames the legacy loadCurrencyData produced.
//   - _currencySupportedBlob: the supported ISO 4217 codes in sorted order, a
//     narrow index so SupportedCodes never decodes the fraction or names blobs.
//
// The fraction and names blobs ride independent sync.Once gates in decode.go,
// so a CurrencyDigits call never forces the (much larger) names decode and vice
// versa.
func encodeCurrencies(input RuntimeInput, table *StringTable) ([]byte, error) {
	data := input.Currencies
	localeIndex := localeIndexMap(input.Locales)

	var fractions blobEncoder
	fractionCodes := slices.Sorted(maps.Keys(data.Fractions))
	fractions.appendUvarint(uint64(len(fractionCodes)))
	for _, code := range fractionCodes {
		f := data.Fractions[code]
		fractions.appendStringRef(table.Add(code))
		fractions.appendUvarint(uint64(f.Digits))
		fractions.appendUvarint(uint64(f.CashDigits))
		fractions.appendUvarint(uint64(f.Rounding))
	}

	var names blobEncoder
	locales := sortedLocaleKeys(data.Currencies)
	names.appendUvarint(uint64(len(locales)))
	for _, locale := range locales {
		names.appendDelta(mustLocaleIndex(localeIndex, locale))
		encodeCurrencyNames(&names, data.Currencies[locale], table)
	}

	var supported blobEncoder
	supportedCodes := supportedCurrencyValues(data)
	supported.appendUvarint(uint64(len(supportedCodes)))
	for _, code := range supportedCodes {
		supported.appendStringRef(table.Add(code))
	}

	var b bytes.Buffer
	b.WriteString("package currency\n\n")
	if err := table.EmitDataConst(&b, "_data"); err != nil {
		return nil, err
	}
	for _, blob := range []struct {
		name  string
		bytes []byte
	}{
		{"_currencyFractionBlob", fractions.bytes()},
		{"_currencyNamesBlob", names.bytes()},
		{"_currencySupportedBlob", supported.bytes()},
	} {
		b.WriteString("\n")
		if err := emitStringConst(&b, blob.name, string(blob.bytes)); err != nil {
			return nil, err
		}
	}
	return FormatFile(b.Bytes())
}

// supportedCurrencyValues returns the supported ISO 4217 currency codes in
// sorted order: every fraction-table code (excluding the DEFAULT sentinel) plus
// every code that carries locale names.
func supportedCurrencyValues(data extract.CurrencyData) []string {
	seen := map[string]bool{}
	for code := range data.Fractions {
		if code != "DEFAULT" {
			seen[code] = true
		}
	}
	for _, currencies := range data.Currencies {
		for code := range currencies {
			seen[code] = true
		}
	}
	return slices.Sorted(maps.Keys(seen))
}

// encodeCurrencyNames serializes one locale's code -> currencyNames map: a
// sorted code map whose values are a display string-map plus the canonical,
// symbol, and narrow scalars. The decoder reads the names back in this exact
// order.
func encodeCurrencyNames(e *blobEncoder, currencies cldr.Currencies, table *StringTable) {
	codes := slices.Sorted(maps.Keys(currencies))
	e.appendUvarint(uint64(len(codes)))
	for _, code := range codes {
		names := currencies[code]
		e.appendStringRef(table.Add(code))
		encodeStringMap(e, names.Display, table)
		e.appendStringRef(table.Add(names.Canonical))
		e.appendStringRef(table.Add(names.Symbol))
		e.appendStringRef(table.Add(names.Narrow))
	}
}
