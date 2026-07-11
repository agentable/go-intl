// Hand-written accessor layer for the currency domain. It exposes fraction
// metadata, localized names, and the narrow supported-code index over lazily
// decoded const blobs.

package currency

// Digits returns the fraction-digit metadata for a currency code, falling back
// to the "DEFAULT" entry when the code is unknown.
func Digits(code string) Data {
	fractionOnce.Do(loadFractions)
	if data, ok := fractions[code]; ok {
		return data
	}
	return fractions["DEFAULT"]
}

// DisplayName returns the localized display name for a (code, plural) pair,
// falling back to the "other" plural form. An empty plural is treated as
// "other".
func DisplayName(loc Locale, code, plural string) string {
	if plural == "" {
		plural = "other"
	}
	names := localeNames(loc, code)
	if name := names.display[plural]; name != "" {
		return name
	}
	return names.display["other"]
}

// CanonicalName returns the canonical currency name for a code, falling back to
// the "other" then "one" display forms.
func CanonicalName(loc Locale, code string) string {
	names := localeNames(loc, code)
	if names.canonical != "" {
		return names.canonical
	}
	if name := names.display["other"]; name != "" {
		return name
	}
	return names.display["one"]
}

// Symbol returns the standard localized currency symbol for a code.
func Symbol(loc Locale, code string) string {
	return localeNames(loc, code).symbol
}

// NarrowSymbol returns the narrow currency symbol for a code, falling back to the
// standard symbol when CLDR has no narrow form.
func NarrowSymbol(loc Locale, code string) string {
	names := localeNames(loc, code)
	if names.narrow != "" {
		return names.narrow
	}
	return names.symbol
}

// SupportedCodes returns the supported ISO 4217 currency codes in sorted order.
// It reads only the narrow supported blob and never triggers the fraction or
// names blob decode.
func SupportedCodes() []string {
	return supported.Get()
}

// localeNames resolves the names record for a (locale, code) pair, gating the
// names blob decode. A missing locale or code yields the zero record.
func localeNames(loc Locale, code string) currencyNames {
	namesOnce.Do(loadNames)
	codes := byLocale[loc]
	if codes == nil {
		return currencyNames{}
	}
	names, ok := codes[code]
	if !ok {
		return currencyNames{}
	}
	return names
}
