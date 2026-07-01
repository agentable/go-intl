package intltest

import (
	"slices"
	"testing"
)

func TestLocaleListJSON(t *testing.T) {
	t.Parallel()

	got := LocaleListJSON(t, []byte(`["en-US","fr"]`))
	if want := []string{"en-US", "fr"}; !slices.Equal(got.Strings(), want) {
		t.Fatalf("LocaleListJSON().Strings() = %v, want %v", got.Strings(), want)
	}
}
