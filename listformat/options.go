package listformat

import (
	"github.com/agentable/go-intl/internal/ecma402"
)

const listFormatOwner = "listformat"

var (
	listTypeValues = [...]string{
		string(Conjunction),
		string(Disjunction),
		string(Unit),
	}
	listStyleValues = [...]string{
		string(LongStyle),
		string(ShortStyle),
		string(NarrowStyle),
	}
)

type Options struct {
	LocaleMatcher *string
	Type          *string
	Style         *string
}

type config struct {
	localeMatcher    string
	hasLocaleMatcher bool
	typ              string
	style            string
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		typ:           string(Conjunction),
		style:         string(LongStyle),
	}
}

func applyOptions(cfg *config, opts Options) {
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.hasLocaleMatcher, opts.LocaleMatcher)
	ecma402.ApplyOption(&cfg.typ, opts.Type)
	ecma402.ApplyOption(&cfg.style, opts.Style)
}

func (cfg config) validate(locName string) error {
	return ecma402.ValidateStringOptions(
		listFormatOwner,
		locName,
		ecma402.LocaleMatcherOptionInput(cfg.localeMatcher, cfg.hasLocaleMatcher),
		listTypeOption(cfg.typ),
		listStyleOption(cfg.style),
	)
}

func listTypeOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("type", value, listTypeValues[:]...)
}

func listStyleOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("style", value, listStyleValues[:]...)
}
