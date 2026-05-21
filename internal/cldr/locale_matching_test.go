package cldr

import (
	"testing"
)

func TestLocaleMatchingAccessors(t *testing.T) {
	t.Parallel()

	if got, want := MatchingDistance("en", "en-GB"), 3; got != want {
		t.Fatalf("MatchingDistance(en, en-GB) = %d, want %d", got, want)
	}
	if got, want := MatchingDistance("en", "en"), 0; got != want {
		t.Fatalf("MatchingDistance(en, en) = %d, want %d", got, want)
	}
	paradigm := ParadigmLocales()
	if len(paradigm) == 0 || paradigm[0] != "en" {
		t.Fatalf("ParadigmLocales = %#v", paradigm)
	}
	vars := MatchVariables()
	if len(vars["$enUS"]) == 0 || vars["$enUS"][0] != "AS" {
		t.Fatalf("MatchVariables[$enUS] = %#v", vars["$enUS"])
	}
}
