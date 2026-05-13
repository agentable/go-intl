package pluralrules

import (
	"fmt"

	"github.com/agentable/go-intl/internal/cachekey"
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

// Options configures PluralRules construction.
//
// The zero value uses cardinal rules with ECMA-402 default digit handling.
type Options struct {
	LocaleMatcher        LocaleMatcher
	Type                 Type
	MinimumIntegerDigits int
	FractionDigits       FractionDigitOptions
	SignificantDigits    SignificantDigitOptions
	RoundingIncrement    int
	RoundingMode         RoundingMode
	RoundingPriority     RoundingPriority
	TrailingZeroDisplay  TrailingZeroDisplay
	Notation             Notation
	CompactDisplay       CompactDisplay
}

// FractionDigitOptions configures visible fraction digit handling.
type FractionDigitOptions struct {
	min    int
	max    int
	hasMin bool
	hasMax bool
}

// FractionDigits fixes the minimum and maximum visible fraction digits.
func FractionDigits(minimum, maximum int) FractionDigitOptions {
	return FractionDigitOptions{min: minimum, max: maximum, hasMin: true, hasMax: true}
}

// MinimumFractionDigits sets the minimum visible fraction digits.
func MinimumFractionDigits(n int) FractionDigitOptions {
	return FractionDigitOptions{min: n, hasMin: true}
}

// MaximumFractionDigits sets the maximum visible fraction digits.
func MaximumFractionDigits(n int) FractionDigitOptions {
	return FractionDigitOptions{max: n, hasMax: true}
}

// SignificantDigitOptions configures significant digit handling.
type SignificantDigitOptions struct {
	min    int
	max    int
	hasMin bool
	hasMax bool
}

func SignificantDigits(minimum, maximum int) SignificantDigitOptions {
	return SignificantDigitOptions{min: minimum, max: maximum, hasMin: true, hasMax: true}
}

func MinimumSignificantDigits(n int) SignificantDigitOptions {
	return SignificantDigitOptions{min: n, hasMin: true}
}

func MaximumSignificantDigits(n int) SignificantDigitOptions {
	return SignificantDigitOptions{max: n, hasMax: true}
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
	if opts.MinimumIntegerDigits != 0 {
		cfg.minIntDigits = opts.MinimumIntegerDigits
	}
	if opts.FractionDigits.hasMin {
		cfg.minFracDigits = opts.FractionDigits.min
		cfg.hasMinFracDigits = true
	}
	if opts.FractionDigits.hasMax {
		cfg.maxFracDigits = opts.FractionDigits.max
		cfg.hasMaxFracDigits = true
	}
	if opts.SignificantDigits.hasMin {
		cfg.minSigDigits = opts.SignificantDigits.min
		cfg.hasMinSigDigits = true
	}
	if opts.SignificantDigits.hasMax {
		cfg.maxSigDigits = opts.SignificantDigits.max
		cfg.hasMaxSigDigits = true
	}
	if opts.RoundingIncrement != 0 {
		cfg.roundingIncrement = opts.RoundingIncrement
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

func (c config) digitOptions() ecma402nf.DigitOptions {
	return ecma402nf.DigitOptions{
		MinimumIntegerDigits:     c.minIntDigits,
		MinimumFractionDigits:    c.minFracDigits,
		MaximumFractionDigits:    c.maxFracDigits,
		MinimumSignificantDigits: c.minSigDigits,
		MaximumSignificantDigits: c.maxSigDigits,
		RoundingIncrement:        c.roundingIncrement,
		RoundingMode:             c.roundingMode,
		RoundingPriority:         c.roundingPriority,
		TrailingZeroDisplay:      c.trailingZeroDisplay,
	}
}

func (c config) validate() error {
	switch c.localeMatcher {
	case string(LookupLocaleMatcher), string(BestFitLocaleMatcher):
	default:
		return invalidOption("localeMatcher", c.localeMatcher)
	}
	switch c.typ {
	case Cardinal, Ordinal:
	default:
		return invalidOption("type", c.typ.String())
	}
	if c.minIntDigits < 1 || c.minIntDigits > 21 {
		return invalidOption("minimumIntegerDigits", fmt.Sprint(c.minIntDigits))
	}
	if c.minFracDigits < 0 || c.minFracDigits > 100 {
		return invalidOption("minimumFractionDigits", fmt.Sprint(c.minFracDigits))
	}
	if c.maxFracDigits < 0 || c.maxFracDigits > 100 || c.maxFracDigits < c.minFracDigits {
		return invalidOption("maximumFractionDigits", fmt.Sprint(c.maxFracDigits))
	}
	if c.hasMinSigDigits && (c.minSigDigits < 1 || c.minSigDigits > 21) {
		return invalidOption("minimumSignificantDigits", fmt.Sprint(c.minSigDigits))
	}
	if c.hasMaxSigDigits && (c.maxSigDigits < 1 || c.maxSigDigits > 21 || (c.hasMinSigDigits && c.maxSigDigits < c.minSigDigits)) {
		return invalidOption("maximumSignificantDigits", fmt.Sprint(c.maxSigDigits))
	}
	if !isValidRoundingIncrement(c.roundingIncrement) {
		return invalidOption("roundingIncrement", fmt.Sprint(c.roundingIncrement))
	}
	if c.roundingIncrement != 1 && (c.hasMinSigDigits || c.hasMaxSigDigits || c.roundingPriority != "auto" || c.minFracDigits != c.maxFracDigits) {
		return invalidOption("roundingIncrement", fmt.Sprint(c.roundingIncrement))
	}
	if !containsString([]string{"ceil", "floor", "expand", "trunc", "halfCeil", "halfFloor", "halfExpand", "halfTrunc", "halfEven"}, c.roundingMode) {
		return invalidOption("roundingMode", c.roundingMode)
	}
	if !containsString([]string{"auto", "morePrecision", "lessPrecision"}, c.roundingPriority) {
		return invalidOption("roundingPriority", c.roundingPriority)
	}
	if !containsString([]string{"auto", "stripIfInteger"}, c.trailingZeroDisplay) {
		return invalidOption("trailingZeroDisplay", c.trailingZeroDisplay)
	}
	if !containsString([]string{"standard", "scientific", "engineering", "compact"}, c.notation) {
		return invalidOption("notation", c.notation)
	}
	if !containsString([]string{"short", "long"}, c.compactDisplay) {
		return invalidOption("compactDisplay", c.compactDisplay)
	}
	return nil
}

func CanonicalKey(opts ...Options) string {
	cfg := defaultConfig()
	if len(opts) > 0 {
		cfg = configFromOptions(opts[0])
	}
	def := defaultConfig()
	parts := make([]string, 0, 4)
	parts = cachekey.AppendString(parts, "compactDisplay", cfg.compactDisplay, def.compactDisplay)
	parts = cachekey.AppendString(parts, "localeMatcher", cfg.localeMatcher, def.localeMatcher)
	if cfg.hasMaxSigDigits {
		parts = cachekey.AppendInt(parts, "maximumSignificantDigits", cfg.maxSigDigits, def.maxSigDigits)
	}
	parts = cachekey.AppendInt(parts, "maximumFractionDigits", cfg.maxFracDigits, def.maxFracDigits)
	if cfg.hasMinSigDigits {
		parts = cachekey.AppendInt(parts, "minimumSignificantDigits", cfg.minSigDigits, def.minSigDigits)
	}
	parts = cachekey.AppendInt(parts, "minimumFractionDigits", cfg.minFracDigits, def.minFracDigits)
	parts = cachekey.AppendInt(parts, "minimumIntegerDigits", cfg.minIntDigits, def.minIntDigits)
	parts = cachekey.AppendString(parts, "notation", cfg.notation, def.notation)
	parts = cachekey.AppendInt(parts, "roundingIncrement", cfg.roundingIncrement, def.roundingIncrement)
	parts = cachekey.AppendString(parts, "roundingMode", cfg.roundingMode, def.roundingMode)
	parts = cachekey.AppendString(parts, "roundingPriority", cfg.roundingPriority, def.roundingPriority)
	parts = cachekey.AppendString(parts, "trailingZeroDisplay", cfg.trailingZeroDisplay, def.trailingZeroDisplay)
	parts = cachekey.AppendString(parts, "type", cfg.typ.String(), def.typ.String())
	return cachekey.Join(parts)
}

func containsString(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func isValidRoundingIncrement(n int) bool {
	switch n {
	case 1, 2, 5, 10, 20, 25, 50, 100, 200, 250, 500, 1000, 2000, 2500, 5000:
		return true
	default:
		return false
	}
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
