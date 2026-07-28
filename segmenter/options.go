package segmenter

import (
	"github.com/agentable/go-intl/internal/ecma402"
)

type LocaleMatcher string
type Granularity string

const segmenterOwner = "segmenter"

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	GraphemeGranularity Granularity = "grapheme"
	WordGranularity     Granularity = "word"
	SentenceGranularity Granularity = "sentence"
)

// Options mirrors the JS Intl.Segmenter options bag.
type Options struct {
	LocaleMatcher *string
	Granularity   *string
}

type config struct {
	localeMatcher    string
	hasLocaleMatcher bool
	granularity      string
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		granularity:   string(GraphemeGranularity),
	}
}

func applyOptions(cfg *config, opts Options) {
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.hasLocaleMatcher, opts.LocaleMatcher)
	ecma402.ApplyOption(&cfg.granularity, opts.Granularity)
}

func (cfg config) validate(locName string) error {
	return ecma402.ValidateStringOptions(
		segmenterOwner,
		locName,
		ecma402.LocaleMatcherOptionInput(cfg.localeMatcher, cfg.hasLocaleMatcher),
		segmenterGranularityOption(cfg.granularity),
	)
}

func segmenterGranularityOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption(
		"granularity",
		value,
		string(GraphemeGranularity),
		string(WordGranularity),
		string(SentenceGranularity),
	)
}
