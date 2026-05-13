package ecma402

import (
	"fmt"
	"strings"

	ecma402types "github.com/agentable/go-intl/internal/ecma402/types"
)

// PartitionPattern mirrors ECMA-402 sec-partitionpattern (formatjs
// PartitionPattern.ts). It splits a pattern of the form "AA{0}BB" into
// alternating literal and placeholder parts:
//
//   - Literal segments carry Type="literal" and Value=raw text.
//   - Placeholder segments carry Type=<name> (e.g. "0", "currency") and
//     Value="" (the value is filled in by the calling Partition* algorithm).
//
// Returns ErrInvalidOption when the pattern contains an unmatched '{'.
func PartitionPattern(pattern string) (ecma402types.Pattern, error) {
	result := make(ecma402types.Pattern, 0, 4)
	beginIndex := strings.IndexByte(pattern, '{')
	nextIndex := 0
	length := len(pattern)
	for beginIndex >= 0 && beginIndex < length {
		endIndex := strings.IndexByte(pattern[beginIndex:], '}')
		if endIndex < 0 {
			return nil, fmt.Errorf(
				"ecma402: invalid pattern %q (unmatched '{'): %w",
				pattern, ErrInvalidOption,
			)
		}
		endIndex += beginIndex
		if beginIndex > nextIndex {
			result = append(result, ecma402types.Part{
				Type:  "literal",
				Value: pattern[nextIndex:beginIndex],
			})
		}
		result = append(result, ecma402types.Part{
			Type:  pattern[beginIndex+1 : endIndex],
			Value: "",
		})
		nextIndex = endIndex + 1
		offset := strings.IndexByte(pattern[nextIndex:], '{')
		if offset < 0 {
			break
		}
		beginIndex = nextIndex + offset
	}
	if nextIndex < length {
		result = append(result, ecma402types.Part{
			Type:  "literal",
			Value: pattern[nextIndex:length],
		})
	}
	return result, nil
}
