package pluralrules

import (
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402nf "github.com/agentable/go-intl/internal/ecma402/numberformat"
)

// Type selects cardinal or ordinal plural rules.
type Type uint8
type LocaleMatcher string
type Notation string
type CompactDisplay string
type RoundingMode string
type RoundingPriority string
type TrailingZeroDisplay string

const (
	// Cardinal selects plural categories for quantities such as "1 item".
	Cardinal Type = iota
	// Ordinal selects plural categories for ordinals such as "1st".
	Ordinal
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

func (t Type) String() string {
	switch t {
	case Cardinal:
		return "cardinal"
	case Ordinal:
		return "ordinal"
	default:
		return "unknown"
	}
}

func (t Type) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

// Options configures PluralRules construction.
//
// The zero value uses cardinal rules with ECMA-402 default digit handling.
type Options struct {
	LocaleMatcher            LocaleMatcher
	Type                     Type
	MinimumIntegerDigits     *int
	MinimumFractionDigits    *int
	MaximumFractionDigits    *int
	MinimumSignificantDigits *int
	MaximumSignificantDigits *int
	RoundingIncrement        *int
	RoundingMode             RoundingMode
	RoundingPriority         RoundingPriority
	TrailingZeroDisplay      TrailingZeroDisplay
	Notation                 Notation
	CompactDisplay           CompactDisplay
}

type config struct {
	typ                 Type
	localeMatcher       string
	minIntDigits        int
	minFracDigits       int
	maxFracDigits       int
	hasMinFracDigits    bool
	hasMaxFracDigits    bool
	minSigDigits        int
	maxSigDigits        int
	hasMinSigDigits     bool
	hasMaxSigDigits     bool
	roundingIncrement   int
	roundingMode        string
	roundingPriority    string
	roundingType        ecma402nf.RoundingType
	trailingZeroDisplay string
	notation            string
	compactDisplay      string
}

func defaultConfig() config {
	return config{
		typ:                 Cardinal,
		localeMatcher:       string(BestFitLocaleMatcher),
		minIntDigits:        1,
		minFracDigits:       0,
		maxFracDigits:       3,
		roundingIncrement:   1,
		roundingMode:        string(HalfExpandRoundingMode),
		roundingPriority:    string(AutoRoundingPriority),
		trailingZeroDisplay: string(AutoTrailingZeroDisplay),
		notation:            string(StandardNotation),
		compactDisplay:      string(ShortCompactDisplay),
	}
}

func configFromOptions(opts Options) config {
	cfg := defaultConfig()
	if opts.LocaleMatcher != "" {
		cfg.localeMatcher = string(opts.LocaleMatcher)
	}
	cfg.typ = opts.Type
	if opts.MinimumIntegerDigits != nil {
		cfg.minIntDigits = *opts.MinimumIntegerDigits
	}
	if opts.MinimumFractionDigits != nil {
		cfg.minFracDigits = *opts.MinimumFractionDigits
		cfg.hasMinFracDigits = true
	}
	if opts.MaximumFractionDigits != nil {
		cfg.maxFracDigits = *opts.MaximumFractionDigits
		cfg.hasMaxFracDigits = true
	}
	if opts.MinimumSignificantDigits != nil {
		cfg.minSigDigits = *opts.MinimumSignificantDigits
		cfg.hasMinSigDigits = true
	}
	if opts.MaximumSignificantDigits != nil {
		cfg.maxSigDigits = *opts.MaximumSignificantDigits
		cfg.hasMaxSigDigits = true
	}
	if opts.RoundingIncrement != nil {
		cfg.roundingIncrement = *opts.RoundingIncrement
	}
	if opts.RoundingMode != "" {
		cfg.roundingMode = string(opts.RoundingMode)
	}
	if opts.RoundingPriority != "" {
		cfg.roundingPriority = string(opts.RoundingPriority)
	}
	if opts.TrailingZeroDisplay != "" {
		cfg.trailingZeroDisplay = string(opts.TrailingZeroDisplay)
	}
	if opts.Notation != "" {
		cfg.notation = string(opts.Notation)
	}
	if opts.CompactDisplay != "" {
		cfg.compactDisplay = string(opts.CompactDisplay)
	}
	return cfg
}

func (c config) digitOptionInput() ecma402nf.DigitOptionInput {
	return ecma402nf.DigitOptionInput{
		MinimumIntegerDigits:        c.minIntDigits,
		MinimumFractionDigits:       c.minFracDigits,
		MaximumFractionDigits:       c.maxFracDigits,
		MinimumSignificantDigits:    c.minSigDigits,
		MaximumSignificantDigits:    c.maxSigDigits,
		RoundingIncrement:           c.roundingIncrement,
		RoundingMode:                c.roundingMode,
		RoundingPriority:            c.roundingPriority,
		TrailingZeroDisplay:         c.trailingZeroDisplay,
		HasMinimumFractionDigits:    c.hasMinFracDigits,
		HasMaximumFractionDigits:    c.hasMaxFracDigits,
		HasMinimumSignificantDigits: c.hasMinSigDigits,
		HasMaximumSignificantDigits: c.hasMaxSigDigits,
	}
}

func (c *config) applyResolvedDigits(digits ecma402nf.ResolvedDigitOptions) {
	c.minIntDigits = digits.MinimumIntegerDigits
	c.minFracDigits = digits.MinimumFractionDigits
	c.maxFracDigits = digits.MaximumFractionDigits
	c.minSigDigits = digits.MinimumSignificantDigits
	c.maxSigDigits = digits.MaximumSignificantDigits
	c.roundingIncrement = digits.RoundingIncrement
	c.roundingMode = digits.RoundingMode
	c.roundingPriority = digits.RoundingPriority
	c.trailingZeroDisplay = digits.TrailingZeroDisplay
	c.roundingType = digits.RoundingType
}

func (c config) canUseIntegerOperands() bool {
	return c.notation == string(StandardNotation) &&
		c.minFracDigits == 0 &&
		!c.hasMinSigDigits &&
		!c.hasMaxSigDigits &&
		c.roundingIncrement == 1 &&
		c.roundingPriority == string(AutoRoundingPriority)
}

func (c config) validate() error {
	if check, ok := ecma402.InvalidStringOption(ecma402.LocaleMatcherOption(c.localeMatcher)); ok {
		return invalidOption(check.Name, check.Value)
	}
	switch c.typ {
	case Cardinal, Ordinal:
	default:
		return invalidOption("type", c.typ.String())
	}
	checks := []ecma402.StringOption{
		ecma402.RequiredStringOption("notation", c.notation, "standard", "scientific", "engineering", "compact"),
		ecma402.RequiredStringOption("compactDisplay", c.compactDisplay, "short", "long"),
	}
	if check, ok := ecma402.InvalidStringOption(checks...); ok {
		return invalidOption(check.Name, check.Value)
	}
	return nil
}

func typeFromString(s string) (Type, bool) {
	switch s {
	case "", "cardinal":
		return Cardinal, true
	case "ordinal":
		return Ordinal, true
	default:
		return 0, false
	}
}
