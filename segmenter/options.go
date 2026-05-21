package segmenter

import (
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/locale"
)

type LocaleMatcher string
type Granularity string

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	GraphemeGranularity Granularity = "grapheme"
	WordGranularity     Granularity = "word"
	SentenceGranularity Granularity = "sentence"
)

// Options mirrors the JS Intl.Segmenter options bag.
type Options struct {
	LocaleMatcher LocaleMatcher
	Granularity   Granularity
}

type config struct {
	localeMatcher string
	granularity   string
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		granularity:   string(GraphemeGranularity),
	}
}

func applyOptions(cfg *config, opts Options) {
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	if opts.Granularity != "" {
		cfg.granularity = string(opts.Granularity)
	}
}

func (cfg config) validate(loc locale.Locale) error {
	if check, ok := ecma402.InvalidStringOption(
		ecma402.LocaleMatcherOption(cfg.localeMatcher),
		ecma402.RequiredStringOption("granularity", cfg.granularity, string(GraphemeGranularity), string(WordGranularity), string(SentenceGranularity)),
	); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	return nil
}

func invalidOption(name, value string, loc locale.Locale) error {
	return ecma402.InvalidOptionError("segmenter", name, value, loc.String(), intlerr.ErrInvalidOption)
}
