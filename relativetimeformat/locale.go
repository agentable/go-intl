package relativetimeformat

import (
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
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:               locales,
		Fallback:              fallback,
		LocaleMatcher:         cfg.localeMatcher,
		Matcher:               relativeTimeLocaleMatcher(),
		RelevantExtensionKeys: ecma402.NumberingSystemExtensionKeys(),
		OptionValues:          ecma402.NumberingSystemExtensionOptions(cfg.numberingSystem),
		LocaleData:            cldrnumber.NumberLocaleData{},
	})
	cldrLoc := ecma402.ResolveDataLocale(resolution, cldrrelativetime.ResolveLocale)
	numberingSystem := resolution.Extensions[ecma402.UnicodeExtensionKeyNumberingSystem]
	if numberingSystem == "" {
		if loc, ok := cldrlocale.ResolveLocale(resolution.DataLocale); ok {
			numberingSystem = loc.DefaultNumberingSystem()
		}
	}
	return resolution.Locale, cldrLoc, numberingSystem
}

var supportedLocales = sync.OnceValue(func() []string {
	return cldrlocale.IntersectSupportedLocales(
		cldrrelativetime.SupportedLocales(),
		cldrnumber.SupportedLocales(),
		cldrplural.SupportedLocales(),
	)
})

var relativeTimeLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(supportedLocales(), cldrlocale.Maximize)
})
