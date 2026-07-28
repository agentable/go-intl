package datetimeformat

import (
	"strings"
	"unicode/utf8"

	"github.com/agentable/go-intl/internal/ecma402"
)

func compileDateTimeProgram(text string) ecma402.Pattern {
	patternParts, err := ecma402.PartitionPattern(text)
	if err != nil {
		return ecma402.Pattern{{Type: ecma402.PatternPartLiteral, Value: normalizePatternLiteral(text)}}
	}
	for i := range patternParts {
		if patternParts[i].Type == ecma402.PatternPartLiteral {
			patternParts[i].Value = normalizeDateTimeConnectorLiteral(patternParts[i].Value)
		}
	}
	return patternParts
}

func interpolateDateTimeParts(patternParts ecma402.Pattern, dateParts, timeParts []Part) []Part {
	parts := make([]Part, 0, len(dateParts)+len(timeParts)+len(patternParts))
	for _, part := range patternParts {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			parts = append(parts, timeParts...)
		case ecma402.PatternPartPlaceholder1:
			parts = append(parts, dateParts...)
		case ecma402.PatternPartLiteral:
			parts = appendLiteralPart(parts, part.Value)
		default:
			parts = appendLiteralPart(parts, "{"+part.Type+"}")
		}
	}
	return parts
}

func interpolateDateTimeRangeParts(patternParts ecma402.Pattern, dateParts, timeParts []RangePart) []RangePart {
	parts := make([]RangePart, 0, len(dateParts)+len(timeParts)+len(patternParts))
	for _, part := range patternParts {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			parts = append(parts, timeParts...)
		case ecma402.PatternPartPlaceholder1:
			parts = append(parts, dateParts...)
		case ecma402.PatternPartLiteral:
			parts = appendRangeLiteralPart(parts, part.Value, SourceShared)
		default:
			parts = appendRangeLiteralPart(parts, "{"+part.Type+"}", SourceShared)
		}
	}
	return parts
}

func normalizeDateTimeConnectorLiteral(pattern string) string {
	var literal strings.Builder
	for pattern != "" {
		if pattern[0] == '\'' {
			value, rest := consumeQuotedPatternLiteral(pattern)
			literal.WriteString(value)
			pattern = rest
			continue
		}
		_, size := utf8.DecodeRuneInString(pattern)
		literal.WriteString(pattern[:size])
		pattern = pattern[size:]
	}
	return normalizePatternLiteral(literal.String())
}
