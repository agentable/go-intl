package numberformat

import (
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

type Options struct {
	Style                Style
	Currency             Currency
	CurrencyDisplay      CurrencyDisplay
	CurrencySign         CurrencySign
	Unit                 Unit
	UnitDisplay          UnitDisplay
	MinimumIntegerDigits int
	FractionDigits       FractionDigitOptions
	SignificantDigits    SignificantDigitOptions
	RoundingIncrement    int
	RoundingPriority     RoundingPriority
	RoundingMode         RoundingMode
	TrailingZeroDisplay  TrailingZeroDisplay
	Notation             Notation
	CompactDisplay       CompactDisplay
	UseGrouping          UseGrouping
	SignDisplay          SignDisplay
	LocaleMatcher        LocaleMatcher
	NumberingSystem      string
}

type FractionDigitOptions struct {
	min    int
	max    int
	hasMin bool
	hasMax bool
}

func FractionDigits(minimum, maximum int) FractionDigitOptions {
	return FractionDigitOptions{min: minimum, max: maximum, hasMin: true, hasMax: true}
}

func MinimumFractionDigits(n int) FractionDigitOptions {
	return FractionDigitOptions{min: n, hasMin: true}
}

func MaximumFractionDigits(n int) FractionDigitOptions {
	return FractionDigitOptions{max: n, hasMax: true}
}

type SignificantDigitOptions struct {
	min    int
	max    int
	hasMin bool
	hasMax bool
}

func SignificantDigits(minimum, maximum int) SignificantDigitOptions {
	return SignificantDigitOptions{min: minimum, max: maximum, hasMin: true, hasMax: true}
}

func MinimumSignificantDigits(n int) SignificantDigitOptions {
	return SignificantDigitOptions{min: n, hasMin: true}
}

func MaximumSignificantDigits(n int) SignificantDigitOptions {
	return SignificantDigitOptions{max: n, hasMax: true}
}

func CurrencyCode(code string) Currency {
	return Currency(strings.ToUpper(code))
}

func UnitIdentifier(unit string) Unit {
	return Unit(unit)
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
	if opts.MinimumIntegerDigits != 0 {
		cfg.minIntDigits = opts.MinimumIntegerDigits
	}
	if opts.FractionDigits.hasMin {
		cfg.minFracDigits = opts.FractionDigits.min
		cfg.hasMinFracDigits = true
	}
	if opts.FractionDigits.hasMax {
		cfg.maxFracDigits = opts.FractionDigits.max
		cfg.hasMaxFracDigits = true
	}
	if opts.SignificantDigits.hasMin {
		cfg.minSigDigits = opts.SignificantDigits.min
		cfg.hasMinSigDigits = true
	}
	if opts.SignificantDigits.hasMax {
		cfg.maxSigDigits = opts.SignificantDigits.max
		cfg.hasMaxSigDigits = true
	}
	if opts.RoundingIncrement != 0 {
		cfg.roundingIncrement = opts.RoundingIncrement
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
	checks := []struct {
		name   string
		value  string
		values []string
	}{
		{name: "style", value: c.style, values: []string{"decimal", "percent", "currency", "unit"}},
		{name: "notation", value: c.notation, values: []string{"standard", "scientific", "engineering", "compact"}},
		{name: "compactDisplay", value: c.compactDisplay, values: []string{"short", "long"}},
		{name: "currencyDisplay", value: c.currencyDisplay, values: []string{"code", "symbol", "narrowSymbol", "name"}},
		{name: "currencySign", value: c.currencySign, values: []string{"standard", "accounting"}},
		{name: "unitDisplay", value: c.unitDisplay, values: []string{"short", "narrow", "long"}},
		{name: "signDisplay", value: c.signDisplay, values: []string{"auto", "always", "exceptZero", "negative", "never"}},
		{name: "roundingMode", value: c.roundingMode, values: []string{"ceil", "floor", "expand", "trunc", "halfCeil", "halfFloor", "halfExpand", "halfTrunc", "halfEven"}},
		{name: "roundingPriority", value: c.roundingPriority, values: []string{"auto", "morePrecision", "lessPrecision"}},
		{name: "trailingZeroDisplay", value: c.trailingZeroDisplay, values: []string{"auto", "stripIfInteger"}},
		{name: "localeMatcher", value: c.localeMatcher, values: []string{"lookup", "best fit"}},
		{name: "useGrouping", value: c.useGrouping, values: []string{"min2", "auto", "always", "false"}},
	}
	for _, check := range checks {
		if !slices.Contains(check.values, check.value) {
			return invalidOption(check.name, check.value, loc)
		}
	}
	if c.numberingSystem != "" && !ecma402.IsWellFormedUnicodeType(c.numberingSystem) {
		return invalidOption("numberingSystem", c.numberingSystem, loc)
	}
	if c.minIntDigits < 1 || c.minIntDigits > 21 {
		return invalidOption("minimumIntegerDigits", fmt.Sprint(c.minIntDigits), loc)
	}
	if c.minFracDigits < 0 || c.minFracDigits > 100 {
		return invalidOption("minimumFractionDigits", fmt.Sprint(c.minFracDigits), loc)
	}
	if c.maxFracDigits < 0 || c.maxFracDigits > 100 || c.maxFracDigits < c.minFracDigits {
		return invalidOption("maximumFractionDigits", fmt.Sprint(c.maxFracDigits), loc)
	}
	if c.hasMinSigDigits && (c.minSigDigits < 1 || c.minSigDigits > 21) {
		return invalidOption("minimumSignificantDigits", fmt.Sprint(c.minSigDigits), loc)
	}
	if c.hasMaxSigDigits && (c.maxSigDigits < 1 || c.maxSigDigits > 21 || (c.hasMinSigDigits && c.maxSigDigits < c.minSigDigits)) {
		return invalidOption("maximumSignificantDigits", fmt.Sprint(c.maxSigDigits), loc)
	}
	if !decimal.IsValidRoundingIncrement(c.roundingIncrement) {
		return invalidOption("roundingIncrement", fmt.Sprint(c.roundingIncrement), loc)
	}
	if c.roundingIncrement != 1 {
		if c.hasMinSigDigits || c.hasMaxSigDigits || c.roundingPriority != "auto" || c.minFracDigits != c.maxFracDigits {
			return invalidOption("roundingIncrement", fmt.Sprint(c.roundingIncrement), loc)
		}
	}
	if c.style == "currency" && c.currency == "" {
		return fmt.Errorf("numberformat: invalid currency %q for locale %q: %w", c.currency, loc.Tag().String(), ErrInvalidOption)
	}
	if c.currency != "" && !validCurrencyCode(c.currency) {
		return invalidOption("currency", c.currency, loc)
	}
	if c.style == "unit" && !ecma402.IsWellFormedUnitIdentifier(c.unit) {
		return invalidOption("unit", c.unit, loc)
	}
	return nil
}

func invalidOption(name, value string, loc locale.Locale) error {
	return fmt.Errorf("numberformat: invalid %s %q for locale %q: %w", name, value, loc.Tag().String(), ErrInvalidOption)
}

func validCurrencyCode(code string) bool {
	if len(code) != 3 {
		return false
	}
	for _, r := range code {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}
