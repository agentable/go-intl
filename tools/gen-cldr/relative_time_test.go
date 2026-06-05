package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

// TestRunGeneratesRelativeTimeDomain asserts the generator emits the
// relativetime domain as a self-contained const-only payload at
// internal/cldr/relativetime/data.go, and no longer emits the retired root
// relative_time.go literal renderer or the relativetime/accessors.go wrapper.
func TestRunGeneratesRelativeTimeDomain(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	writeRelativeTimeCLDRFixture(t, root)
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

	// The retired root literal renderer must not produce relative_time.go anymore.
	if _, err := os.Stat(filepath.Join(out, "relative_time.go")); !os.IsNotExist(err) {
		t.Fatalf("root relative_time.go should no longer be generated, stat err = %v", err)
	}
	// The alias wrapper is retired; the domain owns hand-written accessors.go that
	// the generator never produces.
	if _, err := os.Stat(filepath.Join(out, "relativetime", "accessors.go")); !os.IsNotExist(err) {
		t.Fatalf("relativetime/accessors.go wrapper should no longer be generated, stat err = %v", err)
	}

	// The relativetime domain payload is a const-only data.go carrying the blobs
	// and the private _data table the decoder reads.
	payload, err := os.ReadFile(filepath.Join(out, "relativetime", "data.go"))
	if err != nil {
		t.Fatalf("read relativetime/data.go: %v", err)
	}
	if !containsAll(string(payload), "package relativetime", "const _data", "_relativeTimeFieldBlob", "_relativeTimeSupportedBlob") {
		t.Fatalf("relativetime/data.go missing expected const payload:\n%s", payload)
	}
	// The _data const is emitted in 64-byte chunks, so human-readable patterns can
	// straddle a chunk boundary. Reconstruct the concatenated string literals
	// before asserting the expected patterns are present.
	reconstructed := readGeneratedStringTable(t, filepath.Join(out, "relativetime", "data.go"))
	if !containsAll(reconstructed, "{0} second ago", "in {0} seconds", "yesterday") {
		t.Fatalf("relativetime/data.go _data missing expected relative time strings:\n%s", payload)
	}
}

func writeRelativeTimeCLDRFixture(t *testing.T, root string) {
	t.Helper()
	for _, loc := range []string{"en", "zh", "zh-Hans"} {
		dir := filepath.Join(root, "cldr-dates-full", "main", loc)
		if err := os.MkdirAll(dir, 0o777); err != nil {
			t.Fatalf("mkdir relative time %s: %v", loc, err)
		}
		if err := os.WriteFile(filepath.Join(dir, "dateFields.json"), []byte(relativeTimeJSON(loc)), 0o666); err != nil {
			t.Fatalf("write relative time %s: %v", loc, err)
		}
	}
}

func relativeTimeJSON(locale string) string {
	return `{"main":{"` + locale + `":{"dates":{"fields":{"second":{"relative-type-0":"now","relativeTime-type-future":{"relativeTimePattern-count-one":"in {0} second","relativeTimePattern-count-other":"in {0} seconds"},"relativeTime-type-past":{"relativeTimePattern-count-one":"{0} second ago","relativeTimePattern-count-other":"{0} seconds ago"}},"minute":{"relativeTime-type-future":{"relativeTimePattern-count-one":"in {0} minute","relativeTimePattern-count-other":"in {0} minutes"},"relativeTime-type-past":{"relativeTimePattern-count-one":"{0} minute ago","relativeTimePattern-count-other":"{0} minutes ago"}},"hour":{"relativeTime-type-future":{"relativeTimePattern-count-one":"in {0} hour","relativeTimePattern-count-other":"in {0} hours"},"relativeTime-type-past":{"relativeTimePattern-count-one":"{0} hour ago","relativeTimePattern-count-other":"{0} hours ago"}},"day":{"relative-type--1":"yesterday","relative-type-0":"today","relative-type-1":"tomorrow","relativeTime-type-future":{"relativeTimePattern-count-one":"in {0} day","relativeTimePattern-count-other":"in {0} days"},"relativeTime-type-past":{"relativeTimePattern-count-one":"{0} day ago","relativeTimePattern-count-other":"{0} days ago"}},"week":{"relativeTime-type-future":{"relativeTimePattern-count-one":"in {0} week","relativeTimePattern-count-other":"in {0} weeks"},"relativeTime-type-past":{"relativeTimePattern-count-one":"{0} week ago","relativeTimePattern-count-other":"{0} weeks ago"}},"month":{"relativeTime-type-future":{"relativeTimePattern-count-one":"in {0} month","relativeTimePattern-count-other":"in {0} months"},"relativeTime-type-past":{"relativeTimePattern-count-one":"{0} month ago","relativeTimePattern-count-other":"{0} months ago"}},"quarter":{"relativeTime-type-future":{"relativeTimePattern-count-one":"in {0} quarter","relativeTimePattern-count-other":"in {0} quarters"},"relativeTime-type-past":{"relativeTimePattern-count-one":"{0} quarter ago","relativeTimePattern-count-other":"{0} quarters ago"}},"year":{"relativeTime-type-future":{"relativeTimePattern-count-one":"in {0} year","relativeTimePattern-count-other":"in {0} years"},"relativeTime-type-past":{"relativeTimePattern-count-one":"{0} year ago","relativeTimePattern-count-other":"{0} years ago"}}}}}}}`
}
