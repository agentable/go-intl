package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReportsDuplicateFixtureIDsAcrossPackages(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	localeDir := writeFixtureFile(t, root, "locale", "locale/testdata/conformance/manual/basic.json", `[
		{"id":"shared-id","source":"manual","locale":"en-US","options":{},"input":"en-US","expected":"en-US"}
	]`)
	numberDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"shared-id","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	err := run([]string{localeDir, numberDir})
	if err == nil {
		t.Fatal("run() succeeded, want duplicate fixture ID error")
	}
	if !strings.Contains(err.Error(), "shared-id") {
		t.Fatalf("run() error = %v, want duplicate ID", err)
	}
}

func TestRunReportsMissingRequiredFixtureFieldsWithPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	localeDir := writeFixtureFile(t, root, "locale", "locale/testdata/conformance/manual/basic.json", `[
		{"id":"missing-source","locale":"en-US","options":{},"input":"en-US","expected":"en-US"}
	]`)

	err := run([]string{localeDir})
	if err == nil {
		t.Fatal("run() succeeded, want missing required field error")
	}
	if !strings.Contains(err.Error(), "basic.json") || !strings.Contains(err.Error(), "source") {
		t.Fatalf("run() error = %v, want file path and missing field", err)
	}
}

func TestRunReportsMixedSourcesInOneFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	numberDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/formatjs/basic.json", `[
		{"id":"fmt-number","source":"formatjs:intl-numberformat","locale":"en-US","options":{},"input":1,"expected":"1"},
		{"id":"node-number","source":"node:v76.1:numberFormats","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	err := run([]string{numberDir})
	if err == nil {
		t.Fatal("run() succeeded, want mixed source error")
	}
	if !strings.Contains(err.Error(), "basic.json") || !strings.Contains(err.Error(), "mixed sources") {
		t.Fatalf("run() error = %v, want file path and mixed sources", err)
	}
}

func TestRunRejectsDateTimeEpochInputAndAllowsISOStrings(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	datetimeDir := writeFixtureFile(t, root, "datetimeformat", "datetimeformat/testdata/conformance/manual/basic.json", `[
		{"id":"dtf-epoch","source":"manual","locale":"en-US","options":{},"input":1715169600000,"expected":"May 8, 2026"}
	]`)

	err := run([]string{datetimeDir})
	if err == nil {
		t.Fatal("run() succeeded, want datetime epoch input error")
	}
	if !strings.Contains(err.Error(), "dtf-epoch") || !strings.Contains(err.Error(), "ISO-8601") {
		t.Fatalf("run() error = %v, want fixture ID and ISO-8601 guidance", err)
	}

	root = t.TempDir()
	datetimeDir = writeFixtureFile(t, root, "datetimeformat", "datetimeformat/testdata/conformance/manual/basic.json", `[
		{"id":"dtf-iso","source":"manual","locale":"en-US","options":{},"input":"2026-05-08T12:00:00Z","expected":"May 8, 2026"}
	]`)
	if err := run([]string{datetimeDir}); err != nil {
		t.Fatalf("run() with ISO input error = %v, want nil", err)
	}
}

func TestRunAllowsDateTimeRangeISOInput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	datetimeDir := writeFixtureFile(t, root, "datetimeformat", "datetimeformat/testdata/conformance/manual/basic.json", `[
		{"id":"dtf-range","source":"manual","locale":"en-US","options":{},"input":{"start":"2026-05-08T12:00:00Z","end":"2026-05-10T12:00:00Z"},"expectedRange":"May 8 – 10, 2026"}
	]`)

	if err := run([]string{datetimeDir}); err != nil {
		t.Fatalf("run() error = %v, want nil", err)
	}
}

func TestRunReportsMissingAndExpiredXFailExpiry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	if err := os.WriteFile(filepath.Join(packageDir, "testdata", "xfail.json"), []byte(`[
		{"id":"nf-basic","reason":"pending implementation","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatalf("write xfail: %v", err)
	}

	err := run([]string{packageDir})
	if err == nil {
		t.Fatal("run() succeeded, want missing expires_at error")
	}
	if !strings.Contains(err.Error(), "xfail.json") || !strings.Contains(err.Error(), "expires_at") {
		t.Fatalf("run() error = %v, want xfail path and expires_at", err)
	}

	root = t.TempDir()
	packageDir = writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	if err := os.WriteFile(filepath.Join(packageDir, "testdata", "xfail.json"), []byte(`[
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2000-01-01","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatalf("write xfail: %v", err)
	}

	err = run([]string{packageDir})
	if err == nil {
		t.Fatal("run() succeeded, want expired xfail error")
	}
	if !strings.Contains(err.Error(), "expired") || !strings.Contains(err.Error(), "nf-basic") {
		t.Fatalf("run() error = %v, want expired xfail ID", err)
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
