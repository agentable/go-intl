package durationformat

import (
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
)

type Duration struct {
	Years        float64 `json:"years"`
	Months       float64 `json:"months"`
	Weeks        float64 `json:"weeks"`
	Days         float64 `json:"days"`
	Hours        float64 `json:"hours"`
	Minutes      float64 `json:"minutes"`
	Seconds      float64 `json:"seconds"`
	Milliseconds float64 `json:"milliseconds"`
	Microseconds float64 `json:"microseconds"`
	Nanoseconds  float64 `json:"nanoseconds"`
}

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
	Unit  Unit     `json:"unit,omitempty"`
}

type durationRecord struct {
	values [unitCount]big.Int
}

func durationRecordOf(duration Duration, loc locale.Locale) (*durationRecord, error) {
	numbers := [unitCount]float64{
		yearsIndex:        duration.Years,
		monthsIndex:       duration.Months,
		weeksIndex:        duration.Weeks,
		daysIndex:         duration.Days,
		hoursIndex:        duration.Hours,
		minutesIndex:      duration.Minutes,
		secondsIndex:      duration.Seconds,
		millisecondsIndex: duration.Milliseconds,
		microsecondsIndex: duration.Microseconds,
		nanosecondsIndex:  duration.Nanoseconds,
	}
	record := new(durationRecord)
	var exact big.Float
	for _, spec := range durationUnitSpecs[:] {
		number := numbers[spec.index]
		if math.IsNaN(number) || math.IsInf(number, 0) || math.Trunc(number) != number {
			return nil, invalidValue(spec.unit, durationNumberString(number), expectedDurationIntegerValue, loc)
		}
		exact.SetFloat64(number).Int(&record.values[spec.index])
	}
	return record, nil
}

func durationNumberString(number float64) string {
	switch {
	case math.IsInf(number, 1):
		return "Infinity"
	case math.IsInf(number, -1):
		return "-Infinity"
	default:
		return strconv.FormatFloat(number, 'g', -1, 64)
	}
}

func (r *durationRecord) value(index unitIndex) *big.Int {
	return &r.values[index]
}

// Format is the concatenation of FormatToParts' values. ECMA-402 defines one
// partition per formatter and derives the string from it; keeping a second
// parallel text partition here would be a byte-for-byte drift liability.
func (f *DurationFormat) Format(duration Duration) (string, error) {
	parts, err := f.FormatToParts(duration)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String(), nil
}

func (f *DurationFormat) FormatToParts(duration Duration) ([]Part, error) {
	loc := f.resolved.Locale
	values, err := durationRecordOf(duration, loc)
	if err != nil {
		return nil, err
	}
	sign, err := validateDuration(values, loc)
	if err != nil {
		return nil, err
	}
	groups, err := partitionDurationFormatPattern(values, f, sign, loc)
	if err != nil {
		return nil, err
	}
	return durationListFormatParts(f.listFormatter, groups), nil
}

func partitionDurationFormatPattern(values *durationRecord, f *DurationFormat, sign int, loc locale.Locale) ([][]Part, error) {
	result := make([][]Part, 0, len(durationUnitSpecs))
	signAvailable := true
	for _, spec := range durationUnitSpecs[:] {
		opt := f.unitOptions[spec.index]
		switch opt.style {
		case NumericUnitStyle, TwoDigitUnitStyle:
			parts, err := formatDurationNumericUnits(values, f, spec.index, sign, signAvailable, loc)
			if err != nil {
				return nil, err
			}
			if len(parts) > 0 {
				result = append(result, parts)
			}
			return result, nil
		case LongUnitStyle, ShortUnitStyle, NarrowUnitStyle:
			value := values.value(spec.index)
			fractional := durationNextUnitFractional(f.unitOptions, spec)
			if !durationUnitShown(value, opt.display) && !fractional {
				continue
			}

			formatters := f.unitFormatters[spec.index]
			showNegativeSign := sign < 0 && signAvailable

			if fractional {
				valueString := durationFractionalValueString(values, spec.index, showNegativeSign)
				if opt.display == AutoDisplay && (valueString == "0" || valueString == "-0") {
					return result, nil
				}
				parts, err := formatDurationDecimalNumberParts(formatters, spec.formatUnit, spec.unit, valueString, signAvailable, loc)
				if err != nil {
					return nil, err
				}
				result = append(result, parts)
			} else {
				parts := formatDurationNumberParts(formatters, spec.formatUnit, durationIntegerNumberValue(value, showNegativeSign), signAvailable)
				result = append(result, parts)
			}
			signAvailable = false
			if fractional {
				return result, nil
			}
		case fractionalUnitStyle:
			continue
		}
	}
	return result, nil
}

func formatDurationNumericUnits(values *durationRecord, f *DurationFormat, first unitIndex, sign int, signAvailable bool, loc locale.Locale) ([]Part, error) {
	layout := durationNumericLayoutFor(values, f.unitOptions, first)

	var out []Part
	for i := range layout.count {
		index := layout.units[i]
		if i > 0 {
			out = append(out, Part{Type: PartLiteral, Value: f.separator})
		}

		if index != secondsIndex {
			spec := durationUnitSpecs[index]
			numericValue := durationIntegerNumberValue(values.value(index), sign < 0 && signAvailable)
			parts := formatDurationNumberParts(f.unitFormatters[index], spec.formatUnit, numericValue, signAvailable)
			out = append(out, parts...)
			signAvailable = false
			continue
		}
		secondsValue := durationFractionalValueString(values, secondsIndex, sign < 0 && signAvailable)
		parts, err := formatDurationDecimalNumberParts(f.unitFormatters[index], Second, string(Second), secondsValue, signAvailable, loc)
		if err != nil {
			return nil, err
		}
		out = append(out, parts...)
		signAvailable = false
	}
	return out, nil
}

type durationNumericLayout struct {
	units [3]unitIndex
	count int
}

func (l *durationNumericLayout) add(index unitIndex) {
	l.units[l.count] = index
	l.count++
}

func durationNumericLayoutFor(values *durationRecord, unitOptions [unitCount]resolvedUnitConfig, first unitIndex) durationNumericLayout {
	hoursValue := values.value(hoursIndex)
	minutesValue := values.value(minutesIndex)
	hoursFormatted := first == hoursIndex && durationUnitShown(hoursValue, unitOptions[hoursIndex].display)
	secondsFormatted := durationUnitShown(values.value(secondsIndex), unitOptions[secondsIndex].display) ||
		durationHasSubsecondValue(values)
	minutesAllowed := first == hoursIndex || first == minutesIndex
	minutesRequired := durationUnitShown(minutesValue, unitOptions[minutesIndex].display)
	minutesFormatted := minutesAllowed &&
		((hoursFormatted && secondsFormatted) || minutesRequired)
	var layout durationNumericLayout
	if hoursFormatted {
		layout.add(hoursIndex)
	}
	if minutesFormatted {
		layout.add(minutesIndex)
	}
	if secondsFormatted {
		layout.add(secondsIndex)
	}
	return layout
}

func durationHasSubsecondValue(values *durationRecord) bool {
	for _, spec := range durationUnitSpecs[:] {
		if !spec.fractional {
			continue
		}
		if values.value(spec.index).Sign() != 0 {
			return true
		}
	}
	return false
}

func durationUnitShown(value *big.Int, display Display) bool {
	return display == AlwaysDisplay || value.Sign() != 0
}

func formatDurationNumberParts(formatters durationNumberFormatters, unit Unit, value numberformat.Value, signVisible bool) []Part {
	formatter := formatters.signHidden
	if signVisible {
		formatter = formatters.signVisible
	}
	return durationNumberParts(unit, formatter.FormatToParts(value))
}

func formatDurationDecimalNumberParts(formatters durationNumberFormatters, unit Unit, errorName, value string, signVisible bool, loc locale.Locale) ([]Part, error) {
	numericValue, err := numberformat.ParseValue(value)
	if err != nil {
		return nil, invalidValue(errorName, value, expectedDurationUnitValue, loc)
	}
	return formatDurationNumberParts(formatters, unit, numericValue, signVisible), nil
}

func durationListFormatParts(formatter *listformat.ListFormat, groups [][]Part) []Part {
	switch len(groups) {
	case 0:
		return nil
	case 1:
		return groups[0]
	}
	placeholders := make([]string, len(groups))
	partCount := 0
	for _, group := range groups {
		partCount += len(group)
	}
	listParts := formatter.FormatToParts(placeholders)
	literalCount := len(listParts) - len(groups)
	out := make([]Part, partCount+literalCount)
	outIndex := 0
	groupIndex := 0
	for _, part := range listParts {
		switch part.Type {
		case listformat.PartElement:
			outIndex += copy(out[outIndex:], groups[groupIndex])
			groupIndex++
		case listformat.PartLiteral:
			out[outIndex] = Part{Type: PartLiteral, Value: part.Value}
			outIndex++
		}
	}
	return out
}

func durationNextUnitFractional(unitOptions [unitCount]resolvedUnitConfig, spec durationUnitSpec) bool {
	return spec.hasFractionalChild && unitOptions[spec.fractionalChild].style == fractionalUnitStyle
}

func durationFractionalValueString(values *durationRecord, index unitIndex, showNegativeSign bool) string {
	spec := durationUnitSpecs[index]
	base := spec.nanosecondsPerUnit
	scale := big.NewInt(base)
	var total, term, factor big.Int
	total.Abs(values.value(index))
	total.Mul(&total, scale)
	for spec.hasFractionalChild {
		spec = durationUnitSpecs[spec.fractionalChild]
		child := values.value(spec.index)
		if child.Sign() == 0 {
			continue
		}
		term.Abs(child)
		term.Mul(&term, factor.SetInt64(spec.nanosecondsPerUnit))
		total.Add(&total, &term)
	}
	var integer, fraction big.Int
	integer.QuoRem(&total, scale, &fraction)

	value := integer.String()
	if fraction.Sign() != 0 {
		value += "." + leftPadString(fraction.String(), fractionalDigitWidth(base))
	}
	if showNegativeSign {
		return "-" + value
	}
	return value
}

func fractionalDigitWidth(nanosecondsPerUnit int64) int {
	width := 0
	for scale := nanosecondsPerUnit; scale > nanosecondsPerNanosecond; scale /= 10 {
		width++
	}
	return width
}

func leftPadString(value string, width int) string {
	if len(value) >= width {
		return value
	}
	return strings.Repeat("0", width-len(value)) + value
}

func durationNumberParts(unit Unit, parts []numberformat.Part) []Part {
	out := make([]Part, len(parts))
	for i, part := range parts {
		out[i] = Part{Type: PartType(part.Type), Value: part.Value, Unit: unit}
	}
	return out
}

func validateDuration(values *durationRecord, loc locale.Locale) (int, error) {
	sign := 0
	for _, spec := range durationUnitSpecs[:] {
		valueSign := values.value(spec.index).Sign()
		switch {
		case valueSign < 0:
			if sign > 0 {
				return 0, invalidValue("duration", "mixed signs", expectedDurationMixedSigns, loc)
			}
			sign = -1
		case valueSign > 0:
			if sign < 0 {
				return 0, invalidValue("duration", "mixed signs", expectedDurationMixedSigns, loc)
			}
			sign = 1
		}
	}
	for _, spec := range durationUnitSpecs[:] {
		if spec.maxAbsExclusive == 0 {
			continue
		}
		value := values.value(spec.index)
		var magnitude, limit big.Int
		if magnitude.Abs(value).Cmp(limit.SetUint64(spec.maxAbsExclusive)) >= 0 {
			return 0, invalidValue(spec.unit, value.String(), expectedDurationCalendarUnitValue, loc)
		}
	}
	if normalizedSecondsOutOfRange(values) {
		return 0, invalidValue("duration", "normalized seconds", expectedDurationNormalizedSeconds, loc)
	}
	return sign, nil
}

func normalizedSecondsOutOfRange(values *durationRecord) bool {
	var total, term, scale, limit big.Int
	for _, spec := range durationUnitSpecs[:] {
		value := values.value(spec.index)
		if spec.nanosecondsPerUnit == 0 || value.Sign() == 0 {
			continue
		}
		term.Mul(value, scale.SetInt64(spec.nanosecondsPerUnit))
		total.Add(&total, &term)
	}
	limit.SetInt64(nanosecondsPerSecond).Lsh(&limit, 53)
	return total.Abs(&total).Cmp(&limit) >= 0
}

func durationIntegerNumberValue(value *big.Int, showNegativeSign bool) numberformat.Value {
	if !showNegativeSign {
		return numberformat.BigInt(new(big.Int).Abs(value))
	}
	if value.Sign() == 0 {
		return numberformat.Float(math.Copysign(0, -1))
	}
	return numberformat.BigInt(value)
}

const (
	expectedDurationMixedSigns        = "all non-zero duration fields to have the same sign"
	expectedDurationNormalizedSeconds = "normalized day and smaller fields below 1e9 * 2^53 nanoseconds"
	expectedDurationCalendarUnitValue = "an absolute value less than 2^32"
	expectedDurationIntegerValue      = "a finite integral ECMAScript Number"
	expectedDurationUnitValue         = "a valid duration unit value"
)

func invalidValue(name, value, expected string, loc locale.Locale) error {
	return ecma402.InvalidValueErrorExpected(durationFormatOwner, name, value, loc.String(), expected, nil)
}
