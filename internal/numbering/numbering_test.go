package numbering

import (
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
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

func TestSimpleNumberingSystemsMatchesECMA402SimpleSets(t *testing.T) {
	t.Parallel()

	systems := SimpleNumberingSystems()
	want := []string{
		"adlm", "ahom", "arab", "arabext", "bali", "beng", "bhks", "brah",
		"cakm", "cham", "deva", "diak", "fullwide", "gara", "gong", "gonm",
		"gujr", "gukh", "guru", "hanidec", "hmng", "hmnp", "java", "kali",
		"kawi", "khmr", "knda", "krai", "lana", "lanatham", "laoo", "latn",
		"lepc", "limb", "mathbold", "mathdbl", "mathmono", "mathsanb",
		"mathsans", "mlym", "modi", "mong", "mroo", "mtei", "mymr",
		"mymrepka", "mymrpao", "mymrshan", "mymrtlng", "nagm", "newa",
		"nkoo", "olck", "onao", "orya", "osma", "outlined", "rohg", "saur",
		"segment", "shrd", "sind", "sinh", "sora", "sund", "sunu", "takr",
		"talu", "tamldec", "telu", "thai", "tibt", "tirh", "tnsa", "tols",
		"vaii", "wara", "wcho",
	}
	if !slices.Equal(systems, want) {
		t.Fatalf("SimpleNumberingSystems() = %v, want %v", systems, want)
	}
	testcontract.AssertStringSliceSortedUnique(t, "SimpleNumberingSystems", systems)
}

func TestSimpleNumberingSystemsHaveDigitLocalizationData(t *testing.T) {
	t.Parallel()

	systems := SimpleNumberingSystems()
	for _, numberingSystem := range systems {
		if numberingSystem == "hanidec" {
			continue
		}
		if _, ok := digitZeroFor(numberingSystem); !ok {
			t.Fatalf("SimpleNumberingSystems contains %q without digit localization data", numberingSystem)
		}
	}
}

func TestSimpleNumberingSystemsReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SimpleNumberingSystems", SimpleNumberingSystems)
}
