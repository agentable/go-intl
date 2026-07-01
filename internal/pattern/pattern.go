// Package pattern parses generated CLDR placeholder patterns.
package pattern

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalid classifies malformed placeholder patterns.
var ErrInvalid = errors.New("invalid pattern")

// Literal is the part type for raw pattern text.
const Literal = "literal"

// Part is one literal or placeholder segment of a generated pattern.
type Part struct {
	Type  string
	Value string
}

// Pattern is a parsed sequence of placeholder pattern parts.
type Pattern []Part

// Partition splits text into literal and placeholder parts.
func Partition(text string) (Pattern, error) {
	result := make(Pattern, 0, 4)
	beginIndex := strings.IndexByte(text, '{')
	nextIndex := 0
	length := len(text)
	for beginIndex >= 0 && beginIndex < length {
		endIndex := strings.IndexByte(text[beginIndex:], '}')
		if endIndex < 0 {
			return nil, fmt.Errorf("pattern: unmatched placeholder in %q: %w", text, ErrInvalid)
		}
		endIndex += beginIndex
		if beginIndex > nextIndex {
			result = append(result, Part{
				Type:  Literal,
				Value: text[nextIndex:beginIndex],
			})
		}
		result = append(result, Part{
			Type: text[beginIndex+1 : endIndex],
		})
		nextIndex = endIndex + 1
		offset := strings.IndexByte(text[nextIndex:], '{')
		if offset < 0 {
			break
		}
		beginIndex = nextIndex + offset
	}
	if nextIndex < length {
		result = append(result, Part{
			Type:  Literal,
			Value: text[nextIndex:length],
		})
	}
	return result, nil
}

// FormatIndexed replaces numeric placeholders with the corresponding value.
func FormatIndexed(text string, values ...string) string {
	beginIndex := strings.IndexByte(text, '{')
	if beginIndex < 0 {
		return text
	}

	var out strings.Builder
	out.Grow(len(text))
	nextIndex := 0
	length := len(text)
	for beginIndex >= 0 && beginIndex < length {
		endIndex := strings.IndexByte(text[beginIndex:], '}')
		if endIndex < 0 {
			return text
		}
		endIndex += beginIndex
		out.WriteString(text[nextIndex:beginIndex])
		name := text[beginIndex+1 : endIndex]
		if value, ok := indexedValue(name, values); ok {
			out.WriteString(value)
		} else {
			out.WriteByte('{')
			out.WriteString(name)
			out.WriteByte('}')
		}
		nextIndex = endIndex + 1
		offset := strings.IndexByte(text[nextIndex:], '{')
		if offset < 0 {
			break
		}
		beginIndex = nextIndex + offset
	}
	out.WriteString(text[nextIndex:length])
	return out.String()
}

func indexedValue(name string, values []string) (string, bool) {
	if name == "" {
		return "", false
	}
	index := 0
	maxInt := int(^uint(0) >> 1)
	for i := range len(name) {
		if name[i] < '0' || name[i] > '9' {
			return "", false
		}
		digit := int(name[i] - '0')
		if index > (maxInt-digit)/10 {
			return "", false
		}
		index = index*10 + digit
		if index >= len(values) {
			return "", false
		}
	}
	return values[index], true
}
