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

// formatPriorityCandidate resolves ECMA-402 roundingPriority by comparing the two
// candidates' rounding magnitudes (the decimal place each rounds at), not their
// numeric distance to the source. Per the spec, the fixed (fraction) candidate is
// "more precise" when it rounds at a deeper place: fixedMagnitude < sigMagnitude.
// morePrecision keeps the fixed candidate exactly when it is more precise;
// lessPrecision when it is not. On a distance tie the magnitudes still discriminate.
func formatPriorityCandidate(d decimal.Decimal, opts DigitOptions, more bool) (string, decimal.Decimal) {
	fixedFormatted, fixedRounded := formatFixedCandidate(d, opts)
	sigFormatted, sigRounded := formatSignificantCandidate(d, opts)
	fixedIsMorePrecise := -opts.MaximumFractionDigits < significantRoundingMagnitude(sigRounded, opts)
	if more == fixedIsMorePrecise {
		return fixedFormatted, fixedRounded
	}
	return sigFormatted, sigRounded
}

// significantRoundingMagnitude is ToRawPrecision's [[RoundingMagnitude]] = e-p+1,
// where p is the maximum significant digits and e is the decimal exponent of the
// most significant digit of the rounded value (0 when the value rounds to zero).
func significantRoundingMagnitude(sigRounded decimal.Decimal, opts DigitOptions) int {
	e := 0
	if abs := decimal.Abs(sigRounded); !abs.IsZero() {
		if magnitude, err := decimal.Log10Floor(abs); err == nil {
			e = int(magnitude)
		}
	}
	return e - opts.MaximumSignificantDigits + 1
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

// canUseRoundedString is a fast path that returns the input decimal's string
// verbatim when no digit option requires rounding, padding, or trailing-zero
// changes and the fraction already fits MaximumFractionDigits.
//
// This path is load-bearing for PluralRules: preserving the source decimal's
// visible fraction digits is what lets the operands (notably v) reflect trailing
// zeros — e.g. "1.0" must select "other", not "one". It must therefore stay
// byte-identical to the general fixed path for every input where it fires; that
// agreement is locked by TestCanUseRoundedStringAgreesWithFullPath.
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
	mode = decimal.UnsignedRoundingMode(mode, negative)
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
	mode = decimal.UnsignedRoundingMode(mode, negative)
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
	return decimal.Abs(d), d.Negative()
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
	return decimal.Neg(decimal.Abs(d))
}
