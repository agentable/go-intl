package locale

import "testing"

func boolPtr(v bool) *bool {
	return &v
}

func stringPtr(v string) *string {
	return &v
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	numeric := true
	calendar := "gregory"
	loc, err := New("en", Options{Calendar: &calendar, Numeric: &numeric})
	if err != nil {
		t.Fatal(err)
	}
	calendar = "buddhist"
	numeric = false

	if !loc.Numeric() {
		t.Fatal("Numeric() = false, want true")
	}
	if got := loc.Calendar(); got != "gregory" {
		t.Fatalf("Calendar() = %q, want gregory", got)
	}
}
