package ecma402nf

import (
	"strconv"
	"strings"

	"github.com/agentable/go-intl/internal/decimal"
	"github.com/agentable/go-intl/internal/ecma402"
)

// RoundingType names the ECMA-402 rounding branch selected for digit options.
type RoundingType string

const (
	// RoundingTypeFractionDigits rounds with fraction digit bounds.
	RoundingTypeFractionDigits RoundingType = "fractionDigits"
	// RoundingTypeSignificantDigits rounds with significant digit bounds.
	RoundingTypeSignificantDigits RoundingType = "significantDigits"
	// RoundingTypeMorePrecision chooses the more precise fraction/significant result.
	RoundingTypeMorePrecision RoundingType = "morePrecision"
	// RoundingTypeLessPrecision chooses the less precise fraction/significant result.
	RoundingTypeLessPrecision RoundingType = "lessPrecision"
)

const (
	notationStandard    = "standard"
	notationScientific  = "scientific"
	notationEngineering = "engineering"
	notationCompact     = "compact"

	compactDisplayShort = "short"
	compactDisplayLong  = "long"

	roundingModeCeil       = "ceil"
	roundingModeFloor      = "floor"
	roundingModeExpand     = "expand"
	roundingModeTrunc      = "trunc"
	roundingModeHalfCeil   = "halfCeil"
	roundingModeHalfFloor  = "halfFloor"
	roundingModeHalfExpand = "halfExpand"
	roundingModeHalfTrunc  = "halfTrunc"
	roundingModeHalfEven   = "halfEven"

	roundingPriorityAuto          = "auto"
	roundingPriorityMorePrecision = "morePrecision"
	roundingPriorityLessPrecision = "lessPrecision"

	trailingZeroDisplayAuto           = "auto"
	trailingZeroDisplayStripIfInteger = "stripIfInteger"
)

var (
	notationValues = [...]string{
		notationStandard,
		notationScientific,
		notationEngineering,
		notationCompact,
	}
	compactDisplayValues = [...]string{
		compactDisplayShort,
		compactDisplayLong,
	}
	roundingModeValues = [...]string{
		roundingModeCeil,
		roundingModeFloor,
		roundingModeExpand,
		roundingModeTrunc,
		roundingModeHalfCeil,
		roundingModeHalfFloor,
		roundingModeHalfExpand,
		roundingModeHalfTrunc,
		roundingModeHalfEven,
	}
	roundingPriorityValues = [...]string{
		roundingPriorityAuto,
		roundingPriorityMorePrecision,
		roundingPriorityLessPrecision,
	}
	trailingZeroDisplayValues = [...]string{
		trailingZeroDisplayAuto,
		trailingZeroDisplayStripIfInteger,
	}
)

// DigitOptionConfig carries caller-normalized digit options before
// SetNumberFormatDigitOptions resolves ECMA-402 defaults.
type DigitOptionConfig struct {
	MinimumIntegerDigits     int
	MinimumFractionDigits    int
	MaximumFractionDigits    int
	MinimumSignificantDigits int
	MaximumSignificantDigits int
	RoundingIncrement        int
	RoundingMode             string
	RoundingPriority         string
	TrailingZeroDisplay      string

	HasMinimumFractionDigits    bool
	HasMaximumFractionDigits    bool
	HasMinimumSignificantDigits bool
	HasMaximumSignificantDigits bool
}

// DigitOptionOverrides carries caller-provided digit options. Nil pointers mean
// the option was omitted.
type DigitOptionOverrides struct {
	MinimumIntegerDigits     *int
	MinimumFractionDigits    *int
	MaximumFractionDigits    *int
	MinimumSignificantDigits *int
	MaximumSignificantDigits *int
	RoundingIncrement        *int
	RoundingMode             *string
	RoundingPriority         *string
	TrailingZeroDisplay      *string
}

// DefaultDigitOptionConfig returns the shared ECMA-402 digit option defaults
// used by NumberFormat-family constructors before style-specific defaults apply.
func DefaultDigitOptionConfig() DigitOptionConfig {
	return DigitOptionConfig{
		MinimumIntegerDigits:  1,
		MinimumFractionDigits: 0,
		MaximumFractionDigits: 3,
		RoundingIncrement:     1,
		RoundingMode:          roundingModeHalfExpand,
		RoundingPriority:      roundingPriorityAuto,
		TrailingZeroDisplay:   trailingZeroDisplayAuto,
	}
}

// ApplyOverrides applies caller-provided digit options while preserving
// omitted-vs-zero state for pointer-backed fields.
func (c *DigitOptionConfig) ApplyOverrides(opts DigitOptionOverrides) {
	ecma402.ApplyOption(&c.MinimumIntegerDigits, opts.MinimumIntegerDigits)
	ecma402.ApplyOptionInput(&c.MinimumFractionDigits, &c.HasMinimumFractionDigits, opts.MinimumFractionDigits)
	ecma402.ApplyOptionInput(&c.MaximumFractionDigits, &c.HasMaximumFractionDigits, opts.MaximumFractionDigits)
	ecma402.ApplyOptionInput(&c.MinimumSignificantDigits, &c.HasMinimumSignificantDigits, opts.MinimumSignificantDigits)
	ecma402.ApplyOptionInput(&c.MaximumSignificantDigits, &c.HasMaximumSignificantDigits, opts.MaximumSignificantDigits)
	ecma402.ApplyOption(&c.RoundingIncrement, opts.RoundingIncrement)
	ecma402.ApplyOption(&c.RoundingMode, opts.RoundingMode)
	ecma402.ApplyOption(&c.RoundingPriority, opts.RoundingPriority)
	ecma402.ApplyOption(&c.TrailingZeroDisplay, opts.TrailingZeroDisplay)
}

// ResolvedDigitOptions is the fully resolved digit state and selected rounding
// branch.
type ResolvedDigitOptions struct {
	DigitOptions
	RoundingType RoundingType
}

// ResolvedDigitProperties carries digit resolved-option properties whose
// ECMA-402 visibility depends on the selected rounding branch.
type ResolvedDigitProperties struct {
	MinimumFractionDigits    *int
	MaximumFractionDigits    *int
	MinimumSignificantDigits *int
	MaximumSignificantDigits *int
}

// ResolvedProperties returns the digit resolved-option properties visible for
// this resolved digit state.
func (d ResolvedDigitOptions) ResolvedProperties() ResolvedDigitProperties {
	switch d.RoundingType {
	case RoundingTypeFractionDigits:
		return d.resolvedFractionDigitProperties()
	case RoundingTypeSignificantDigits:
		return d.resolvedSignificantDigitProperties()
	default:
		return d.resolvedAllDigitProperties()
	}
}

// ResolvedPluralRulesProperties returns the digit resolved-option properties
// visible for Intl.PluralRules.
func (d ResolvedDigitOptions) ResolvedPluralRulesProperties() ResolvedDigitProperties {
	if d.RoundingType == RoundingTypeFractionDigits {
		return d.resolvedFractionDigitProperties()
	}
	return d.resolvedSignificantDigitProperties()
}

// CanUseIntegerOperands reports whether integer numeric bridges can bypass
// decimal operand construction for Intl.PluralRules without changing selected
// categories.
func (d ResolvedDigitOptions) CanUseIntegerOperands(notation string) bool {
	return notation == notationStandard &&
		d.RoundingType == RoundingTypeFractionDigits &&
		d.MinimumFractionDigits == 0 &&
		d.RoundingIncrement == 1 &&
		d.RoundingPriority == roundingPriorityAuto
}

func (d ResolvedDigitOptions) resolvedFractionDigitProperties() ResolvedDigitProperties {
	return ResolvedDigitProperties{
		MinimumFractionDigits: ecma402.ResolvedScalar(d.MinimumFractionDigits),
		MaximumFractionDigits: ecma402.ResolvedScalar(d.MaximumFractionDigits),
	}
}

func (d ResolvedDigitOptions) resolvedSignificantDigitProperties() ResolvedDigitProperties {
	return ResolvedDigitProperties{
		MinimumSignificantDigits: ecma402.ResolvedScalar(d.MinimumSignificantDigits),
		MaximumSignificantDigits: ecma402.ResolvedScalar(d.MaximumSignificantDigits),
	}
}

func (d ResolvedDigitOptions) resolvedAllDigitProperties() ResolvedDigitProperties {
	return ResolvedDigitProperties{
		MinimumFractionDigits:    ecma402.ResolvedScalar(d.MinimumFractionDigits),
		MaximumFractionDigits:    ecma402.ResolvedScalar(d.MaximumFractionDigits),
		MinimumSignificantDigits: ecma402.ResolvedScalar(d.MinimumSignificantDigits),
		MaximumSignificantDigits: ecma402.ResolvedScalar(d.MaximumSignificantDigits),
	}
}

// InvalidDigitOption identifies the ECMA-402 digit option that failed
// validation.
type InvalidDigitOption struct {
	Name     string
	Value    string
	Expected string
}

// NotationOption returns the shared NumberFormat-family notation option rule.
func NotationOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("notation", value, notationValues[:]...)
}

// CompactDisplayOption returns the shared NumberFormat-family compactDisplay
// option rule.
func CompactDisplayOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("compactDisplay", value, compactDisplayValues[:]...)
}

func roundingModeOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("roundingMode", value, roundingModeValues[:]...)
}

func roundingPriorityOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("roundingPriority", value, roundingPriorityValues[:]...)
}

func trailingZeroDisplayOption(value string) ecma402.StringOption {
	return ecma402.RequiredStringOption("trailingZeroDisplay", value, trailingZeroDisplayValues[:]...)
}

// InvalidDigitOptionError returns an invalid-option error for a failed shared
// NumberFormat digit option.
func InvalidDigitOptionError(owner string, invalid InvalidDigitOption, loc string) error {
	return ecma402.InvalidOptionErrorExpected(owner, invalid.Name, invalid.Value, loc, invalid.Expected, nil)
}

// SetNumberFormatDigitOptions resolves ECMA-402 digit options for NumberFormat
// and other constructors that reuse the same abstract operation.
func SetNumberFormatDigitOptions(in DigitOptionConfig, mnfdDefault, mxfdDefault int, notation string) (ResolvedDigitOptions, InvalidDigitOption, bool) {
	if invalid, ok := validateDigitConfig(in); ok {
		return ResolvedDigitOptions{}, invalid, true
	}
	if in.RoundingIncrement != 1 {
		mxfdDefault = mnfdDefault
	}

	out := ResolvedDigitOptions{
		DigitOptions: DigitOptions{
			MinimumIntegerDigits: in.MinimumIntegerDigits,
			RoundingIncrement:    in.RoundingIncrement,
			RoundingMode:         in.RoundingMode,
			TrailingZeroDisplay:  in.TrailingZeroDisplay,
		},
	}

	hasSd := in.HasMinimumSignificantDigits || in.HasMaximumSignificantDigits
	hasFd := in.HasMinimumFractionDigits || in.HasMaximumFractionDigits
	needSd, needFd := true, true
	if in.RoundingPriority == roundingPriorityAuto {
		needSd = hasSd
		if hasSd || (!hasFd && notation == notationCompact) {
			needFd = false
		}
	}

	if needSd {
		if hasSd {
			out.MinimumSignificantDigits = in.MinimumSignificantDigits
			if !in.HasMinimumSignificantDigits {
				out.MinimumSignificantDigits = 1
			}
			out.MaximumSignificantDigits = in.MaximumSignificantDigits
			if !in.HasMaximumSignificantDigits {
				out.MaximumSignificantDigits = 21
			}
			if out.MinimumSignificantDigits > out.MaximumSignificantDigits {
				return ResolvedDigitOptions{}, invalidDigitOption(
					"maximumSignificantDigits",
					strconv.Itoa(out.MaximumSignificantDigits),
					"greater than or equal to minimumSignificantDigits",
				), true
			}
		} else {
			out.MinimumSignificantDigits = 1
			out.MaximumSignificantDigits = 21
		}
	}
	if needFd {
		if hasFd {
			mnfd, hasMnfd := in.MinimumFractionDigits, in.HasMinimumFractionDigits
			mxfd, hasMxfd := in.MaximumFractionDigits, in.HasMaximumFractionDigits
			switch {
			case !hasMnfd:
				mnfd = min(mnfdDefault, mxfd)
			case !hasMxfd:
				mxfd = max(mxfdDefault, mnfd)
			case mnfd > mxfd:
				return ResolvedDigitOptions{}, invalidDigitOption(
					"maximumFractionDigits",
					strconv.Itoa(mxfd),
					"greater than or equal to minimumFractionDigits",
				), true
			}
			out.MinimumFractionDigits = mnfd
			out.MaximumFractionDigits = mxfd
		} else {
			out.MinimumFractionDigits = mnfdDefault
			out.MaximumFractionDigits = mxfdDefault
		}
	}

	switch {
	case !needSd && !needFd:
		out.MinimumFractionDigits = 0
		out.MaximumFractionDigits = 0
		out.MinimumSignificantDigits = 1
		out.MaximumSignificantDigits = 2
		out.RoundingType = RoundingTypeMorePrecision
		out.RoundingPriority = roundingPriorityMorePrecision
	case in.RoundingPriority == roundingPriorityMorePrecision:
		out.RoundingType = RoundingTypeMorePrecision
		out.RoundingPriority = roundingPriorityMorePrecision
	case in.RoundingPriority == roundingPriorityLessPrecision:
		out.RoundingType = RoundingTypeLessPrecision
		out.RoundingPriority = roundingPriorityLessPrecision
	case hasSd:
		out.RoundingType = RoundingTypeSignificantDigits
		out.RoundingPriority = roundingPriorityAuto
	default:
		out.RoundingType = RoundingTypeFractionDigits
		out.RoundingPriority = roundingPriorityAuto
	}

	if in.RoundingIncrement != 1 {
		if out.RoundingType != RoundingTypeFractionDigits || out.MaximumFractionDigits != out.MinimumFractionDigits {
			return ResolvedDigitOptions{}, invalidDigitOption(
				"roundingIncrement",
				strconv.Itoa(in.RoundingIncrement),
				"roundingIncrement 1 unless fraction digit rounding uses equal minimumFractionDigits and maximumFractionDigits",
			), true
		}
	}
	return out, InvalidDigitOption{}, false
}

func validateDigitConfig(in DigitOptionConfig) (InvalidDigitOption, bool) {
	if check, ok := ecma402.InvalidIntegerOption(
		ecma402.IntegerOption{Name: "minimumIntegerDigits", Value: in.MinimumIntegerDigits, Min: 1, Max: 21, Set: true},
		ecma402.IntegerOption{Name: "minimumFractionDigits", Value: in.MinimumFractionDigits, Min: 0, Max: 100, Set: in.HasMinimumFractionDigits},
		ecma402.IntegerOption{Name: "maximumFractionDigits", Value: in.MaximumFractionDigits, Min: 0, Max: 100, Set: in.HasMaximumFractionDigits},
		ecma402.IntegerOption{Name: "minimumSignificantDigits", Value: in.MinimumSignificantDigits, Min: 1, Max: 21, Set: in.HasMinimumSignificantDigits},
		ecma402.IntegerOption{Name: "maximumSignificantDigits", Value: in.MaximumSignificantDigits, Min: 1, Max: 21, Set: in.HasMaximumSignificantDigits},
		ecma402.IntegerOption{Name: "roundingIncrement", Value: in.RoundingIncrement, Min: 1, Max: 5000, Set: true},
	); ok {
		return invalidIntegerDigitOption(check), true
	}
	if !decimal.IsValidRoundingIncrement(in.RoundingIncrement) {
		return invalidDigitOption("roundingIncrement", strconv.Itoa(in.RoundingIncrement), expectedRoundingIncrement()), true
	}
	if check, ok := ecma402.InvalidStringOption(
		roundingModeOption(in.RoundingMode),
		roundingPriorityOption(in.RoundingPriority),
		trailingZeroDisplayOption(in.TrailingZeroDisplay),
	); ok {
		return invalidStringDigitOption(check), true
	}
	return InvalidDigitOption{}, false
}

func invalidDigitOption(name, value, expected string) InvalidDigitOption {
	return InvalidDigitOption{Name: name, Value: value, Expected: expected}
}

func invalidIntegerDigitOption(check ecma402.IntegerOption) InvalidDigitOption {
	return invalidDigitOption(check.Name, strconv.Itoa(check.Value), check.Expected())
}

func invalidStringDigitOption(check ecma402.StringOption) InvalidDigitOption {
	return invalidDigitOption(check.Name, check.Value, check.Expected())
}

func expectedRoundingIncrement() string {
	return "one of " + intValues(decimal.RoundingIncrements())
}

func intValues(values []int) string {
	var b strings.Builder
	for i, value := range values {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(strconv.Itoa(value))
	}
	return b.String()
}
