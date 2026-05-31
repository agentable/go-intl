package relativetimeformat

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
)

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
	Unit  Unit     `json:"unit,omitempty"`
}

func (f *RelativeTimeFormat) FormatInt(value int, unit Unit) (string, error) {
	return f.FormatInt64(int64(value), unit)
}

func (f *RelativeTimeFormat) FormatInt64(value int64, unit Unit) (string, error) {
	parts, err := f.formatInt64ToParts(value, unit)
	if err != nil {
		return "", err
	}
	return joinParts(parts), nil
}

func (f *RelativeTimeFormat) FormatIntToParts(value int, unit Unit) ([]Part, error) {
	return f.FormatInt64ToParts(int64(value), unit)
}

func (f *RelativeTimeFormat) FormatInt64ToParts(value int64, unit Unit) ([]Part, error) {
	return f.formatInt64ToParts(value, unit)
}

func (f *RelativeTimeFormat) formatInt64ToParts(value int64, unit Unit) ([]Part, error) {
	numberParts, err := f.formatAbsInt64ToParts(value)
	if err != nil {
		return nil, invalidValue("value", strconv.FormatInt(value, 10))
	}
	category, err := f.plural.Select(pluralrules.Int(value))
	if err != nil {
		return nil, invalidValue("value", strconv.FormatInt(value, 10))
	}
	return f.numericParts(strconv.FormatInt(value, 10), value < 0, category.String(), numberParts, unit)
}

func (f *RelativeTimeFormat) FormatUint(value uint, unit Unit) (string, error) {
	return f.FormatUint64(uint64(value), unit)
}

func (f *RelativeTimeFormat) FormatUint64(value uint64, unit Unit) (string, error) {
	parts, err := f.FormatUint64ToParts(value, unit)
	if err != nil {
		return "", err
	}
	return joinParts(parts), nil
}

func (f *RelativeTimeFormat) FormatUintToParts(value uint, unit Unit) ([]Part, error) {
	return f.FormatUint64ToParts(uint64(value), unit)
}

func (f *RelativeTimeFormat) FormatUint64ToParts(value uint64, unit Unit) ([]Part, error) {
	category, err := f.plural.Select(pluralrules.Uint(value))
	if err != nil {
		return nil, invalidValue("value", strconv.FormatUint(value, 10))
	}
	return f.numericParts(strconv.FormatUint(value, 10), false, category.String(), f.number.FormatToParts(numberformat.Uint(value)), unit)
}

func (f *RelativeTimeFormat) FormatFloat64(value float64, unit Unit) (string, error) {
	parts, err := f.formatFloat64ToParts(value, unit)
	if err != nil {
		return "", err
	}
	return joinParts(parts), nil
}

func (f *RelativeTimeFormat) FormatFloat64ToParts(value float64, unit Unit) ([]Part, error) {
	return f.formatFloat64ToParts(value, unit)
}

func (f *RelativeTimeFormat) formatFloat64ToParts(value float64, unit Unit) ([]Part, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return nil, invalidValue("value", fmt.Sprint(value))
	}
	category, err := f.plural.Select(pluralrules.Float(value))
	if err != nil {
		return nil, invalidValue("value", fmt.Sprint(value))
	}
	return f.numericParts(strconv.FormatFloat(value, 'f', -1, 64), math.Signbit(value), category.String(), f.number.FormatToParts(numberformat.Float(math.Abs(value))), unit)
}

func (f *RelativeTimeFormat) FormatDecimal(value string, unit Unit) (string, error) {
	parts, err := f.formatDecimalToParts(value, unit)
	if err != nil {
		return "", err
	}
	return joinParts(parts), nil
}

func (f *RelativeTimeFormat) FormatDecimalToParts(value string, unit Unit) ([]Part, error) {
	return f.formatDecimalToParts(value, unit)
}

func (f *RelativeTimeFormat) formatDecimalToParts(value string, unit Unit) ([]Part, error) {
	pluralValue, err := pluralrules.Decimal(value)
	if err != nil {
		return nil, invalidValue("value", value)
	}
	category, err := f.plural.Select(pluralValue)
	if err != nil {
		return nil, invalidValue("value", value)
	}
	absValue := strings.TrimPrefix(value, "-")
	numberValue, err := numberformat.Decimal(absValue)
	if err != nil {
		return nil, invalidValue("value", value)
	}
	numberParts := f.number.FormatToParts(numberValue)
	return f.numericParts(decimalRelativeLiteralKey(value), strings.HasPrefix(value, "-"), category.String(), numberParts, unit)
}

func decimalRelativeLiteralKey(value string) string {
	d, err := ecma402.ParseFiniteDecimalInput(value)
	if err != nil {
		return value
	}
	return d.String()
}

func (f *RelativeTimeFormat) numericParts(literalKey string, past bool, plural string, numberParts []numberformat.Part, unit Unit) ([]Part, error) {
	resolvedUnit, err := singularUnit(unit)
	if err != nil {
		return nil, err
	}
	field, ok := f.relativeTimeField(resolvedUnit)
	if !ok {
		return nil, invalidValue("unit", string(unit))
	}
	if f.resolved.Numeric == NumericAuto {
		if literal := field.Relative[literalKey]; literal != "" {
			return []Part{{Type: PartLiteral, Value: literal}}, nil
		}
	}
	patterns := field.Future
	if past {
		patterns = field.Past
	}
	pattern := patterns[plural]
	if pattern == "" {
		pattern = patterns["other"]
	}
	if pattern == "" {
		return nil, invalidValue("unit", string(unit))
	}
	return makePartsList(pattern, resolvedUnit, numberParts)
}

func (f *RelativeTimeFormat) relativeTimeField(unit Unit) (cldrrelativetime.RelativeTimeField, bool) {
	styles := f.fields[string(unit)]
	if len(styles) == 0 {
		return cldrrelativetime.RelativeTimeField{}, false
	}
	if field, ok := styles[string(f.resolved.Style)]; ok {
		return field, true
	}
	if field, ok := styles[string(LongStyle)]; ok {
		return field, true
	}
	return cldrrelativetime.RelativeTimeField{}, false
}

func makePartsList(pattern string, unit Unit, numberParts []numberformat.Part) ([]Part, error) {
	patternParts, err := ecma402.PartitionPattern(pattern)
	if err != nil {
		return nil, fmt.Errorf("relativetimeformat: malformed CLDR pattern: %w", err)
	}
	out := make([]Part, 0, len(patternParts)+len(numberParts))
	for _, part := range patternParts {
		switch part.Type {
		case "literal":
			out = append(out, Part{Type: PartLiteral, Value: part.Value})
		case "0":
			for _, numberPart := range numberParts {
				out = append(out, Part{
					Type:  PartType(numberPart.Type),
					Value: numberPart.Value,
					Unit:  unit,
				})
			}
		}
	}
	return out, nil
}

func singularUnit(unit Unit) (Unit, error) {
	switch unit {
	case Seconds:
		return Second, nil
	case Minutes:
		return Minute, nil
	case Hours:
		return Hour, nil
	case Days:
		return Day, nil
	case Weeks:
		return Week, nil
	case Months:
		return Month, nil
	case Quarters:
		return Quarter, nil
	case Years:
		return Year, nil
	case Second, Minute, Hour, Day, Week, Month, Quarter, Year:
		return unit, nil
	default:
		return "", invalidValue("unit", string(unit))
	}
}

func (f *RelativeTimeFormat) formatAbsInt64ToParts(value int64) ([]numberformat.Part, error) {
	if value >= 0 {
		return f.number.FormatToParts(numberformat.Int(value)), nil
	}
	numberValue, err := numberformat.Decimal(strings.TrimPrefix(strconv.FormatInt(value, 10), "-"))
	if err != nil {
		return nil, err
	}
	return f.number.FormatToParts(numberValue), nil
}

func joinParts(parts []Part) string {
	size := 0
	for _, part := range parts {
		size += len(part.Value)
	}
	var b strings.Builder
	b.Grow(size)
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func invalidValue(name, value string) error {
	return intlerr.New(intlerr.InvalidValue, "relativetimeformat", name, value, "", intlerr.ErrInvalidValue)
}
