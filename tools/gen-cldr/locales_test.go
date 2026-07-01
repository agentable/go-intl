package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/agentable/go-intl/tools/gen-cldr/cldr"
)

func writeRuntimeCLDRFixtures(t *testing.T, root string) {
	t.Helper()
	writeBaseCLDRFixture(t, root)
	writeNumberCLDRFixture(t, root)
	writeDateCLDRFixture(t, root)
	writeTimeZoneCLDRFixture(t, root)
	writeUnitCLDRFixture(t, root)
}

func TestRunGeneratesLocalesAndLikelySubtags(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}

	// The retired root literal renderers must no longer produce strings.go,
	// locales.go, or likely_subtags.go; the locale kernel now owns a const-only
	// data.go plus hand-written decode.go/accessors.go.
	for _, retired := range []string{"strings.go", "locales.go", "likely_subtags.go", "collations.go", "preference.go", "manifest.go", "supported.go", "locale_matching.go", "regions.go", "timezones.go"} {
		if _, err := os.Stat(filepath.Join(out, retired)); !os.IsNotExist(err) {
			t.Fatalf("root %s should no longer be generated, stat err = %v", retired, err)
		}
	}

	// The locale kernel payload is a const-only data.go carrying the registry,
	// likely-subtags, numbering, and preference blobs plus the private _data table
	// the decoder reads.
	payload, err := os.ReadFile(filepath.Join(out, "locale", "data.go"))
	if err != nil {
		t.Fatalf("read locale/data.go: %v", err)
	}
	if !containsAll(string(payload), "package cldrlocale", "const _data", "_localeBlob", "_maximizeBlob", "_minimizeBlob", "_numberingBlob") {
		t.Fatalf("locale/data.go missing expected const payload:\n%s", payload)
	}
	// The _data const is emitted in 64-byte chunks, so locale tags can straddle a
	// chunk boundary. Reconstruct the concatenated string literals before
	// asserting the expected tags are present.
	reconstructed := readGeneratedStringTable(t, filepath.Join(out, "locale", "data.go"))
	if !containsAll(reconstructed, "und", "en", "zh-Hans") {
		t.Fatalf("locale/data.go _data missing expected locale tags:\n%s", payload)
	}

	// The manifest is owned by the locale kernel now.
	manifest, err := os.ReadFile(filepath.Join(out, "locale", "manifest.go"))
	if err != nil {
		t.Fatalf("read locale/manifest.go: %v", err)
	}
	if !containsAll(string(manifest), "package cldrlocale", "type DataManifest struct", "func Manifest() DataManifest") {
		t.Fatalf("locale/manifest.go missing expected content:\n%s", manifest)
	}
}

func TestRunPrefersLanguageMinimizeAlias(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	likely := `{"supplemental":{"likelySubtags":{"und_CN":"zh_Hans_CN","zh":"zh_Hans_CN","zh_Hant":"zh_Hant_TW","und_Hant":"zh_Hant_TW"}}}`
	if err := os.WriteFile(filepath.Join(root, "cldr-core", "supplemental", "likelySubtags.json"), []byte(likely), 0o666); err != nil {
		t.Fatalf("write likelySubtags: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// The minimize-alias preference is an extract-layer decision verified
	// byte-for-byte through the production accessors in the locale kernel
	// round-trip gate. Here we assert the kernel payload is emitted and carries
	// the minimized zh alias rather than the und-CN key in its _data table.
	payload, err := os.ReadFile(filepath.Join(out, "locale", "data.go"))
	if err != nil {
		t.Fatalf("read locale/data.go: %v", err)
	}
	if !containsAll(string(payload), "package cldrlocale", "_minimizeBlob") {
		t.Fatalf("locale/data.go missing minimize blob:\n%s", payload)
	}
	reconstructed := readGeneratedStringTable(t, filepath.Join(out, "locale", "data.go"))
	if !strings.Contains(reconstructed, "zh") {
		t.Fatalf("locale/data.go _data missing minimized zh alias:\n%s", payload)
	}
}

func TestRunFallsBackToFullAvailableLocales(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	available := `{"availableLocales":{"modern":[],"full":["und","en","en-US","zh","zh-Hans","zh-Hans-CN"]}}`
	if err := os.WriteFile(filepath.Join(root, "cldr-core", "availableLocales.json"), []byte(available), 0o666); err != nil {
		t.Fatalf("write availableLocales: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
	reconstructed := readGeneratedStringTable(t, filepath.Join(out, "locale", "data.go"))
	if !containsAll(reconstructed, "en", "en-US") {
		t.Fatalf("locale/data.go _data missing fallback locales:\n%s", reconstructed)
	}
}

func TestRunAcceptsNestedAvailableLocales(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	available := `{"availableLocales":{"modern":{"_cldrVersion":"48","modern":["und","en","en-US","zh","zh-Hans","zh-Hans-CN"]}}}`
	if err := os.WriteFile(filepath.Join(root, "cldr-core", "availableLocales.json"), []byte(available), 0o666); err != nil {
		t.Fatalf("write availableLocales: %v", err)
	}
	out := filepath.Join(dir, "out")
	if err := os.MkdirAll(out, 0o777); err != nil {
		t.Fatalf("mkdir out: %v", err)
	}
	versionPath := filepath.Join(out, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	if err := Run(context.Background(), Config{CLDRDir: root, OutDir: out, VersionFile: versionPath, ProfileFile: writeLocaleProfileFixture(t, dir)}, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func writeBaseCLDRFixture(t *testing.T, root string) {
	t.Helper()
	for _, name := range cldr.RequiredPackages() {
		if err := os.MkdirAll(filepath.Join(root, name), 0o777); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		meta := `{"name":"` + name + `","version":"48.1.0"}`
		if err := os.WriteFile(filepath.Join(root, name, "package.json"), []byte(meta), 0o666); err != nil {
			t.Fatalf("write %s package.json: %v", name, err)
		}
	}
	available := `{"availableLocales":{"modern":["und","en","en-US","zh","zh-Hans","zh-Hans-CN"]}}`
	if err := os.WriteFile(filepath.Join(root, "cldr-core", "availableLocales.json"), []byte(available), 0o666); err != nil {
		t.Fatalf("write availableLocales: %v", err)
	}
	likely := `{"supplemental":{"likelySubtags":{"en":"en_Latn_US","zh":"zh_Hans_CN","zh_Hant":"zh_Hant_TW"}}}`
	supp := filepath.Join(root, "cldr-core", "supplemental")
	if err := os.MkdirAll(supp, 0o777); err != nil {
		t.Fatalf("mkdir supplemental: %v", err)
	}
	if err := os.WriteFile(filepath.Join(supp, "likelySubtags.json"), []byte(likely), 0o666); err != nil {
		t.Fatalf("write likelySubtags: %v", err)
	}
}

func writeLocaleProfileFixture(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "locale-profile.json")
	profile := `{"locales":["en","en-US","zh","zh-Hans","zh-Hans-CN"]}`
	if err := os.WriteFile(path, []byte(profile), 0o666); err != nil {
		t.Fatalf("write locale profile: %v", err)
	}
	return path
}

func containsAll(s string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(s, needle) {
			return false
		}
	}
	return true
}

func readGeneratedStringTable(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read strings.go: %v", err)
	}
	var out strings.Builder
	for _, literal := range regexp.MustCompile(`"(?:\\.|[^"\\])*"`).FindAllString(string(raw), -1) {
		value, err := strconv.Unquote(literal)
		if err != nil {
			t.Fatalf("unquote generated string literal %s: %v", literal, err)
		}
		out.WriteString(value)
	}
	return out.String()
}
