package ecma402nf

import (
	"strconv"

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

// DigitOptionInput is the caller-normalized option record consumed by
// SetNumberFormatDigitOptions.
type DigitOptionInput struct {
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

// ResolvedDigitOptions is the fully resolved digit state and selected rounding
// branch.
type ResolvedDigitOptions struct {
	DigitOptions
	RoundingType RoundingType
}

// InvalidDigitOption identifies the ECMA-402 digit option that failed
// validation.
type InvalidDigitOption struct {
	Name  string
	Value string
}

// SetNumberFormatDigitOptions resolves ECMA-402 digit options for NumberFormat
// and other constructors that reuse the same abstract operation.
func SetNumberFormatDigitOptions(in DigitOptionInput, mnfdDefault, mxfdDefault int, notation string) (ResolvedDigitOptions, InvalidDigitOption, bool) {
	if invalid, ok := validateDigitInput(in); ok {
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
	if in.RoundingPriority == "auto" {
		needSd = hasSd
		if hasSd || (!hasFd && notation == "compact") {
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
				return ResolvedDigitOptions{}, InvalidDigitOption{Name: "maximumSignificantDigits", Value: strconv.Itoa(out.MaximumSignificantDigits)}, true
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
				return ResolvedDigitOptions{}, InvalidDigitOption{Name: "maximumFractionDigits", Value: strconv.Itoa(mxfd)}, true
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
		out.RoundingPriority = "morePrecision"
	case in.RoundingPriority == "morePrecision":
		out.RoundingType = RoundingTypeMorePrecision
		out.RoundingPriority = "morePrecision"
	case in.RoundingPriority == "lessPrecision":
		out.RoundingType = RoundingTypeLessPrecision
		out.RoundingPriority = "lessPrecision"
	case hasSd:
		out.RoundingType = RoundingTypeSignificantDigits
		out.RoundingPriority = "auto"
	default:
		out.RoundingType = RoundingTypeFractionDigits
		out.RoundingPriority = "auto"
	}

	if in.RoundingIncrement != 1 {
		if out.RoundingType != RoundingTypeFractionDigits || out.MaximumFractionDigits != out.MinimumFractionDigits {
			return ResolvedDigitOptions{}, InvalidDigitOption{Name: "roundingIncrement", Value: strconv.Itoa(in.RoundingIncrement)}, true
		}
	}
	return out, InvalidDigitOption{}, false
}

func validateDigitInput(in DigitOptionInput) (InvalidDigitOption, bool) {
	integerChecks := []ecma402.IntegerOption{
		{Name: "minimumIntegerDigits", Value: in.MinimumIntegerDigits, Min: 1, Max: 21, Set: true},
		{Name: "minimumFractionDigits", Value: in.MinimumFractionDigits, Min: 0, Max: 100, Set: in.HasMinimumFractionDigits},
		{Name: "maximumFractionDigits", Value: in.MaximumFractionDigits, Min: 0, Max: 100, Set: in.HasMaximumFractionDigits},
		{Name: "minimumSignificantDigits", Value: in.MinimumSignificantDigits, Min: 1, Max: 21, Set: in.HasMinimumSignificantDigits},
		{Name: "maximumSignificantDigits", Value: in.MaximumSignificantDigits, Min: 1, Max: 21, Set: in.HasMaximumSignificantDigits},
		{Name: "roundingIncrement", Value: in.RoundingIncrement, Min: 1, Max: 5000, Set: true},
	}
	if check, ok := ecma402.InvalidIntegerOption(integerChecks...); ok {
		return InvalidDigitOption{Name: check.Name, Value: strconv.Itoa(check.Value)}, true
	}
	if !decimal.IsValidRoundingIncrement(in.RoundingIncrement) {
		return InvalidDigitOption{Name: "roundingIncrement", Value: strconv.Itoa(in.RoundingIncrement)}, true
	}
	stringChecks := []ecma402.StringOption{
		ecma402.RequiredStringOption("roundingMode", in.RoundingMode, "ceil", "floor", "expand", "trunc", "halfCeil", "halfFloor", "halfExpand", "halfTrunc", "halfEven"),
		ecma402.RequiredStringOption("roundingPriority", in.RoundingPriority, "auto", "morePrecision", "lessPrecision"),
		ecma402.RequiredStringOption("trailingZeroDisplay", in.TrailingZeroDisplay, "auto", "stripIfInteger"),
	}
	if check, ok := ecma402.InvalidStringOption(stringChecks...); ok {
		return InvalidDigitOption{Name: check.Name, Value: check.Value}, true
	}
	return InvalidDigitOption{}, false
}
