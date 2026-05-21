package collation

// Default sensitivity table for Intl.Collator usage="search".
//
// ECMA-402 §10.2.3 step 24-25 (InitializeCollator) requires that when
// `usage` is "search" and `sensitivity` is unset, the default sensitivity
// comes from the resolved locale's [[SearchLocaleData]].[[sensitivity]].
// The active CLDR pin does not expose this per-locale field through the
// generated layer, so this file holds a small bootstrap table. When the
// CLDR generator is extended to emit search.collation/sensitivity, this
// file should be regenerated and treated as data, not source.

// SearchSensitivity returns the spec-default sensitivity for a Collator
// constructed with usage="search" and no explicit sensitivity option. The
// returned value is always one of "base", "accent", "case", or "variant".
func SearchSensitivity(dataLocale string) string {
	if s, ok := searchSensitivityByLocale[dataLocale]; ok {
		return s
	}
	// CLDR's "search" collator type is documented to behave with no
	// secondary-level differences for the vast majority of locales; "base"
	// is the broadest compatible default until per-locale data lands.
	return "base"
}

// searchSensitivityByLocale carries locale-specific overrides. Empty today;
// populate as CLDR data confirms divergences from the "base" default.
var searchSensitivityByLocale = map[string]string{}
