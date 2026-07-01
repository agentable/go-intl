package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
	"github.com/agentable/go-intl/tools/gen-cldr/codegen"
	"github.com/agentable/go-intl/tools/internal/localeprofile"
)

func TestRunValidatesVersionBeforeReadingCLDRInputs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	versionPath := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}
	for _, name := range cldr.RequiredPackages() {
		pkgDir := filepath.Join(dir, "node_modules", name)
		if err := os.MkdirAll(pkgDir, 0o777); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		meta := `{"name":"` + name + `","version":"48.1.0"}`
		if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(meta), 0o666); err != nil {
			t.Fatalf("write %s package.json: %v", name, err)
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

func TestBuildManifestRecordsInputHashes(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	versionContent := "cldr=48.1.0\nicu=78\ntzdata=2025b\n"
	versionPath := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(versionPath, []byte(versionContent), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}
	profileContent := `{"locales":["en","fr"]}`
	profilePath := filepath.Join(dir, "locale-profile.json")
	if err := os.WriteFile(profilePath, []byte(profileContent), 0o666); err != nil {
		t.Fatalf("write locale-profile.json: %v", err)
	}

	cldrRoot := filepath.Join(dir, "node_modules")
	packages := cldr.RequiredPackages()
	packageContents := make(map[string]string)
	for _, name := range packages {
		pkgDir := filepath.Join(cldrRoot, name)
		if err := os.MkdirAll(pkgDir, 0o777); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		content := fmt.Sprintf(`{"name":%q,"version":"48.1.0"}`, name)
		packageContents[name] = content
		if err := os.WriteFile(filepath.Join(pkgDir, "package.json"), []byte(content), 0o666); err != nil {
			t.Fatalf("write %s package.json: %v", name, err)
		}
	}

	got, err := buildManifest(
		versionPath,
		profilePath,
		cldrRoot,
		cldr.Versions{CLDR: "48.1.0", ICU: "78", TZData: "2025b"},
		localeprofile.Profile{Locales: []string{"en", "fr"}},
	)
	if err != nil {
		t.Fatalf("buildManifest: %v", err)
	}
	if got.Generator != "tools/gen-cldr" || got.CLDR != "48.1.0" || got.ICU != "78" || got.TZData != "2025b" {
		t.Fatalf("buildManifest versions = %#v", got)
	}
	if !slices.Equal(got.LocaleProfile, []string{"en", "fr"}) {
		t.Fatalf("LocaleProfile = %v, want [en fr]", got.LocaleProfile)
	}

	wantHashes := make([]codegen.ManifestHash, 2+len(packages))
	wantHashes[0] = codegen.ManifestHash{Name: "internal/cldr/VERSION", SHA256: testSHA256(versionContent)}
	wantHashes[1] = codegen.ManifestHash{Name: "tools/locale-profile.json", SHA256: testSHA256(profileContent)}
	for i, name := range packages {
		wantHashes[2+i] = codegen.ManifestHash{
			Name:   filepath.ToSlash(filepath.Join("cldr-json", name, "package.json")),
			SHA256: testSHA256(packageContents[name]),
		}
	}
	if !slices.Equal(got.InputHashes, wantHashes) {
		t.Fatalf("InputHashes = %#v, want %#v", got.InputHashes, wantHashes)
	}
}

func testSHA256(content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", sum)
}
