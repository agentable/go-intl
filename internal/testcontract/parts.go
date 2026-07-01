package testcontract

import (
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

// AssertParts verifies conformance part records while formatter packages keep
// ownership of their package-local Part and PartType definitions.
func AssertParts[P any](
	t testing.TB,
	name string,
	got []P,
	want []conformance.Part,
	record func(P) conformance.Part,
) {
	t.Helper()

	assertRecords(t, name, got, want, record)
}

// AssertRangeParts verifies conformance range-part records while preserving
// formatter-owned range source types.
func AssertRangeParts[P any](
	t testing.TB,
	name string,
	got []P,
	want []conformance.RangePart,
	record func(P) conformance.RangePart,
) {
	t.Helper()

	assertRecords(t, name, got, want, record)
}

func assertRecords[P any, R comparable](
	t testing.TB,
	name string,
	got []P,
	want []R,
	record func(P) R,
) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d", name, len(got), len(want))
	}
	for i, part := range got {
		gotPart := record(part)
		wantPart := want[i]
		if gotPart != wantPart {
			t.Fatalf("%s[%d] = %#v, want %#v", name, i, gotPart, wantPart)
		}
	}
}

// AssertExpectedRange verifies a conformance expectedRange string.
func AssertExpectedRange(t testing.TB, name, got string, want *string) {
	t.Helper()

	if want == nil {
		t.Fatalf("%s expectedRange is required", name)
		return
	}
	if got != *want {
		t.Fatalf("%s = %q, want %q", name, got, *want)
	}
}
