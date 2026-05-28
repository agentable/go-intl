package main

import (
	"errors"
	"os"
	"os/exec"
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

func TestBuildCaseProducesBinary(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp(".", "sizecheck-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Fatalf("remove temp package: %v", err)
		}
	})
	pkgDir, err := writeCase(filepath.Join(root, "cases"), binaryCase{
		name: "hello",
		source: `package main

func main() {}`,
	})
	if err != nil {
		t.Fatalf("writeCase() error = %v, want nil", err)
	}
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o777); err != nil {
		t.Fatal(err)
	}

	got, err := buildCase(t.Context(), "hello", pkgDir, filepath.Join(binDir, "hello"))
	if err != nil {
		t.Fatalf("buildCase() error = %v, want nil", err)
	}
	if got.name != "hello" || got.bytes <= 0 {
		t.Fatalf("buildCase() = %+v, want named binary result", got)
	}
}

func TestBuildCaseReturnsGoBuildFailure(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp(".", "sizecheck-test-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Fatalf("remove temp package: %v", err)
		}
	})
	pkgDir, err := writeCase(filepath.Join(root, "cases"), binaryCase{
		name: "broken",
		source: `package main

func main() {`,
	})
	if err != nil {
		t.Fatalf("writeCase() error = %v, want nil", err)
	}
	binDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(binDir, 0o777); err != nil {
		t.Fatal(err)
	}

	_, err = buildCase(t.Context(), "broken", pkgDir, filepath.Join(binDir, "broken"))
	if err == nil {
		t.Fatal("buildCase() error = nil, want go build failure")
	}
	if _, ok := errors.AsType[*exec.ExitError](err); !ok {
		t.Fatalf("buildCase() error = %T, want *exec.ExitError in chain", err)
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

func TestReadVersionFileRejectsMissingPins(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "VERSION")
	if err := os.WriteFile(path, []byte("cldr=48.1.0\nicu=78\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	_, err := readVersionFile(path)
	if !errors.Is(err, errMissingDataVersion) {
		t.Fatalf("readVersionFile() error = %v, want errMissingDataVersion", err)
	}
}

func TestReadLocaleProfileCountRejectsInvalidProfiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    error
	}{
		{name: "empty locales", content: `{"locales":[]}`, want: errEmptyLocaleProfile},
		{name: "multiple values", content: `{"locales":["en"]} {}`, want: errMultipleProfileValues},
		{name: "unknown field", content: `{"locales":["en"],"extra":true}`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "locale-profile.json")
			if err := os.WriteFile(path, []byte(tc.content), 0o666); err != nil {
				t.Fatal(err)
			}
			_, err := readLocaleProfileCount(path)
			if err == nil {
				t.Fatal("readLocaleProfileCount() error = nil, want error")
			}
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("readLocaleProfileCount() error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestWriteCaseCreatesTrimmedMainFile(t *testing.T) {
	t.Parallel()

	pkgDir, err := writeCase(t.TempDir(), binaryCase{
		name: "hello",
		source: `
package main

func main() {}
`,
	})
	if err != nil {
		t.Fatalf("writeCase() error = %v, want nil", err)
	}
	got, err := os.ReadFile(filepath.Join(pkgDir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "package main\n\nfunc main() {}\n" {
		t.Fatalf("writeCase() content = %q", got)
	}
}

func TestRunReturnsDataProfileErrorsBeforeBuilding(t *testing.T) {
	t.Parallel()

	err := run(t.Context(), filepath.Join(t.TempDir(), "out"), runOptions{Root: t.TempDir()})
	if err == nil {
		t.Fatal("run() error = nil, want missing data profile error")
	}
}
