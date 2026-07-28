package pluralrules

import (
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
)

const pluralRulesOwner = "pluralrules"

// Type selects cardinal or ordinal plural rules.
type Type string
type LocaleMatcher string
type Notation string
type CompactDisplay string
type RoundingMode string
type RoundingPriority string
type TrailingZeroDisplay string

const (
	// Cardinal selects plural categories for quantities such as "1 item".
	Cardinal Type = "cardinal"
	// Ordinal selects plural categories for ordinals such as "1st".
	Ordinal Type = "ordinal"
)

const (
	LookupLocaleMatcher  LocaleMatcher = "lookup"
	BestFitLocaleMatcher LocaleMatcher = "best fit"

	StandardNotation    Notation = "standard"
	ScientificNotation  Notation = "scientific"
	EngineeringNotation Notation = "engineering"
	CompactNotation     Notation = "compact"

	ShortCompactDisplay CompactDisplay = "short"
	LongCompactDisplay  CompactDisplay = "long"

	CeilRoundingMode       RoundingMode = "ceil"
	FloorRoundingMode      RoundingMode = "floor"
	ExpandRoundingMode     RoundingMode = "expand"
	TruncRoundingMode      RoundingMode = "trunc"
	HalfCeilRoundingMode   RoundingMode = "halfCeil"
	HalfFloorRoundingMode  RoundingMode = "halfFloor"
	HalfExpandRoundingMode RoundingMode = "halfExpand"
	HalfTruncRoundingMode  RoundingMode = "halfTrunc"
	HalfEvenRoundingMode   RoundingMode = "halfEven"

	AutoRoundingPriority          RoundingPriority = "auto"
	MorePrecisionRoundingPriority RoundingPriority = "morePrecision"
	LessPrecisionRoundingPriority RoundingPriority = "lessPrecision"

	AutoTrailingZeroDisplay           TrailingZeroDisplay = "auto"
	StripIfIntegerTrailingZeroDisplay TrailingZeroDisplay = "stripIfInteger"
)

var pluralRuleTypeValues = [...]string{
	string(Cardinal),
	string(Ordinal),
}

func (t Type) String() string {
	if t == "" {
		return string(Cardinal)
	}
	return string(t)
}

func (t Type) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// Options configures PluralRules construction.
//
// The zero value uses cardinal rules with ECMA-402 default digit handling.
type Options struct {
	LocaleMatcher            *string
	Type                     *string
	MinimumIntegerDigits     *int
	MinimumFractionDigits    *int
	MaximumFractionDigits    *int
	MinimumSignificantDigits *int
	MaximumSignificantDigits *int
	RoundingIncrement        *int
	RoundingMode             *string
	RoundingPriority         *string
	TrailingZeroDisplay      *string
	Notation                 *string
	CompactDisplay           *string
}

type config struct {
	typ              string
	localeMatcher    string
	hasLocaleMatcher bool
	digits           ecma402nf.DigitOptionConfig
	notation         string
	compactDisplay   string
}

func defaultConfig() config {
	return config{
		typ:            string(Cardinal),
		localeMatcher:  string(BestFitLocaleMatcher),
		digits:         ecma402nf.DefaultDigitOptionConfig(),
		notation:       string(StandardNotation),
		compactDisplay: string(ShortCompactDisplay),
	}
}

func configFromOptions(opts Options) config {
	cfg := defaultConfig()
	ecma402.ApplyOptionInput(&cfg.localeMatcher, &cfg.hasLocaleMatcher, opts.LocaleMatcher)
	ecma402.ApplyOption(&cfg.typ, opts.Type)
	cfg.digits.ApplyOverrides(ecma402nf.DigitOptionOverrides{
		MinimumIntegerDigits:     opts.MinimumIntegerDigits,
		MinimumFractionDigits:    opts.MinimumFractionDigits,
		MaximumFractionDigits:    opts.MaximumFractionDigits,
		MinimumSignificantDigits: opts.MinimumSignificantDigits,
		MaximumSignificantDigits: opts.MaximumSignificantDigits,
		RoundingIncrement:        opts.RoundingIncrement,
		RoundingMode:             opts.RoundingMode,
		RoundingPriority:         opts.RoundingPriority,
		TrailingZeroDisplay:      opts.TrailingZeroDisplay,
	})
	ecma402.ApplyOption(&cfg.notation, opts.Notation)
	ecma402.ApplyOption(&cfg.compactDisplay, opts.CompactDisplay)
	return cfg
}

func pluralRuleTypeOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("type", value, pluralRuleTypeValues[:]...)
}

func (c config) validate(locName string) error {
	return ecma402.ValidateStringOptions(
		pluralRulesOwner,
		locName,
		ecma402.LocaleMatcherOptionInput(c.localeMatcher, c.hasLocaleMatcher),
		pluralRuleTypeOption(c.typ),
		ecma402nf.NotationOption(c.notation),
		ecma402nf.CompactDisplayOption(c.compactDisplay),
	)
}
