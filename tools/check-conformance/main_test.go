package main

import (
	"bytes"
	"io"
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

	err := run([]string{localeDir, numberDir}, io.Discard)
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

	err := run([]string{localeDir}, io.Discard)
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
	numberDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"fmt-number","source":"manual:numberformat-one","locale":"en-US","options":{},"input":1,"expected":"1"},
		{"id":"node-number","source":"manual:numberformat-two","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	err := run([]string{numberDir}, io.Discard)
	if err == nil {
		t.Fatal("run() succeeded, want mixed source error")
	}
	if !strings.Contains(err.Error(), "basic.json") || !strings.Contains(err.Error(), "mixed sources") {
		t.Fatalf("run() error = %v, want file path and mixed sources", err)
	}
}

func TestRunReportsSourceDirectoryMismatch(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	numberDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"nf-wrong-dir","source":"formatjs:packages/intl-numberformat/tests/basic.test.ts","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	err := run([]string{numberDir}, io.Discard)
	if err == nil {
		t.Fatal("run() succeeded, want source directory mismatch")
	}
	if !strings.Contains(err.Error(), "nf-wrong-dir") || !strings.Contains(err.Error(), "manual") || !strings.Contains(err.Error(), "formatjs:") {
		t.Fatalf("run() error = %v, want fixture ID and source directory mismatch", err)
	}
}

func TestRunRequiresErrorFixturesInErrorsFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	numberDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"nf-invalid","source":"manual","locale":"en-US","options":{},"input":1,"errorCode":"invalid_option"}
	]`)

	err := run([]string{numberDir}, io.Discard)
	if err == nil {
		t.Fatal("run() succeeded, want misplaced error fixture")
	}
	if !strings.Contains(err.Error(), "errors.json") || !strings.Contains(err.Error(), "nf-invalid") {
		t.Fatalf("run() error = %v, want fixture ID and errors.json guidance", err)
	}

	root = t.TempDir()
	numberDir = writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/errors.json", `[
		{"id":"nf-positive","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	err = run([]string{numberDir}, io.Discard)
	if err == nil {
		t.Fatal("run() succeeded, want positive fixture rejected from errors.json")
	}
	if !strings.Contains(err.Error(), "errors.json") || !strings.Contains(err.Error(), "errorCode") {
		t.Fatalf("run() error = %v, want errors.json and errorCode guidance", err)
	}
}

func TestRunRejectsDateTimeEpochInputAndAllowsISOStrings(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	datetimeDir := writeFixtureFile(t, root, "datetimeformat", "datetimeformat/testdata/conformance/manual/basic.json", `[
		{"id":"dtf-epoch","source":"manual","locale":"en-US","options":{},"input":1715169600000,"expected":"May 8, 2026"}
	]`)

	err := run([]string{datetimeDir}, io.Discard)
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
	if err := run([]string{datetimeDir}, io.Discard); err != nil {
		t.Fatalf("run() with ISO input error = %v, want nil", err)
	}
}

func TestRunAllowsDateTimeRangeISOInput(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	datetimeDir := writeFixtureFile(t, root, "datetimeformat", "datetimeformat/testdata/conformance/manual/basic.json", `[
		{"id":"dtf-range","source":"manual","locale":"en-US","options":{},"input":{"start":"2026-05-08T12:00:00Z","end":"2026-05-10T12:00:00Z"},"expectedRange":"May 8 – 10, 2026"}
	]`)

	if err := run([]string{datetimeDir}, io.Discard); err != nil {
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

	err := run([]string{packageDir}, io.Discard)
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

	err = run([]string{packageDir}, io.Discard)
	if err == nil {
		t.Fatal("run() succeeded, want expired xfail error")
	}
	if !strings.Contains(err.Error(), "expired") || !strings.Contains(err.Error(), "nf-basic") {
		t.Fatalf("run() error = %v, want expired xfail ID", err)
	}
}

func TestRunReportsDivergenceIDMissingFromFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeDivergences(t, packageDir, "id: nf-missing\nsource: manual:missing\nowner: numberformat\nstatus: accepted\nreason: known upstream difference\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")

	err := run([]string{packageDir}, io.Discard)
	if err == nil {
		t.Fatal("run() succeeded, want missing divergence ID error")
	}
	if !strings.Contains(err.Error(), "nf-missing") {
		t.Fatalf("run() error = %v, want missing divergence ID", err)
	}
}

func TestRunReportsDivergenceSourceMismatch(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/manual/basic.json", `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeDivergences(t, packageDir, "id: nf-basic\nsource: formatjs:wrong\nowner: numberformat\nstatus: accepted\nreason: known upstream difference\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")

	err := run([]string{packageDir}, io.Discard)
	if err == nil || !strings.Contains(err.Error(), "source") || !strings.Contains(err.Error(), "manual") {
		t.Fatalf("run() error = %v, want source mismatch", err)
	}
}

func TestRunValidatesSkipListAndWritesCoverageReport(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/formatjs/basic.json", `[
		{"id":"nf-formatjs","source":"formatjs:packages/intl-numberformat/tests/basic.test.ts","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	skipListPath := filepath.Join(root, ".skip-list.json")
	if err := os.WriteFile(skipListPath, []byte(`[
		{"source":"formatjs:unsupported","category":"unsupported-extractor-shape","route":"extractor","reason":"source uses assertions the extractor cannot reduce"}
	]`), 0o666); err != nil {
		t.Fatalf("write skip-list: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"-skip-list", skipListPath, "-coverage", packageDir}, &out)
	if err != nil {
		t.Fatalf("run(-coverage) error = %v, want nil", err)
	}
	for _, want := range []string{"conformance coverage:", "numberformat:", "formatjs=1", "unsupported-extractor-shape=1", "routes extractor=1"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("coverage output = %q, want %q", out.String(), want)
		}
	}
}

func TestRunValidatesNodeWitnessCoverage(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/node-v26/smoke.json", `[
		{"id":"numberformat-node-v26-smoke","source":"node:v26.0.0:numberformat","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	err := run([]string{"-node-witness", packageDir}, io.Discard)
	if err == nil {
		t.Fatal("run(-node-witness) succeeded, want missing coverage error")
	}
	if !strings.Contains(err.Error(), "numberformat") || !strings.Contains(err.Error(), "resolved-options") {
		t.Fatalf("run(-node-witness) error = %v, want numberformat resolved-options guidance", err)
	}

	writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/node-v26/resolved-options.json", `[
		{"id":"numberformat-node-v26-resolved","source":"node:v26.0.0:numberformat:resolved-options","locale":"en-US","options":{},"input":1,"expected":"1","expectedResolvedOptions":{"locale":"en-US"}}
	]`)
	writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/node-v26/errors.json", `[
		{"id":"numberformat-node-v26-invalid-style","source":"node:v26.0.0:numberformat:errors","locale":"en-US","options":{"style":"invalid"},"input":1,"errorCode":"invalid_option"},
		{"id":"numberformat-node-v26-unit-casing-rejected","source":"node:v26.0.0:numberformat:errors","locale":"en","options":{"style":"unit","unit":"METER"},"input":1,"errorCode":"invalid_option"}
	]`)
	writeFixtureFile(t, root, "numberformat", "numberformat/testdata/conformance/node-v26/edge.json", `[
		{"id":"numberformat-node-v26-negative-zero-sign","source":"node:v26.0.0:numberformat:edge","locale":"en","options":{"signDisplay":"auto"},"input":"-0","expected":"-0","expectedParts":[{"type":"minusSign","value":"-"},{"type":"integer","value":"0"}],"expectedResolvedOptions":{"locale":"en"}},
		{"id":"numberformat-node-v26-rounding-increment","source":"node:v26.0.0:numberformat:edge","locale":"en","options":{"minimumFractionDigits":2,"maximumFractionDigits":2,"roundingIncrement":5},"input":1.234,"expected":"1.25","expectedParts":[{"type":"integer","value":"1"}],"expectedResolvedOptions":{"locale":"en"}},
		{"id":"numberformat-node-v26-rounding-priority-more-precision","source":"node:v26.0.0:numberformat:edge","locale":"en","options":{"minimumSignificantDigits":2,"maximumFractionDigits":0,"roundingPriority":"morePrecision"},"input":1.234,"expected":"1.234","expectedParts":[{"type":"integer","value":"1"}],"expectedResolvedOptions":{"locale":"en"}},
		{"id":"numberformat-node-v26-compact-plural-few","source":"node:v26.0.0:numberformat:edge","locale":"ru","options":{"notation":"compact","compactDisplay":"long"},"input":2000,"expected":"2 тысячи","expectedParts":[{"type":"integer","value":"2"},{"type":"compact","value":"тысячи"}],"expectedResolvedOptions":{"locale":"ru"}},
		{"id":"numberformat-node-v26-range-collapse","source":"node:v26.0.0:numberformat:edge","locale":"en","options":{"maximumFractionDigits":0},"input":{"start":1.2,"end":1.4},"expectedRange":"~1","expectedRangeParts":[{"type":"approximatelySign","value":"~","source":"shared"},{"type":"integer","value":"1","source":"shared"}],"expectedResolvedOptions":{"locale":"en"}},
		{"id":"numberformat-node-v26-czech-plural-range-unit","source":"node:v26.0.0:numberformat:edge","locale":"cs","options":{"style":"unit","unit":"meter","unitDisplay":"long"},"input":{"start":2,"end":1},"expectedRange":"2–1 metrů","expectedRangeParts":[{"type":"integer","value":"2","source":"startRange"},{"type":"literal","value":"–","source":"shared"},{"type":"integer","value":"1","source":"endRange"},{"type":"literal","value":" ","source":"shared"},{"type":"unit","value":"metrů","source":"shared"}],"expectedResolvedOptions":{"locale":"cs"}},
		{"id":"numberformat-node-v26-czech-plural-range-currency-name","source":"node:v26.0.0:numberformat:edge","locale":"cs","options":{"style":"currency","currency":"USD","currencyDisplay":"name","maximumFractionDigits":0},"input":{"start":2,"end":1},"expectedRange":"2–1 amerických dolarů","expectedRangeParts":[{"type":"integer","value":"2","source":"startRange"},{"type":"literal","value":"–","source":"shared"},{"type":"integer","value":"1","source":"endRange"},{"type":"literal","value":" ","source":"shared"},{"type":"currency","value":"amerických dolarů","source":"shared"}],"expectedResolvedOptions":{"locale":"cs"}},
		{"id":"numberformat-node-v26-negative-percent-range-affixes","source":"node:v26.0.0:numberformat:edge","locale":"en","options":{"style":"percent","maximumFractionDigits":0},"input":{"start":-0.01,"end":-0.02},"expectedRange":"-1–2%","expectedRangeParts":[{"type":"minusSign","value":"-","source":"shared"},{"type":"integer","value":"1","source":"startRange"},{"type":"literal","value":"–","source":"shared"},{"type":"integer","value":"2","source":"endRange"},{"type":"percentSign","value":"%","source":"shared"}],"expectedResolvedOptions":{"locale":"en"}}
	]`)
	if err := run([]string{"-node-witness", packageDir}, io.Discard); err != nil {
		t.Fatalf("run(-node-witness) error = %v, want nil", err)
	}
}

func TestRunRejectsInvalidFlags(t *testing.T) {
	t.Parallel()

	if err := run([]string{"-definitely-not-a-flag"}, io.Discard); err == nil {
		t.Fatal("run(invalid flag) error = nil, want error")
	}
}

func TestMainExitReportsErrorsWithoutExitingTestProcess(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := mainExit([]string{"-definitely-not-a-flag"}, &stdout, &stderr); code != 1 {
		t.Fatalf("mainExit(invalid flag) = %d, want 1", code)
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
	if !strings.Contains(stderr.String(), "flag provided but not defined") {
		t.Fatalf("stderr = %q, want flag parse error", stderr.String())
	}
}

func TestMainExitReturnsZeroForValidRoots(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	if code := mainExit(nil, &stdout, &stderr); code != 0 {
		t.Fatalf("mainExit(nil) = %d, stderr %q; want 0", code, stderr.String())
	}
	if stdout.Len() != 0 || stderr.Len() != 0 {
		t.Fatalf("stdout/stderr = %q/%q, want empty", stdout.String(), stderr.String())
	}
}

func TestMainUsesExitCode(t *testing.T) {
	// main mutates package globals and os.Args, so this test must stay serial.
	type exitCalled struct{}

	oldArgs := os.Args
	oldExit := exitProcess
	t.Cleanup(func() {
		os.Args = oldArgs
		exitProcess = oldExit
	})

	os.Args = []string{"check-conformance"}
	var exitCode int
	exitProcess = func(code int) {
		exitCode = code
		panic(exitCalled{})
	}
	defer func() {
		if got := recover(); got != (exitCalled{}) {
			t.Fatalf("main recover = %v, want exit called", got)
		}
		if exitCode != 0 {
			t.Fatalf("main exit code = %d, want 0", exitCode)
		}
	}()
	main()
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
