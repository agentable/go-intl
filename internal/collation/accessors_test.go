package collation

import (
	"slices"
	"testing"
)

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	locales := SupportedLocales()
	if len(locales) == 0 {
		t.Fatal("SupportedLocales returned no locale tags")
	}
	locales[0] = "mutated"
	if slices.Contains(SupportedLocales(), "mutated") {
		t.Fatal("SupportedLocales returned a shared slice; callers can corrupt the backend capability cache")
	}
}

func TestSupportedCollationsReturnsCopy(t *testing.T) {
	t.Parallel()

	collations := SupportedCollations()
	if len(collations) == 0 {
		return
	}
	collations[0] = "mutated"
	if slices.Contains(SupportedCollations(), "mutated") {
		t.Fatal("SupportedCollations returned a shared slice; callers can corrupt the backend capability cache")
	}
}
