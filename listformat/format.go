package listformat

import (
	"github.com/agentable/go-intl/internal/pattern"
)

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
}

func (f *ListFormat) Format(list []string) string {
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	case 2:
		return applyListPattern(f.pattern.Pair, list[0], list[1])
	}

	result := applyListPattern(f.pattern.End, list[len(list)-2], list[len(list)-1])
	for i := len(list) - 3; i > 0; i-- {
		result = applyListPattern(f.pattern.Middle, list[i], result)
	}
	return applyListPattern(f.pattern.Start, list[0], result)
}

func applyListPattern(text, first, second string) string {
	return pattern.FormatIndexed(text, first, second)
}

func (f *ListFormat) FormatToParts(list []string) []Part {
	switch len(list) {
	case 0:
		return nil
	case 1:
		return []Part{{Type: PartElement, Value: list[0]}}
	case 2:
		return listPatternParts(f.pattern.Pair, elementPart(list[0]), elementPart(list[1]))
	}

	result := listPatternParts(f.pattern.End, elementPart(list[len(list)-2]), elementPart(list[len(list)-1]))
	for i := len(list) - 3; i > 0; i-- {
		result = listPatternParts(f.pattern.Middle, elementPart(list[i]), result)
	}
	return listPatternParts(f.pattern.Start, elementPart(list[0]), result)
}

func elementPart(value string) []Part {
	return []Part{{Type: PartElement, Value: value}}
}

func listPatternParts(text string, first, second []Part) []Part {
	patternParts, err := pattern.Partition(text)
	if err != nil {
		return nil
	}
	out := make([]Part, 0, len(patternParts)+len(first)+len(second))
	for _, part := range patternParts {
		switch part.Type {
		case "0":
			out = append(out, first...)
		case "1":
			out = append(out, second...)
		case pattern.Literal:
			out = append(out, Part{Type: PartLiteral, Value: part.Value})
		}
	}
	return out
}
