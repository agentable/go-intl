package cldr

import (
	"os"
	"path/filepath"
	"testing"
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

	pkgDir := filepath.Join(dir, "cldr-core")
	if err := os.MkdirAll(pkgDir, 0o777); err != nil {
		t.Fatalf("mkdir cldr-core: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{"name":"cldr-core","version":"48.1.0"}`), 0o666); err != nil {
		t.Fatalf("write package.json: %v", err)
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
