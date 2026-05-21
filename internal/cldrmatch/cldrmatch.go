// Package cldrmatch maps requested language tags to generated CLDR data locales for each formatter family.
package cldrmatch

import (
	"golang.org/x/text/language"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/localematcher"
)

type Kind int

const (
	KindNumber Kind = iota
	KindDate
)

func Number(tag language.Tag, matcher string) cldrnumber.Locale {
	loc, _ := cldrnumber.ResolveLocale(resolveDataLocale(tag, matcher, KindNumber))
	return loc
}

func Date(tag language.Tag, matcher string) cldrdate.Locale {
	loc, _ := cldrdate.ResolveLocale(resolveDataLocale(tag, matcher, KindDate))
	return loc
}

func resolveDataLocale(tag language.Tag, matcher string, kind Kind) string {
	if directString(tag.String(), kind) {
		return tag.String()
	}
	alg := localematcher.AlgorithmBestFit
	if matcher == "lookup" {
		alg = localematcher.AlgorithmLookup
	}
	matched := localematcher.MatchWithMaximizer([]string{tag.String()}, supportedLocales(kind), "en", alg, cldrlocale.Maximize)
	if directString(matched.DataLocale, kind) {
		return matched.DataLocale
	}
	return defaultLocale(kind)
}

func directString(tag string, kind Kind) bool {
	if _, ok := cldrlocale.ResolveLocale(tag); !ok {
		return false
	}
	return supportsLocale(kind, tag)
}

func supportedLocales(kind Kind) []string {
	if kind == KindDate {
		return cldrdate.SupportedLocales()
	}
	return cldrnumber.SupportedLocales()
}

func supportsLocale(kind Kind, tag string) bool {
	if kind == KindDate {
		return dateLocaleSet[tag]
	}
	return numberLocaleSet[tag]
}

func defaultLocale(kind Kind) string {
	if kind == KindDate {
		return dateLocaleSet.defaultLocale()
	}
	return numberLocaleSet.defaultLocale()
}

type localeSet map[string]bool

var numberLocaleSet = newLocaleSet(cldrnumber.SupportedLocales())
var dateLocaleSet = newLocaleSet(cldrdate.SupportedLocales())

func newLocaleSet(tags []string) localeSet {
	out := make(localeSet, len(tags))
	for _, tag := range tags {
		if _, ok := cldrlocale.ResolveLocale(tag); ok {
			out[tag] = true
		}
	}
	return out
}

func (s localeSet) defaultLocale() string {
	if s["en"] {
		return "en"
	}
	return ""
}
