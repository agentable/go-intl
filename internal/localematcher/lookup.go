package localematcher

import "strings"

func LookupMatcher(requested, supported []string, defaultLocale string) Result {
	for _, loc := range requested {
		noExtensionLocale, extension := removeUnicodeExtension(loc)
		availableLocale := BestAvailableLocale(supported, noExtensionLocale)
		if availableLocale != "" {
			return Result{Locale: availableLocale, DataLocale: availableLocale, Extension: extension}
		}
	}
	return Result{Locale: defaultLocale, DataLocale: defaultLocale}
}

func BestAvailableLocale(supported []string, locale string) string {
	available := make(map[string]struct{}, len(supported))
	for _, loc := range supported {
		available[loc] = struct{}{}
	}
	candidate := locale
	for {
		if _, ok := available[candidate]; ok {
			return candidate
		}
		pos := strings.LastIndex(candidate, "-")
		if pos < 0 {
			return ""
		}
		if pos >= 2 && candidate[pos-2] == '-' {
			pos -= 2
		}
		candidate = candidate[:pos]
	}
}

func LookupSupportedLocales(supported, requested []string) []string {
	out := make([]string, 0, len(requested))
	for _, loc := range requested {
		noExtensionLocale, _ := removeUnicodeExtension(loc)
		if BestAvailableLocale(supported, noExtensionLocale) != "" {
			out = append(out, noExtensionLocale)
		}
	}
	return out
}

func removeUnicodeExtension(locale string) (string, string) {
	start := strings.Index(locale, "-u-")
	if start < 0 {
		return locale, ""
	}
	end := len(locale)
	parts := strings.Split(locale[start+1:], "-")
	pos := start + 1
	for i, part := range parts {
		if i > 0 {
			pos++
		}
		if i > 0 && len(part) == 1 {
			end = pos - 1
			break
		}
		pos += len(part)
	}
	return locale[:start] + locale[end:], locale[start:end]
}
