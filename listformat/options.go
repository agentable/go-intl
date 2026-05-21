package listformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

type Options struct {
	LocaleMatcher LocaleMatcher
	Type          Type
	Style         Style
}

type config struct {
	localeMatcher string
	typ           string
	style         string
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		typ:           string(Conjunction),
		style:         string(LongStyle),
	}
}

func applyOptions(cfg *config, opts Options) {
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	if opts.Type != "" {
		cfg.typ = string(opts.Type)
	}
	if opts.Style != "" {
		cfg.style = string(opts.Style)
	}
}

func (cfg config) validate(loc locale.Locale) error {
	if check, ok := ecma402.InvalidStringOption(
		ecma402.LocaleMatcherOption(cfg.localeMatcher),
		ecma402.RequiredStringOption("type", cfg.typ, string(Conjunction), string(Disjunction), string(Unit)),
		ecma402.RequiredStringOption("style", cfg.style, string(LongStyle), string(ShortStyle), string(NarrowStyle)),
	); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	return nil
}

func invalidOption(name, value string, loc locale.Locale) error {
	return ecma402.InvalidOptionError("listformat", name, value, loc.String(), intlerr.ErrInvalidOption)
}
