package collator

import (
	"github.com/agentable/go-intl/internal/ecma402"
)

type LocaleMatcher string
type Usage string
type Sensitivity string
type CaseFirst string

const collatorOwner = "collator"

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

var (
	collatorUsageValues = [...]string{
		string(SortUsage),
		string(SearchUsage),
	}
	supportedCollatorUsageValues = [...]string{
		string(SortUsage),
	}
	collatorSensitivityValues = [...]string{
		string(BaseSensitivity),
		string(AccentSensitivity),
		string(CaseSensitivity),
		string(VariantSensitivity),
	}
	collatorCaseFirstValues = [...]string{
		string(FalseCaseFirst),
		string(UpperCaseFirst),
		string(LowerCaseFirst),
	}
)

// Options mirrors the JS Intl.Collator options bag.
//
// A zero value means "no caller preference" for every field; defaults defined
// by ECMA-402 (Usage=SortUsage, Sensitivity=VariantSensitivity for sort) are
// applied during construction.
type Options struct {
	LocaleMatcher     *string
	Usage             *string
	Sensitivity       *string
	CaseFirst         *string
	Numeric           *bool
	IgnorePunctuation *bool
	Collation         *string
}

type config struct {
	localeMatcher     string
	hasLocaleMatcher  bool
	usage             string
	hasSensitivity    bool
	sensitivity       string
	hasCaseFirst      bool
	caseFirst         string
	numeric           bool
	hasNumeric        bool
	ignorePunctuation bool
	collation         string
	hasCollation      bool
}

func defaultConfig() config {
	return config{
		localeMatcher: string(BestFitLocaleMatcher),
		usage:         string(SortUsage),
	}
}

func applyOptions(cfg *config, opts Options) {
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.hasLocaleMatcher, opts.LocaleMatcher)
	ecma402.ApplyOption(&cfg.usage, opts.Usage)
	ecma402.ApplyOptionInput(&cfg.sensitivity, &cfg.hasSensitivity, opts.Sensitivity)
	ecma402.ApplyOptionInput(&cfg.caseFirst, &cfg.hasCaseFirst, opts.CaseFirst)
	ecma402.ApplyOptionInput(&cfg.numeric, &cfg.hasNumeric, opts.Numeric)
	ecma402.ApplyOption(&cfg.ignorePunctuation, opts.IgnorePunctuation)
	ecma402.ApplyUnicodeTypeOptionInput(&cfg.collation, &cfg.hasCollation, "co", opts.Collation)
}

func (cfg config) validate(locName string) error {
	if err := ecma402.ValidateStringOptions(
		collatorOwner,
		locName,
		ecma402.LocaleMatcherOptionInput(cfg.localeMatcher, cfg.hasLocaleMatcher),
		usageOption(cfg.usage),
	); err != nil {
		return err
	}
	if err := ecma402.ValidateSupportedStringOptions(collatorOwner, locName, supportedUsageOption(cfg.usage)); err != nil {
		return err
	}
	if err := ecma402.ValidateStringOptions(
		collatorOwner,
		locName,
		sensitivityOptionInput(cfg.sensitivity, cfg.hasSensitivity),
		caseFirstOptionInput(cfg.caseFirst, cfg.hasCaseFirst),
	); err != nil {
		return err
	}
	if cfg.caseFirst != "" {
		if err := ecma402.ValidateSupportedStringOptions(collatorOwner, locName, supportedCaseFirstOption(cfg.caseFirst)); err != nil {
			return err
		}
	}
	if cfg.ignorePunctuation {
		return ecma402.UnsupportedOptionErrorExpected(
			collatorOwner,
			"ignorePunctuation",
			"true",
			locName,
			"false or omitted until the active backend supports UCA alternate-shifted handling",
			nil,
		)
	}
	return ecma402.ValidateUnicodeTypeOptionInput(collatorOwner, "collation", cfg.collation, locName, cfg.hasCollation)
}

func usageOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("usage", value, collatorUsageValues[:]...)
}

func supportedUsageOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("usage", value, supportedCollatorUsageValues[:]...)
}

func sensitivityOptionInput(value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput(
		"sensitivity",
		value,
		present,
		collatorSensitivityValues[:]...,
	)
}

func caseFirstOptionInput(value string, present bool) ecma402.StringOption {
	return ecma402.OptionalStringOptionInput(
		"caseFirst",
		value,
		present,
		collatorCaseFirstValues[:]...,
	)
}

func supportedCaseFirstOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("caseFirst", value, supportedCollatorCaseFirstValues[:]...)
}
