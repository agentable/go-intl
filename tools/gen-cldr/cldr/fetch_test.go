package cldr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveLocalDirRequiresCorePackages(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	for _, name := range RequiredPackages {
		if err := os.MkdirAll(filepath.Join(dir, name), 0o777); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
	}

	got, err := ResolveLocalDir(dir)
	if err != nil {
		t.Fatalf("ResolveLocalDir: %v", err)
	}
	if got != dir {
		t.Fatalf("ResolveLocalDir = %q, want %q", got, dir)
	}
}

func TestResolveLocalDirRejectsMissingPackage(t *testing.T) {
	t.Parallel()

	if _, err := ResolveLocalDir(t.TempDir()); err == nil {
		t.Fatal("ResolveLocalDir succeeded with no packages")
	}
}
