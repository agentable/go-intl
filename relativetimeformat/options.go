package relativetimeformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

type Options struct {
	LocaleMatcher   LocaleMatcher
	NumberingSystem string
	Style           Style
	Numeric         Numeric
}

type config struct {
	localeMatcher   string
	numberingSystem string
	style           string
	numeric         string
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		style:         string(LongStyle),
		numeric:       string(NumericAlways),
	}
}

func applyOptions(cfg *config, opts Options) {
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	if opts.NumberingSystem != "" {
		cfg.numberingSystem = opts.NumberingSystem
	}
	if opts.Style != "" {
		cfg.style = string(opts.Style)
	}
	if opts.Numeric != "" {
		cfg.numeric = string(opts.Numeric)
	}
}

func (cfg config) validate(loc locale.Locale) error {
	if check, ok := ecma402.InvalidStringOption(
		ecma402.LocaleMatcherOption(cfg.localeMatcher),
		ecma402.RequiredStringOption("style", cfg.style, string(LongStyle), string(ShortStyle), string(NarrowStyle)),
		ecma402.RequiredStringOption("numeric", cfg.numeric, string(NumericAlways), string(NumericAuto)),
	); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	if cfg.numberingSystem != "" && !ecma402.IsWellFormedUnicodeType(cfg.numberingSystem) {
		return invalidOption("numberingSystem", cfg.numberingSystem, loc)
	}
	return nil
}

func invalidOption(name, value string, loc locale.Locale) error {
	return ecma402.InvalidOptionError("relativetimeformat", name, value, loc.String(), intlerr.ErrInvalidOption)
}
