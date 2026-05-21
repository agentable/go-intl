package collator

import (
	"strings"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/locale"
)

type LocaleMatcher string
type Usage string
type Sensitivity string
type CaseFirst string

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	SortUsage   Usage = "sort"
	SearchUsage Usage = "search"

	BaseSensitivity    Sensitivity = "base"
	AccentSensitivity  Sensitivity = "accent"
	CaseSensitivity    Sensitivity = "case"
	VariantSensitivity Sensitivity = "variant"

	UpperCaseFirst CaseFirst = "upper"
	LowerCaseFirst CaseFirst = "lower"
	FalseCaseFirst CaseFirst = "false"
)

// Options mirrors the JS Intl.Collator options bag.
//
// A zero value means "no caller preference" for every field; defaults defined
// by ECMA-402 (Usage=SortUsage, Sensitivity=VariantSensitivity for sort) are
// applied during construction.
type Options struct {
	LocaleMatcher     LocaleMatcher
	Usage             Usage
	Sensitivity       Sensitivity
	CaseFirst         CaseFirst
	Numeric           *bool
	IgnorePunctuation *bool
	Collation         string
}

type config struct {
	localeMatcher     string
	usage             string
	sensitivity       string
	caseFirst         string
	numeric           bool
	numericSet        bool
	ignorePunctuation bool
	collation         string
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		usage:         string(SortUsage),
	}
}

func applyOptions(cfg *config, opts Options) {
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	if opts.Usage != "" {
		cfg.usage = string(opts.Usage)
	}
	if opts.Sensitivity != "" {
		cfg.sensitivity = string(opts.Sensitivity)
	}
	if opts.CaseFirst != "" {
		cfg.caseFirst = string(opts.CaseFirst)
	}
	if opts.Numeric != nil {
		cfg.numeric = *opts.Numeric
		cfg.numericSet = true
	}
	if opts.IgnorePunctuation != nil {
		cfg.ignorePunctuation = *opts.IgnorePunctuation
	}
	if opts.Collation != "" {
		cfg.collation = strings.ToLower(opts.Collation)
	}
}

func (cfg config) validate(loc locale.Locale) error {
	if check, ok := ecma402.InvalidStringOption(ecma402.LocaleMatcherOption(cfg.localeMatcher)); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	switch cfg.usage {
	case string(SortUsage):
	case string(SearchUsage):
		return unsupportedOption("usage", cfg.usage, loc)
	default:
		return invalidOption("usage", cfg.usage, loc)
	}
	if check, ok := ecma402.InvalidStringOption(
		ecma402.OptionalStringOption("sensitivity", cfg.sensitivity, string(BaseSensitivity), string(AccentSensitivity), string(CaseSensitivity), string(VariantSensitivity)),
		ecma402.OptionalStringOption("caseFirst", cfg.caseFirst, string(FalseCaseFirst), string(UpperCaseFirst), string(LowerCaseFirst)),
	); ok {
		return invalidOption(check.Name, check.Value, loc)
	}
	if cfg.caseFirst == string(UpperCaseFirst) || cfg.caseFirst == string(LowerCaseFirst) {
		return unsupportedOption("caseFirst", cfg.caseFirst, loc)
	}
	if cfg.collation != "" {
		if !ecma402.IsWellFormedUnicodeType(cfg.collation) {
			return invalidOption("collation", cfg.collation, loc)
		}
		if !isDefaultCollation(cfg.collation) {
			return unsupportedOption("collation", cfg.collation, loc)
		}
	}
	return nil
}

func isDefaultCollation(value string) bool {
	return value == "default" || value == "standard" || value == "search"
}

func invalidOption(name, value string, loc locale.Locale) error {
	return ecma402.InvalidOptionError("collator", name, value, loc.String(), intlerr.ErrInvalidOption)
}

func unsupportedOption(name, value string, loc locale.Locale) error {
	return ecma402.UnsupportedOptionError("collator", name, value, loc.String(), intlerr.ErrUnsupportedOption)
}
