package locale_test

import (
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
)

// ECMA-402 §14.3 (`Intl.Locale.prototype.calendars` and siblings) returns a
// fresh Array on every call so JS callers can mutate the result without
// corrupting the Locale's shared CLDR data. The Go bridge must preserve
// that invariant: the slice returned by each info getter is the caller's
// to mutate, and a second call must return identical-but-independent data.
func TestLocaleInfoGettersReturnIndependentSlices(t *testing.T) {
	t.Parallel()
	loc := intltest.Locale(t, "en-US")

	cases := []struct {
		name string
		get  func() []string
	}{
		{"Calendars", loc.GetCalendars},
		{"Collations", loc.GetCollations},
		{"HourCycles", loc.GetHourCycles},
		{"NumberingSystems", loc.GetNumberingSystems},
		{"TimeZones", loc.GetTimeZones},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			first := tc.get()
			if len(first) == 0 {
				t.Skipf("%s returned empty slice for en-US; mutation invariant trivially holds", tc.name)
			}
			// Snapshot to compare against, then mutate the returned slice.
			snapshot := slices.Clone(first)
			first[0] = "__mutated__"
			second := tc.get()
			if slices.Equal(second, first) {
				t.Fatalf("%s second call reflects in-place mutation: %v", tc.name, second)
			}
			if !slices.Equal(second, snapshot) {
				t.Fatalf("%s second call diverged from snapshot:\n got=%v\n want=%v", tc.name, second, snapshot)
			}
		})
	}
}
