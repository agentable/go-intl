package datetimeformat

import (
	"strings"
	"unicode"
)

func appendLiteralPart(parts []Part, value string) []Part {
	if value == "" {
		return parts
	}
	if len(parts) > 0 && parts[len(parts)-1].Type == PartLiteral {
		parts[len(parts)-1].Value += value
		return parts
	}
	return append(parts, Part{Type: PartLiteral, Value: value})
}

func consumeQuotedPatternLiteral(pattern string) (string, string) {
	var b strings.Builder
	for i := 1; i < len(pattern); i++ {
		if pattern[i] != '\'' {
			b.WriteByte(pattern[i])
			continue
		}
		if i+1 < len(pattern) && pattern[i+1] == '\'' {
			b.WriteByte('\'')
			i++
			continue
		}
		return b.String(), pattern[i+1:]
	}
	return b.String(), ""
}

func isDatePatternField(r rune) bool {
	switch r {
	case 'G', 'y', 'M', 'L', 'd', 'E', 'e', 'c':
		return true
	}
	return false
}

func isTimePatternField(r rune) bool {
	switch r {
	case 'h', 'H', 'K', 'k', 'm', 's', 'S', 'a', 'b', 'B', 'z', 'Z', 'O', 'v', 'X', 'x':
		return true
	}
	return false
}

func trimTrailingLiteralSpace(parts []Part) []Part {
	if len(parts) == 0 || parts[len(parts)-1].Type != PartLiteral {
		return parts
	}
	parts[len(parts)-1].Value = strings.TrimRightFunc(parts[len(parts)-1].Value, unicode.IsSpace)
	if parts[len(parts)-1].Value == "" {
		return parts[:len(parts)-1]
	}
	return parts
}
