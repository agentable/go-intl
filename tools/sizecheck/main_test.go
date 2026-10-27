package main

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
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

	root, err := os.MkdirTemp(".", ".sizecheck-test-*")
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

	got, err := buildCase(t.Context(), "hello", pkgDir, filepath.Join(binDir, "hello"), false)
	if err != nil {
		t.Fatalf("buildCase() error = %v, want nil", err)
	}
	if got.name != "hello" || got.bytes <= 0 {
		t.Fatalf("buildCase() = %+v, want named binary result", got)
	}
}

func TestBuildCaseReturnsGoBuildFailure(t *testing.T) {
	t.Parallel()

	root, err := os.MkdirTemp(".", ".sizecheck-test-*")
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

	_, err = buildCase(t.Context(), "broken", pkgDir, filepath.Join(binDir, "broken"), false)
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

func installGoBuildTracer(t *testing.T) string {
	t.Helper()
	realGo, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}
	shim := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	name := "go"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	if err := os.WriteFile(filepath.Join(shim, name), data, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GO_INTL_SIZECHECK_REAL_GO", realGo)
	t.Setenv("GO_INTL_SIZECHECK_TRACE", filepath.Join(shim, "trace"))
	t.Setenv("PATH", shim+string(os.PathListSeparator)+os.Getenv("PATH"))
	return shim
}

func TestColdBuildIsolation(t *testing.T) {
	if os.Getenv("GO_INTL_TEST_COLD_BUILDS") != "1" {
		t.Skip("run with GO_INTL_TEST_COLD_BUILDS=1 for real cold compiler traces")
	}
	shim := installGoBuildTracer(t)
	callerCache := t.TempDir()
	sentinel := filepath.Join(callerCache, "sentinel")
	if err := os.WriteFile(sentinel, []byte("caller-owned"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOCACHE", callerCache)
	cacheParent := t.TempDir()
	for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(key, cacheParent)
	}
	root := t.TempDir()
	writeSizecheckDataProfile(t, root)
	out, err := os.MkdirTemp(".", ".sizecheck-run-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(out); err != nil {
			t.Error(err)
		}
	})
	cases := []binaryCase{
		{name: "first", source: "package main\nimport \"errors\"\nfunc main(){ _ = errors.New(\"first\") }"},
		{name: "second", source: "package main\nimport \"errors\"\nfunc main(){ _ = errors.New(\"second\") }"},
	}
	seen := map[string]bool{}
	for range 2 {
		var stdout bytes.Buffer
		if err := runWithDependencies(t.Context(), out, runOptions{Root: root, Cold: true}, runDependencies{cases: cases, stdout: &stdout}); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"cold=true", "cache=isolated-per-case", "go=go", "target=", "first", "second"} {
			if !strings.Contains(stdout.String(), want) {
				t.Fatalf("report missing %q: %s", want, &stdout)
			}
		}
		traces, err := os.ReadDir(filepath.Join(shim, "trace"))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range traces {
			path := filepath.Join(shim, "trace", entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var trace buildTrace
			if err := json.Unmarshal(data, &trace); err != nil {
				t.Fatal(err)
			}
			if trace.Cache == callerCache || seen[trace.Cache] || trace.Entries != 0 {
				t.Fatalf("build did not start with its own empty cache: %+v", trace)
			}
			seen[trace.Cache] = true
			if _, err := os.Stat(trace.Cache); !os.IsNotExist(err) {
				t.Fatalf("cache remains after build: %s: %v", trace.Cache, err)
			}
			if !strings.Contains(trace.Output, "-p errors ") {
				t.Fatalf("shared errors dependency was not compiled: %s", trace.Output)
			}
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		}
		if len(traces) != 2 {
			t.Fatalf("got %d build traces, want 2", len(traces))
		}
		slices.Reverse(cases)
	}
	if data, err := os.ReadFile(sentinel); err != nil || string(data) != "caller-owned" {
		t.Fatalf("caller cache changed: %s, %v", data, err)
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

// The helper observes the real Go subprocess boundary without replacing compilation.
type buildTrace struct {
	Cache   string
	Entries int
	Output  string
}

func TestMain(m *testing.M) {
	realGo := os.Getenv("GO_INTL_SIZECHECK_REAL_GO")
	if realGo == "" {
		os.Exit(m.Run())
	}
	args := os.Args[1:]
	trace := buildTrace{Cache: os.Getenv("GOCACHE")}
	isBuild := len(args) > 0 && args[0] == "build"
	if isBuild {
		entries, err := os.ReadDir(trace.Cache)
		if err != nil {
			_, _ = os.Stderr.WriteString(err.Error())
			os.Exit(2)
		}
		trace.Entries = len(entries)
		args = append([]string{"build", "-x"}, args[1:]...)
	}
	output, err := exec.Command(realGo, args...).CombinedOutput()
	if isBuild {
		trace.Output = string(output)
		dir := os.Getenv("GO_INTL_SIZECHECK_TRACE")
		if mkdirErr := os.MkdirAll(dir, 0o700); mkdirErr != nil {
			os.Exit(2)
		}
		file, createErr := os.CreateTemp(dir, "build-*.json")
		if createErr != nil {
			os.Exit(2)
		}
		data, marshalErr := json.Marshal(trace)
		if marshalErr != nil {
			os.Exit(2)
		}
		if _, writeErr := file.Write(data); writeErr != nil {
			os.Exit(2)
		}
		if closeErr := file.Close(); closeErr != nil {
			os.Exit(2)
		}
	}
	if isBuild && os.Getenv("GO_INTL_SIZECHECK_DENY_CLEANUP") == "1" {
		if err := os.WriteFile(filepath.Join(trace.Cache, "retained"), []byte("x"), 0o600); err != nil {
			os.Exit(2)
		}
		if err := os.Chmod(trace.Cache, 0o500); err != nil {
			os.Exit(2)
		}
	}

	if _, writeErr := os.Stdout.Write(output); writeErr != nil {
		os.Exit(2)
	}
	if err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}

func TestRunWithDependenciesBuildsCasesAndWritesOutput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeSizecheckDataProfile(t, root)
	outDir, err := os.MkdirTemp(".", ".sizecheck-run-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(outDir); err != nil {
			t.Fatalf("remove temp output: %v", err)
		}
	})
	var stdout bytes.Buffer
	err = runWithDependencies(t.Context(), outDir, runOptions{Root: root}, runDependencies{
		cases: []binaryCase{{
			name: "hello",
			source: `package main

func main() {}`,
		}},
		stdout: &stdout,
	})
	if err != nil {
		t.Fatalf("runWithDependencies() error = %v, want nil", err)
	}
	for _, want := range []string{
		"data-profile: locales=2 cldr=48.1.0 icu=78 tzdata=2025b cold=false",
		"hello",
		"cache=inherited",
	} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("runWithDependencies() stdout = %q, want substring %q", stdout.String(), want)
		}
	}
}

func TestBuildCaseColdCleanupFailurePreservesBuildError(t *testing.T) {
	if runtime.GOOS == "windows" || os.Geteuid() == 0 {
		t.Skip("directory permission failure requires an unprivileged Unix user")
	}
	installGoBuildTracer(t)
	parent := t.TempDir()
	t.Setenv("TMPDIR", parent)
	t.Setenv("GOFLAGS", "-sizecheck-invalid-flag")
	t.Setenv("GO_INTL_SIZECHECK_DENY_CLEANUP", "1")
	t.Cleanup(func() {
		entries, err := os.ReadDir(parent)
		if err != nil {
			t.Error(err)
			return
		}
		for _, entry := range entries {
			if err := os.Chmod(filepath.Join(parent, entry.Name()), 0o700); err != nil {
				t.Error(err)
			}
		}
	})
	_, err := buildCase(t.Context(), "failure", ".", filepath.Join(parent, "out"), true)
	if _, ok := errors.AsType[*exec.ExitError](err); !ok {
		t.Fatalf("lost build error: %v", err)
	}
	if !errors.Is(err, os.ErrPermission) || !strings.Contains(err.Error(), "remove build cache for failure") {
		t.Fatalf("lost cleanup error: %v", err)
	}
}

func TestBuildCaseColdFailureAndCancellationCleanup(t *testing.T) {
	parent := t.TempDir()
	for _, key := range []string{"TMPDIR", "TMP", "TEMP"} {
		t.Setenv(key, parent)
	}
	t.Setenv("GOFLAGS", "-sizecheck-invalid-flag")
	_, err := buildCase(t.Context(), "failure", ".", filepath.Join(parent, "out"), true)
	if _, ok := errors.AsType[*exec.ExitError](err); !ok {
		t.Fatalf("lost build error: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = buildCase(ctx, "canceled", ".", filepath.Join(parent, "canceled"), true)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
	entries, err := os.ReadDir(parent)
	if err != nil || len(entries) != 0 {
		t.Fatalf("cache remains: %v, %v", entries, err)
	}
}
