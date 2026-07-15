package cldr

import (
	"fmt"
	"strings"
)

type datePatternWidth struct {
	min   int
	max   int
	oneOf map[int]struct{}
}

var executableDatePatternFields = map[byte]datePatternWidth{
	'G': {min: 1, max: 5},
	'y': {min: 1, max: 20},
	'M': {min: 1, max: 5},
	'L': {min: 1, max: 5},
	'd': {min: 1, max: 2},
	'E': {min: 1, max: 6},
	'e': {min: 1, max: 6},
	'c': {min: 1, max: 6},
	'h': {min: 1, max: 2},
	'H': {min: 1, max: 2},
	'K': {min: 1, max: 2},
	'k': {min: 1, max: 2},
	'm': {min: 1, max: 2},
	's': {min: 1, max: 2},
	'S': {min: 1, max: 9},
	'a': {min: 1, max: 5},
	'b': {min: 1, max: 5},
	'B': {min: 1, max: 5},
	'z': {min: 1, max: 4},
	'Z': {min: 1, max: 5},
	'O': {oneOf: widthSet(1, 4)},
	'v': {oneOf: widthSet(1, 4)},
	'X': {min: 1, max: 5},
	'x': {min: 1, max: 5},
}

func widthSet(values ...int) map[int]struct{} {
	out := make(map[int]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func validateExecutableDatePattern(pattern string) error {
	for i := 0; i < len(pattern); {
		if pattern[i] == '\'' {
			next, err := skipDatePatternLiteral(pattern, i)
			if err != nil {
				return err
			}
			i = next
			continue
		}
		field := pattern[i]
		if !isASCIIAlpha(field) {
			i++
			continue
		}
		end := i + 1
		for end < len(pattern) && pattern[end] == field {
			end++
		}
		width := end - i
		allowed, ok := executableDatePatternFields[field]
		if !ok {
			return fmt.Errorf("unknown executable field %q in %q", field, pattern)
		}
		if !allowed.accepts(width) {
			return fmt.Errorf("unsupported width %d for field %q in %q", width, field, pattern)
		}
		i = end
	}
	return nil
}

func (w datePatternWidth) accepts(width int) bool {
	if w.oneOf != nil {
		_, ok := w.oneOf[width]
		return ok
	}
	return width >= w.min && width <= w.max
}

func skipDatePatternLiteral(pattern string, start int) (int, error) {
	if start+1 < len(pattern) && pattern[start+1] == '\'' {
		return start + 2, nil
	}
	for i := start + 1; i < len(pattern); i++ {
		if pattern[i] != '\'' {
			continue
		}
		if i+1 < len(pattern) && pattern[i+1] == '\'' {
			i++
			continue
		}
		return i + 1, nil
	}
	return 0, fmt.Errorf("unterminated quote in %q", pattern)
}

func isExecutableDateSkeleton(skeleton string) bool {
	skeleton, _, _ = strings.Cut(skeleton, "-count-")
	found := false
	for i := 0; i < len(skeleton); {
		if skeleton[i] == '\'' {
			next, err := skipDatePatternLiteral(skeleton, i)
			if err != nil {
				return false
			}
			i = next
			continue
		}
		if !isASCIIAlpha(skeleton[i]) {
			i++
			continue
		}
		field := skeleton[i]
		if _, ok := executableDatePatternFields[field]; !ok {
			return false
		}
		found = true
		for i < len(skeleton) && skeleton[i] == field {
			i++
		}
	}
	return found
}

func isASCIIAlpha(value byte) bool {
	return value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z'
}

func validateIndexedTemplate(value string, allowThird bool) error {
	counts := [3]int{}
	for i := 0; i < len(value); {
		switch value[i] {
		case '{':
			if i+2 >= len(value) || value[i+2] != '}' || value[i+1] < '0' || value[i+1] > '9' {
				return fmt.Errorf("malformed placeholder at byte %d in %q", i, value)
			}
			index := int(value[i+1] - '0')
			if index > 2 || index == 2 && !allowThird {
				return fmt.Errorf("unexpected placeholder {%d} in %q", index, value)
			}
			counts[index]++
			i += 3
		case '}':
			return fmt.Errorf("unmatched closing brace at byte %d in %q", i, value)
		default:
			i++
		}
	}
	if counts[0] != 1 || counts[1] != 1 {
		return fmt.Errorf("expected one {0} and one {1}, got %q", value)
	}
	if counts[2] > 1 {
		return fmt.Errorf("expected at most one {2}, got %q", value)
	}
	return nil
}
