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

	if !slices.IsSorted(SimpleNumberingSystems) {
		t.Fatalf("SimpleNumberingSystems = %v, want sorted ECMA-402 table", SimpleNumberingSystems)
	}
	for i := 1; i < len(SimpleNumberingSystems); i++ {
		if SimpleNumberingSystems[i] == SimpleNumberingSystems[i-1] {
			t.Fatalf("SimpleNumberingSystems contains duplicate %q", SimpleNumberingSystems[i])
		}
	}
	for _, want := range []string{"arab", "hanidec", "latn", "thai"} {
		if !slices.Contains(SimpleNumberingSystems, want) {
			t.Fatalf("SimpleNumberingSystems missing %q", want)
		}
	}
}

func TestSimpleNumberingSystemsHaveDigitLocalizationData(t *testing.T) {
	t.Parallel()

	for _, numberingSystem := range SimpleNumberingSystems {
		if numberingSystem == "hanidec" {
			continue
		}
		if _, ok := digitZeroByNumberingSystem[numberingSystem]; !ok {
			t.Fatalf("SimpleNumberingSystems contains %q without digit localization data", numberingSystem)
		}
	}
	for numberingSystem := range digitZeroByNumberingSystem {
		if !slices.Contains(SimpleNumberingSystems, numberingSystem) {
			t.Fatalf("digit localization data contains non-simple numbering system %q", numberingSystem)
		}
	}
}
