package extract

import (
	"slices"
	"testing"
)

func TestExtractLocalesSortsUniqueTagsAndPinsUndefined(t *testing.T) {
	t.Parallel()

	got := ExtractLocales([]string{"fr", "und", "en", "fr", "", "zh-Hans"}).Tags
	want := []string{"und", "en", "fr", "zh-Hans"}
	if !slices.Equal(got, want) {
		t.Fatalf("ExtractLocales().Tags = %#v, want %#v", got, want)
	}
}

func TestExtractLocalesIncludesUndefinedForEmptyInput(t *testing.T) {
	t.Parallel()

	got := ExtractLocales(nil).Tags
	want := []string{"und"}
	if !slices.Equal(got, want) {
		t.Fatalf("ExtractLocales().Tags = %#v, want %#v", got, want)
	}
}
