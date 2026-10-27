package cldr

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/agentable/go-intl/tools/internal/datapin"
)

func TestReadVersionFileAndCrossCheck(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	versionPath := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	got, err := ReadVersionFile(versionPath)
	if err != nil {
		t.Fatalf("ReadVersionFile: %v", err)
	}
	if got != (Versions{CLDR: "48.1.0", ICU: "78", TZData: "2025b"}) {
		t.Fatalf("ReadVersionFile = %+v", got)
	}

	for _, name := range datapin.RequiredCLDRPackages() {
		pkgDir := filepath.Join(dir, name)
		if err := os.MkdirAll(pkgDir, 0o777); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{"name":"`+name+`","version":"48.1.0"}`), 0o666); err != nil {
			t.Fatal(err)
		}
	}
	if err := CrossCheck(dir, got); err != nil {
		t.Fatalf("CrossCheck: %v", err)
	}
}

func TestCrossCheckRejectsVersionMismatch(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "cldr-core")
	if err := os.MkdirAll(pkgDir, 0o777); err != nil {
		t.Fatalf("mkdir cldr-core: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{"name":"cldr-core","version":"47.0.0"}`), 0o666); err != nil {
		t.Fatalf("write package.json: %v", err)
	}
	if err := CrossCheck(dir, Versions{CLDR: "48.1.0"}); err == nil {
		t.Fatal("CrossCheck succeeded for mismatched version")
	}
}

func TestCrossCheckRequiresEveryPackageIdentity(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, pkg, body string }{
		{"non-core mismatch", "cldr-numbers-full", `{"name":"cldr-numbers-full","version":"47.0.0"}`},
		{"wrong name", "cldr-units-full", `{"name":"unrelated","version":"48.1.0"}`},
		{"missing metadata", "cldr-misc-full", ""},
		{"malformed metadata", "cldr-dates-full", "{"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for _, name := range datapin.RequiredCLDRPackages() {
				dir := filepath.Join(root, name)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"`+name+`","version":"48.1.0"}`), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(root, tc.pkg, "package.json")
			if tc.body == "" {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(path, []byte(tc.body), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := CrossCheck(root, Versions{CLDR: "48.1.0"}); err == nil {
				t.Fatal("CrossCheck accepted invalid package identity")
			}
		})
	}
}

func TestReadVersionFileRejectsMalformedPins(t *testing.T) {
	t.Parallel()
	const valid = "cldr=48.1.0\nicu=78\ntzdata=2025b\n"
	for _, raw := range []string{valid + "cldr=48.1.0\n", valid + "bad line\n", valid + "=empty\n", "cldr=48.1.0\nicu=78\n", "cldr=bad\nicu=78\ntzdata=2025b\n"} {
		t.Run(raw, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "VERSION")
			mustWriteFile(t, path, raw)
			if _, err := ReadVersionFile(path); err == nil {
				t.Fatalf("accepted %q", raw)
			}
		})
	}
}
