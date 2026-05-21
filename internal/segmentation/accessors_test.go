package segmentation

import (
	"slices"
	"testing"
)

func TestSupportedLocales(t *testing.T) {
	t.Parallel()

	got := SupportedLocales()
	want := []string{
		"ar",
		"de",
		"en",
		"en-GB",
		"en-US",
		"es",
		"fr",
		"hi",
		"it",
		"pt",
		"ru",
	}

	if !slices.Equal(got, want) {
		t.Fatalf("SupportedLocales() = %v, want %v", got, want)
	}
}

func TestSupportedLocalesExcludesTailoredLocales(t *testing.T) {
	t.Parallel()

	got := SupportedLocales()
	for _, unsupported := range []string{"ja", "km", "lo", "my", "th", "zh", "zh-Hans", "zh-Hant"} {
		if slices.Contains(got, unsupported) {
			t.Fatalf("SupportedLocales() includes %q, want dictionary/tailoring locale excluded", unsupported)
		}
	}
}

func TestSupportedLocalesReturnsSnapshot(t *testing.T) {
	t.Parallel()

	got := SupportedLocales()
	if len(got) == 0 {
		t.Fatal("SupportedLocales() is empty")
	}
	got[0] = "ja"

	if slices.Contains(SupportedLocales(), "ja") {
		t.Fatal("SupportedLocales() leaked mutable backing storage")
	}
}
