package durationformat

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/numberformat"
)

type Duration struct {
	Years        int64 `json:"years"`
	Months       int64 `json:"months"`
	Weeks        int64 `json:"weeks"`
	Days         int64 `json:"days"`
	Hours        int64 `json:"hours"`
	Minutes      int64 `json:"minutes"`
	Seconds      int64 `json:"seconds"`
	Milliseconds int64 `json:"milliseconds"`
	Microseconds int64 `json:"microseconds"`
	Nanoseconds  int64 `json:"nanoseconds"`
}

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
	Unit  Unit     `json:"unit,omitempty"`
}

func (f *DurationFormat) Format(duration Duration) (string, error) {
	parts, err := f.FormatToParts(duration)
	if err != nil {
		return "", err
	}
	return joinParts(parts), nil
}

func (f *DurationFormat) FormatToParts(duration Duration) ([]Part, error) {
	if err := validateDuration(duration); err != nil {
		return nil, err
	}
	groups, err := f.partitionDurationFormatPattern(duration)
	if err != nil {
		return nil, err
	}
	return f.listFormatParts(groups)
}

func (f *DurationFormat) partitionDurationFormatPattern(duration Duration) ([][]Part, error) {
	result := make([][]Part, 0, len(durationUnitSpecs))
	signDisplayed := true
	sign := durationSign(duration)
	for _, spec := range durationUnitSpecs {
		opt := f.unitOptions[spec.index]
		value := durationValue(duration, spec.index)
		switch opt.style {
		case NumericUnitStyle, TwoDigitUnitStyle:
			parts, err := f.formatNumericUnits(duration, spec.index, sign, signDisplayed)
			if err != nil {
				return nil, err
			}
			if len(parts) > 0 {
				result = append(result, parts)
			}
			return result, nil
		case LongUnitStyle, ShortUnitStyle, NarrowUnitStyle:
			fractional := f.nextUnitFractional(spec.index)
			unitVisible := opt.display == AlwaysDisplay || value != 0 || fractional
			if !unitVisible {
				continue
			}

			showSign := sign < 0 && signDisplayed
			valueString := int64ValueString(value, showSign)
			formatters := f.formatters.unit[spec.index]
			switch {
			case fractional:
				valueString = f.fractionalValueString(duration, spec.index, showSign)
				formatters = f.formatters.unitFraction[spec.index]
			case !signDisplayed:
				valueString = uint64ValueString(absInt64(value), false)
			}
			value, err := numberformat.Decimal(valueString)
			if err != nil {
				return nil, invalidValue(spec.unit, valueString)
			}
			numberParts := formatters.formatter(signDisplayed).FormatToParts(value)
			result = append(result, durationNumberParts(spec.formatUnit, numberParts))
			signDisplayed = false
			if fractional {
				return result, nil
			}
		case fractionalUnitStyle:
			continue
		}
	}
	return result, nil
}

func (f *DurationFormat) formatNumericUnits(duration Duration, first unitIndex, sign int, signDisplayed bool) ([]Part, error) {
	hoursValue := duration.Hours
	minutesValue := duration.Minutes

	hoursFormatted := first == hoursIndex && (hoursValue != 0 || f.unitOptions[hoursIndex].display == AlwaysDisplay)
	secondsFormatted := duration.Seconds != 0 ||
		duration.Milliseconds != 0 ||
		duration.Microseconds != 0 ||
		duration.Nanoseconds != 0 ||
		f.unitOptions[secondsIndex].display == AlwaysDisplay
	minutesAllowed := first == hoursIndex || first == minutesIndex
	minutesRequired := minutesValue != 0 || f.unitOptions[minutesIndex].display == AlwaysDisplay
	minutesFormatted := minutesAllowed &&
		((hoursFormatted && secondsFormatted) || minutesRequired)

	var out []Part
	if hoursFormatted {
		value := int64ValueString(hoursValue, sign < 0 && signDisplayed)
		parts, err := f.formatNumericPart(hoursIndex, Hour, value, signDisplayed)
		if err != nil {
			return nil, err
		}
		out = append(out, parts...)
		signDisplayed = false
	}
	if minutesFormatted {
		if hoursFormatted {
			out = append(out, Part{Type: PartLiteral, Value: f.separator})
		}
		value := int64ValueString(minutesValue, sign < 0 && signDisplayed)
		parts, err := f.formatNumericPart(minutesIndex, Minute, value, signDisplayed)
		if err != nil {
			return nil, err
		}
		out = append(out, parts...)
		signDisplayed = false
	}
	if secondsFormatted {
		if minutesFormatted {
			out = append(out, Part{Type: PartLiteral, Value: f.separator})
		}
		secondsValue := f.fractionalValueString(duration, secondsIndex, sign < 0 && signDisplayed)
		parts, err := f.formatNumericPart(secondsIndex, Second, secondsValue, signDisplayed)
		if err != nil {
			return nil, err
		}
		out = append(out, parts...)
	}
	return out, nil
}

func (f *DurationFormat) formatNumericPart(index unitIndex, unit Unit, value string, signDisplayed bool) ([]Part, error) {
	formatters := f.formatters.numeric[index]
	if !signDisplayed {
		value = strings.TrimPrefix(value, "-")
	}
	if index == secondsIndex {
		formatters = f.formatters.numericFraction[index]
	}
	numericValue, err := numberformat.Decimal(value)
	if err != nil {
		return nil, invalidValue(string(unit), value)
	}
	parts := formatters.formatter(signDisplayed).FormatToParts(numericValue)
	return durationNumberParts(unit, parts), nil
}

func (f *DurationFormat) listFormatParts(groups [][]Part) ([]Part, error) {
	if len(groups) == 0 {
		return nil, nil
	}
	elements := make([]string, len(groups))
	partCount := 0
	for i, group := range groups {
		elements[i] = joinParts(group)
		partCount += len(group)
	}
	listParts := f.formatters.list.FormatToParts(elements)
	literalCount := len(listParts) - len(groups)
	out := make([]Part, 0, partCount+literalCount)
	groupIndex := 0
	for _, part := range listParts {
		switch part.Type {
		case listformat.PartElement:
			out = append(out, groups[groupIndex]...)
			groupIndex++
		case listformat.PartLiteral:
			out = append(out, Part{Type: PartLiteral, Value: part.Value})
		}
	}
	return out, nil
}

func (f *DurationFormat) nextUnitFractional(index unitIndex) bool {
	return durationNextUnitFractional(f.unitOptions, index)
}

func durationNextUnitFractional(unitOptions [unitCount]resolvedUnitConfig, index unitIndex) bool {
	switch index {
	case secondsIndex:
		return unitOptions[millisecondsIndex].style == fractionalUnitStyle
	case millisecondsIndex:
		return unitOptions[microsecondsIndex].style == fractionalUnitStyle
	case microsecondsIndex:
		return unitOptions[nanosecondsIndex].style == fractionalUnitStyle
	case yearsIndex, monthsIndex, weeksIndex, daysIndex, hoursIndex, minutesIndex, nanosecondsIndex, unitCount:
	}
	return false
}

func (f *DurationFormat) fractionalValueString(duration Duration, index unitIndex, negative bool) string {
	switch index {
	case secondsIndex:
		return decimalValueString(duration.Seconds, negative, 9, []fractionalPart{
			{value: duration.Milliseconds, denominator: 1_000},
			{value: duration.Microseconds, denominator: 1_000_000},
			{value: duration.Nanoseconds, denominator: 1_000_000_000},
		})
	case millisecondsIndex:
		return decimalValueString(duration.Milliseconds, negative, 6, []fractionalPart{
			{value: duration.Microseconds, denominator: 1_000},
			{value: duration.Nanoseconds, denominator: 1_000_000},
		})
	case microsecondsIndex:
		return decimalValueString(duration.Microseconds, negative, 3, []fractionalPart{
			{value: duration.Nanoseconds, denominator: 1_000},
		})
	case yearsIndex, monthsIndex, weeksIndex, daysIndex, hoursIndex, minutesIndex, nanosecondsIndex, unitCount:
	}
	return int64ValueString(durationValue(duration, index), negative)
}

type fractionalPart struct {
	value       int64
	denominator int64
}

func decimalValueString(whole int64, negative bool, width int, parts []fractionalPart) string {
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(width)), nil)
	total := new(big.Int).SetUint64(absInt64(whole))
	total.Mul(total, scale)
	for _, part := range parts {
		term := new(big.Int).SetUint64(absInt64(part.value))
		term.Mul(term, scale)
		term.Div(term, big.NewInt(part.denominator))
		total.Add(total, term)
	}

	integer := new(big.Int)
	fraction := new(big.Int)
	integer.QuoRem(total, scale, fraction)

	value := integer.String()
	if fraction.Sign() != 0 {
		value += "." + leftPadString(fraction.String(), width)
	}
	if negative {
		return "-" + value
	}
	return value
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

func joinParts(parts []Part) string {
	valueLen := 0
	for _, part := range parts {
		valueLen += len(part.Value)
	}
	var b strings.Builder
	b.Grow(valueLen)
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func durationValue(duration Duration, index unitIndex) int64 {
	switch index {
	case yearsIndex:
		return duration.Years
	case monthsIndex:
		return duration.Months
	case weeksIndex:
		return duration.Weeks
	case daysIndex:
		return duration.Days
	case hoursIndex:
		return duration.Hours
	case minutesIndex:
		return duration.Minutes
	case secondsIndex:
		return duration.Seconds
	case millisecondsIndex:
		return duration.Milliseconds
	case microsecondsIndex:
		return duration.Microseconds
	case nanosecondsIndex:
		return duration.Nanoseconds
	case unitCount:
	}
	return 0
}

func durationSign(duration Duration) int {
	for _, spec := range durationUnitSpecs {
		value := durationValue(duration, spec.index)
		if value < 0 {
			return -1
		}
		if value > 0 {
			return 1
		}
	}
	return 0
}

func validateDuration(duration Duration) error {
	sign := 0
	for _, spec := range durationUnitSpecs {
		value := durationValue(duration, spec.index)
		switch {
		case value < 0:
			if sign > 0 {
				return invalidValue("duration", "mixed signs")
			}
			sign = -1
		case value > 0:
			if sign < 0 {
				return invalidValue("duration", "mixed signs")
			}
			sign = 1
		}
	}
	for _, tc := range []struct {
		name  string
		value int64
	}{
		{name: "years", value: duration.Years},
		{name: "months", value: duration.Months},
		{name: "weeks", value: duration.Weeks},
	} {
		if absInt64(tc.value) >= 1<<32 {
			return invalidValue(tc.name, strconv.FormatInt(tc.value, 10))
		}
	}
	if normalizedSecondsOutOfRange(duration) {
		return invalidValue("duration", "normalized seconds")
	}
	return nil
}

func normalizedSecondsOutOfRange(duration Duration) bool {
	total := big.NewInt(0)
	addScaled := func(value int64, scale int64) {
		term := big.NewInt(value)
		term.Mul(term, big.NewInt(scale))
		total.Add(total, term)
	}
	addScaled(duration.Days, 86_400_000_000_000)
	addScaled(duration.Hours, 3_600_000_000_000)
	addScaled(duration.Minutes, 60_000_000_000)
	addScaled(duration.Seconds, 1_000_000_000)
	addScaled(duration.Milliseconds, 1_000_000)
	addScaled(duration.Microseconds, 1_000)
	addScaled(duration.Nanoseconds, 1)
	limit := new(big.Int).Lsh(big.NewInt(1_000_000_000), 53)
	return new(big.Int).Abs(total).Cmp(limit) >= 0
}

func int64ValueString(value int64, negative bool) string {
	return uint64ValueString(absInt64(value), negative)
}

func uint64ValueString(value uint64, negative bool) string {
	out := strconv.FormatUint(value, 10)
	if negative {
		return "-" + out
	}
	return out
}

func absInt64(value int64) uint64 {
	const minInt64 = -1 << 63
	if value >= 0 {
		return uint64(value)
	}
	if value == minInt64 {
		return 1 << 63
	}
	return uint64(-value)
}

func invalidValue(name, value string) error {
	return intlerr.New(intlerr.InvalidValue, "durationformat", name, value, "", intlerr.ErrInvalidValue)
}
