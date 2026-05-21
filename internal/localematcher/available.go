package localematcher

import "strings"

type availableLocale struct {
	locale     string
	dataLocale string
	derived    bool
}

func availableLocalesFor(supported []string) []availableLocale {
	raw := make(map[string]struct{}, len(supported))
	for _, loc := range supported {
		noExtensionLocale, _ := removeUnicodeExtension(loc)
		raw[noExtensionLocale] = struct{}{}
	}

	seen := make(map[string]struct{}, len(supported))
	out := make([]availableLocale, 0, len(supported))
	for _, loc := range supported {
		noExtensionLocale, _ := removeUnicodeExtension(loc)
		out = appendAvailableLocale(out, seen, noExtensionLocale, noExtensionLocale, false)
		if alias, ok := languageRegionAlias(noExtensionLocale); ok {
			if _, rawAlias := raw[alias]; !rawAlias {
				out = appendAvailableLocale(out, seen, alias, noExtensionLocale, true)
			}
			out = appendFallbackLocales(out, seen, raw, alias, noExtensionLocale)
		}
		out = appendFallbackLocales(out, seen, raw, noExtensionLocale, noExtensionLocale)
	}
	return out
}

func appendFallbackLocales(out []availableLocale, seen, raw map[string]struct{}, loc, dataLocale string) []availableLocale {
	for {
		pos := truncationPosition(loc)
		if pos < 0 {
			return out
		}
		loc = loc[:pos]
		if _, rawLocale := raw[loc]; rawLocale {
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
	parts := strings.Split(loc, "-")
	if len(parts) < 3 || !isScriptSubtag(parts[1]) || !isRegionSubtag(parts[2]) {
		return "", false
	}
	alias := parts[0] + "-" + parts[2]
	if len(parts) > 3 {
		alias += "-" + strings.Join(parts[3:], "-")
	}
	return alias, true
}

func isScriptSubtag(s string) bool {
	return len(s) == 4 && isASCIIAlpha(s)
}

func isRegionSubtag(s string) bool {
	return len(s) == 2 && isASCIIAlpha(s) || len(s) == 3 && isASCIIDigit(s)
}

func isASCIIAlpha(s string) bool {
	for i := range len(s) {
		c := s[i]
		if c < 'A' || c > 'Z' && c < 'a' || c > 'z' {
			return false
		}
	}
	return true
}

func isASCIIDigit(s string) bool {
	for i := range len(s) {
		c := s[i]
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
