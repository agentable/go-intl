package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSizeCasesCoverP4MeasurementBoundaries(t *testing.T) {
	t.Parallel()

	cases := sizeCases()
	names := make([]string, len(cases))
	for i, tc := range cases {
		names[i] = tc.name
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
	if err := os.WriteFile(filepath.Join(root, "tools", "locale-profile.json"), []byte(`{"locales":["fr","en","und","en"]}`), 0o666); err != nil {
		t.Fatal(err)
	}

	got, err := readDataProfile(root)
	if err != nil {
		t.Fatalf("readDataProfile() error = %v, want nil", err)
	}
	want := dataProfile{CLDR: "48.1.0", ICU: "78", TZData: "2025b", LocaleCount: 2}
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

func TestReadLocaleProfileCountUsesSharedProfileContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		want    int
		wantErr string
	}{
		{name: "normalizes duplicates and sentinel", content: `{"locales":["fr","en","und","en"]}`, want: 2},
		{name: "empty locales", content: `{"locales":["und",""]}`, wantErr: "locales is empty"},
		{name: "multiple values", content: `{"locales":["en"]} {}`, wantErr: "multiple JSON values"},
		{name: "unknown field", content: `{"locales":["en"],"extra":true}`, wantErr: `unknown field "extra"`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "locale-profile.json")
			if err := os.WriteFile(path, []byte(tc.content), 0o666); err != nil {
				t.Fatal(err)
			}
			got, err := readLocaleProfileCount(path)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("readLocaleProfileCount() error = %v, want nil", err)
				}
				if got != tc.want {
					t.Fatalf("readLocaleProfileCount() = %d, want %d", got, tc.want)
				}
				return
			}
			if err == nil {
				t.Fatal("readLocaleProfileCount() error = nil, want error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("readLocaleProfileCount() error = %v, want containing %q", err, tc.wantErr)
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

func TestRunMainRejectsInvalidArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if got := runMain([]string{"-unknown"}, &stdout, &stderr); got != 2 {
		t.Fatalf("runMain() exit = %d, want 2", got)
	}
	if stdout.Len() != 0 {
		t.Fatalf("runMain() stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("runMain() stderr = %q, want flag parse error", stderr.String())
	}
}

func TestRunMainRejectsPositionalArguments(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if got := runMain([]string{"extra"}, &stdout, &stderr); got != 2 {
		t.Fatalf("runMain() exit = %d, want 2", got)
	}
	if stdout.Len() != 0 {
		t.Fatalf("runMain() stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), `unknown argument "extra"`) {
		t.Fatalf("runMain() stderr = %q, want unknown argument", stderr.String())
	}
}

func TestRunMainReturnsRunError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	t.Cleanup(func() {
		if err := os.RemoveAll(outputDir); err != nil {
			t.Fatalf("remove output dir: %v", err)
		}
	})

	if got := runMain(nil, &stdout, &stderr); got != 1 {
		t.Fatalf("runMain() exit = %d, want 1", got)
	}
	if stdout.Len() != 0 {
		t.Fatalf("runMain() stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "read version file") {
		t.Fatalf("runMain() stderr = %q, want data profile error", stderr.String())
	}
}

func TestRunWithDependenciesBuildsCasesAndWritesOutput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeSizecheckDataProfile(t, root)
	outDir, err := os.MkdirTemp(".", "sizecheck-run-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Fatalf("remove temp output: %v", err)
		}
	})
	var stdout bytes.Buffer
	cleaned := false
	err = runWithDependencies(t.Context(), outDir, runOptions{Root: root, Cold: true}, runDependencies{
		cases: []binaryCase{{
			name: "hello",
			source: `package main

func main() {}`,
		}},
		cleanBuildCache: func(context.Context) error {
			cleaned = true
			return nil
		},
		stdout: &stdout,
	})
	if err != nil {
		t.Fatalf("runWithDependencies() error = %v, want nil", err)
	}
	if !cleaned {
		t.Fatal("runWithDependencies() did not clean build cache")
	}
	for _, want := range []string{
		"data-profile: locales=2 cldr=48.1.0 icu=78 tzdata=2025b cold=true",
		"hello",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("runWithDependencies() stdout = %q, want substring %q", stdout.String(), want)
		}
	}
}

func TestRunWithDependenciesReturnsCleanBuildCacheError(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeSizecheckDataProfile(t, root)
	wantErr := errors.New("clean failed")
	err := runWithDependencies(t.Context(), filepath.Join(t.TempDir(), "out"), runOptions{Root: root, Cold: true}, runDependencies{
		cases: []binaryCase{},
		cleanBuildCache: func(context.Context) error {
			return wantErr
		},
		stdout: io.Discard,
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("runWithDependencies() error = %v, want %v", err, wantErr)
	}
}

func TestWriteCaseReturnsCreateDirectoryError(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(root, []byte("not a directory"), 0o666); err != nil {
		t.Fatal(err)
	}
	_, err := writeCase(root, binaryCase{name: "hello", source: "package main"})
	if err == nil {
		t.Fatal("writeCase() error = nil, want create directory error")
	}
}

func TestCleanBuildCacheWithCommand(t *testing.T) {
	t.Parallel()

	if err := cleanBuildCacheWithCommand(t.Context(), helperCommand(0, "")); err != nil {
		t.Fatalf("cleanBuildCacheWithCommand() error = %v, want nil", err)
	}
}

func TestCleanBuildCacheWithCommandWrapsGoCleanFailure(t *testing.T) {
	t.Parallel()

	err := cleanBuildCacheWithCommand(t.Context(), helperCommand(7, "boom"))
	if err == nil {
		t.Fatal("cleanBuildCacheWithCommand() error = nil, want failure")
	}
	if _, ok := errors.AsType[*exec.ExitError](err); !ok {
		t.Fatalf("cleanBuildCacheWithCommand() error = %T, want *exec.ExitError in chain", err)
	}
}

func writeSizecheckDataProfile(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "internal", "cldr"), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "cldr", "VERSION"), []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "tools", "locale-profile.json"), []byte(`{"locales":["en","fr"]}`), 0o666); err != nil {
		t.Fatal(err)
	}
}

func helperCommand(exitCode int, stderr string) commandContextFunc {
	return func(ctx context.Context, name string, args ...string) *exec.Cmd {
		cmdArgs := append([]string{"-test.run=TestHelperProcess", "--", name}, args...)
		cmd := exec.CommandContext(ctx, os.Args[0], cmdArgs...)
		cmd.Env = append(os.Environ(),
			"GO_WANT_HELPER_PROCESS=1",
			fmt.Sprintf("GO_HELPER_EXIT_CODE=%d", exitCode),
			"GO_HELPER_STDERR="+stderr,
		)
		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	fmt.Fprint(os.Stderr, os.Getenv("GO_HELPER_STDERR"))
	code, err := strconv.Atoi(os.Getenv("GO_HELPER_EXIT_CODE"))
	if err != nil {
		os.Exit(2)
	}
	os.Exit(code)
}
