package codegen

import "testing"

func TestRenderManifest(t *testing.T) {
	t.Parallel()

	src, err := renderManifest(ManifestInput{
		Generator:     "tools/gen-cldr",
		CLDR:          "48.1.0",
		ICU:           "78",
		TZData:        "2025b",
		LocaleProfile: []string{"en", "fr"},
		InputHashes: []ManifestHash{
			{Name: "internal/cldr/VERSION", SHA256: "abc"},
			{Name: "tools/locale-profile.json", SHA256: "def"},
		},
	})
	if err != nil {
		t.Fatalf("renderManifest: %v", err)
	}
	got := string(src)
	assertSourceContainsAll(t, "renderManifest output", got,
		"type DataManifest struct",
		`CLDR:      "48.1.0"`,
		`"en"`,
		`Name: "tools/locale-profile.json", SHA256: "def"`,
		"func Manifest() DataManifest",
	)
}
