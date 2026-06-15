package ecma402

import (
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

// ConstructorLocaleOptions carries the shared ECMA-402 locale negotiation inputs
// every Intl service constructor prepares before formatter-specific data lookup.
type ConstructorLocaleOptions struct {
	Locales               locale.List
	Fallback              locale.Locale
	LocaleMatcher         string
	Matcher               *localematcher.Matcher
	RelevantExtensionKeys []string
	OptionValues          []localematcher.Option
	LocaleData            localematcher.LocaleDataLookup
}

// ConstructorLocaleResolution is the shared constructor locale result. Formatter
// packages still own CLDR data fallback, option-derived fields, and errors.
type ConstructorLocaleResolution struct {
	Locale     locale.Locale
	DataLocale string
	Extensions map[string]string
}

// ResolveConstructorLocale applies the ECMA-402 ResolveLocale path used by Intl
// service constructors after each formatter has validated its typed options.
func ResolveConstructorLocale(opts ConstructorLocaleOptions) ConstructorLocaleResolution {
	algorithm, _ := LocaleMatcherAlgorithm(opts.LocaleMatcher)
	result := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:             algorithm,
		Matcher:               opts.Matcher,
		Requested:             RequestedLocaleStrings(opts.Locales),
		DefaultLocale:         DefaultLocale(),
		RelevantExtensionKeys: opts.RelevantExtensionKeys,
		OptionValues:          opts.OptionValues,
		LocaleData:            opts.LocaleData,
	})
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = opts.Fallback
	}
	return ConstructorLocaleResolution{
		Locale:     resolvedLocale,
		DataLocale: result.DataLocale,
		Extensions: result.Extensions,
	}
}
