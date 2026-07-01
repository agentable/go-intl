package codegen

import (
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestAppendRelativeTimeField(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	var e blobEncoder
	appendRelativeTimeField(&e, cldr.RelativeTimeField{
		Future:   map[string]string{"one": "a"},
		Past:     map[string]string{"other": "bb"},
		Relative: map[string]string{"0": "ccc"},
	}, table)

	assertBytesEqual(t, "appendRelativeTimeField() bytes", e.bytes(), []byte{1, 0, 3, 3, 1, 1, 4, 5, 9, 2, 1, 11, 1, 12, 3})
}
