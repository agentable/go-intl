package listformat

import (
	"strings"

	"github.com/agentable/go-intl/internal/pattern"
)

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
}

func (f *ListFormat) Format(list []string) string {
	return joinParts(f.FormatToParts(list))
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
