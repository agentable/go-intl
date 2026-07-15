package relativetimeformat

import (
	"fmt"
	"maps"
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
	selection, err := resolveRelativeTimePattern(value, unit, f)
	if err != nil {
		return "", err
	}
	if selection.literal != "" {
		return selection.literal, nil
	}
	return relativeTimePatternText(selection.pattern, f.number.Format(value.numberFormatValue())), nil
}

func (f *RelativeTimeFormat) FormatToParts(value Value, unit Unit) ([]Part, error) {
	selection, err := resolveRelativeTimePattern(value, unit, f)
	if err != nil {
		return nil, err
	}
	if selection.literal != "" {
		return []Part{{Type: PartLiteral, Value: selection.literal}}, nil
	}
	return relativeTimePatternParts(selection.pattern, selection.unit, f.number.FormatToParts(value.numberFormatValue())), nil
}

type relativeTimePatternSelection struct {
	unit    Unit
	pattern ecma402.Pattern
	literal string
}

func resolveRelativeTimePattern(value Value, unit Unit, f *RelativeTimeFormat) (relativeTimePatternSelection, error) {
	if value.invalidErr != nil {
		return relativeTimePatternSelection{}, invalidRelativeTimeValue(value.errValue, f.resolved.Locale.String(), value.invalidErr)
	}
	resolvedUnit, ok := singularUnit(unit)
	if !ok {
		return relativeTimePatternSelection{}, invalidRelativeTimeUnit(unit, f.resolved.Locale.String())
	}
	field, ok := f.fields[resolvedUnit]
	if !ok {
		return relativeTimePatternSelection{}, invalidRelativeTimeUnit(unit, f.resolved.Locale.String())
	}
	if f.resolved.Numeric == NumericAuto {
		if literal := field.relative[value.literalKey()]; literal != "" {
			return relativeTimePatternSelection{unit: resolvedUnit, literal: literal}, nil
		}
	}
	category := f.plural.Select(value.pluralRulesValue())
	patterns := field.future
	if value.isPast() {
		patterns = field.past
	}
	// pattern is always non-empty: compileRelativeTimePatternSet requires a
	// non-empty "other" pattern at construction and backfills every category from
	// it, so no format-time empty-pattern check is needed.
	return relativeTimePatternSelection{unit: resolvedUnit, pattern: patterns.pattern(category)}, nil
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
			continue
		}
		compiled, err := compileRelativeTimeField(field)
		if err != nil {
			return nil, err
		}
		out[unit] = compiled
	}
	return out, nil
}

var relativeTimeUnits = [...]Unit{Second, Minute, Hour, Day, Week, Month, Quarter, Year}

const expectedRelativeTimeUnit = `one of "second", "minute", "hour", "day", "week", "month", "quarter", "year", or their plural forms`

func selectRelativeTimeField(fields cldrrelativetime.RelativeTimeFields, unit Unit, style Style) (cldrrelativetime.RelativeTimeField, bool) {
	styles := fields[string(unit)]
	if len(styles) == 0 {
		return cldrrelativetime.RelativeTimeField{}, false
	}
	if field, ok := styles[string(style)]; ok {
		return field, true
	}
	if field, ok := styles[string(LongStyle)]; ok {
		return field, true
	}
	return cldrrelativetime.RelativeTimeField{}, false
}

func compileRelativeTimeField(field cldrrelativetime.RelativeTimeField) (relativeTimeField, error) {
	future, err := compileRelativeTimePatternSet(field.Future)
	if err != nil {
		return relativeTimeField{}, err
	}
	past, err := compileRelativeTimePatternSet(field.Past)
	if err != nil {
		return relativeTimeField{}, err
	}
	return relativeTimeField{
		future:   future,
		past:     past,
		relative: maps.Clone(field.Relative),
	}, nil
}

func compileRelativeTimePatternSet(patterns map[string]string) (relativeTimePatternSet, error) {
	other, err := compileRelativeTimePattern(patterns[pluralrules.Other.String()])
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
	if pattern == "" {
		return nil, nil
	}
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

func relativeTimePatternText(pattern ecma402.Pattern, number string) string {
	size := 0
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartLiteral:
			size += len(part.Value)
		case ecma402.PatternPartPlaceholder0:
			size += len(number)
		}
	}
	var b strings.Builder
	b.Grow(size)
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartLiteral:
			b.WriteString(part.Value)
		case ecma402.PatternPartPlaceholder0:
			b.WriteString(number)
		}
	}
	return b.String()
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
