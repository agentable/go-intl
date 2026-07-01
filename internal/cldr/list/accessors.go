// Hand-written accessor layer for the list domain. It exposes list patterns and
// the narrow supported-locale index over lazily decoded const blobs.

package list

import (
	"slices"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

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
// type defaults to "conjunction" and an empty style to "long". An unknown tuple
// yields a zero ListPattern (empty fields).
func Pattern(locale Locale, typ, style string) ListPattern {
	if typ == "" {
		typ = "conjunction"
	}
	if style == "" {
		style = "long"
	}
	patternOnce.Do(loadPatterns)
	var data listPatternRefs
	if byType := patternsByLocale[locale]; byType != nil {
		if byStyle := byType[typ]; byStyle != nil {
			data = byStyle[style]
		}
	}
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
	return slices.Clone(supportedTags)
}
