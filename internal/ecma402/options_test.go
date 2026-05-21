package ecma402

import (
	"testing"
)

func TestInvalidStringOption(t *testing.T) {
	t.Parallel()

	checks := []StringOption{
		RequiredStringOption("required", "a", "a", "b"),
		OptionalStringOption("optional", "", "x"),
		RequiredStringOption("bad", "z", "x", "y"),
	}
	got, ok := InvalidStringOption(checks...)
	if !ok {
		t.Fatal("InvalidStringOption() ok=false, want true")
	}
	if got.Name != "bad" || got.Value != "z" {
		t.Fatalf("InvalidStringOption() = %+v, want bad/z", got)
	}
}

func TestInvalidStringOptionAllValid(t *testing.T) {
	t.Parallel()

	_, ok := InvalidStringOption(
		RequiredStringOption("required", "b", "a", "b"),
		OptionalStringOption("optional", "", "x"),
	)
	if ok {
		t.Fatal("InvalidStringOption() ok=true, want false")
	}
}

func TestInvalidIntegerOption(t *testing.T) {
	t.Parallel()

	checks := []IntegerOption{
		{Name: "unset", Value: 100, Min: 1, Max: 10},
		{Name: "ok", Value: 5, Min: 1, Max: 10, Set: true},
		{Name: "bad", Value: 11, Min: 1, Max: 10, Set: true},
	}
	got, ok := InvalidIntegerOption(checks...)
	if !ok {
		t.Fatal("InvalidIntegerOption() ok=false, want true")
	}
	if got.Name != "bad" || got.Value != 11 {
		t.Fatalf("InvalidIntegerOption() = %+v, want bad/11", got)
	}
}

func TestInvalidIntegerOptionAllValid(t *testing.T) {
	t.Parallel()

	_, ok := InvalidIntegerOption(
		IntegerOption{Name: "unset", Value: 100, Min: 1, Max: 10},
		IntegerOption{Name: "ok", Value: 10, Min: 1, Max: 10, Set: true},
	)
	if ok {
		t.Fatal("InvalidIntegerOption() ok=true, want false")
	}
}
