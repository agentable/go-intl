package collation

import (
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
)

func TestSupportedLocalesReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedLocales", SupportedLocales)
}

func TestSupportedCollationsReturnsCopy(t *testing.T) {
	t.Parallel()

	testcontract.AssertStringSliceReturnsCopy(t, "SupportedCollations", SupportedCollations)
}

func TestSupportedCollationsReflectBackendExtensions(t *testing.T) {
	t.Parallel()

	got := SupportedCollations()
	testcontract.AssertStringSliceSortedUnique(t, "SupportedCollations", got)
	testcontract.AssertStringSliceContainsAll(t, "SupportedCollations", got, "phonebk")
	for _, forbidden := range []string{"default", "search", "standard"} {
		if slices.Contains(got, forbidden) {
			t.Fatalf("SupportedCollations contains %q; ECMA-402 forbids it as an advertised collation value: %v", forbidden, got)
		}
	}
}

func TestSupportedCollationFromExtensionFiltersCollatorValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		extension string
		want      string
	}{
		{name: "specialization", extension: "-u-co-phonebk", want: "phonebk"},
		{name: "case insensitive", extension: "-u-co-PHONEBK", want: "phonebk"},
		{name: "missing co", extension: "-u-kn-true"},
		{name: "default hidden", extension: "-u-co-default"},
		{name: "search hidden", extension: "-u-co-search"},
		{name: "standard hidden", extension: "-u-co-standard"},
		{name: "invalid extension", extension: "co-phonebk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := supportedCollationFromExtension(tt.extension); got != tt.want {
				t.Fatalf("supportedCollationFromExtension(%q) = %q, want %q", tt.extension, got, tt.want)
			}
		})
	}
}

func TestSupportedCollationsForLocaleUsesRelevantExtensionValues(t *testing.T) {
	t.Parallel()

	got := SupportedCollationsForLocale("de")
	if len(got) == 0 || got[0] != "default" {
		t.Fatalf(`SupportedCollationsForLocale("de") = %v, want default first`, got)
	}
	testcontract.AssertStringSliceContainsAll(t, "SupportedCollationsForLocale", got, "phonebk")
	seen := map[string]bool{}
	for _, value := range got {
		if seen[value] {
			t.Fatalf(`SupportedCollationsForLocale("de") contains duplicate %q: %v`, value, got)
		}
		seen[value] = true
	}

	got[0] = "mutated"
	again := SupportedCollationsForLocale("de")
	if again[0] != "default" {
		t.Fatalf(`SupportedCollationsForLocale("de") returned shared storage: %v`, again)
	}
}
