// Package collation exposes the locale tags supported by the embedded collator.
//
// The data backing this package is sourced from golang.org/x/text/collate.
package collation

import (
	"maps"
	"slices"
	"strings"
	"sync"

	"github.com/agentable/go-intl/internal/localeid"

	"golang.org/x/text/collate"
)

type supportedCapability struct {
	locales            []string
	collations         []string
	collationsByLocale map[string][]string
}

var supportedCapabilities = sync.OnceValue(func() supportedCapability {
	tags := collate.Supported()
	locales := make([]string, 0, len(tags))
	seenLocales := make(map[string]bool, len(tags))
	collations := map[string]bool{}
	collationsByLocale := map[string]map[string]bool{}
	for _, t := range tags {
		s := t.String()
		if s == "und" {
			continue
		}
		base := s
		if tagBase, extension, ok := strings.Cut(s, "-u-"); ok {
			base = tagBase
			if co := supportedCollationFromExtension("-u-" + extension); co != "" {
				collations[co] = true
				if collationsByLocale[base] == nil {
					collationsByLocale[base] = map[string]bool{}
				}
				collationsByLocale[base][co] = true
			}
		}
		if !seenLocales[base] {
			seenLocales[base] = true
			locales = append(locales, base)
		}
	}
	perLocale := make(map[string][]string, len(collationsByLocale))
	for loc, values := range collationsByLocale {
		perLocale[loc] = slices.Sorted(maps.Keys(values))
	}
	return supportedCapability{
		locales:            locales,
		collations:         slices.Sorted(maps.Keys(collations)),
		collationsByLocale: perLocale,
	}
})

// SupportedLocales returns the canonical locale tags with collator data.
func SupportedLocales() []string {
	return slices.Clone(supportedCapabilities().locales)
}

// SupportedCollations returns collation identifiers the active collator can
// apply through explicit ECMA-402 collation requests.
func SupportedCollations() []string {
	return slices.Clone(supportedCapabilities().collations)
}

// SupportedCollationsForLocale returns ECMA-402 Collator [[co]] locale data for a
// canonical locale, using prefix fallback to the nearest backend data locale.
// The default collation is always first; specialization values are only included
// when the active x/text backend advertises a matching BCP 47 "co" extension.
func SupportedCollationsForLocale(locale string) []string {
	byLocale := supportedCapabilities().collationsByLocale
	var values []string
	for candidate := locale; candidate != ""; {
		if values = byLocale[candidate]; values != nil {
			break
		}
		pos := strings.LastIndexByte(candidate, '-')
		if pos < 0 {
			break
		}
		candidate = candidate[:pos]
	}
	return localeid.RelevantExtensionValues("default", values...)
}

func supportedCollationFromExtension(extension string) string {
	ext, err := localeid.ParseUnicodeExtension(localeid.LowercaseUnicodeLocaleID(extension))
	if err != nil {
		return ""
	}
	co, ok := ext.TypeForKey("co")
	if !ok || co == "" {
		return ""
	}
	switch co {
	case "default", "search", "standard":
		return ""
	default:
		return co
	}
}
