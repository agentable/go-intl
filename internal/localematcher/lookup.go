package localematcher

import "strings"

func LookupMatcher(requested, supported []string, defaultLocale string) Result {
	return NewMatcher(supported, nil).lookup(requested, defaultLocale)
}

func BestAvailableLocale(supported []string, locale string) string {
	return NewMatcher(supported, nil).bestAvailableLocale(locale)
}

func truncationPosition(locale string) int {
	pos := strings.LastIndex(locale, "-")
	if pos >= 2 && locale[pos-2] == '-' {
		pos -= 2
	}
	return pos
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
