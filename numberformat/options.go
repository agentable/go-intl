package numberformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
)

const numberFormatOwner = "numberformat"

type Options struct {
	Style                    *string
	Currency                 *string
	CurrencyDisplay          *string
	CurrencySign             *string
	Unit                     *string
	UnitDisplay              *string
	MinimumIntegerDigits     *int
	MinimumFractionDigits    *int
	MaximumFractionDigits    *int
	MinimumSignificantDigits *int
	MaximumSignificantDigits *int
	RoundingIncrement        *int
	RoundingPriority         *string
	RoundingMode             *string
	TrailingZeroDisplay      *string
	Notation                 *string
	CompactDisplay           *string
	UseGrouping              *string
	SignDisplay              *string
	LocaleMatcher            *string
	NumberingSystem          *string
}

type config struct {
	style              string
	currency           string
	hasCurrency        bool
	currencyDisplay    string
	currencySign       string
	unit               string
	hasUnit            bool
	unitDisplay        string
	digits             ecma402nf.DigitOptionConfig
	notation           string
	compactDisplay     string
	useGrouping        string
	signDisplay        string
	localeMatcher      string
	hasLocaleMatcher   bool
	numberingSystem    string
	hasNumberingSystem bool
}

func defaultConfig() config {
	return config{
		style:           string(DecimalStyle),
		currencyDisplay: string(CurrencyDisplaySymbol),
		currencySign:    string(StandardCurrencySign),
		unitDisplay:     string(ShortUnitDisplay),
		digits:          ecma402nf.DefaultDigitOptionConfig(),
		notation:        string(StandardNotation),
		compactDisplay:  string(ShortCompactDisplay),
		useGrouping:     string(UseGroupingAuto),
		signDisplay:     string(AutoSignDisplay),
		localeMatcher:   string(BestFitLocaleMatcher),
	}
}

func applyOptions(cfg *config, opts Options) {
	ecma402.ApplyOption(&cfg.style, opts.Style)
	ecma402.ApplyCurrencyCodeOptionInput(&cfg.currency, &cfg.hasCurrency, opts.Currency)
	ecma402.ApplyOption(&cfg.currencyDisplay, opts.CurrencyDisplay)
	ecma402.ApplyOption(&cfg.currencySign, opts.CurrencySign)
	ecma402.ApplyOptionInput(&cfg.unit, &cfg.hasUnit, opts.Unit)
	ecma402.ApplyOption(&cfg.unitDisplay, opts.UnitDisplay)
	cfg.digits.ApplyOverrides(ecma402nf.DigitOptionOverrides{
		MinimumIntegerDigits:     opts.MinimumIntegerDigits,
		MinimumFractionDigits:    opts.MinimumFractionDigits,
		MaximumFractionDigits:    opts.MaximumFractionDigits,
		MinimumSignificantDigits: opts.MinimumSignificantDigits,
		MaximumSignificantDigits: opts.MaximumSignificantDigits,
		RoundingIncrement:        opts.RoundingIncrement,
		RoundingPriority:         opts.RoundingPriority,
		RoundingMode:             opts.RoundingMode,
		TrailingZeroDisplay:      opts.TrailingZeroDisplay,
	})
	ecma402.ApplyOption(&cfg.notation, opts.Notation)
	ecma402.ApplyOption(&cfg.compactDisplay, opts.CompactDisplay)
	ecma402.ApplyOption(&cfg.useGrouping, opts.UseGrouping)
	ecma402.ApplyOption(&cfg.signDisplay, opts.SignDisplay)
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.hasLocaleMatcher, opts.LocaleMatcher)
	ecma402.ApplyUnicodeTypeOptionInput(&cfg.numberingSystem, &cfg.hasNumberingSystem, "nu", opts.NumberingSystem)
}

func (c config) validate(locName string) error {
	if err := ecma402.ValidateStringOptions(
		numberFormatOwner,
		locName,
		styleOption(c.style),
		ecma402nf.NotationOption(c.notation),
		ecma402nf.CompactDisplayOption(c.compactDisplay),
		currencyDisplayOption(c.currencyDisplay),
		currencySignOption(c.currencySign),
		unitDisplayOption(c.unitDisplay),
		signDisplayOption(c.signDisplay),
		useGroupingOption(c.useGrouping),
		ecma402.LocaleMatcherOptionInput(c.localeMatcher, c.hasLocaleMatcher),
	); err != nil {
		return err
	}
	if err := ecma402.ValidateUnicodeTypeOptionInput(numberFormatOwner, "numberingSystem", c.numberingSystem, locName, c.hasNumberingSystem); err != nil {
		return err
	}
	if c.hasCurrency && !ecma402.IsWellFormedCurrencyCode(c.currency) {
		return ecma402.InvalidCurrencyCodeOptionError(numberFormatOwner, "currency", c.currency, locName)
	}
	if c.style == string(CurrencyStyle) && !c.hasCurrency {
		return missingStyleOptionError("currency", `a currency code when style is "currency"`, locName)
	}
	if c.hasUnit && !ecma402.IsWellFormedUnitIdentifier(c.unit) {
		return ecma402.InvalidUnitIdentifierOptionError(numberFormatOwner, "unit", c.unit, locName)
	}
	if c.style == string(UnitStyle) && !c.hasUnit {
		return missingStyleOptionError("unit", `a sanctioned unit identifier when style is "unit"`, locName)
	}
	return nil
}

func styleOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption(
		"style",
		value,
		string(DecimalStyle),
		string(PercentStyle),
		string(CurrencyStyle),
		string(UnitStyle),
	)
}

func currencyDisplayOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption(
		"currencyDisplay",
		value,
		string(CurrencyDisplayCode),
		string(CurrencyDisplaySymbol),
		string(CurrencyDisplayNarrowSymbol),
		string(CurrencyDisplayName),
	)
}

func currencySignOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption(
		"currencySign",
		value,
		string(StandardCurrencySign),
		string(AccountingCurrencySign),
	)
}

func unitDisplayOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption(
		"unitDisplay",
		value,
		string(ShortUnitDisplay),
		string(NarrowUnitDisplay),
		string(LongUnitDisplay),
	)
}

func signDisplayOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption(
		"signDisplay",
		value,
		string(AutoSignDisplay),
		string(AlwaysSignDisplay),
		string(ExceptZeroSignDisplay),
		string(NegativeSignDisplay),
		string(NeverSignDisplay),
	)
}

func useGroupingOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption(
		"useGrouping",
		value,
		string(UseGroupingMin2),
		string(UseGroupingAuto),
		string(UseGroupingAlways),
		string(UseGroupingFalse),
	)
}

func missingStyleOptionError(name, expected, loc string) error {
	return ecma402.InvalidOptionErrorExpected(numberFormatOwner, name, "", loc, expected, nil)
}
