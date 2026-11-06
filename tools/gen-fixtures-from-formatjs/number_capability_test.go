package main

import (
	"bytes"
	"encoding/json/v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

func TestNumberFormatRecoveredAssertionsReachFormatter(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	source := filepath.Join(root, "formatjs")
	tests := formatJSPackageTestsRoot(source, formatJSNumberFormatPackageDir)
	mustWriteFile(t, filepath.Join(tests, "capability.test.ts"), `
expect(new Intl.NumberFormat('en', {}).format(1)).toBe('1')
expect(new Intl.NumberFormat('en', {signDisplay: 'always'}).format(1)).toBe('+1')
expect(new Intl.NumberFormat('fr', {style: 'currency', currency: 'EUR'}).format(42)).toBe('42,00 €')
`)
	out := filepath.Join(root, "out")
	skips, err := importFormatJSNumberFormat(source, out)
	if err != nil {
		t.Fatal(err)
	}
	got, err := conformance.LoadFixtures(filepath.Join(out, "numberformat"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || len(skips) != 0 {
		t.Fatalf("recovered %d assertions, skips=%v, want all three", len(got), skips)
	}
	first, err := os.ReadFile(formatJSFixtureFile(filepath.Join(out, "numberformat", "testdata", "conformance", "formatjs"), "capability.test.ts", fixtureSlug))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := importFormatJSNumberFormat(source, out); err != nil {
		t.Fatal(err)
	}
	second, err := os.ReadFile(formatJSFixtureFile(filepath.Join(out, "numberformat", "testdata", "conformance", "formatjs"), "capability.test.ts", fixtureSlug))
	if err != nil || !bytes.Equal(first, second) {
		t.Fatalf("output not deterministic: %v", err)
	}

	binary := filepath.Join(root, "numberformat.test")
	build := exec.CommandContext(t.Context(), "go", "test", "-c", "-o", binary, "./numberformat")
	build.Dir = filepath.Join("..", "..")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build real NumberFormat callback: %v\n%s", err, output)
	}
	run := func(t *testing.T, fixtureRoot string) (string, error) {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), binary, "-test.run=^TestNumberConformanceObservations$")
		cmd.Env = append(os.Environ(), "GO_INTL_FIXTURE_ADAPTER_ROOT="+fixtureRoot)
		output, err := cmd.CombinedOutput()
		return string(output), err
	}
	if output, err := run(t, filepath.Join(out, "numberformat")); err != nil {
		t.Fatalf("formatter rejected recovered assertions: %v\n%s", err, output)
	}
	for _, f := range got {
		t.Run(f.ID, func(t *testing.T) {
			t.Parallel()
			f.Expected = new("incorrect recovered expectation")
			data, err := json.Marshal([]conformance.Fixture{f})
			if err != nil {
				t.Fatal(err)
			}
			fixtureRoot := filepath.Join(t.TempDir(), "numberformat")
			mustWriteFile(t, filepath.Join(fixtureRoot, "testdata", "conformance", "formatjs", "capability-test-ts.json"), string(data))
			output, err := run(t, fixtureRoot)
			if err == nil || !strings.Contains(output, f.ID) || !strings.Contains(output, "Format()") {
				t.Fatalf("formatter did not reject wrong recovered assertion %s: %v\n%s", f.ID, err, output)
			}
		})
	}
}
