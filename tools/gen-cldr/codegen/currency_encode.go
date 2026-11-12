package codegen

import (
	"maps"
	"slices"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/extract"
)

// encodeCurrencies renders the const-only payload for the currency domain. It
// emits a private _data string table plus four independent blobs, each
// prefixed by its record count:
//
//   - _currencyFractionBlob:  the locale-independent fraction table, sorted by
//     ISO code. Each record is a code StringRef followed by one default-digits
//     uvarint. Used by CurrencyDigits.
//   - _currencyNamesBlob:     per-locale display/canonical/symbol/narrow data.
//     Locale indices are written as a sorted delta stream; each locale carries a
//     sorted code map whose values are a display string-map plus the three
//     scalar StringRefs. The decoder rebuilds the runtime
//     map[Locale]map[string]currencyNames.
//   - _currencySupportedBlob: the supported ISO 4217 codes in sorted order, a
//     narrow index so SupportedCodes never decodes the fraction or names blobs.
//   - _currencyFormatBlob: sparse per-locale currency-specific pattern and
//     separator overrides, decoded independently of names and fractions.
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

	formats := make(map[string]cldr.Currencies)
	for locale, currencies := range data.Currencies {
		for code, names := range currencies {
			if names.Pattern == "" && names.Decimal == "" && names.Group == "" {
				continue
			}
			if formats[locale] == nil {
				formats[locale] = make(cldr.Currencies)
			}
			formats[locale][code] = names
		}
	}
	var overrides blobEncoder
	if err := overrides.appendLocaleDeltaRecords(sortedLocaleKeys(formats), localeIndex, func(locale string) {
		appendStringRefKeyMap(&overrides, formats[locale], table, func(names cldr.CurrencyNames) {
			overrides.appendStringRef(table.Add(names.Pattern))
			overrides.appendStringRef(table.Add(names.Decimal))
			overrides.appendStringRef(table.Add(names.Group))
		})
	}); err != nil {
		return nil, err
	}

	return renderPayloadFile("currency", table,
		payloadBlob{"_currencyFractionBlob", fractions.bytes()},
		payloadBlob{"_currencyNamesBlob", names.bytes()},
		payloadBlob{"_currencySupportedBlob", supported.bytes()},
		payloadBlob{"_currencyFormatBlob", overrides.bytes()},
	)
}

// appendCurrencyFraction owns the default fraction-digit wire encoding for one
// currency fraction row.
func appendCurrencyFraction(e *blobEncoder, f cldr.CurrencyFraction) {
	e.appendUvarint(uint64(f.Digits))
}

// supportedCurrencyValues lists codes with a name in the selected CLDR profile.
// Fraction exceptions affect precision, not supported-values membership.
func supportedCurrencyValues(data extract.CurrencyData) []string {
	seen := map[string]bool{}
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
