package testcontract

import (
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

type partFixture struct {
	typ, value, unit string
}

type rangePartFixture struct {
	typ, value, source string
}

func TestAssertParts(t *testing.T) {
	t.Parallel()

	got := []partFixture{{typ: "integer", value: "1"}, {typ: "unit", value: "m", unit: "meter"}}
	want := []conformance.Part{{Type: "integer", Value: "1"}, {Type: "unit", Value: "m", Unit: "meter"}}
	AssertParts(t, "parts", got, want, func(part partFixture) conformance.Part {
		return conformance.Part{Type: part.typ, Value: part.value, Unit: part.unit}
	})
}

func TestAssertRangeParts(t *testing.T) {
	t.Parallel()

	got := []rangePartFixture{{typ: "integer", value: "1", source: "startRange"}}
	want := []conformance.RangePart{{Type: "integer", Value: "1", Source: "startRange"}}
	AssertRangeParts(t, "range parts", got, want, func(part rangePartFixture) conformance.RangePart {
		return conformance.RangePart{Type: part.typ, Value: part.value, Source: part.source}
	})
}

func TestAssertExpectedRange(t *testing.T) {
	t.Parallel()

	want := "1-2"
	AssertExpectedRange(t, "FormatRange", "1-2", &want)
}
