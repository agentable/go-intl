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
	writeMatchingCLDRFixture(t, root)
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

	for _, name := range []string{"strings.go", "locales.go", "likely_subtags.go", "collations.go"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatalf("expected generated %s: %v", name, err)
		}
	}
	locales, err := os.ReadFile(filepath.Join(out, "locales.go"))
	if err != nil {
		t.Fatalf("read locales.go: %v", err)
	}
	if string(locales) == "" || !containsAll(string(locales), "\"und\"", "\"en\"", "\"zh-Hans\"") {
		t.Fatalf("locales.go missing expected tags:\n%s", locales)
	}
	likely, err := os.ReadFile(filepath.Join(out, "likely_subtags.go"))
	if err != nil {
		t.Fatalf("read likely_subtags.go: %v", err)
	}
	if !containsAll(string(likely), "MaximizeSubtags", "MinimizeSubtags", "maximizeSubtagRecord", "searchLikelySubtag") {
		t.Fatalf("likely_subtags.go missing expected generated content:\n%s", likely)
	}
	collations, err := os.ReadFile(filepath.Join(out, "collations.go"))
	if err != nil {
		t.Fatalf("read collations.go: %v", err)
	}
	if !containsAll(string(collations), `"compat"`, `"emoji"`, `"phonebk"`) || strings.Contains(string(collations), `"big5han"`) || strings.Contains(string(collations), `"search"`) || strings.Contains(string(collations), `"standard"`) {
		t.Fatalf("collations.go did not contain supported canonical collations only:\n%s", collations)
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
	generated, err := os.ReadFile(filepath.Join(out, "likely_subtags.go"))
	if err != nil {
		t.Fatalf("read likely_subtags.go: %v", err)
	}
	if !strings.Contains(string(generated), "searchMinimizeSubtag") || strings.Contains(string(generated), "und-CN") {
		t.Fatalf("likely_subtags.go did not prefer zh minimize alias:\n%s", generated)
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
	locales, err := os.ReadFile(filepath.Join(out, "locales.go"))
	if err != nil {
		t.Fatalf("read locales.go: %v", err)
	}
	if !containsAll(string(locales), "\"en\"", "\"en-US\"") {
		t.Fatalf("locales.go missing fallback locales:\n%s", locales)
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
	for _, name := range cldr.RequiredPackages {
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
	bcp47 := filepath.Join(root, "cldr-bcp47", "bcp47")
	if err := os.MkdirAll(bcp47, 0o777); err != nil {
		t.Fatalf("mkdir bcp47: %v", err)
	}
	collations := `{"keyword":{"u":{"co":{"_description":"Collation type key","big5han":{"_deprecated":true},"compat":{},"ducet":{},"emoji":{},"eor":{},"phonebk":{"_alias":"phonebook"},"search":{},"standard":{}}}}}`
	if err := os.WriteFile(filepath.Join(bcp47, "collation.json"), []byte(collations), 0o666); err != nil {
		t.Fatalf("write collation.json: %v", err)
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
