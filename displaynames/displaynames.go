package displaynames

import (
	"sync"

	cldrdn "github.com/agentable/go-intl/internal/cldr/displaynames"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
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
		Matcher:       displayNamesLocaleMatcher(),
	})
	resolved := ResolvedOptions{
		Locale:   resolution.Locale,
		Style:    Style(cfg.style),
		Type:     Type(cfg.typ),
		Fallback: Fallback(cfg.fallback),
	}
	if resolved.Type == Language {
		resolved.LanguageDisplay = ecma402.ResolvedScalar(LanguageDisplay(cfg.languageDisplay))
	}
	return &DisplayNames{
		resolved:   resolved,
		dataLocale: ecma402.ResolveDataLocaleTag(resolution),
	}, nil
}

// Of returns the localized display name for a code. Invalid code shape returns
// gointl.ErrInvalidCode. When no name exists and Fallback is CodeFallback (the
// default), the canonicalized code is returned with ok=true; when Fallback is
// NoneFallback, the empty string is returned with ok=false. The (string, bool)
// pair is the Go bridge for the JS `string | undefined` return.
func (d *DisplayNames) Of(code string) (string, bool, error) {
	resolved := d.resolved
	dataLocale := d.dataLocale
	typ := string(resolved.Type)
	style := string(resolved.Style)
	allowCodeFallback := resolved.Fallback == CodeFallback

	canonical, err := ecma402.CanonicalCodeForDisplayNames(typ, code)
	if err != nil {
		localeName := resolved.Locale.String()
		return "", false, ecma402.InvalidDisplayNamesCodeError(
			displayNamesOwner,
			typ,
			code,
			localeName,
			err,
		)
	}
	languageDisplay := string(ecma402.ResolvedScalarValue(resolved.LanguageDisplay))
	if value, ok := cldrdn.Of(dataLocale, typ, style, languageDisplay, canonical, allowCodeFallback); ok && value != "" {
		return value, true, nil
	}
	if !allowCodeFallback {
		return "", false, nil
	}
	return canonical, true, nil
}
