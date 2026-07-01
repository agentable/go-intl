package codegen

import (
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestEncodeNumberSymbols(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	encodeNumberSymbols(&e, map[string]cldr.NumberSymbols{
		"latn": {
			Decimal:                "a",
			Group:                  "b",
			Percent:                "c",
			Plus:                   "d",
			Minus:                  "e",
			NaN:                    "f",
			Infinity:               "g",
			ApproxSign:             "h",
			RangeSign:              "i",
			PerMille:               "j",
			Exponential:            "k",
			SuperscriptingExponent: "l",
			TimeSeparator:          "m",
		},
	}, table)

	want := []byte{
		1, 0, 4,
		4, 1, 5, 1, 6, 1, 7, 1, 8, 1, 9, 1, 10, 1,
		11, 1, 12, 1, 13, 1, 14, 1, 15, 1, 16, 1,
	}
	assertBytesEqual(t, "encodeNumberSymbols() bytes", e.bytes(), want)
}

func TestEncodeCompactExponentPatterns(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	encodeCompactExponentPatterns(&e, map[int]map[string]string{
		2: {"one": "a"},
		1: {"other": "bb"},
	}, table)

	want := []byte{
		2,
		1, 1, 0, 5, 5, 2,
		2, 1, 7, 3, 10, 1,
	}
	assertBytesEqual(t, "encodeCompactExponentPatterns() bytes", e.bytes(), want)
}
