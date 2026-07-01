package localematcher

import "testing"

func TestAvailableLocalesForDerivesFallbacksAndDataLocales(t *testing.T) {
	t.Parallel()

	got := availableLocalesFor([]string{
		"zh-Hant-HK",
		"zh-Hant-HK-u-nu-hanidec",
		"en-US",
		"en-US",
	})
	want := map[string]availableLocale{
		"zh-Hant-HK": {locale: "zh-Hant-HK", dataLocale: "zh-Hant-HK"},
		"zh-HK":      {locale: "zh-HK", dataLocale: "zh-Hant-HK", derived: true},
		"zh-Hant":    {locale: "zh-Hant", dataLocale: "zh-Hant-HK", derived: true},
		"zh":         {locale: "zh", dataLocale: "zh-Hant-HK", derived: true},
		"en-US":      {locale: "en-US", dataLocale: "en-US"},
		"en":         {locale: "en", dataLocale: "en-US", derived: true},
	}
	if len(got) != len(want) {
		t.Fatalf("availableLocalesFor() = %#v, want %d records", got, len(want))
	}
	seen := map[string]bool{}
	for _, locale := range got {
		if seen[locale.locale] {
			t.Fatalf("availableLocalesFor() returned duplicate locale %q in %#v", locale.locale, got)
		}
		seen[locale.locale] = true
		if wantLocale, ok := want[locale.locale]; !ok || locale != wantLocale {
			t.Fatalf("availableLocalesFor() record = %#v, want %#v", locale, wantLocale)
		}
	}
}
