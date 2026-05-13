package cldr

import "testing"

func TestVersion(t *testing.T) {
	t.Parallel()

	got := Version()
	if got.CLDR != "48.1.0" || got.ICU != "78" || got.TZData != "2025b" {
		t.Fatalf("Version() = %#v", got)
	}
}
