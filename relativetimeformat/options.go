package relativetimeformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
)

const relativeTimeFormatOwner = "relativetimeformat"

var (
	relativeTimeStyleValues = [...]string{
		string(LongStyle),
		string(ShortStyle),
		string(NarrowStyle),
	}
	relativeTimeNumericValues = [...]string{
		string(NumericAlways),
		string(NumericAuto),
	}
)

type Options struct {
	LocaleMatcher   *string
	NumberingSystem *string
	Style           *string
	Numeric         *string
}

type config struct {
	localeMatcher      string
	localeMatcherSet   bool
	numberingSystem    string
	numberingSystemSet bool
	style              string
	numeric            string
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		style:         string(LongStyle),
		numeric:       string(NumericAlways),
	}
}

func applyOptions(cfg *config, opts Options) {
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.localeMatcherSet, opts.LocaleMatcher)
	ecma402.ApplyOptionInput(&cfg.numberingSystem, &cfg.numberingSystemSet, opts.NumberingSystem)
	ecma402.ApplyOption(&cfg.style, opts.Style)
	ecma402.ApplyOption(&cfg.numeric, opts.Numeric)
}

func (cfg config) validate(locName string) error {
	if err := ecma402.ValidateStringOptions(
		relativeTimeFormatOwner,
		locName,
		ecma402.LocaleMatcherOptionInput(cfg.localeMatcher, cfg.localeMatcherSet),
		relativeTimeStyleOption(cfg.style),
		relativeTimeNumericOption(cfg.numeric),
	); err != nil {
		return err
	}
	return ecma402.ValidateUnicodeTypeOptionInput(relativeTimeFormatOwner, "numberingSystem", cfg.numberingSystem, locName, cfg.numberingSystemSet)
}

func relativeTimeStyleOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("style", value, relativeTimeStyleValues[:]...)
}

func relativeTimeNumericOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("numeric", value, relativeTimeNumericValues[:]...)
}
