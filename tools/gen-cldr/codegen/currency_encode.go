package codegen

import (
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
//     scalar StringRefs. The decoder rebuilds the runtime
//     map[Locale]map[string]currencyNames.
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
	appendStringRefKeyMap(&fractions, data.Fractions, table, func(f cldr.CurrencyFraction) {
		appendCurrencyFraction(&fractions, f)
	})

	var names blobEncoder
	locales := sortedLocaleKeys(data.Currencies)
	if err := names.appendLocaleDeltaRecords(locales, localeIndex, func(locale string) {
		encodeCurrencyNames(&names, data.Currencies[locale], table)
	}); err != nil {
		return nil, err
	}

	var supported blobEncoder
	supportedCodes := supportedCurrencyValues(data)
	supported.appendStringRefSlice(supportedCodes, table)

	return renderPayloadFile("currency", table,
		payloadBlob{"_currencyFractionBlob", fractions.bytes()},
		payloadBlob{"_currencyNamesBlob", names.bytes()},
		payloadBlob{"_currencySupportedBlob", supported.bytes()},
	)
}

// appendCurrencyFraction owns the digits/cashDigits/rounding wire order for one
// currency fraction row.
func appendCurrencyFraction(e *blobEncoder, f cldr.CurrencyFraction) {
	e.appendUvarint(uint64(f.Digits))
	e.appendUvarint(uint64(f.CashDigits))
	e.appendUvarint(uint64(f.Rounding))
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
	appendStringRefKeyMap(e, currencies, table, func(names cldr.CurrencyNames) {
		appendCurrencyNames(e, names, table)
	})
}

// appendCurrencyNames owns the display/canonical/symbol/narrow wire order for
// one localized currency-name row.
func appendCurrencyNames(e *blobEncoder, names cldr.CurrencyNames, table *StringTable) {
	e.appendStringRefMap(names.Display, table)
	e.appendStringRef(table.Add(names.Canonical))
	e.appendStringRef(table.Add(names.Symbol))
	e.appendStringRef(table.Add(names.Narrow))
}
