package cldr

import (
	"testing"

	"golang.org/x/text/language"
)

func TestUnitAccessors(t *testing.T) {
	t.Parallel()

	loc, ok := ResolveLocale(language.MustParse("en-US"))
	if !ok {
		t.Fatal("ResolveLocale(en-US) ok=false")
	}
	if got, want := loc.UnitPattern("meter", "long", "one"), "{0} meter"; got != want {
		t.Fatalf("UnitPattern(meter, long, one) = %q, want %q", got, want)
	}
	if got, want := loc.UnitPattern("meter", "short", "other"), "{0} m"; got != want {
		t.Fatalf("UnitPattern(meter, short, other) = %q, want %q", got, want)
	}
	if got, want := loc.CompoundUnitPattern("long"), "{0} per {1}"; got != want {
		t.Fatalf("CompoundUnitPattern(long) = %q, want %q", got, want)
	}
}
