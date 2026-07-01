package ecma402

import (
	"fmt"

	"github.com/agentable/go-intl/internal/pattern"
)

const (
	// PatternPartLiteral is the part type for literal text returned by PartitionPattern.
	PatternPartLiteral = pattern.Literal
	// PatternPartPlaceholder0 is the placeholder part type for "{0}".
	PatternPartPlaceholder0 = "0"
	// PatternPartPlaceholder1 is the placeholder part type for "{1}".
	PatternPartPlaceholder1 = "1"
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
func PartitionPattern(text string) (Pattern, error) {
	parts, err := pattern.Partition(text)
	if err != nil {
		return nil, fmt.Errorf("ecma402: invalid pattern %q: %w: %w", text, ErrInvalidOption, err)
	}
	result := make(Pattern, len(parts))
	for i, part := range parts {
		result[i] = Part{Type: part.Type, Value: part.Value}
	}
	return result, nil
}
