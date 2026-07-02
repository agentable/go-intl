package cldr

import (
	"encoding/json"
	"path/filepath"
	"slices"
	"testing"
)

func TestLoadAvailableLocalesPrefersModern(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "availableLocales.json"), `{
		"availableLocales": {
			"modern": ["en", "fr"],
			"full": ["en", "fr", "zh"]
		}
	}`)

	got, err := loadAvailableLocales(root)
	if err != nil {
		t.Fatalf("loadAvailableLocales() error = %v", err)
	}
	assertStrings(t, "loadAvailableLocales()", got, []string{"en", "fr"})
}

func TestLoadAvailableLocalesFallsBackToFull(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "availableLocales.json"), `{
		"availableLocales": {
			"full": {"modern": ["en", "fr", "zh-Hant"]}
		}
	}`)

	got, err := loadAvailableLocales(root)
	if err != nil {
		t.Fatalf("loadAvailableLocales() error = %v", err)
	}
	assertStrings(t, "loadAvailableLocales()", got, []string{"en", "fr", "zh-Hant"})
}

func TestLoadAvailableLocalesRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "availableLocales.json"), `{`)

	if _, err := loadAvailableLocales(root); err == nil {
		t.Fatal("loadAvailableLocales() succeeded, want error")
	}
}

func TestLoadAvailableLocalesRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "missing availableLocales",
			raw:  `{}`,
		},
		{
			name: "empty availableLocales",
			raw:  `{"availableLocales":{}}`,
		},
		{
			name: "empty modern and full",
			raw:  `{"availableLocales":{"modern":[],"full":[]}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-core", "availableLocales.json"), tc.raw)
			if _, err := loadAvailableLocales(root); err == nil {
				t.Fatalf("loadAvailableLocales(%s) succeeded, want error", tc.name)
			}
		})
	}
}

func TestLoadTimeZoneAliasesFlattensZoneAlias(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "aliases.json"), `{
		"supplemental": {
			"metadata": {
				"alias": {
					"zoneAlias": {
						"America": {
							"Montreal": {
								"_reason": "deprecated",
								"_replacement": "America/Toronto"
							}
						},
						"EST5EDT": {
							"_reason": "deprecated",
							"_replacement": "America/New_York"
						}
					}
				}
			}
		}
	}`)

	got, err := loadTimeZoneAliases(root)
	if err != nil {
		t.Fatalf("loadTimeZoneAliases() error = %v", err)
	}
	want := []TimeZoneAlias{
		{Alias: "America/Montreal", Canonical: "America/Toronto"},
		{Alias: "EST5EDT", Canonical: "America/New_York"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("loadTimeZoneAliases() = %#v, want %#v", got, want)
	}
}

func TestAvailableLocaleListSupportsFlatAndNested(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "flat",
			raw:  `["en", "fr"]`,
			want: []string{"en", "fr"},
		},
		{
			name: "nested modern",
			raw:  `{"modern": ["en", "fr"]}`,
			want: []string{"en", "fr"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got availableLocaleList
			if err := json.Unmarshal([]byte(tc.raw), &got); err != nil {
				t.Fatalf("json.Unmarshal(%s) error = %v", tc.raw, err)
			}
			assertStrings(t, tc.name, []string(got), tc.want)
		})
	}
}

func TestAvailableLocaleListRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	var got availableLocaleList
	if err := json.Unmarshal([]byte(`{`), &got); err == nil {
		t.Fatal("json.Unmarshal({) succeeded, want error")
	}
}

func TestAvailableLocaleListRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{name: "null", raw: `null`},
		{name: "scalar", raw: `"en"`},
		{name: "nested without modern", raw: `{"other": ["en"]}`},
		{name: "nested wrong type", raw: `{"modern": "en"}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var got availableLocaleList
			if err := json.Unmarshal([]byte(tc.raw), &got); err == nil {
				t.Fatalf("json.Unmarshal(%s) succeeded, want error", tc.raw)
			}
		})
	}
}

func TestFilterAvailableLocalesKeepsUndAndAvailableOrder(t *testing.T) {
	t.Parallel()

	got := filterAvailableLocales(
		[]string{"fr", "en", "zh"},
		[]string{"en", "fr"},
	)
	assertStrings(t, "filterAvailableLocales()", got, []string{undefinedLocale, "fr", "en"})
}

func TestFilterAvailableLocalesWithoutAllowlistKeepsAvailableLocales(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		allowlist []string
	}{
		{name: "nil"},
		{name: "empty", allowlist: []string{}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := filterAvailableLocales(
				[]string{"fr", undefinedLocale, "fr", "en", "", "zh", "en"},
				tc.allowlist,
			)
			assertStrings(t, "filterAvailableLocales()", got, []string{undefinedLocale, "fr", "en", "zh"})
		})
	}
}

func TestFilterAvailableLocalesDeduplicatesAndPinsUnd(t *testing.T) {
	t.Parallel()

	got := filterAvailableLocales(
		[]string{undefinedLocale, "fr", "fr", "en", "", "zh", "en"},
		[]string{undefinedLocale, "en", "fr", "fr", ""},
	)
	assertStrings(t, "filterAvailableLocales()", got, []string{undefinedLocale, "fr", "en"})
}

func TestLoadLikelySubtags(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "likelySubtags.json"), `{
		"supplemental": {
			"likelySubtags": {
				"en": "en-Latn-US",
				"zh": "zh-Hans-CN"
			}
		}
	}`)

	got, err := loadLikelySubtags(root)
	if err != nil {
		t.Fatalf("loadLikelySubtags() error = %v", err)
	}
	assertStringMap(t, "loadLikelySubtags()", got, map[string]string{
		"en": "en-Latn-US",
		"zh": "zh-Hans-CN",
	})
}

func TestLoadLikelySubtagsRejectsInvalidJSON(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "likelySubtags.json"), `{`)

	if _, err := loadLikelySubtags(root); err == nil {
		t.Fatal("loadLikelySubtags() succeeded, want error")
	}
}

func TestLoadLikelySubtagsRejectsInvalidShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "missing likelySubtags",
			raw:  `{"supplemental":{}}`,
		},
		{
			name: "null likelySubtags",
			raw:  `{"supplemental":{"likelySubtags":null}}`,
		},
		{
			name: "empty likelySubtags",
			raw:  `{"supplemental":{"likelySubtags":{}}}`,
		},
		{
			name: "wrong likelySubtags type",
			raw:  `{"supplemental":{"likelySubtags":"en-Latn-US"}}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			mustWriteFile(t, filepath.Join(root, "cldr-core", "supplemental", "likelySubtags.json"), tc.raw)
			if _, err := loadLikelySubtags(root); err == nil {
				t.Fatalf("loadLikelySubtags(%s) succeeded, want error", tc.name)
			}
		})
	}
}
