package gointl

import "testing"

func TestOptionPointerHelpers(t *testing.T) {
	t.Parallel()

	if got := Int(2); got == nil || *got != 2 {
		t.Fatalf("Int(2) = %v, want pointer to 2", got)
	}
	if got := Bool(false); got == nil || *got {
		t.Fatalf("Bool(false) = %v, want pointer to false", got)
	}
	if got := String("USD"); got == nil || *got != "USD" {
		t.Fatalf("String(USD) = %v, want pointer to USD", got)
	}
	type optionString string
	if got := String(optionString("meter")); got == nil || *got != "meter" {
		t.Fatalf("String(optionString(meter)) = %v, want pointer to meter", got)
	}
}
