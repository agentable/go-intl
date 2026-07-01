package datetimeformat

import (
	"slices"
	"testing"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
)

func TestDateLocaleDataForwardsDateExtensionKeys(t *testing.T) {
	t.Parallel()

	adapter := dateLocaleData{}
	source := cldrdate.DateLocaleData{}
	tests := []struct {
		name   string
		locale string
		key    string
	}{
		{name: "calendar", locale: "en-US", key: "ca"},
		{name: "hour-cycle", locale: "en-US", key: "hc"},
		{name: "numbering-system", locale: "en-US", key: "nu"},
		{name: "locale-numbering-system", locale: "th", key: "nu"},
		{name: "unknown", locale: "en-US", key: "co"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := adapter.For(tc.locale, tc.key)
			want := source.For(tc.locale, tc.key)
			if (got == nil) != (want == nil) || !slices.Equal(got, want) {
				t.Fatalf("dateLocaleData.For(%q, %q) = %v, want %v", tc.locale, tc.key, got, want)
			}
		})
	}
}

func TestDateLocaleDataReturnsFreshSlices(t *testing.T) {
	adapter := dateLocaleData{}

	first := adapter.For("en-US", "ca")
	if len(first) == 0 {
		t.Fatal(`dateLocaleData.For("en-US", "ca") returned no calendar data`)
	}
	first[0] = "mutated"

	got := adapter.For("en-US", "ca")
	want := cldrdate.DateLocaleData{}.For("en-US", "ca")
	if !slices.Equal(got, want) {
		t.Fatalf(`dateLocaleData.For("en-US", "ca") after caller mutation = %v, want %v`, got, want)
	}
}
