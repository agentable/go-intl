// Package extract turns parsed cldr-json into compact intermediate data for
// codegen.
package extract

import (
	"maps"
	"slices"
)

const undefinedLocale = "und"

type Locales struct {
	Tags []string
}

func ExtractLocales(available []string) Locales {
	seen := make(map[string]bool, len(available))
	for _, tag := range available {
		if tag == "" || tag == undefinedLocale {
			continue
		}
		seen[tag] = true
	}
	sorted := slices.Sorted(maps.Keys(seen))
	tags := make([]string, len(sorted)+1)
	tags[0] = undefinedLocale
	copy(tags[1:], sorted)
	return Locales{Tags: tags}
}
