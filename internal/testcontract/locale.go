package testcontract

import (
	"slices"
	"testing"
)

// AssertLocaleListStrings verifies that a locale list matches the expected
// canonical string sequence exactly.
func AssertLocaleListStrings[S ~[]E, E interface{ String() string }](t testing.TB, name string, got S, want []string) {
	t.Helper()

	gotStrings := make([]string, len(got))
	for i, value := range got {
		gotStrings[i] = value.String()
	}
	if !slices.Equal(gotStrings, want) {
		t.Fatalf("%s = %v, want %v", name, gotStrings, want)
	}
}
