package codegen

import (
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestAppendListPattern(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	appendListPattern(&e, cldr.ListPattern{
		Pair:   "a",
		Start:  "bb",
		Middle: "ccc",
		End:    "dddd",
	}, table)

	assertBytesEqual(t, "appendListPattern() bytes", e.bytes(), []byte{0, 1, 1, 2, 3, 3, 6, 4})
}
