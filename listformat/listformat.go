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
	matcher := string(BestFitLocaleMatcher)
	kind := string(Conjunction)
	style := string(LongStyle)
	ecma402.ApplyOption(&matcher, opts.LocaleMatcher)
	ecma402.ApplyOption(&kind, opts.Type)
	ecma402.ApplyOption(&style, opts.Style)

	if err := ecma402.ValidateStringOptions(
		listFormatOwner,
		validationLocale.String(),
		ecma402.LocaleMatcherOptionInput(matcher, opts.LocaleMatcher != nil),
		ecma402.RequiredStringOption("type", kind, listTypeValues[:]...),
		ecma402.RequiredStringOption("style", style, listStyleValues[:]...),
	); err != nil {
		return nil, err
	}
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:       locales,
		Fallback:      validationLocale,
		LocaleMatcher: matcher,
		Matcher:       listLocaleMatcher(),
	})
	cldrLoc := ecma402.ResolveDataLocale(resolution, cldrlist.ResolveLocale)
	return &ListFormat{
		resolved: ResolvedOptions{
			Locale: resolution.Locale,
			Type:   Type(kind),
			Style:  Style(style),
		},
		templates: compileListTemplates(cldrlist.Pattern(cldrLoc, kind, style), ecma402.ResolveDataLocaleTag(resolution)),
	}, nil
}
