package main

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestRunGeneratesRelativeTimeFields(t *testing.T) {
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

	generated, err := os.ReadFile(filepath.Join(out, "relative_time.go"))
	if err != nil {
		t.Fatalf("read relative_time.go: %v", err)
	}
	if !containsAll(string(generated), "type RelativeTimeField struct", "func RelativeTimeSupportedLocales", "func (l Locale) RelativeTimeFields") {
		t.Fatalf("relative_time.go missing expected generated content:\n%s", generated)
	}
	wrapper, err := os.ReadFile(filepath.Join(out, "relativetime", "accessors.go"))
	if err != nil {
		t.Fatalf("read relativetime/accessors.go: %v", err)
	}
	if !containsAll(string(wrapper), "func SupportedLocales() []string", "func FieldsFor(loc Locale) RelativeTimeFields") {
		t.Fatalf("relativetime/accessors.go missing expected generated content:\n%s", wrapper)
	}
	stringsData := readGeneratedStringTable(t, filepath.Join(out, "strings.go"))
	if !containsAll(stringsData, "{0} second ago", "in {0} seconds", "yesterday") {
		t.Fatalf("strings.go missing expected relative time strings:\n%s", stringsData)
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
