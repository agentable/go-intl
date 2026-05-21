package locale

import "testing"

func boolPtr(v bool) *bool {
	return &v
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	numeric := true
	loc, err := New("en", Options{Numeric: &numeric})
	if err != nil {
		t.Fatal(err)
	}
	numeric = false

	if !loc.Numeric() {
		t.Fatal("Numeric() = false, want true")
	}
}
