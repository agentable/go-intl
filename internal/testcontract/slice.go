// Package testcontract provides small reusable assertions for repository-wide
// test contracts.
package testcontract

import (
	"slices"
	"testing"
)

// AssertStringSliceReturnsCopy verifies that values returns a fresh slice whose
// caller-visible mutations do not corrupt the accessor's cached state.
func AssertStringSliceReturnsCopy(t testing.TB, name string, values func() []string) {
	t.Helper()

	first := requireStringSliceValues(t, name, values(), "")
	assertStringSliceReturnsCopy(t, name, first, values)
}

// AssertOptionalStringSliceReturnsCopy verifies the same caller-owned snapshot
// contract when an accessor is allowed to report no values.
func AssertOptionalStringSliceReturnsCopy(t testing.TB, name string, values func() []string) {
	t.Helper()

	first := values()
	if len(first) == 0 {
		return
	}
	assertStringSliceReturnsCopy(t, name, first, values)
}

func assertStringSliceReturnsCopy(t testing.TB, name string, first []string, values func() []string) {
	t.Helper()

	want := first[0]
	first[0] = "mutated"
	got := requireStringSliceValues(t, name, values(), " after caller mutation")
	if got[0] != want {
		t.Fatalf("%s reused caller storage; first value = %q, want %q", name, got[0], want)
	}
}

func requireStringSliceValues(t testing.TB, name string, values []string, suffix string) []string {
	t.Helper()

	if len(values) == 0 {
		t.Fatalf("%s returned no values%s", name, suffix)
	}
	return values
}

// AssertStringSliceSortedUnique verifies the stable ordering contract for public
// capability lists and generated supported-value tables.
func AssertStringSliceSortedUnique(t testing.TB, name string, values []string) {
	t.Helper()

	if !slices.IsSorted(values) {
		t.Fatalf("%s = %v, want sorted", name, values)
	}
	for i := 1; i < len(values); i++ {
		if values[i] == values[i-1] {
			t.Fatalf("%s contains duplicate %q: %v", name, values[i], values)
		}
	}
}

// AssertStringSliceContainsAll verifies the stable anchor values a capability
// list or generated identifier table must keep advertising.
func AssertStringSliceContainsAll(t testing.TB, name string, values []string, required ...string) {
	t.Helper()

	for _, want := range required {
		if !slices.Contains(values, want) {
			t.Fatalf("%s missing %q in %v", name, want, values)
		}
	}
}

// AssertStringSliceSubset verifies every advertised value belongs to an
// authoritative profile list.
func AssertStringSliceSubset(t testing.TB, name string, values []string, profileName string, profile []string) {
	t.Helper()

	requireStringSliceValues(t, name, values, "")
	allowed := make(map[string]bool, len(profile))
	for _, value := range profile {
		allowed[value] = true
	}
	for _, value := range values {
		if !allowed[value] {
			t.Errorf("%s value %q is not in %s", name, value, profileName)
		}
	}
}
