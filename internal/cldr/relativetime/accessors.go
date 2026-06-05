// Hand-written accessor layer for the relativetime domain. The query semantics
// mirror the legacy root cldr relative-time accessors exactly, so
// RelativeTimeFormat output is byte-for-byte unchanged.

package relativetime

import cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"

// ResolveLocale resolves a tag to its kernel locale handle, forwarding to the
// shared locale kernel so relativetime handles index identically to every other
// domain.
func ResolveLocale(tag string) (Locale, bool) {
	return cldrlocale.ResolveLocale(tag)
}

// FieldsFor returns the relative-time field data for the locale, or nil when the
// locale carries none. The returned maps mirror the legacy
// loc.RelativeTimeFields() value exactly.
func FieldsFor(loc Locale) RelativeTimeFields {
	fieldOnce.Do(loadFields)
	return fieldsByLocale[loc]
}

// SupportedLocales returns the relative-time-supported locale tags in
// sorted-locale order. It reads only the narrow supported blob and never
// triggers the field blob decode.
func SupportedLocales() []string {
	supportedOnce.Do(loadSupported)
	tags := make([]string, len(supportedTags))
	copy(tags, supportedTags)
	return tags
}
