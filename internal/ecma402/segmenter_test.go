package ecma402_test

import (
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestUTF16CodeUnitCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want int
	}{
		{name: "empty", in: "", want: 0},
		{name: "ascii", in: "abc", want: 3},
		{name: "bmp non-ascii", in: "\u00e9", want: 1},
		{name: "combining mark", in: "e\u0301", want: 2},
		{name: "supplementary", in: "\U0001F600", want: 2},
		{name: "mixed", in: "a\U0001F600b", want: 4},
		{name: "regional indicator pair", in: "\U0001F1FA\U0001F1F8", want: 4},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := ecma402.UTF16CodeUnitCount(tc.in); got != tc.want {
				t.Fatalf("UTF16CodeUnitCount(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}
