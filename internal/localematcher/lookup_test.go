package localematcher

import "testing"

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

func TestLookupSupportedLocales(t *testing.T) {
	t.Parallel()

	got := LookupSupportedLocales([]string{"en", "fr", "zh-Hant"}, []string{"en-US-u-ca-gregory", "fr-FR", "zh-Hant-TW", "de"})
	want := []string{"en-US", "fr-FR", "zh-Hant-TW"}
	if len(got) != len(want) {
		t.Fatalf("LookupSupportedLocales() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("LookupSupportedLocales() = %#v, want %#v", got, want)
		}
	}
}
