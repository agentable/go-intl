package listformat

import (
	"sync"

	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

type ListFormat struct {
	resolved ResolvedOptions
	pattern  cldrlist.ListPattern
}

var listLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrlist.SupportedLocales(), cldrlist.Maximize)
})

func New(locales locale.List, opts Options) (*ListFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}
	resolvedLocale, _, pattern := resolveLocale(locales, validationLocale, cfg)
	return &ListFormat{pattern: pattern, resolved: ResolvedOptions{
		Locale: resolvedLocale,
		Type:   Type(cfg.typ),
		Style:  Style(cfg.style),
	}}, nil
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, cldrlist.Locale, cldrlist.ListPattern) {
	defaultLocale := ecma402.DefaultLocale()
	matcher, _ := ecma402.LocaleMatcherAlgorithm(cfg.localeMatcher)
	result := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:     matcher,
		Matcher:       listLocaleMatcher(),
		Requested:     ecma402.RequestedLocaleStrings(locales),
		DefaultLocale: defaultLocale,
	})
	cldrLoc, ok := cldrlist.ResolveLocale(result.DataLocale)
	if !ok {
		cldrLoc, _ = cldrlist.ResolveLocale(defaultLocale)
	}
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = fallback
	}
	return resolvedLocale, cldrLoc, cldrlist.Pattern(cldrLoc, cfg.typ, cfg.style)
}
