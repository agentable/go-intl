package codegen

import (
	"strings"
	"testing"
)

type sampleLiteral struct {
	Name string
	Rank int
}

func TestDayPeriodRulesLiteral_EmptyIsTyped(t *testing.T) {
	t.Parallel()

	got := dayPeriodRulesLiteral(nil, NewStringTable())
	want := "map[Locale][]DayPeriodRange{}"
	if got != want {
		t.Fatalf("dayPeriodRulesLiteral(nil) = %q, want %q", got, want)
	}
}

func TestEmitLiteral_DeterministicMapAndStruct(t *testing.T) {
	t.Parallel()

	value := map[string]sampleLiteral{
		"b": {Name: "beta", Rank: 2},
		"a": {Name: "alpha", Rank: 1},
	}

	var out strings.Builder
	if err := EmitLiteral(&out, value); err != nil {
		t.Fatalf("EmitLiteral: %v", err)
	}
	got := out.String()
	want := "map[string]codegen.sampleLiteral{\n\t\"a\": {Name: \"alpha\", Rank: 1},\n\t\"b\": {Name: \"beta\", Rank: 2},\n}"
	if got != want {
		t.Fatalf("EmitLiteral output mismatch:\ngot:  %s\nwant: %s", got, want)
	}
}
