package listformat

import (
	"strings"

	cldrlist "github.com/agentable/go-intl/internal/cldr/list"
	"github.com/agentable/go-intl/internal/ecma402"
)

type Part struct {
	Type  PartType `json:"type"`
	Value string   `json:"value"`
}

type listTemplates struct {
	pair   ecma402.Pattern
	start  ecma402.Pattern
	middle ecma402.Pattern
	end    ecma402.Pattern
}

func compileListTemplates(pattern cldrlist.ListPattern) listTemplates {
	return listTemplates{
		pair:   compileListTemplate(pattern.Pair),
		start:  compileListTemplate(pattern.Start),
		middle: compileListTemplate(pattern.Middle),
		end:    compileListTemplate(pattern.End),
	}
}

func compileListTemplate(text string) ecma402.Pattern {
	parts, err := ecma402.PartitionPattern(text)
	if err != nil {
		return nil
	}
	return parts
}

func (f *ListFormat) Format(list []string) string {
	templates := f.templates
	switch len(list) {
	case 0:
		return ""
	case 1:
		return list[0]
	case 2:
		var b strings.Builder
		b.Grow(listPatternTextSize(templates.pair, len(list[0]), len(list[1])))
		writeListPatternText(&b, templates.pair, list[0], list[1])
		return b.String()
	}

	lastPairIndex := len(list) - 2
	var b strings.Builder
	b.Grow(listTextSize(templates, list, 0, lastPairIndex))
	writeListText(&b, templates, list, 0, lastPairIndex)
	return b.String()
}

func (f *ListFormat) FormatToParts(list []string) []Part {
	templates := f.templates
	switch len(list) {
	case 0:
		return nil
	case 1:
		return []Part{{Type: PartElement, Value: list[0]}}
	case 2:
		return listPatternParts(templates.pair, list[0], []Part{{Type: PartElement, Value: list[1]}})
	}

	lastPairIndex := len(list) - 2
	result := listPatternParts(templates.end, list[lastPairIndex], []Part{{Type: PartElement, Value: list[len(list)-1]}})
	for i := lastPairIndex - 1; i > 0; i-- {
		result = listPatternParts(templates.middle, list[i], result)
	}
	return listPatternParts(templates.start, list[0], result)
}

func listPatternParts(pattern ecma402.Pattern, first string, second []Part) []Part {
	partCount := 0
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			partCount++
		case ecma402.PatternPartPlaceholder1:
			partCount += len(second)
		case ecma402.PatternPartLiteral:
			partCount++
		}
	}

	out := make([]Part, partCount)
	outIndex := 0
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			out[outIndex] = Part{Type: PartElement, Value: first}
			outIndex++
		case ecma402.PatternPartPlaceholder1:
			outIndex += copy(out[outIndex:], second)
		case ecma402.PatternPartLiteral:
			out[outIndex] = Part{Type: PartLiteral, Value: part.Value}
			outIndex++
		}
	}
	return out
}

func writeListText(b *strings.Builder, templates listTemplates, list []string, index, lastPairIndex int) {
	pattern := listNestedTemplate(templates, index, lastPairIndex)
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			b.WriteString(list[index])
		case ecma402.PatternPartPlaceholder1:
			if index == lastPairIndex {
				b.WriteString(list[index+1])
			} else {
				writeListText(b, templates, list, index+1, lastPairIndex)
			}
		case ecma402.PatternPartLiteral:
			b.WriteString(part.Value)
		}
	}
}

func writeListPatternText(b *strings.Builder, pattern ecma402.Pattern, first, second string) {
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			b.WriteString(first)
		case ecma402.PatternPartPlaceholder1:
			b.WriteString(second)
		case ecma402.PatternPartLiteral:
			b.WriteString(part.Value)
		}
	}
}

func listTextSize(templates listTemplates, list []string, index, lastPairIndex int) int {
	secondLen := len(list[index+1])
	if index < lastPairIndex {
		secondLen = listTextSize(templates, list, index+1, lastPairIndex)
	}
	return listPatternTextSize(listNestedTemplate(templates, index, lastPairIndex), len(list[index]), secondLen)
}

func listPatternTextSize(pattern ecma402.Pattern, firstLen, secondLen int) int {
	size := 0
	for _, part := range pattern {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			size += firstLen
		case ecma402.PatternPartPlaceholder1:
			size += secondLen
		case ecma402.PatternPartLiteral:
			size += len(part.Value)
		}
	}
	return size
}

func listNestedTemplate(templates listTemplates, index, lastPairIndex int) ecma402.Pattern {
	if index == lastPairIndex {
		return templates.end
	}
	if index == 0 {
		return templates.start
	}
	return templates.middle
}
