// Package cldrmatch maps requested language tags to generated CLDR data locales for each formatter family.
package cldrmatch

import (
	"golang.org/x/text/language"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/localematcher"
)

type Kind int

const (
	KindNumber Kind = iota
	KindDate
)

func Number(tag language.Tag, matcher string) cldr.Locale {
	return resolve(tag, matcher, KindNumber)
}

func Date(tag language.Tag, matcher string) cldr.Locale {
	return resolve(tag, matcher, KindDate)
}

func resolve(tag language.Tag, matcher string, kind Kind) cldr.Locale {
	if cldrLoc, ok := direct(tag, kind); ok {
		return cldrLoc
	}
	alg := localematcher.AlgorithmBestFit
	if matcher == "lookup" {
		alg = localematcher.AlgorithmLookup
	}
	matched := localematcher.Match([]string{tag.String()}, supportedLocales(kind), "en", alg)
	if cldrLoc, ok := direct(language.Make(matched.DataLocale), kind); ok {
		return cldrLoc
	}
	return defaultLocale(kind)
}

func direct(tag language.Tag, kind Kind) (cldr.Locale, bool) {
	loc, ok := cldr.ResolveLocale(tag)
	if !ok || !supportsLocale(kind, loc) {
		return cldr.Undefined, false
	}
	return loc, true
}

func supportedLocales(kind Kind) []string {
	if kind == KindDate {
		return cldr.DateSupportedLocales()
	}
	return cldr.NumberSupportedLocales()
}

func supportsLocale(kind Kind, loc cldr.Locale) bool {
	if kind == KindDate {
		return dateLocaleSet[loc]
	}
	return numberLocaleSet[loc]
}

func defaultLocale(kind Kind) cldr.Locale {
	if kind == KindDate {
		return dateLocaleSet.defaultLocale()
	}
	return numberLocaleSet.defaultLocale()
}

type localeSet map[cldr.Locale]bool

var numberLocaleSet = newLocaleSet(cldr.NumberSupportedLocales())
var dateLocaleSet = newLocaleSet(cldr.DateSupportedLocales())

func newLocaleSet(tags []string) localeSet {
	out := make(localeSet, len(tags))
	for _, tag := range tags {
		loc, ok := cldr.ResolveLocale(language.Make(tag))
		if ok {
			out[loc] = true
		}
	}
	return out
}

func (s localeSet) defaultLocale() cldr.Locale {
	loc, ok := cldr.ResolveLocale(language.Make("en"))
	if ok && s[loc] {
		return loc
	}
	return cldr.Undefined
}
