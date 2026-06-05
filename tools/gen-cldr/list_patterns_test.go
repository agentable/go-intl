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

	// The retired root literal renderer must no longer produce list_patterns.go,
	// and the retired root string table must no longer produce strings.go; the
	// list domain now owns a const-only data.go plus hand-written
	// decode.go/accessors.go.
	for _, retired := range []string{"list_patterns.go", "strings.go"} {
		if _, err := os.Stat(filepath.Join(out, retired)); !os.IsNotExist(err) {
			t.Fatalf("root %s should no longer be generated, stat err = %v", retired, err)
		}
	}
	for _, retired := range []string{"patterns.go", "strings.go", "likely_subtags.go", "locale.go"} {
		if _, err := os.Stat(filepath.Join(out, "list", retired)); !os.IsNotExist(err) {
			t.Fatalf("list/%s should no longer be generated, stat err = %v", retired, err)
		}
	}

	// The list domain payload is a const-only data.go carrying the blobs and the
	// private _data table the decoder reads.
	payload, err := os.ReadFile(filepath.Join(out, "list", "data.go"))
	if err != nil {
		t.Fatalf("read list/data.go: %v", err)
	}
	if !containsAll(string(payload), "package list", "const _data", "_listPatternBlob", "_listSupportedBlob") {
		t.Fatalf("list/data.go missing expected const payload:\n%s", payload)
	}
	// The _data const is emitted in 64-byte chunks, so human-readable patterns can
	// straddle a chunk boundary. Reconstruct the concatenated string literals
	// before asserting the expected patterns are present.
	reconstructed := readGeneratedStringTable(t, filepath.Join(out, "list", "data.go"))
	if !containsAll(reconstructed, "{0} and {1}", "{0}, and {1}", "{0} or {1}", "{0} {1}") {
		t.Fatalf("list/data.go _data missing expected list pattern strings:\n%s", payload)
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
