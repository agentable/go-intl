package cldr

import "strings"

func inheritedLocaleData[T any](locales []string, loaded map[string]T) map[string]T {
	out := make(map[string]T, len(locales))
	for _, locale := range locales {
		if locale == undefinedLocale {
			continue
		}
		if data, ok := localeDataOrParent(locale, loaded); ok {
			out[locale] = data
		}
	}
	return out
}

func localeDataOrParent[T any](locale string, loaded map[string]T) (T, bool) {
	if data, ok := loaded[locale]; ok {
		return data, true
	}
	for parent := parentLocale(locale); parent != ""; parent = parentLocale(parent) {
		if data, ok := loaded[parent]; ok {
			return data, true
		}
	}
	var zero T
	return zero, false
}

func parentLocale(locale string) string {
	pos := strings.LastIndex(locale, "-")
	if pos < 0 {
		return ""
	}
	return locale[:pos]
}
