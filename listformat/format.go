package listformat

import (
	"fmt"
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

// compileListTemplate parses a generated CLDR list pattern at construction. The
// generator validates every pattern (tools/gen-cldr/cldr/list_patterns.go
// validateListPattern), so a parse failure means the embedded data is corrupt —
// a broken invariant we fail loudly on (Must* idiom) rather than silently
// degrading to empty output.
func compileListTemplate(text string) ecma402.Pattern {
	parts, err := ecma402.PartitionPattern(text)
	if err != nil {
		panic(fmt.Sprintf("listformat: malformed embedded CLDR list pattern %q: %v", text, err))
	}
	return parts
}

// Format is the concatenation of FormatToParts' values. ECMA-402 defines one
// partition per formatter and derives the string from it; a second recursive
// text traversal would be a byte-for-byte drift liability.
func (f *ListFormat) Format(list []string) string {
	parts := f.FormatToParts(list)
	size := 0
	for _, p := range parts {
		size += len(p.Value)
	}
	var b strings.Builder
	b.Grow(size)
	for _, p := range parts {
		b.WriteString(p.Value)
	}
	return b.String()
}

func (f *ListFormat) FormatToParts(list []string) []Part {
	templates := f.templates
	switch len(list) {
	case 0:
		// Empty, non-nil so it marshals to "[]" (matching native), never "null".
		return []Part{}
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
