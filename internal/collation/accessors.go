// Package collation exposes the locale tags supported by the embedded collator.
//
// The data backing this package is sourced from golang.org/x/text/collate.
package collation

import (
	"slices"
	"strings"
	"sync"

	"golang.org/x/text/collate"
)

var supportedLocales = sync.OnceValue(func() []string {
	tags := collate.Supported()
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		s := t.String()
		if s == "und" {
			continue
		}
		// Drop locale-extension forms ("de-u-co-phonebk") from the public list;
		// they describe collation specializations, not base-name locales.
		if i := strings.Index(s, "-u-"); i >= 0 {
			s = s[:i]
		}
		out = append(out, s)
	}
	return dedupe(out)
})

// SupportedLocales returns the canonical locale tags with collator data.
func SupportedLocales() []string {
	return supportedLocales()
}

// SupportedCollations returns collation identifiers the active collator can
// apply through explicit ECMA-402 collation requests.
func SupportedCollations() []string {
	return slices.Clone(supportedCollations)
}

func dedupe(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	out := tags[:0]
	for _, t := range tags {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

var supportedCollations []string
