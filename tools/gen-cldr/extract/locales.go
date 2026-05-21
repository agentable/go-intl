// Package extract turns parsed cldr-json into compact intermediate data for
// codegen.
package extract

import (
	"maps"
	"slices"
)

type Locales struct {
	Tags []string
}

func ExtractLocales(available, allowlist []string) Locales {
	allowed := make(map[string]bool, len(allowlist)+1)
	allowed["und"] = true
	for _, tag := range allowlist {
		allowed[tag] = true
	}
	seen := make(map[string]bool, len(available)+1)
	seen["und"] = true
	for _, tag := range available {
		if allowed[tag] {
			seen[tag] = true
		}
	}
	tags := slices.Sorted(maps.Keys(seen))
	if i := slices.Index(tags, "und"); i > 0 {
		tags = slices.Insert(slices.Delete(tags, i, i+1), 0, "und")
	}
	return Locales{Tags: tags}
}
