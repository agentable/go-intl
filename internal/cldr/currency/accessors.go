// Hand-written accessor layer for the currency domain. The query semantics
// mirror the legacy root cldr currency accessors exactly, so NumberFormat and
// DisplayNames output is byte-for-byte unchanged.

package currency

import cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"

// ResolveLocale resolves a tag to its kernel locale handle, forwarding to the
// shared locale kernel so currency handles index identically to every other
// domain.
func ResolveLocale(tag string) (Locale, bool) {
	return cldrlocale.ResolveLocale(tag)
}

// Digits returns the fraction-digit metadata for a currency code, falling back
// to the "DEFAULT" entry when the code is unknown. It mirrors the legacy
// CurrencyDigits accessor.
func Digits(code string) Data {
	fractionOnce.Do(loadFractions)
	if data, ok := fractions[code]; ok {
		return data
	}
	return fractions["DEFAULT"]
}

// DisplayName returns the localized display name for a (code, plural) pair,
// falling back to the "other" plural form. An empty plural is treated as
// "other". It mirrors the legacy Locale.CurrencyDisplayName accessor.
func DisplayName(loc Locale, code, plural string) string {
	if plural == "" {
		plural = "other"
	}
	data := localeNames(loc, code)
	if name := data.display[plural]; name != "" {
		return name
	}
	return data.display["other"]
}

// CanonicalName returns the canonical currency name for a code, falling back to
// the "other" then "one" display forms. It mirrors the legacy
// Locale.CurrencyCanonicalName accessor.
func CanonicalName(loc Locale, code string) string {
	data := localeNames(loc, code)
	if data.canonical != "" {
		return data.canonical
	}
	if name := data.display["other"]; name != "" {
		return name
	}
	return data.display["one"]
}

// Symbol returns the currency symbol for a code, preferring the narrow form
// when width == "narrow" and a narrow symbol exists. It mirrors the legacy
// Locale.CurrencySymbol accessor.
func Symbol(loc Locale, code, width string) string {
	data := localeNames(loc, code)
	if width == "narrow" && data.narrow != "" {
		return data.narrow
	}
	return data.symbol
}

// SupportedCodes returns the supported ISO 4217 currency codes in sorted order.
// It reads only the narrow supported blob and never triggers the fraction or
// names blob decode.
func SupportedCodes() []string {
	supportedOnce.Do(loadSupported)
	codes := make([]string, len(supportedTags))
	copy(codes, supportedTags)
	return codes
}

// localeNames resolves the names record for a (locale, code) pair, gating the
// names blob decode. A missing locale or code yields the zero record, whose
// accessors degrade to "" exactly as the legacy map indexing did.
func localeNames(loc Locale, code string) names {
	namesOnce.Do(loadNames)
	return byLocale[loc][code]
}
