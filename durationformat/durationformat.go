package durationformat

import (
	"sync"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
)

type DurationFormat struct {
	resolved                        ResolvedOptions
	unitOptions                     [unitCount]resolvedUnitConfig
	listFormatter                   *listformat.ListFormat
	unitFormatters                  [unitCount]durationNumberFormatters
	numericFormatters               [unitCount]durationNumberFormatters
	secondsNumericFractionFormatter durationNumberFormatters
	separator                       string
}

var durationLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(supportedLocales(), cldrlocale.Maximize)
})

func New(locales locale.List, opts Options) (*DurationFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	validationLocaleName := validationLocale.String()
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validateLocaleOptions(validationLocaleName); err != nil {
		return nil, err
	}
	resolvedLocale, cldrLoc, numberingSystem := resolveLocale(locales, validationLocale, cfg)
	if err := ecma402.ValidateStringOptions(durationFormatOwner, validationLocaleName, durationStyleOption(cfg.style)); err != nil {
		return nil, err
	}
	unitOptions, err := resolveUnitOptions(cfg, validationLocaleName)
	if err != nil {
		return nil, err
	}
	if err := ecma402.ValidateIntegerOptions(durationFormatOwner, validationLocaleName, ecma402.IntegerOption{
		Name:  "fractionalDigits",
		Value: cfg.fractionalDigits,
		Min:   0,
		Max:   9,
		Set:   cfg.hasFractionalDigits,
	}); err != nil {
		return nil, err
	}
	resolved := resolvedOptionsForDurationFormat(resolvedLocale, numberingSystem, cfg, unitOptions)
	symbols := cldrLoc.NumberSymbols(numberingSystem)
	separator := symbols.TimeSeparator
	format := &DurationFormat{
		resolved:    resolved,
		unitOptions: unitOptions,
		separator:   separator,
	}
	if err := buildDurationFormatters(format); err != nil {
		return nil, err
	}
	return format, nil
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, cldrnumber.Locale, string) {
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:               locales,
		Fallback:              fallback,
		LocaleMatcher:         cfg.localeMatcher,
		Matcher:               durationLocaleMatcher(),
		RelevantExtensionKeys: ecma402.NumberingSystemExtensionKeys(),
		OptionValues:          ecma402.NumberingSystemExtensionOptions(cfg.numberingSystem),
		LocaleData:            cldrnumber.NumberLocaleData{},
	})
	cldrLoc := ecma402.ResolveDataLocale(resolution, cldrnumber.ResolveLocale)
	numberingSystem := resolution.Extensions[ecma402.UnicodeExtensionKeyNumberingSystem]
	if numberingSystem == "" {
		numberingSystem = cldrLoc.DefaultNumberingSystem()
	}
	return resolution.Locale, cldrLoc, numberingSystem
}
