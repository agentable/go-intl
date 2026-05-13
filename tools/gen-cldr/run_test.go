package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func TestRunValidatesVersionBeforeReadingPhase3Inputs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	versionPath := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}
	for _, name := range cldr.RequiredPackages {
		pkgDir := filepath.Join(dir, "node_modules", name)
		if err := os.MkdirAll(pkgDir, 0o777); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		if name == "cldr-core" {
			if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(`{"name":"cldr-core","version":"48.1.0"}`), 0o666); err != nil {
				t.Fatalf("write package.json: %v", err)
			}
		}
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	err := Run(context.Background(), Config{
		CLDRDir:     filepath.Join(dir, "node_modules"),
		OutDir:      filepath.Join(dir, "out"),
		VersionFile: versionPath,
		ProfileFile: writeLocaleProfileFixture(t, dir),
	}, log)
	if err == nil || !strings.Contains(err.Error(), "availableLocales.json") {
		t.Fatalf("Run error = %v, want missing availableLocales.json after version validation", err)
	}
}
