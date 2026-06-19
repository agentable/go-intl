package numberformat

import (
	"strings"

	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

type Options struct {
	Style                    Style
	Currency                 Currency
	CurrencyDisplay          CurrencyDisplay
	CurrencySign             CurrencySign
	Unit                     Unit
	UnitDisplay              UnitDisplay
	MinimumIntegerDigits     *int
	MinimumFractionDigits    *int
	MaximumFractionDigits    *int
	MinimumSignificantDigits *int
	MaximumSignificantDigits *int
	RoundingIncrement        *int
	RoundingPriority         RoundingPriority
	RoundingMode             RoundingMode
	TrailingZeroDisplay      TrailingZeroDisplay
	Notation                 Notation
	CompactDisplay           CompactDisplay
	UseGrouping              UseGrouping
	SignDisplay              SignDisplay
	LocaleMatcher            LocaleMatcher
	NumberingSystem          string
}

type config struct {
	style               string
	currency            string
	currencyDisplay     string
	currencySign        string
	unit                string
	unitDisplay         string
	minIntDigits        int
	minFracDigits       int
	maxFracDigits       int
	hasMinFracDigits    bool
	hasMaxFracDigits    bool
	minSigDigits        int
	maxSigDigits        int
	hasMinSigDigits     bool
	hasMaxSigDigits     bool
	roundingPriority    string
	roundingIncrement   int
	roundingMode        string
	roundingType        ecma402nf.RoundingType
	trailingZeroDisplay string
	notation            string
	compactDisplay      string
	useGrouping         string
	signDisplay         string
	localeMatcher       string
	numberingSystem     string
}

func defaultConfig() config {
	return config{
		style:               string(DecimalStyle),
		currencyDisplay:     string(CurrencyDisplaySymbol),
		currencySign:        string(StandardCurrencySign),
		unitDisplay:         string(ShortUnitDisplay),
		minIntDigits:        1,
		minFracDigits:       0,
		maxFracDigits:       3,
		roundingPriority:    string(AutoRoundingPriority),
		roundingIncrement:   1,
		roundingMode:        string(HalfExpandRoundingMode),
		trailingZeroDisplay: string(AutoTrailingZeroDisplay),
		notation:            string(StandardNotation),
		compactDisplay:      string(ShortCompactDisplay),
		useGrouping:         string(UseGroupingAuto),
		signDisplay:         string(AutoSignDisplay),
		localeMatcher:       string(BestFitLocaleMatcher),
	}
}

func applyOptions(cfg *config, opts Options) {
	if opts.Style != "" {
		cfg.style = string(opts.Style)
	}
	if opts.Currency != "" {
		cfg.currency = strings.ToUpper(string(opts.Currency))
	}
	if opts.CurrencyDisplay != "" {
		cfg.currencyDisplay = string(opts.CurrencyDisplay)
	}
	if opts.CurrencySign != "" {
		cfg.currencySign = string(opts.CurrencySign)
	}
	if opts.Unit != "" {
		cfg.unit = string(opts.Unit)
	}
	if opts.UnitDisplay != "" {
		cfg.unitDisplay = string(opts.UnitDisplay)
	}
	if opts.MinimumIntegerDigits != nil {
		cfg.minIntDigits = *opts.MinimumIntegerDigits
	}
	if opts.MinimumFractionDigits != nil {
		cfg.minFracDigits = *opts.MinimumFractionDigits
		cfg.hasMinFracDigits = true
	}
	if opts.MaximumFractionDigits != nil {
		cfg.maxFracDigits = *opts.MaximumFractionDigits
		cfg.hasMaxFracDigits = true
	}
	if opts.MinimumSignificantDigits != nil {
		cfg.minSigDigits = *opts.MinimumSignificantDigits
		cfg.hasMinSigDigits = true
	}
	if opts.MaximumSignificantDigits != nil {
		cfg.maxSigDigits = *opts.MaximumSignificantDigits
		cfg.hasMaxSigDigits = true
	}
	if opts.RoundingIncrement != nil {
		cfg.roundingIncrement = *opts.RoundingIncrement
	}
	if opts.RoundingPriority != "" {
		cfg.roundingPriority = string(opts.RoundingPriority)
	}
	if opts.RoundingMode != "" {
		cfg.roundingMode = string(opts.RoundingMode)
	}
	if opts.TrailingZeroDisplay != "" {
		cfg.trailingZeroDisplay = string(opts.TrailingZeroDisplay)
	}
	if opts.Notation != "" {
		cfg.notation = string(opts.Notation)
	}
	if opts.CompactDisplay != "" {
		cfg.compactDisplay = string(opts.CompactDisplay)
	}
	if opts.UseGrouping != "" {
		cfg.useGrouping = string(opts.UseGrouping)
	}
	if opts.SignDisplay != "" {
		cfg.signDisplay = string(opts.SignDisplay)
	}
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	if opts.NumberingSystem != "" {
		cfg.numberingSystem = opts.NumberingSystem
	}
}

func (c config) validate(loc locale.Locale) error {
	checks := []ecma402.StringOption{
		ecma402.RequiredStringOption("style", c.style, "decimal", "percent", "currency", "unit"),
		ecma402.RequiredStringOption("notation", c.notation, "standard", "scientific", "engineering", "compact"),
		ecma402.RequiredStringOption("compactDisplay", c.compactDisplay, "short", "long"),
		ecma402.RequiredStringOption("currencyDisplay", c.currencyDisplay, "code", "symbol", "narrowSymbol", "name"),
		ecma402.RequiredStringOption("currencySign", c.currencySign, "standard", "accounting"),
		ecma402.RequiredStringOption("unitDisplay", c.unitDisplay, "short", "narrow", "long"),
		ecma402.RequiredStringOption("signDisplay", c.signDisplay, "auto", "always", "exceptZero", "negative", "never"),
		ecma402.RequiredStringOption("useGrouping", c.useGrouping, "min2", "auto", "always", "false"),
		ecma402.LocaleMatcherOption(c.localeMatcher),
	}
	if check, ok := ecma402.InvalidStringOption(checks...); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	if c.numberingSystem != "" && !ecma402.IsWellFormedUnicodeType(c.numberingSystem) {
		return invalidOption("numberingSystem", c.numberingSystem, loc)
	}
	if c.style == "currency" && c.currency == "" {
		return invalidOption("currency", c.currency, loc)
	}
	if c.currency != "" && !ecma402.IsWellFormedCurrencyCode(c.currency) {
		return invalidOption("currency", c.currency, loc)
	}
	if c.style == "unit" && c.unit == "" {
		return invalidOption("unit", c.unit, loc)
	}
	if c.unit != "" && !ecma402.IsWellFormedUnitIdentifier(c.unit) {
		return invalidOption("unit", c.unit, loc)
	}
	return nil
}

func (c config) digitOptionInput() ecma402nf.DigitOptionInput {
	return ecma402nf.DigitOptionInput{
		MinimumIntegerDigits:        c.minIntDigits,
		MinimumFractionDigits:       c.minFracDigits,
		MaximumFractionDigits:       c.maxFracDigits,
		MinimumSignificantDigits:    c.minSigDigits,
		MaximumSignificantDigits:    c.maxSigDigits,
		RoundingIncrement:           c.roundingIncrement,
		RoundingMode:                c.roundingMode,
		RoundingPriority:            c.roundingPriority,
		TrailingZeroDisplay:         c.trailingZeroDisplay,
		HasMinimumFractionDigits:    c.hasMinFracDigits,
		HasMaximumFractionDigits:    c.hasMaxFracDigits,
		HasMinimumSignificantDigits: c.hasMinSigDigits,
		HasMaximumSignificantDigits: c.hasMaxSigDigits,
	}
}

func (c *config) applyResolvedDigits(digits ecma402nf.ResolvedDigitOptions) {
	c.minIntDigits = digits.MinimumIntegerDigits
	c.minFracDigits = digits.MinimumFractionDigits
	c.maxFracDigits = digits.MaximumFractionDigits
	c.minSigDigits = digits.MinimumSignificantDigits
	c.maxSigDigits = digits.MaximumSignificantDigits
	c.roundingIncrement = digits.RoundingIncrement
	c.roundingMode = digits.RoundingMode
	c.roundingPriority = digits.RoundingPriority
	c.trailingZeroDisplay = digits.TrailingZeroDisplay
	c.roundingType = digits.RoundingType
}

func invalidOption(name, value string, loc locale.Locale) error {
	return ecma402.InvalidOptionError("numberformat", name, value, loc.Tag().String(), intlerr.ErrInvalidOption)
}
