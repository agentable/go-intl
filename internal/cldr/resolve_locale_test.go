package cldr

import (
	"testing"

	"golang.org/x/text/language"
)

func TestResolveLocale_GeneratedLocales(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale(language.MustParse("en-US"))
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	if loc == Undefined {
		t.Fatal("ResolveLocale(en-US) returned Undefined")
	}

	loc, ok = ResolveLocale(language.MustParse("en-Latn-US"))
	if !ok {
		t.Fatal("ResolveLocale(en-Latn-US) should fall back to en")
	}
	if loc == Undefined {
		t.Fatal("ResolveLocale(en-Latn-US) returned Undefined")
	}
}

func TestAvailableLocales_GeneratedLocales(t *testing.T) {
	t.Parallel()

	got := AvailableLocales()
	if len(got) == 0 {
		t.Fatal("AvailableLocales returned empty list")
	}
	if got[0] != "und" {
		t.Fatalf("AvailableLocales()[0] = %q, want und", got[0])
	}
}
