package datetimeformat

import (
	"strings"

	"github.com/agentable/go-intl/internal/pattern"
)

func interpolateDateTimeParts(text string, dateParts, timeParts []Part) []Part {
	patternParts, err := pattern.Partition(text)
	if err != nil {
		return []Part{{Type: PartLiteral, Value: text}}
	}
	parts := make([]Part, 0, len(dateParts)+len(timeParts)+len(patternParts))
	for _, part := range patternParts {
		switch part.Type {
		case "0":
			parts = append(parts, timeParts...)
		case "1":
			parts = append(parts, dateParts...)
		case pattern.Literal:
			parts = appendDateTimeConnectorLiteral(parts, part.Value)
		default:
			parts = appendLiteralPart(parts, "{"+part.Type+"}")
		}
	}
	return parts
}

func interpolateDateTimeRangeParts(text string, dateParts, timeParts []RangePart) []RangePart {
	patternParts, err := pattern.Partition(text)
	if err != nil {
		return []RangePart{{Type: PartLiteral, Value: text, Source: SourceShared}}
	}
	parts := make([]RangePart, 0, len(dateParts)+len(timeParts)+len(patternParts))
	for _, part := range patternParts {
		switch part.Type {
		case "0":
			parts = append(parts, timeParts...)
		case "1":
			parts = append(parts, dateParts...)
		case pattern.Literal:
			parts = appendDateTimeConnectorRangeLiteral(parts, part.Value)
		default:
			parts = appendRangeLiteralPart(parts, "{"+part.Type+"}", SourceShared)
		}
	}
	return parts
}

func appendDateTimeConnectorLiteral(parts []Part, text string) []Part {
	literal, _ := consumeDateTimeConnectorLiteral(text)
	return appendLiteralPart(parts, literal)
}

func appendDateTimeConnectorRangeLiteral(parts []RangePart, text string) []RangePart {
	literal, _ := consumeDateTimeConnectorLiteral(text)
	return appendRangeLiteralPart(parts, literal, SourceShared)
}

func consumeDateTimeConnectorLiteral(pattern string) (string, string) {
	var literal strings.Builder
	for pattern != "" && pattern[0] != '{' {
		if pattern[0] == '\'' {
			value, rest := consumeQuotedPatternLiteral(pattern)
			literal.WriteString(value)
			pattern = rest
			continue
		}
		literal.WriteByte(pattern[0])
		pattern = pattern[1:]
	}
	return literal.String(), pattern
}
