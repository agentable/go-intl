// Hand-written accessor layer for the list domain. The query semantics mirror
// the legacy root cldr list accessors exactly, so ListFormat output is
// byte-for-byte unchanged.

package list

import cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"

// Maximize forwards to the shared locale kernel so list maximizes tags
// identically to every other domain.
func Maximize(tag string) string {
	return cldrlocale.Maximize(tag)
}

// ResolveLocale resolves a tag to its kernel locale handle, forwarding to the
// shared locale kernel so list handles index identically to every other domain.
func ResolveLocale(tag string) (Locale, bool) {
	return cldrlocale.ResolveLocale(tag)
}

// Pattern returns the list pattern for a (locale, type, style) tuple. An empty
// type defaults to "conjunction" and an empty style to "long", matching the
// legacy accessor. An unknown tuple yields a zero ListPattern (empty fields),
// the same value the legacy map indexing produced.
func Pattern(locale Locale, typ, style string) ListPattern {
	if typ == "" {
		typ = "conjunction"
	}
	if style == "" {
		style = "long"
	}
	patternOnce.Do(loadPatterns)
	data := patternsByLocale[locale][typ][style]
	return ListPattern{
		Pair:   data.pair,
		Start:  data.start,
		Middle: data.middle,
		End:    data.end,
	}
}

// SupportedLocales returns the list-supported locale tags in sorted-locale
// order. It reads only the narrow supported blob and never triggers the pattern
// blob decode.
func SupportedLocales() []string {
	supportedOnce.Do(loadSupported)
	tags := make([]string, len(supportedTags))
	copy(tags, supportedTags)
	return tags
}
