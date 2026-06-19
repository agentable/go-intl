package relativetimeformat

import (
	"fmt"
	"strings"

	cldrrelativetime "github.com/agentable/go-intl/internal/cldr/relativetime"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/numberformat"
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
	return joinParts(parts), nil
}

func (f *RelativeTimeFormat) FormatToParts(value Value, unit Unit) ([]Part, error) {
	if value.kind == valueInvalid {
		return nil, invalidValue("value", value.errValue)
	}
	category, err := f.plural.Select(value.pluralValue)
	if err != nil {
		return nil, invalidValue("value", value.errValue)
	}
	return f.numericParts(value.literalKey, value.past, category.String(), f.number.FormatToParts(value.numberValue), unit)
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
	case "seconds":
		return Second, nil
	case "minutes":
		return Minute, nil
	case "hours":
		return Hour, nil
	case "days":
		return Day, nil
	case "weeks":
		return Week, nil
	case "months":
		return Month, nil
	case "quarters":
		return Quarter, nil
	case "years":
		return Year, nil
	case Second, Minute, Hour, Day, Week, Month, Quarter, Year:
		return unit, nil
	default:
		return "", invalidValue("unit", string(unit))
	}
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
