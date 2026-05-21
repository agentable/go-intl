package numbering

import (
	"slices"
	"testing"
)

func TestLocalizeDigits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		in              string
		numberingSystem string
		want            string
	}{
		{name: "empty numbering system", in: "012", want: "012"},
		{name: "latn", in: "012", numberingSystem: "latn", want: "012"},
		{name: "no digits", in: "abc", numberingSystem: "arab", want: "abc"},
		{name: "arab", in: "A012", numberingSystem: "arab", want: "A٠١٢"},
		{name: "hanidec", in: "90", numberingSystem: "hanidec", want: "九〇"},
		{name: "unsupported", in: "123", numberingSystem: "unknown", want: "123"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := LocalizeDigits(tc.in, tc.numberingSystem); got != tc.want {
				t.Fatalf("LocalizeDigits(%q, %q) = %q, want %q", tc.in, tc.numberingSystem, got, tc.want)
			}
		})
	}
}

func TestSimpleNumberingSystemsIncludesECMA402SimpleSets(t *testing.T) {
	t.Parallel()

	for _, want := range []string{"arab", "hanidec", "latn", "thai"} {
		if !slices.Contains(SimpleNumberingSystems, want) {
			t.Fatalf("SimpleNumberingSystems missing %q", want)
		}
	}
}
