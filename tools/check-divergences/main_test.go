package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsDivergenceIDMissingFromFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeDivergences(t, packageDir, "id: nf-missing\nreason: known upstream difference\n")

	err := run([]string{packageDir})
	if err == nil {
		t.Fatal("run() succeeded, want missing divergence ID error")
	}
	if !strings.Contains(err.Error(), "nf-missing") {
		t.Fatalf("run() error = %v, want missing divergence ID", err)
	}
}

func TestRunAllowsEmptyDivergencesWithNoFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := filepath.Join(root, "locale")
	writeDivergences(t, packageDir, "")

	if err := run([]string{packageDir}); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
}

func TestRunAllowsResolvedDivergenceMissingFromFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeDivergences(t, packageDir, "id: nf-old\nstatus: resolved\nreason: historical difference\n")

	if err := run([]string{packageDir}); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
}

func writeFixtureFile(t *testing.T, root, packageName, rel, data string) string {
	t.Helper()

	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatalf("mkdir fixture dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return filepath.Join(root, packageName)
}

func writeDivergences(t *testing.T, packageDir, data string) {
	t.Helper()

	path := filepath.Join(packageDir, "testdata", "divergences.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatalf("mkdir divergences dir: %v", err)
	}
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatalf("write divergences: %v", err)
	}
}
