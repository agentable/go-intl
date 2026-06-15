package displaynames

import (
	"errors"
	"sync"

	cldrdn "github.com/agentable/go-intl/internal/cldr/displaynames"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

type DisplayNames struct {
	resolved   ResolvedOptions
	dataLocale string
}

var displayNamesLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrdn.SupportedLocales(), cldrlocale.Maximize)
})

func New(locales locale.List, opts Options) (*DisplayNames, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}
	resolvedLocale, dataLocale := resolveLocale(locales, validationLocale, cfg)
	resolved := ResolvedOptions{
		Locale:   resolvedLocale,
		Style:    Style(cfg.style),
		Type:     Type(cfg.typ),
		Fallback: Fallback(cfg.fallback),
	}
	if resolved.Type == Language {
		ld := LanguageDisplay(cfg.languageDisplay)
		resolved.LanguageDisplay = &ld
	}
	return &DisplayNames{dataLocale: dataLocale, resolved: resolved}, nil
}

// Of returns the localized display name for a code. Invalid code shape returns
// gointl.ErrInvalidCode. When no name exists and Fallback is CodeFallback (the
// default), the canonicalized code is returned with ok=true; when Fallback is
// NoneFallback, the empty string is returned with ok=false. The (string, bool)
// pair is the Go bridge for the JS `string | undefined` return.
func (d *DisplayNames) Of(code string) (string, bool, error) {
	canonical, err := ecma402.CanonicalCodeForDisplayNames(string(d.resolved.Type), code)
	if err != nil {
		return "", false, intlerr.New(
			intlerr.InvalidCode,
			"displaynames",
			string(d.resolved.Type),
			code,
			"",
			errors.Join(intlerr.ErrInvalidCode, err),
		)
	}
	var languageDisplay string
	if d.resolved.LanguageDisplay != nil {
		languageDisplay = string(*d.resolved.LanguageDisplay)
	}
	if value, ok := cldrdn.Of(d.dataLocale, string(d.resolved.Type), string(d.resolved.Style), languageDisplay, canonical, string(d.resolved.Fallback)); ok && value != "" {
		return value, true, nil
	}
	if d.resolved.Fallback == NoneFallback {
		return "", false, nil
	}
	return canonical, true, nil
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, string) {
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:       locales,
		Fallback:      fallback,
		LocaleMatcher: cfg.localeMatcher,
		Matcher:       displayNamesLocaleMatcher(),
	})
	dataLocale := resolution.DataLocale
	if dataLocale == "" {
		dataLocale = ecma402.DefaultLocale()
	}
	return resolution.Locale, dataLocale
}
