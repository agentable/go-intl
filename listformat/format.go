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
	pair   contextualListTemplate
	start  ecma402.Pattern
	middle ecma402.Pattern
	end    contextualListTemplate
}

func compileListTemplates(pattern cldrlist.ListPattern, dataLocale string) listTemplates {
	return listTemplates{
		pair:   compileContextualListTemplate(pattern.Pair, dataLocale),
		start:  compileListTemplate(pattern.Start),
		middle: compileListTemplate(pattern.Middle),
		end:    compileContextualListTemplate(pattern.End, dataLocale),
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
		panic(fmt.Errorf("listformat: malformed embedded CLDR list pattern: %w", err))
	}
	return parts
}

// Format concatenates the values from the same partition as FormatToParts.
func (f *ListFormat) Format(list []string) string {
	var scratch [16]Part
	parts := f.formatParts(list, scratch[:0])
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
	// A non-nil empty result marshals to "[]", matching native.
	return f.formatParts(list, []Part{})
}

func (f *ListFormat) formatParts(list []string, out []Part) []Part {
	switch len(list) {
	case 0:
		return out
	case 1:
		return append(out, Part{Type: PartElement, Value: list[0]})
	}

	templates := f.templates
	last := list[len(list)-1]
	var pattern, end ecma402.Pattern
	var count int
	if len(list) == 2 {
		pattern = templates.pair.forElement(last)
		count = len(pattern)
	} else {
		pattern = templates.start
		end = templates.end.forElement(last)
		count = len(pattern) + (len(list)-3)*(len(templates.middle)-1) + len(end) - 1
	}
	if count > cap(out) {
		out = make([]Part, count)
	} else {
		out = out[:count]
	}
	left, right := 0, len(out)-1
	// Each fold wraps the remaining suffix at {1}. Fill the wrapper from both
	// ends, leaving its suffix in place instead of copying it at every level.
	for i, first := range list[:len(list)-1] {
		for j, part := range pattern {
			if part.Type == ecma402.PatternPartPlaceholder1 {
				for k := len(pattern) - 1; k > j; k-- {
					out[right] = listPatternPart(pattern[k], first)
					right--
				}
				break
			}
			out[left] = listPatternPart(part, first)
			left++
		}
		pattern = templates.middle
		if i == len(list)-3 {
			pattern = end
		}
	}
	out[left] = Part{Type: PartElement, Value: last}
	return out
}

func listPatternPart(part ecma402.Part, first string) Part {
	if part.Type == ecma402.PatternPartPlaceholder0 {
		return Part{Type: PartElement, Value: first}
	}
	return Part{Type: PartLiteral, Value: part.Value}
}
