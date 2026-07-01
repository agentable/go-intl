package ecma402

import "testing"

func TestResolvedScalarReturnsFreshPointer(t *testing.T) {
	t.Parallel()

	got := ResolvedScalar(3)
	if got == nil || *got != 3 {
		t.Fatalf("ResolvedScalar(3) = %v, want pointer to 3", got)
	}
	*got = 4
	if next := ResolvedScalar(3); next == nil || *next != 3 {
		t.Fatalf("ResolvedScalar(3) after caller mutation = %v, want pointer to 3", next)
	}
}

func TestCloneResolvedScalar(t *testing.T) {
	t.Parallel()

	value := 7
	got := CloneResolvedScalar(&value)
	if got == nil || *got != 7 {
		t.Fatalf("CloneResolvedScalar(&7) = %v, want pointer to 7", got)
	}
	if got == &value {
		t.Fatal("CloneResolvedScalar returned the original pointer")
	}
	*got = 8
	if value != 7 {
		t.Fatalf("caller mutation changed source value to %d, want 7", value)
	}
	if got := CloneResolvedScalar[int](nil); got != nil {
		t.Fatalf("CloneResolvedScalar(nil) = %v, want nil", got)
	}
}

func TestResolvedScalarValue(t *testing.T) {
	t.Parallel()

	type style string
	value := style("short")
	if got := ResolvedScalarValue(&value); got != style("short") {
		t.Fatalf("ResolvedScalarValue(&short) = %q, want short", got)
	}
	if got := ResolvedScalarValue[style](nil); got != "" {
		t.Fatalf("ResolvedScalarValue(nil) = %q, want zero value", got)
	}
}
