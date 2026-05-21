package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGeneratesListPatterns(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	writeListPatternCLDRFixture(t, root)
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

	generated, err := os.ReadFile(filepath.Join(out, "list_patterns.go"))
	if err != nil {
		t.Fatalf("read list_patterns.go: %v", err)
	}
	if !containsAll(string(generated), "type ListPattern struct", "func ListSupportedLocales", "func (l Locale) ListPattern") {
		t.Fatalf("list_patterns.go missing expected generated content:\n%s", generated)
	}
	if !containsAll(string(generated), `localeIndex["en-US"]`, `localeIndex["zh-Hans-CN"]`) {
		t.Fatalf("list_patterns.go missing inherited profile locales:\n%s", generated)
	}
	leaf, err := os.ReadFile(filepath.Join(out, "list", "patterns.go"))
	if err != nil {
		t.Fatalf("read list/patterns.go: %v", err)
	}
	if !containsAll(string(leaf), "package list", "func SupportedLocales() []string", "func Pattern(locale Locale, typ, style string) ListPattern") {
		t.Fatalf("list/patterns.go missing expected generated content:\n%s", leaf)
	}
	stringsData := readGeneratedStringTable(t, filepath.Join(out, "strings.go"))
	if !containsAll(stringsData, "{0} and {1}", "{0}, and {1}", "{0} or {1}", "{0} {1}") {
		t.Fatalf("strings.go missing expected list pattern strings:\n%s", stringsData)
	}
	listStringsData := readGeneratedStringTable(t, filepath.Join(out, "list", "strings.go"))
	if !containsAll(listStringsData, "{0} and {1}", "{0}, and {1}", "{0} or {1}", "{0} {1}") {
		t.Fatalf("list/strings.go missing expected list pattern strings:\n%s", listStringsData)
	}
}

func writeListPatternCLDRFixture(t *testing.T, root string) {
	t.Helper()
	for _, loc := range []string{"en", "zh", "zh-Hans"} {
		dir := filepath.Join(root, "cldr-misc-full", "main", loc)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			t.Fatalf("mkdir list patterns %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "listPatterns.json"), []byte(listPatternsJSON(loc)), 0o666); err != nil {
			t.Fatalf("write list patterns %s: %v", loc, err)
		}
	}
}

func listPatternsJSON(locale string) string {
	return `{"main":{"` + locale + `":{"listPatterns":{"listPattern-type-standard":{"2":"{0} and {1}","start":"{0}, {1}","middle":"{0}, {1}","end":"{0}, and {1}"},"listPattern-type-standard-short":{"2":"{0} & {1}","start":"{0}, {1}","middle":"{0}, {1}","end":"{0}, & {1}"},"listPattern-type-standard-narrow":{"2":"{0}, {1}","start":"{0}, {1}","middle":"{0}, {1}","end":"{0}, {1}"},"listPattern-type-or":{"2":"{0} or {1}","start":"{0}, {1}","middle":"{0}, {1}","end":"{0}, or {1}"},"listPattern-type-or-short":{"2":"{0} or {1}","start":"{0}, {1}","middle":"{0}, {1}","end":"{0}, or {1}"},"listPattern-type-or-narrow":{"2":"{0} or {1}","start":"{0}, {1}","middle":"{0}, {1}","end":"{0}, or {1}"},"listPattern-type-unit":{"2":"{0}, {1}","start":"{0}, {1}","middle":"{0}, {1}","end":"{0}, {1}"},"listPattern-type-unit-short":{"2":"{0}, {1}","start":"{0}, {1}","middle":"{0}, {1}","end":"{0}, {1}"},"listPattern-type-unit-narrow":{"2":"{0} {1}","start":"{0} {1}","middle":"{0} {1}","end":"{0} {1}"}}}}}`
}
