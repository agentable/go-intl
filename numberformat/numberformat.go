package numberformat

import (
	"fmt"
	"strings"

	"golang.org/x/text/language"

	"github.com/agentable/go-intl/internal/cldr"
	"github.com/agentable/go-intl/internal/localematcher"
	"github.com/agentable/go-intl/locale"
)

type NumberFormat struct {
	resolved ResolvedOptions
	cldrLoc  cldr.Locale
	grouping digitGrouping
}

func New(loc locale.Locale, opts ...Options) (*NumberFormat, error) {
	cfg := defaultConfig()
	if len(opts) > 1 {
		return nil, fmt.Errorf("numberformat: multiple options for locale %q: %w", loc.Tag().String(), ErrInvalidOption)
	}
	if len(opts) > 0 {
		applyOptions(&cfg, opts[0])
	}
	if cfg.style == "percent" && !cfg.hasMaxFracDigits {
		cfg.maxFracDigits = 0
	}
	if cfg.style == "currency" && cfg.currency != "" {
		currency := cldr.CurrencyDigits(strings.ToUpper(cfg.currency))
		cfg.currency = strings.ToUpper(cfg.currency)
		if !cfg.hasMinFracDigits {
			cfg.minFracDigits = currency.DefaultDigits
		}
		if !cfg.hasMaxFracDigits {
			cfg.maxFracDigits = currency.DefaultDigits
		}
	}
	if err := cfg.validate(loc); err != nil {
		return nil, err
	}
	cfg = resolveDigitDefaults(cfg)
	resolvedLocale, cldrLoc, numberingSystem := resolveLocale(loc, cfg)
	minFracDigits, maxFracDigits := cfg.minFracDigits, cfg.maxFracDigits
	if cfg.roundingPriority == "auto" && (cfg.hasMinSigDigits || cfg.hasMaxSigDigits) {
		minFracDigits, maxFracDigits = 0, 0
	}
	resolved := ResolvedOptions{
		Locale:                   resolvedLocale,
		NumberingSystem:          numberingSystem,
		Style:                    Style(cfg.style),
		Currency:                 cfg.currency,
		CurrencyDisplay:          CurrencyDisplay(cfg.currencyDisplay),
		CurrencySign:             CurrencySign(cfg.currencySign),
		Unit:                     cfg.unit,
		UnitDisplay:              UnitDisplay(cfg.unitDisplay),
		MinimumIntegerDigits:     cfg.minIntDigits,
		MinimumFractionDigits:    minFracDigits,
		MaximumFractionDigits:    maxFracDigits,
		MinimumSignificantDigits: cfg.minSigDigits,
		MaximumSignificantDigits: cfg.maxSigDigits,
		UseGrouping:              UseGrouping(cfg.useGrouping),
		Notation:                 Notation(cfg.notation),
		CompactDisplay:           CompactDisplay(cfg.compactDisplay),
		SignDisplay:              SignDisplay(cfg.signDisplay),
		RoundingIncrement:        cfg.roundingIncrement,
		RoundingMode:             RoundingMode(cfg.roundingMode),
		RoundingPriority:         RoundingPriority(cfg.roundingPriority),
		TrailingZeroDisplay:      TrailingZeroDisplay(cfg.trailingZeroDisplay),
	}
	return &NumberFormat{cldrLoc: cldrLoc, resolved: resolved, grouping: groupingForNumberFormat(cldrLoc, resolved)}, nil
}

func resolveLocale(loc locale.Locale, cfg config) (locale.Locale, cldr.Locale, string) {
	result := localematcher.ResolveLocale(localematcher.ResolveOptions{
		Algorithm:             localeMatcherAlgorithm(cfg.localeMatcher),
		Requested:             []string{loc.String()},
		Supported:             cldr.NumberSupportedLocales(),
		DefaultLocale:         "en",
		RelevantExtensionKeys: []string{"nu"},
		Options:               map[string]string{"nu": cfg.numberingSystem},
		LocaleData:            cldr.NumberLocaleData{},
	})
	cldrLoc, ok := cldr.ResolveLocale(language.Make(result.DataLocale))
	if !ok {
		cldrLoc, _ = cldr.ResolveLocale(language.English)
	}
	numberingSystem := result.Extensions["nu"]
	if numberingSystem == "" {
		numberingSystem = cldrLoc.DefaultNumberingSystem()
	}
	resolvedLocale, err := locale.Parse(result.Locale)
	if err != nil {
		resolvedLocale = loc
	}
	return resolvedLocale, cldrLoc, numberingSystem
}

func localeMatcherAlgorithm(matcher string) localematcher.Algorithm {
	if matcher == string(LookupLocaleMatcher) {
		return localematcher.AlgorithmLookup
	}
	return localematcher.AlgorithmBestFit
}

func resolveDigitDefaults(cfg config) config {
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
