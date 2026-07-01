package localematcher

import (
	"strings"

	"github.com/agentable/go-intl/internal/localeid"
)

type availableLocale struct {
	locale     string
	dataLocale string
	derived    bool
}

func availableLocalesFor(supported []string) []availableLocale {
	backed := make(map[string]struct{}, len(supported))
	for _, loc := range supported {
		noExtensionLocale, _ := removeUnicodeExtension(loc)
		backed[noExtensionLocale] = struct{}{}
	}

	seen := make(map[string]struct{}, len(supported))
	out := make([]availableLocale, 0, len(supported))
	for _, loc := range supported {
		noExtensionLocale, _ := removeUnicodeExtension(loc)
		out = appendAvailableLocale(out, seen, noExtensionLocale, noExtensionLocale, false)
		if alias, ok := languageRegionAlias(noExtensionLocale); ok {
			if _, backedAlias := backed[alias]; !backedAlias {
				out = appendAvailableLocale(out, seen, alias, noExtensionLocale, true)
			}
			out = appendFallbackLocales(out, seen, backed, alias, noExtensionLocale)
		}
		out = appendFallbackLocales(out, seen, backed, noExtensionLocale, noExtensionLocale)
	}
	return out
}

func appendFallbackLocales(out []availableLocale, seen, backed map[string]struct{}, loc, dataLocale string) []availableLocale {
	for {
		pos := truncationPosition(loc)
		if pos < 0 {
			return out
		}
		loc = loc[:pos]
		if _, backedLocale := backed[loc]; backedLocale {
			continue
		}
		out = appendAvailableLocale(out, seen, loc, dataLocale, true)
	}
}

func appendAvailableLocale(out []availableLocale, seen map[string]struct{}, loc, dataLocale string, derived bool) []availableLocale {
	if loc == "" {
		return out
	}
	if _, ok := seen[loc]; ok {
		return out
	}
	seen[loc] = struct{}{}
	return append(out, availableLocale{locale: loc, dataLocale: dataLocale, derived: derived})
}

func languageRegionAlias(loc string) (string, bool) {
	language, rest, ok := strings.Cut(loc, "-")
	if !ok {
		return "", false
	}
	script, rest, ok := strings.Cut(rest, "-")
	if !ok || !localeid.IsUnicodeScriptSubtag(script) {
		return "", false
	}
	region, suffix, hasSuffix := strings.Cut(rest, "-")
	if !localeid.IsUnicodeRegionSubtag(region) {
		return "", false
	}
	alias := language + "-" + region
	if hasSuffix {
		alias += "-" + suffix
	}
	return alias, true
}
