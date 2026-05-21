package codegen

import (
	"strings"
	"testing"
)

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
	for _, want := range []string{
		"type DataManifest struct",
		`CLDR:      "48.1.0"`,
		`"en"`,
		`Name: "tools/locale-profile.json", SHA256: "def"`,
		"func Manifest() DataManifest",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("renderManifest output missing %q:\n%s", want, got)
		}
	}
}
