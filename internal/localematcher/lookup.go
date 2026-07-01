package localematcher

import (
	"strings"

	"github.com/agentable/go-intl/internal/localeid"
)

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
	matcher := NewMatcher(supported, nil)
	out := make([]string, 0, len(requested))
	for _, loc := range requested {
		noExtensionLocale, _ := removeUnicodeExtension(loc)
		if matcher.bestAvailableLocale(noExtensionLocale) != "" {
			out = append(out, noExtensionLocale)
		}
	}
	return out
}

func removeUnicodeExtension(locale string) (string, string) {
	base, extension, err := localeid.RemoveUnicodeExtension(locale)
	if err != nil {
		return locale, ""
	}
	return base, extension
}
