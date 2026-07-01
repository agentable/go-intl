package localematcher

import (
	"slices"
	"testing"
)

func TestLocaleDistanceOverridesAreSorted(t *testing.T) {
	t.Parallel()

	if !slices.IsSortedFunc(localeDistanceOverrides[:], compareLocaleDistanceRecord) {
		t.Fatal("localeDistanceOverrides must stay sorted for BinarySearchFunc")
	}
}

func TestLocaleDistanceOverride(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		desired   string
		supported string
		want      int
		ok        bool
	}{
		{name: "exact maximized chinese traditional", desired: "zh-TW", supported: "zh-Hant", want: 0, ok: true},
		{name: "canada prefers us over gb", desired: "en-CA", supported: "en-US", want: 39, ok: true},
		{name: "caribbean spanish prefers americas", desired: "es-KY", supported: "es-419", want: 39, ok: true},
		{name: "unknown pair", desired: "fr-CA", supported: "fr-FR"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := localeDistanceOverride(tc.desired, tc.supported)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("localeDistanceOverride(%q, %q) = %d, %v; want %d, %v", tc.desired, tc.supported, got, ok, tc.want, tc.ok)
			}
		})
	}
}
