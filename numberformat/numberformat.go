package numberformat

import (
	"strings"
	"sync"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	cldrnumber "github.com/agentable/go-intl/internal/cldr/number"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

type NumberFormat struct {
	resolved      ResolvedOptions
	digits        digitState
	cldrLoc       cldrnumber.Locale
	numberSymbols cldrnumber.NumberSymbols
	grouping      digitGrouping
	currency      currencyPatternSet
	unit          unitPatternSet
	compact       compactPatternSet
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
	if cfg.style == "percent" && !cfg.hasMaxFracDigits {
		cfg.maxFracDigits = 0
	}
	if cfg.style == "currency" && cfg.currency != "" {
		cfg.currency = strings.ToUpper(cfg.currency)
		if cfg.notation == "standard" {
			currency := cldrnumber.CurrencyDigits(cfg.currency)
			if !cfg.hasMinFracDigits {
				cfg.minFracDigits = currency.DefaultDigits
			}
			if !cfg.hasMaxFracDigits {
				cfg.maxFracDigits = currency.DefaultDigits
			}
		}
	}
	// ECMA-402 SetNumberFormatDigitOptions: when only one of mnfd/mxfd is
	// provided, the other defaults are adjusted to keep mnfd <= mxfd
	// (step 18.a/b: max(mxfdDefault, mnfd) / min(mnfdDefault, mxfd)).
	if cfg.hasMinFracDigits && !cfg.hasMaxFracDigits && cfg.minFracDigits > cfg.maxFracDigits {
		cfg.maxFracDigits = cfg.minFracDigits
	}
	if cfg.hasMaxFracDigits && !cfg.hasMinFracDigits && cfg.maxFracDigits < cfg.minFracDigits {
		cfg.minFracDigits = cfg.maxFracDigits
	}
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}
	cfg = resolveDigitDefaults(cfg)
	resolvedLocale, cldrLoc, numberingSystem := resolveLocale(locales, validationLocale, cfg)
	digits := digitState{
		minInt:  cfg.minIntDigits,
		minFrac: cfg.minFracDigits,
		maxFrac: cfg.maxFracDigits,
		minSig:  cfg.minSigDigits,
		maxSig:  cfg.maxSigDigits,
	}
	// ECMA-402 SetNumberFormatDigitOptions: when sig digits are set and
	// rounding priority is "auto", fraction digits step out of the effective
	// rounding path (kept as 0/0 internally).
	if cfg.roundingPriority == "auto" && (cfg.hasMinSigDigits || cfg.hasMaxSigDigits) {
		digits.minFrac, digits.maxFrac = 0, 0
	}
	roundingType := resolvedRoundingType(cfg)
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
	switch roundingType {
	case "fractionDigits":
		minFrac, maxFrac := digits.minFrac, digits.maxFrac
		resolved.MinimumFractionDigits = &minFrac
		resolved.MaximumFractionDigits = &maxFrac
	case "significantDigits":
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
		cldrLoc:       cldrLoc,
		resolved:      resolved,
		digits:        digits,
		numberSymbols: numberSymbolsForNumberFormat(cldrLoc, numberingSystem),
		grouping:      groupingForNumberFormat(cldrLoc, resolved),
		currency:      currencyPatternsForNumberFormat(cldrLoc, resolved),
		unit:          unitPatternsForNumberFormat(cldrLoc, resolved),
		compact:       compactPatternsForNumberFormat(cldrLoc, resolved),
	}, nil
}

// resolvedRoundingType mirrors ECMA-402 §15.5 ResolveRoundingType:
//   - morePrecision / lessPrecision when roundingPriority selects either;
//   - significantDigits when only sig digit options are set;
//   - fractionDigits otherwise.
func resolvedRoundingType(cfg config) string {
	switch cfg.roundingPriority {
	case "morePrecision":
		return "morePrecision"
	case "lessPrecision":
		return "lessPrecision"
	}
	if cfg.hasMinSigDigits || cfg.hasMaxSigDigits {
		return "significantDigits"
	}
	return "fractionDigits"
}

func resolveLocale(locales locale.List, fallback locale.Locale, cfg config) (locale.Locale, cldrnumber.Locale, string) {
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
	numberingSystem := result.Extensions["nu"]
	if numberingSystem == "" {
		numberingSystem = cldrLoc.DefaultNumberingSystem()
	}
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = fallback
	}
	return resolvedLocale, cldrLoc, numberingSystem
}

func resolveDigitDefaults(cfg config) config {
	if cfg.roundingPriority == "auto" &&
		cfg.notation == "compact" &&
		!cfg.hasMinSigDigits && !cfg.hasMaxSigDigits &&
		!cfg.hasMinFracDigits && !cfg.hasMaxFracDigits {
		cfg.minFracDigits = 0
		cfg.maxFracDigits = 0
		cfg.minSigDigits = 1
		cfg.maxSigDigits = 2
		cfg.roundingPriority = "morePrecision"
		return cfg
	}
	if !cfg.hasMinSigDigits && !cfg.hasMaxSigDigits && cfg.roundingPriority == "auto" {
		return cfg
	}
	if !cfg.hasMinSigDigits {
		cfg.minSigDigits = 1
	}
	if !cfg.hasMaxSigDigits {
		cfg.maxSigDigits = 21
	}
	return cfg
}
