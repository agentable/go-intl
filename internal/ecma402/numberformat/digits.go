package ecma402nf

import (
	"strings"

	"github.com/agentable/go-intl/internal/decimal"
)

// DigitOptions is the resolved ECMA-402 digit-formatting state consumed by
// FormatNumericToString-style operations.
type DigitOptions struct {
	MinimumIntegerDigits     int
	MinimumFractionDigits    int
	MaximumFractionDigits    int
	MinimumSignificantDigits int
	MaximumSignificantDigits int
	RoundingIncrement        int
	RoundingMode             string
	RoundingPriority         string
	TrailingZeroDisplay      string
}

// FormattedNumeric is the result of ECMA-402 FormatNumericToString.
type FormattedNumeric struct {
	Formatted string
	Rounded   decimal.Decimal
}

// FormatNumericToString applies ECMA-402 digit rounding and zero-padding to d.
func FormatNumericToString(d decimal.Decimal, opts DigitOptions) FormattedNumeric {
	if !d.IsFinite() {
		return FormattedNumeric{Formatted: d.String(), Rounded: d}
	}
	if canUseRoundedString(d, opts) {
		return FormattedNumeric{Formatted: d.String(), Rounded: d}
	}
	switch roundingType(opts) {
	case decimal.RoundingTypeSignificantDigits:
		formatted, rounded := formatSignificantCandidate(d, opts)
		return FormattedNumeric{Formatted: formatted, Rounded: rounded}
	case decimal.RoundingTypeMorePrecision:
		formatted, rounded := formatPriorityCandidate(d, opts, true)
		return FormattedNumeric{Formatted: formatted, Rounded: rounded}
	case decimal.RoundingTypeLessPrecision:
		formatted, rounded := formatPriorityCandidate(d, opts, false)
		return FormattedNumeric{Formatted: formatted, Rounded: rounded}
	case decimal.RoundingTypeFractionDigits:
	}
	formatted, rounded := formatFixedCandidate(d, opts)
	return FormattedNumeric{Formatted: formatted, Rounded: rounded}
}

// FormatDecimal is the public typed bridge for callers that only need the
// formatted string.
func FormatDecimal(d decimal.Decimal, opts DigitOptions) string {
	return FormatNumericToString(d, opts).Formatted
}

func formatPriorityCandidate(d decimal.Decimal, opts DigitOptions, more bool) (string, decimal.Decimal) {
	fixedFormatted, fixedRounded := formatFixedCandidate(d, opts)
	sigFormatted, sigRounded := formatSignificantCandidate(d, opts)
	cmp := decimal.AbsDiffCmp(d, sigRounded, fixedRounded)
	if more {
		if cmp <= 0 {
			return sigFormatted, sigRounded
		}
		return fixedFormatted, fixedRounded
	}
	if cmp >= 0 {
		return sigFormatted, sigRounded
	}
	return fixedFormatted, fixedRounded
}

func formatFixedCandidate(d decimal.Decimal, opts DigitOptions) (string, decimal.Decimal) {
	rounded, negative := roundFixed(d, opts)
	formatted := trimFraction(roundedFixedString(rounded, opts.MaximumFractionDigits), opts)
	formatted = padMinimumIntegerDigits(formatted, opts.MinimumIntegerDigits)
	if negative {
		return "-" + formatted, signedDecimal(rounded, true)
	}
	return formatted, rounded
}

func formatSignificantCandidate(d decimal.Decimal, opts DigitOptions) (string, decimal.Decimal) {
	rounded, negative := roundSignificant(d, opts)
	formatted := trimSignificantFraction(rounded.String(), opts.MinimumSignificantDigits)
	formatted = padMinimumSignificantDigits(formatted, opts.MinimumSignificantDigits)
	if opts.TrailingZeroDisplay == "stripIfInteger" {
		formatted = stripIntegerFraction(formatted)
	}
	formatted = padMinimumIntegerDigits(formatted, opts.MinimumIntegerDigits)
	if negative {
		return "-" + formatted, signedDecimal(rounded, true)
	}
	return formatted, rounded
}

func roundingType(opts DigitOptions) decimal.RoundingType {
	priority := decimal.PriorityAuto
	switch opts.RoundingPriority {
	case "morePrecision":
		priority = decimal.PriorityMorePrecision
	case "lessPrecision":
		priority = decimal.PriorityLessPrecision
	}
	return decimal.ApplyRoundingPriority(opts.MaximumSignificantDigits > 0, true, priority)
}

func canUseRoundedString(d decimal.Decimal, opts DigitOptions) bool {
	switch {
	case opts.MinimumIntegerDigits != 1,
		opts.MinimumFractionDigits != 0,
		opts.RoundingIncrement != 1,
		opts.RoundingMode != "halfExpand",
		opts.TrailingZeroDisplay != "auto",
		opts.MaximumSignificantDigits > 0,
		opts.RoundingPriority != "auto":
		return false
	}
	_, fraction, ok := strings.Cut(strings.TrimPrefix(d.String(), "-"), ".")
	return !ok || len(fraction) <= opts.MaximumFractionDigits
}

func roundFixed(d decimal.Decimal, opts DigitOptions) (decimal.Decimal, bool) {
	unsigned, negative := unsignedDecimal(d)
	mode, err := decimal.ParseRoundingMode(opts.RoundingMode)
	if err != nil {
		mode = decimal.RoundHalfExpand
	}
	mode = decimal.GetUnsignedRoundingMode(mode, negative)
	scale := -int32(opts.MaximumFractionDigits) // #nosec G115 -- validated to ECMA-402 fraction digit range before construction.
	rounded := decimal.QuantizeToIncrement(unsigned, opts.RoundingIncrement, scale, mode)
	return rounded, negative
}

func roundSignificant(d decimal.Decimal, opts DigitOptions) (decimal.Decimal, bool) {
	unsigned, negative := unsignedDecimal(d)
	if unsigned.IsZero() {
		return unsigned, negative
	}
	if significantDigitCount(unsigned.String()) <= opts.MaximumSignificantDigits {
		return unsigned, negative
	}
	magnitude, err := decimal.Log10Floor(unsigned)
	if err != nil {
		return unsigned, negative
	}
	mode, err := decimal.ParseRoundingMode(opts.RoundingMode)
	if err != nil {
		mode = decimal.RoundHalfExpand
	}
	mode = decimal.GetUnsignedRoundingMode(mode, negative)
	scale := magnitude - int32(opts.MaximumSignificantDigits) + 1 // #nosec G115 -- validated significant digit range is 1..21.
	return decimal.QuantizeToIncrement(unsigned, 1, scale, mode), negative
}

func roundedFixedString(d decimal.Decimal, maximumFractionDigits int) string {
	if d.IsZero() && maximumFractionDigits > 0 {
		return "0." + strings.Repeat("0", maximumFractionDigits)
	}
	return strings.TrimPrefix(d.String(), "-")
}

func unsignedDecimal(d decimal.Decimal) (decimal.Decimal, bool) {
	negative := d.Negative()
	if !negative {
		return d, false
	}
	unsigned, err := decimal.ParseString(strings.TrimPrefix(d.String(), "-"))
	if err != nil {
		return d, true
	}
	return unsigned, true
}

func trimFraction(formatted string, opts DigitOptions) string {
	if opts.TrailingZeroDisplay == "stripIfInteger" {
		return stripIntegerFraction(formatted)
	}
	cut := opts.MaximumFractionDigits - opts.MinimumFractionDigits
	for cut > 0 && strings.HasSuffix(formatted, "0") {
		formatted = strings.TrimSuffix(formatted, "0")
		cut--
	}
	return strings.TrimSuffix(formatted, ".")
}

func stripIntegerFraction(formatted string) string {
	integer, fraction, ok := strings.Cut(formatted, ".")
	if ok && strings.Trim(fraction, "0") == "" {
		return integer
	}
	return formatted
}

func padMinimumIntegerDigits(formatted string, minimum int) string {
	integer, fraction, ok := strings.Cut(formatted, ".")
	if len(integer) < minimum {
		integer = strings.Repeat("0", minimum-len(integer)) + integer
	}
	if !ok {
		return integer
	}
	return integer + "." + fraction
}

func trimSignificantFraction(formatted string, minimum int) string {
	for strings.Contains(formatted, ".") && strings.HasSuffix(formatted, "0") && significantDigitCount(formatted) > minimum {
		formatted = strings.TrimSuffix(formatted, "0")
	}
	return strings.TrimSuffix(formatted, ".")
}

func padMinimumSignificantDigits(formatted string, minimum int) string {
	for significantDigitCount(formatted) < minimum {
		if !strings.Contains(formatted, ".") {
			formatted += "."
		}
		formatted += "0"
	}
	return formatted
}

func significantDigitCount(formatted string) int {
	formatted = strings.TrimPrefix(formatted, "-")
	hasNonZero := strings.IndexFunc(formatted, isNonZeroDigit) >= 0
	if !hasNonZero {
		digits := 0
		for _, r := range formatted {
			if isDigit(r) {
				digits++
			}
		}
		if digits == 0 {
			return 1
		}
		return digits
	}
	count := 0
	started := false
	for _, r := range formatted {
		if !isDigit(r) {
			continue
		}
		if r != '0' || started {
			started = true
			count++
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isNonZeroDigit(r rune) bool {
	return r >= '1' && r <= '9'
}

func signedDecimal(d decimal.Decimal, negative bool) decimal.Decimal {
	if !negative || d.IsZero() {
		return d
	}
	out, err := decimal.ParseString("-" + strings.TrimPrefix(d.String(), "-"))
	if err != nil {
		return d
	}
	return out
}
