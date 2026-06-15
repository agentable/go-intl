package relativetimeformat

import (
	"slices"
	"sync"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	cldrplural "github.com/agentable/go-intl/internal/cldr/plural"
	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, cldrrelativetime.Locale, string) {
	defaultLocale := ecma402.DefaultLocale()
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:               locales,
		Fallback:              fallback,
		LocaleMatcher:         cfg.localeMatcher,
		Matcher:               relativeTimeLocaleMatcher(),
		RelevantExtensionKeys: []string{"nu"},
		OptionValues:          []localematcher.Option{{Key: "nu", Value: cfg.numberingSystem}},
		LocaleData:            cldrnumber.NumberLocaleData{},
	})
	cldrLoc, ok := cldrrelativetime.ResolveLocale(resolution.DataLocale)
	if !ok {
		cldrLoc, _ = cldrrelativetime.ResolveLocale(defaultLocale)
	}
	numberingSystem := resolution.Extensions["nu"]
	if numberingSystem == "" {
		if loc, ok := cldrlocale.ResolveLocale(resolution.DataLocale); ok {
			numberingSystem = loc.DefaultNumberingSystem()
		}
	}
	return resolution.Locale, cldrLoc, numberingSystem
}

var supportedLocales = sync.OnceValue(func() []string {
	numbers := cldrnumber.SupportedLocales()
	plurals := cldrplural.SupportedLocales()

	out := slices.Clone(cldrrelativetime.SupportedLocales())
	return slices.DeleteFunc(out, func(loc string) bool {
		return !slices.Contains(numbers, loc) || !slices.Contains(plurals, loc)
	})
})

var relativeTimeLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(supportedLocales(), cldrlocale.Maximize)
})
