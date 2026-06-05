package numberformat

import (
	"strings"
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
	resolved               ResolvedOptions
	digitOptions           ecma402nf.DigitOptions
	cardinalRule           func(pluralop.OperandsRecord) pluralop.Category
	digits                 digitState
	cldrLoc                cldrnumber.Locale
	currencyLoc            cldrcurrency.Locale
	numberSymbols          cldrnumber.NumberSymbols
	grouping               digitGrouping
	currency               currencyPatternSet
	unit                   unitPatternSet
	compact                compactPatternSet
	numberingSystem        string
	localizeDigits         bool
	decimalIntegerFastPath bool
}

var numberLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrnumber.SupportedLocales(), cldrlocale.Maximize)
})

// digitState carries the effective integer/fraction/significant digit values
// used by hot-path formatting. It is populated independently of the public
// ResolvedOptions, which may nil out fraction- or significant-digit pairs
// per ECMA-402 §15.5.2 visibility rules.
type digitState struct {
	minInt           int
	minFrac, maxFrac int
	minSig, maxSig   int
}

func New(locales locale.List, opts Options) (*NumberFormat, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if cfg.notation == "compact" && opts.UseGrouping == "" {
		cfg.useGrouping = string(UseGroupingMin2)
	}
	if cfg.style == "currency" && cfg.currency != "" {
		cfg.currency = strings.ToUpper(cfg.currency)
	}
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}
	mnfdDefault, mxfdDefault := digitDefaults(cfg)
	resolvedDigits, invalid, ok := ecma402nf.SetNumberFormatDigitOptions(cfg.digitOptionInput(), mnfdDefault, mxfdDefault, cfg.notation)
	if ok {
		return nil, invalidOption(invalid.Name, invalid.Value, validationLocale)
	}
	cfg.applyResolvedDigits(resolvedDigits)
	resolvedLocale, cldrLoc, unitLoc, numberingSystem := resolveLocale(locales, validationLocale, cfg)
	digits := digitState{
		minInt:  cfg.minIntDigits,
		minFrac: cfg.minFracDigits,
		maxFrac: cfg.maxFracDigits,
		minSig:  cfg.minSigDigits,
		maxSig:  cfg.maxSigDigits,
	}
	resolved := ResolvedOptions{
		Locale:               resolvedLocale,
		NumberingSystem:      numberingSystem,
		Style:                Style(cfg.style),
		MinimumIntegerDigits: digits.minInt,
		UseGrouping:          UseGrouping(cfg.useGrouping),
		Notation:             Notation(cfg.notation),
		CompactDisplay:       CompactDisplay(cfg.compactDisplay),
		SignDisplay:          SignDisplay(cfg.signDisplay),
		RoundingIncrement:    cfg.roundingIncrement,
		RoundingMode:         RoundingMode(cfg.roundingMode),
		RoundingPriority:     RoundingPriority(cfg.roundingPriority),
		TrailingZeroDisplay:  TrailingZeroDisplay(cfg.trailingZeroDisplay),
	}
	// ECMA-402 §15.5.2: only the active rounding-type pair appears in
	// resolvedOptions. The opposite pair must remain absent (nil here).
	switch cfg.roundingType {
	case ecma402nf.RoundingTypeFractionDigits:
		minFrac, maxFrac := digits.minFrac, digits.maxFrac
		resolved.MinimumFractionDigits = &minFrac
		resolved.MaximumFractionDigits = &maxFrac
	case ecma402nf.RoundingTypeSignificantDigits:
		minSig, maxSig := digits.minSig, digits.maxSig
		resolved.MinimumSignificantDigits = &minSig
		resolved.MaximumSignificantDigits = &maxSig
	default: // morePrecision, lessPrecision — both pairs visible
		minFrac, maxFrac := digits.minFrac, digits.maxFrac
		minSig, maxSig := digits.minSig, digits.maxSig
		resolved.MinimumFractionDigits = &minFrac
		resolved.MaximumFractionDigits = &maxFrac
		resolved.MinimumSignificantDigits = &minSig
		resolved.MaximumSignificantDigits = &maxSig
	}
	if cfg.style == "currency" {
		resolved.Currency = cfg.currency
		resolved.CurrencyDisplay = CurrencyDisplay(cfg.currencyDisplay)
		resolved.CurrencySign = CurrencySign(cfg.currencySign)
	}
	if cfg.style == "unit" {
		resolved.Unit = cfg.unit
		resolved.UnitDisplay = UnitDisplay(cfg.unitDisplay)
	}
	return &NumberFormat{
		cldrLoc:                cldrLoc,
		currencyLoc:            unitLoc,
		resolved:               resolved,
		digitOptions:           resolvedDigits.DigitOptions,
		cardinalRule:           cardinalRuleForNumberFormat(resolvedLocale),
		digits:                 digits,
		numberSymbols:          numberSymbolsForNumberFormat(cldrLoc, numberingSystem),
		grouping:               groupingForNumberFormat(cldrLoc, resolved),
		currency:               currencyPatternsForNumberFormat(cldrLoc, unitLoc, resolved),
		unit:                   unitPatternsForNumberFormat(unitLoc, resolved),
		compact:                compactPatternsForNumberFormat(cldrLoc, resolved),
		numberingSystem:        numberingSystem,
		localizeDigits:         numberingSystem != "" && numberingSystem != "latn",
		decimalIntegerFastPath: canUseDecimalIntegerFastPath(resolved, digits),
	}, nil
}

func cardinalRuleForNumberFormat(loc locale.Locale) func(pluralop.OperandsRecord) pluralop.Category {
	if rule, ok := plural.CardinalRule(loc.Tag().String()); ok {
		return rule
	}
	if rule, ok := plural.CardinalRule("en"); ok {
		return rule
	}
	return func(pluralop.OperandsRecord) pluralop.Category {
		return pluralop.Other
	}
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, cldrnumber.Locale, cldrlocale.Locale, string) {
	defaultLocale := ecma402.DefaultLocale()
	matcher, _ := ecma402.LocaleMatcherAlgorithm(cfg.localeMatcher)
	result := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:             matcher,
		Matcher:               numberLocaleMatcher(),
		Requested:             ecma402.RequestedLocaleStrings(locales),
		DefaultLocale:         defaultLocale,
		RelevantExtensionKeys: []string{"nu"},
		OptionValues:          []localematcher.Option{{Key: "nu", Value: cfg.numberingSystem}},
		LocaleData:            cldrnumber.NumberLocaleData{},
	})
	cldrLoc, ok := cldrnumber.ResolveLocale(result.DataLocale)
	if !ok {
		cldrLoc, _ = cldrnumber.ResolveLocale(defaultLocale)
	}
	unitLoc, ok := cldrlocale.ResolveLocale(result.DataLocale)
	if !ok {
		unitLoc, _ = cldrlocale.ResolveLocale(defaultLocale)
	}
	numberingSystem := result.Extensions["nu"]
	if numberingSystem == "" {
		numberingSystem = cldrLoc.DefaultNumberingSystem()
	}
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = fallback
	}
	return resolvedLocale, cldrLoc, unitLoc, numberingSystem
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
