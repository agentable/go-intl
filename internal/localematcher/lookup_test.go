package localematcher

import (
	"slices"
	"testing"
)

func TestLookupMatcher(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		requested []string
		supported []string
		def       string
		locale    string
		extension string
	}{
		{name: "truncates requested locale", requested: []string{"fr-XX", "en"}, supported: []string{"fr", "en"}, def: "en", locale: "fr"},
		{name: "truncates script request", requested: []string{"zh-Hans"}, supported: []string{"zh", "zh-Hant"}, def: "en", locale: "zh"},
		{name: "preserves unicode extension", requested: []string{"th-u-ca-gregory"}, supported: []string{"th"}, def: "en", locale: "th", extension: "-u-ca-gregory"},
		{name: "preserves private use after unicode extension", requested: []string{"en-US-u-nu-thai-x-private"}, supported: []string{"en"}, def: "fr", locale: "en", extension: "-u-nu-thai"},
		{name: "truncates private use", requested: []string{"en-x-private"}, supported: []string{"en"}, def: "fr", locale: "en"},
		{name: "falls back for empty requested", requested: nil, supported: []string{"fr"}, def: "en", locale: "en"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := LookupMatcher(tc.requested, tc.supported, tc.def)
			if got.Locale != tc.locale || got.Extension != tc.extension {
				t.Fatalf("LookupMatcher() = %#v, want locale %q extension %q", got, tc.locale, tc.extension)
			}
		})
	}
}

func TestLookupMatcherUsesDerivedAvailableLocale(t *testing.T) {
	t.Parallel()

	supported := []string{"zh-Hant-HK"}
	if got := BestAvailableLocale(supported, "zh-HK"); got != "zh-HK" {
		t.Fatalf("BestAvailableLocale() = %q, want zh-HK", got)
	}

	got := LookupMatcher([]string{"zh-HK-u-nu-hanidec"}, supported, "en")
	if got.Locale != "zh-HK" || got.DataLocale != "zh-Hant-HK" || got.Extension != "-u-nu-hanidec" {
		t.Fatalf("LookupMatcher() = %#v, want locale zh-HK data zh-Hant-HK extension -u-nu-hanidec", got)
	}
}

func TestRemoveUnicodeExtensionForMatcher(t *testing.T) {
	t.Parallel()

	base, extension := removeUnicodeExtension("en-US-u-ca-gregory-x-private")
	if base != "en-US-x-private" || extension != "-u-ca-gregory" {
		t.Fatalf("removeUnicodeExtension(valid) = %q, %q; want base without u and extension", base, extension)
	}

	base, extension = removeUnicodeExtension("u-ca-gregory")
	if base != "u-ca-gregory" || extension != "" {
		t.Fatalf("removeUnicodeExtension(invalid) = %q, %q; want original locale and empty extension", base, extension)
	}
}

func TestLookupSupportedLocales(t *testing.T) {
	t.Parallel()

	got := LookupSupportedLocales([]string{"en", "fr", "zh-Hant"}, []string{"en-US-u-ca-gregory", "fr-FR", "zh-Hant-TW", "de"})
	want := []string{"en-US", "fr-FR", "zh-Hant-TW"}
	if !slices.Equal(got, want) {
		t.Fatalf("LookupSupportedLocales() = %#v, want %#v", got, want)
	}

	got = LookupSupportedLocales([]string{"zh-Hant-HK"}, []string{"zh-HK-u-nu-hanidec", "en"})
	want = []string{"zh-HK"}
	if !slices.Equal(got, want) {
		t.Fatalf("LookupSupportedLocales(derived) = %#v, want %#v", got, want)
	}
}
