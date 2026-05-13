package datetimeformat

import "strings"

func interpolateDateTimeParts(pattern string, dateParts, timeParts []Part) []Part {
	parts := make([]Part, 0, len(dateParts)+len(timeParts)+strings.Count(pattern, "{")+1)
	for pattern != "" {
		if rest, ok := strings.CutPrefix(pattern, "{0}"); ok {
			parts = append(parts, timeParts...)
			pattern = rest
			continue
		}
		if rest, ok := strings.CutPrefix(pattern, "{1}"); ok {
			parts = append(parts, dateParts...)
			pattern = rest
			continue
		}
		literal, rest := consumeDateTimeConnectorLiteral(pattern)
		if literal != "" {
			parts = appendLiteralPart(parts, literal)
			pattern = rest
			continue
		}
		parts = appendLiteralPart(parts, pattern[:1])
		pattern = pattern[1:]
	}
	return parts
}

func interpolateDateTimeRangeParts(pattern string, dateParts, timeParts []RangePart) []RangePart {
	parts := make([]RangePart, 0, len(dateParts)+len(timeParts)+strings.Count(pattern, "{")+1)
	for pattern != "" {
		if rest, ok := strings.CutPrefix(pattern, "{0}"); ok {
			parts = append(parts, timeParts...)
			pattern = rest
			continue
		}
		if rest, ok := strings.CutPrefix(pattern, "{1}"); ok {
			parts = append(parts, dateParts...)
			pattern = rest
			continue
		}
		literal, rest := consumeDateTimeConnectorLiteral(pattern)
		if literal != "" {
			parts = appendRangeLiteralPart(parts, literal, SourceShared)
			pattern = rest
			continue
		}
		parts = appendRangeLiteralPart(parts, pattern[:1], SourceShared)
		pattern = pattern[1:]
	}
	return parts
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
