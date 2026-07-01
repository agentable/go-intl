package testcontract

import "testing"

// LoadProbe names package state that must remain unloaded after a narrow-index
// accessor runs.
type LoadProbe struct {
	Name   string
	Loaded func() bool
}

// AssertNarrowStringIndexDoesNotLoad verifies that values reads only its narrow
// string index and leaves heavier payload state untouched.
func AssertNarrowStringIndexDoesNotLoad(t testing.TB, name string, values func() []string, probes ...LoadProbe) {
	t.Helper()

	requireStringSliceValues(t, name, values(), "")
	for _, probe := range probes {
		if probe.Loaded() {
			t.Errorf("%s decoded the %s; narrow index must not", name, probe.Name)
		}
	}
}
