package listformat

import (
	"sync"

	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

type ListFormat struct {
	resolved  ResolvedOptions
	templates listTemplates
}

var listLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrlist.SupportedLocales(), cldrlist.Maximize)
})

func New(locales locale.List, opts Options) (*ListFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	validationLocaleName := validationLocale.String()
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocaleName); err != nil {
		return nil, err
	}
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:       locales,
		Fallback:      validationLocale,
		LocaleMatcher: cfg.localeMatcher,
		Matcher:       listLocaleMatcher(),
	})
	cldrLoc := ecma402.ResolveDataLocale(resolution, cldrlist.ResolveLocale)
	return &ListFormat{
		resolved: ResolvedOptions{
			Locale: resolution.Locale,
			Type:   Type(cfg.typ),
			Style:  Style(cfg.style),
		},
		templates: compileListTemplates(cldrlist.Pattern(cldrLoc, cfg.typ, cfg.style)),
	}, nil
}
