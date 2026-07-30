package listformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

// TestFormatEqualsConcatOfParts locks the ECMA-402 invariant that Format is the
// concatenation of FormatToParts' values. Format now derives from the parts walk,
// so this can never drift; the cases span the 0/1/2/n list-size branches.
func TestFormatEqualsConcatOfParts(t *testing.T) {
	t.Parallel()

	f, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	lists := [][]string{
		nil,
		{"a"},
		{"a", "b"},
		{"a", "b", "c"},
		{"a", "b", "c", "d", "e"},
	}
	for _, list := range lists {
		got := f.Format(list)
		want := ""
		for _, p := range f.FormatToParts(list) {
			want += p.Value
		}
		if got != want {
			t.Errorf("Format(%v) = %q, concat(FormatToParts) = %q", list, got, want)
		}
	}
}
