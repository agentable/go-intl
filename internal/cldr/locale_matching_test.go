package cldr

import (
	"testing"

	"golang.org/x/text/language"
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

func TestPluralAccessorsUseGeneratedRules(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale(language.MustParse("en"))
	if !ok {
		t.Fatal("ResolveLocale(en) ok=false")
	}
	if got := loc.Cardinal(NewOperand("1")); got != One {
		t.Fatalf("Cardinal(1) = %q, want %q", got, One)
	}
	if got := loc.Ordinal(NewOperand("2")); got != Two {
		t.Fatalf("Ordinal(2) = %q, want %q", got, Two)
	}
}
