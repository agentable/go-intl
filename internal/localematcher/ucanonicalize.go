package localematcher

import (
	"cmp"
	"slices"
	"strings"
)

type keyword struct {
	key   string
	value string
}

func UnicodeExtensionValue(extension, key string) string {
	body, ok := strings.CutPrefix(extension, "-u-")
	if !ok {
		return ""
	}
	parts := strings.Split(body, "-")
	for i := range len(parts) {
		if len(parts[i]) != 2 || parts[i] != key {
			continue
		}
		start := i + 1
		end := start
		for end < len(parts) && len(parts[end]) >= 3 {
			end++
		}
		if start == end {
			return "true"
		}
		return strings.Join(parts[start:end], "-")
	}
	return ""
}

func InsertUnicodeExtensionAndCanonicalize(loc string, keywords []keyword) string {
	if len(keywords) == 0 {
		return loc
	}
	slices.SortFunc(keywords, func(a, b keyword) int { return cmp.Compare(a.key, b.key) })
	parts := []string{loc, "u"}
	for _, kw := range keywords {
		parts = append(parts, kw.key)
		if kw.value != "" && kw.value != "true" {
			parts = append(parts, strings.Split(kw.value, "-")...)
		}
	}
	return strings.Join(parts, "-")
}
