package datapin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadVersionsCommentsAndExtensions(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "VERSION")
	if err := os.WriteFile(path, []byte("# source pins\n\n cldr = 48.1.0\n icu=78\n tzdata=2025b\n future=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := ReadVersions(path)
	if err != nil || got != (Versions{CLDR: "48.1.0", ICU: "78", TZData: "2025b"}) {
		t.Fatalf("ReadVersions = %+v, %v", got, err)
	}
}

func TestReadTZDataJSONShapes(t *testing.T) {
	t.Parallel()
	const hash = "11810413345fc7805017e27ea9fa4885fd74cd61b2911711ad038f5d28d71474"
	for _, raw := range []string{
		`{"version":"2025b","url":"https://example.test/a","sha256":"` + hash + `","license":"public-domain"}`,
		"{\n\"version\":\n\"2025b\",\n\"url\":\"https:\\/\\/example.test\\/a\",\n\"sha256\":\"" + strings.ToUpper(hash) + "\",\n\"license\":\"public-domain\"\n}",
	} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "tzdata.json")
			if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
				t.Fatal(err)
			}
			got, err := ReadTZData(path)
			if err != nil || got.URL != "https://example.test/a" || got.SHA256 != hash {
				t.Fatalf("ReadTZData = %+v, %v", got, err)
			}
		})
	}
}

func TestRequiredCLDRPackagesReturnsSnapshot(t *testing.T) {
	t.Parallel()

	got := RequiredCLDRPackages()
	if len(got) == 0 {
		t.Fatal("RequiredCLDRPackages returned no packages")
	}
	wantFirst := got[0]
	got[0] = "mutated"

	if again := RequiredCLDRPackages(); again[0] != wantFirst {
		t.Fatalf("RequiredCLDRPackages()[0] = %q, want %q", again[0], wantFirst)
	}
}
