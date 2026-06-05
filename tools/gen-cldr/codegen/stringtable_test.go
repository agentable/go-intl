package codegen

import (
	"strings"
	"testing"
)

func TestStringTable_DeduplicatesAndEmitsData(t *testing.T) {
	t.Parallel()

	table := NewStringTable()
	jan := table.Add("January")
	feb := table.Add("February")
	janAgain := table.Add("January")

	if jan != janAgain {
		t.Fatalf("Add returned different refs for duplicate string: %v != %v", jan, janAgain)
	}
	if got, want := jan.String(), "January"; got != want {
		t.Fatalf("jan.String() = %q, want %q", got, want)
	}
	if got, want := feb.String(), "February"; got != want {
		t.Fatalf("feb.String() = %q, want %q", got, want)
	}

	var out strings.Builder
	if err := table.EmitDataConst(&out, "_data"); err != nil {
		t.Fatalf("EmitDataConst: %v", err)
	}
	src := out.String()
	if !strings.Contains(src, "const _data =") {
		t.Fatalf("EmitDataConst output missing _data const:\n%s", src)
	}
	if strings.Count(src, "January") != 1 {
		t.Fatalf("EmitDataConst output should contain January exactly once:\n%s", src)
	}
}
