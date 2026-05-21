package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestSizeCasesCoverP4MeasurementBoundaries(t *testing.T) {
	t.Parallel()

	names := make([]string, 0, len(sizeCases()))
	for _, tc := range sizeCases() {
		names = append(names, tc.name)
	}
	for _, name := range []string{
		"empty",
		"cldr-available",
		"root-facade",
		"numberformat",
		"datetimeformat",
		"pluralrules",
		"listformat",
		"relativetimeformat",
		"durationformat",
		"displaynames",
		"collator",
		"segmenter",
		"all-formatters",
	} {
		if !slices.Contains(names, name) {
			t.Fatalf("sizeCases() missing %q: %v", name, names)
		}
	}
}

func TestReadDataProfile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "cldr"), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "cldr", "VERSION"), []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "locale-profile.json"), []byte(`{"locales":["en","fr","zh-Hant-TW"]}`), 0o666); err != nil {
		t.Fatal(err)
	}

	got, err := readDataProfile(root)
	if err != nil {
		t.Fatalf("readDataProfile() error = %v, want nil", err)
	}
	want := dataProfile{CLDR: "48.1.0", ICU: "78", TZData: "2025b", LocaleCount: 3}
	if got != want {
		t.Fatalf("readDataProfile() = %+v, want %+v", got, want)
	}
}

func TestFormatResultsIncludesDataProfileAndBuildDuration(t *testing.T) {
	t.Parallel()

	output := formatResults(
		dataProfile{CLDR: "48.1.0", ICU: "78", TZData: "2025b", LocaleCount: 104, Cold: true},
		[]result{
			{name: "empty", bytes: 10, duration: 100 * time.Millisecond},
			{name: "all-formatters", bytes: 42, duration: 1500 * time.Millisecond},
		},
	)
	for _, want := range []string{
		"data-profile: locales=104 cldr=48.1.0 icu=78 tzdata=2025b cold=true",
		"case",
		"bytes",
		"delta",
		"build",
		"all-formatters",
		"32",
		"1.5s",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("formatResults() = %q, want substring %q", output, want)
		}
	}
}
