package numberformat

import (
	"sync"

	cldrcurrency "github.com/agentable/go-intl/internal/cldr/currency"
	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/cldr/plural"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	"github.com/agentable/go-intl/internal/localematcher"
	pluralop "github.com/agentable/go-intl/internal/plural"
	"github.com/agentable/go-intl/locale"
)

type NumberFormat struct {
	formatState decimalFormatState
}

type pluralRuleFunc func(pluralop.OperandsRecord) pluralop.Category

var numberLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrnumber.SupportedLocales(), cldrlocale.Maximize)
})

func New(locales locale.List, opts Options) (*NumberFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	validationLocaleName := validationLocale.String()
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if cfg.notation == "compact" && opts.UseGrouping == nil {
		cfg.useGrouping = string(UseGroupingMin2)
	}
	if err := cfg.validate(validationLocaleName); err != nil {
		return nil, err
	}
	mnfdDefault, mxfdDefault := digitDefaults(cfg)
	resolvedDigits, invalid, ok := ecma402nf.SetNumberFormatDigitOptions(cfg.digits, mnfdDefault, mxfdDefault, cfg.notation)
	if ok {
		return nil, ecma402nf.InvalidDigitOptionError(numberFormatOwner, invalid, validationLocaleName)
	}
	resolvedLocale, dataLocale, cldrLoc, unitLoc, numberingSystem := resolveLocale(locales, validationLocale, cfg)
	digitOptions := resolvedDigits.DigitOptions
	digitProperties := resolvedDigits.ResolvedProperties()
	resolved := ResolvedOptions{
		Locale:                   resolvedLocale,
		NumberingSystem:          numberingSystem,
		Style:                    Style(cfg.style),
		MinimumIntegerDigits:     digitOptions.MinimumIntegerDigits,
		MinimumFractionDigits:    digitProperties.MinimumFractionDigits,
		MaximumFractionDigits:    digitProperties.MaximumFractionDigits,
		MinimumSignificantDigits: digitProperties.MinimumSignificantDigits,
		MaximumSignificantDigits: digitProperties.MaximumSignificantDigits,
		UseGrouping:              UseGrouping(cfg.useGrouping),
		Notation:                 Notation(cfg.notation),
		SignDisplay:              SignDisplay(cfg.signDisplay),
		RoundingIncrement:        digitOptions.RoundingIncrement,
		RoundingMode:             RoundingMode(digitOptions.RoundingMode.String()),
		RoundingPriority:         RoundingPriority(digitOptions.RoundingPriority),
		TrailingZeroDisplay:      TrailingZeroDisplay(digitOptions.TrailingZeroDisplay),
	}
	if cfg.style == "currency" {
		resolved.Currency = ecma402.ResolvedScalar(cfg.currency)
		resolved.CurrencyDisplay = ecma402.ResolvedScalar(CurrencyDisplay(cfg.currencyDisplay))
		resolved.CurrencySign = ecma402.ResolvedScalar(CurrencySign(cfg.currencySign))
	}
	if cfg.style == "unit" {
		resolved.Unit = ecma402.ResolvedScalar(cfg.unit)
		resolved.UnitDisplay = ecma402.ResolvedScalar(UnitDisplay(cfg.unitDisplay))
	}
	if cfg.notation == "compact" {
		resolved.CompactDisplay = ecma402.ResolvedScalar(CompactDisplay(cfg.compactDisplay))
	}
	symbols := cldrLoc.NumberSymbols(numberingSystem)
	grouping := groupingForNumberFormat(cldrLoc, resolved)
	cardinalRule, err := plural.Rule(dataLocale, "cardinal")
	if err != nil {
		return nil, err
	}
	return &NumberFormat{
		formatState: decimalFormatState{
			resolved:     resolved,
			symbols:      symbols,
			grouping:     grouping,
			digitOptions: resolvedDigits,
			cardinalRule: cardinalRule,
			currencyLoc:  unitLoc,
			currency:     currencyPatternsForNumberFormat(cldrLoc, unitLoc, resolved),
			unit:         unitPatternsForNumberFormat(unitLoc, resolved),
			compact:      compactPatternsForNumberFormat(cldrLoc, resolved),
		},
	}, nil
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, string, cldrnumber.Locale, cldrlocale.Locale, string) {
	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:               locales,
		Fallback:              fallback,
		LocaleMatcher:         cfg.localeMatcher,
		Matcher:               numberLocaleMatcher(),
		RelevantExtensionKeys: ecma402.NumberingSystemExtensionKeys(),
		OptionValues:          ecma402.NumberingSystemExtensionOptions(cfg.numberingSystem),
		LocaleData:            cldrnumber.NumberLocaleData{},
	})
	cldrLoc := ecma402.ResolveDataLocale(resolution, cldrnumber.ResolveLocale)
	unitLoc := ecma402.ResolveDataLocale(resolution, cldrlocale.ResolveLocale)
	numberingSystem := resolution.Extensions[ecma402.UnicodeExtensionKeyNumberingSystem]
	if numberingSystem == "" {
		numberingSystem = cldrLoc.DefaultNumberingSystem()
	}
	return resolution.Locale, ecma402.ResolveDataLocaleTag(resolution), cldrLoc, unitLoc, numberingSystem
}

func digitDefaults(cfg config) (minimumFractionDigits, maximumFractionDigits int) {
	if cfg.style == "currency" && cfg.notation == "standard" {
		currency := cldrcurrency.Digits(cfg.currency)
		return currency.DefaultDigits, currency.DefaultDigits
	}
	if cfg.style == "percent" {
		return 0, 0
	}
	return 0, 3
}
