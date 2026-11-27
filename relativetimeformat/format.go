package relativetimeformat

import (
	"fmt"
	"strings"

	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
	Unit  Unit     `json:"unit,omitempty"`
}

func (f *RelativeTimeFormat) Format(value Value, unit Unit) (string, error) {
	parts, err := f.FormatToParts(value, unit)
	if err != nil {
		return "", err
	}
	if len(parts) == 1 {
		return parts[0].Value, nil
	}
	size := 0
	for _, part := range parts {
		size += len(part.Value)
	}
	var text strings.Builder
	text.Grow(size)
	for _, part := range parts {
		text.WriteString(part.Value)
	}
	return text.String(), nil
}

func (f *RelativeTimeFormat) FormatToParts(value Value, unit Unit) ([]Part, error) {
	if value.invalidErr != nil {
		return nil, invalidRelativeTimeValue(value.errValue, f.resolved.Locale.String(), value.invalidErr)
	}
	resolvedUnit, ok := singularUnit(unit)
	if !ok {
		return nil, invalidRelativeTimeUnit(unit, f.resolved.Locale.String())
	}
	field, ok := f.fields[resolvedUnit]
	if !ok {
		return nil, invalidRelativeTimeUnit(unit, f.resolved.Locale.String())
	}
	if f.resolved.Numeric == NumericAuto {
		if literal := field.relative[value.literalKey()]; literal != "" {
			return []Part{{Type: PartLiteral, Value: literal}}, nil
		}
	}
	category := f.plural.Select(value.pluralRulesValue())
	patterns := field.future
	if value.isPast() {
		patterns = field.past
	}
	return relativeTimePatternParts(patterns.pattern(category), resolvedUnit, f.number.FormatToParts(value.numberFormatValue())), nil
}

type relativeTimeField struct {
	future   relativeTimePatternSet
	past     relativeTimePatternSet
	relative map[string]string
}

const relativeTimePatternCategoryCount = int(pluralrules.Other) + 1

type relativeTimePatternSet [relativeTimePatternCategoryCount]ecma402.Pattern

var relativeTimeFallbackCategories = [...]pluralrules.Category{
	pluralrules.Zero,
	pluralrules.One,
	pluralrules.Two,
	pluralrules.Few,
	pluralrules.Many,
}

func compileRelativeTimeFields(raw cldrrelativetime.RelativeTimeFields, style Style) (map[Unit]relativeTimeField, error) {
	out := make(map[Unit]relativeTimeField, len(relativeTimeUnits))
	for _, unit := range relativeTimeUnits {
		field, ok := selectRelativeTimeField(raw, unit, style)
		if !ok {
			return nil, fmt.Errorf("relativetimeformat: CLDR relative-time unit %s style %s missing after long fallback", unit, style)
		}
		compiled, err := compileRelativeTimeField(field)
		if err != nil {
			return nil, fmt.Errorf("relativetimeformat: CLDR relative-time unit %s style %s: %w", unit, style, err)
		}
		out[unit] = compiled
	}
	return out, nil
}

var relativeTimeUnits = [...]Unit{Second, Minute, Hour, Day, Week, Month, Quarter, Year}

const expectedRelativeTimeUnit = `one of "second", "minute", "hour", "day", "week", "month", "quarter", "year", or their plural forms`

func selectRelativeTimeField(fields cldrrelativetime.RelativeTimeFields, unit Unit, style Style) (cldrrelativetime.RelativeTimeField, bool) {
	styles := fields[string(unit)]
	if field, ok := styles[string(style)]; ok {
		return field, true
	}
	field, ok := styles[string(LongStyle)]
	return field, ok
}

func compileRelativeTimeField(field cldrrelativetime.RelativeTimeField) (relativeTimeField, error) {
	future, err := compileRelativeTimePatternSet(field.Future)
	if err != nil {
		return relativeTimeField{}, fmt.Errorf("future: %w", err)
	}
	past, err := compileRelativeTimePatternSet(field.Past)
	if err != nil {
		return relativeTimeField{}, fmt.Errorf("past: %w", err)
	}
	return relativeTimeField{
		future:   future,
		past:     past,
		relative: field.Relative,
	}, nil
}

func compileRelativeTimePatternSet(patterns map[string]string) (relativeTimePatternSet, error) {
	text := patterns[pluralrules.Other.String()]
	if text == "" {
		return relativeTimePatternSet{}, fmt.Errorf("missing CLDR other pattern")
	}
	other, err := compileRelativeTimePattern(text)
	if err != nil {
		return relativeTimePatternSet{}, err
	}
	var out relativeTimePatternSet
	out[pluralrules.Other] = other
	for _, category := range relativeTimeFallbackCategories {
		text := patterns[category.String()]
		if text == "" {
			out[category] = other
			continue
		}
		pattern, err := compileRelativeTimePattern(text)
		if err != nil {
			return relativeTimePatternSet{}, err
		}
		out[category] = pattern
	}
	return out, nil
}

func compileRelativeTimePattern(pattern string) (ecma402.Pattern, error) {
	parts, err := ecma402.PartitionPattern(pattern)
	if err != nil {
		return nil, fmt.Errorf("relativetimeformat: malformed CLDR pattern: %w", err)
	}
	return parts, nil
}

func (p relativeTimePatternSet) pattern(category pluralrules.Category) ecma402.Pattern {
	// category is always in 0..Other and the set is sized Other+1, so the index
	// is always valid.
	return p[category]
}

func relativeTimePatternParts(pattern ecma402.Pattern, unit Unit, numberParts []numberformat.Part) []Part {
	partCount := 0
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartLiteral:
			partCount++
		case ecma402.PatternPartPlaceholder0:
			partCount += len(numberParts)
		}
	}
	out := make([]Part, partCount)
	outIndex := 0
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartLiteral:
			out[outIndex] = Part{Type: PartLiteral, Value: part.Value}
			outIndex++
		case ecma402.PatternPartPlaceholder0:
			for _, numberPart := range numberParts {
				out[outIndex] = Part{
					Type:  PartType(numberPart.Type),
					Value: numberPart.Value,
					Unit:  unit,
				}
				outIndex++
			}
		}
	}
	return out
}

func singularUnit(unit Unit) (Unit, bool) {
	switch unit {
	case Second, Unit("seconds"):
		return Second, true
	case Minute, Unit("minutes"):
		return Minute, true
	case Hour, Unit("hours"):
		return Hour, true
	case Day, Unit("days"):
		return Day, true
	case Week, Unit("weeks"):
		return Week, true
	case Month, Unit("months"):
		return Month, true
	case Quarter, Unit("quarters"):
		return Quarter, true
	case Year, Unit("years"):
		return Year, true
	}
	return "", false
}

func invalidRelativeTimeUnit(unit Unit, loc string) error {
	return invalidValue("unit", string(unit), expectedRelativeTimeUnit, loc)
}

func invalidRelativeTimeValue(value, loc string, err error) error {
	return ecma402.InvalidFiniteNumericValueError(relativeTimeFormatOwner, "value", value, loc, err)
}

func invalidValue(name, value, expected, loc string) error {
	return ecma402.InvalidValueErrorExpected(relativeTimeFormatOwner, name, value, loc, expected, nil)
}
