package conformance

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const runFixturesFailureRoot = "GO_INTL_RUN_FIXTURES_FAILURE_ROOT"

func TestRunFixturesRejectsMalformedDivergenceBeforeCallbacks(t *testing.T) {
	if runFixturesFailureChild(t) {
		return
	}

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeDivergenceFile(t, root, "malformed divergence line\n")
	assertRunFixturesFailsBeforeCallbacks(t, root, "TestRunFixturesRejectsMalformedDivergenceBeforeCallbacks", errMalformedDivergenceLine.Error())
}

func TestRunFixturesRejectsMalformedXFailBeforeCallbacks(t *testing.T) {
	if runFixturesFailureChild(t) {
		return
	}

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeXFailFile(t, root, `{`)
	assertRunFixturesFailsBeforeCallbacks(t, root, "TestRunFixturesRejectsMalformedXFailBeforeCallbacks", xfailPath(root))
}

func TestRunFixturesRejectsExpiredXFailBeforeCallbacks(t *testing.T) {
	if runFixturesFailureChild(t) {
		return
	}

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeXFailFile(t, root, `[
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2000-01-01","tracking_issue":"SPEC-70"}
	]`)
	assertRunFixturesFailsBeforeCallbacks(t, root, "TestRunFixturesRejectsExpiredXFailBeforeCallbacks", errExpiredXFail.Error())
}

func TestRunFixturesRejectsUnknownDivergenceBeforeCallbacks(t *testing.T) {
	if runFixturesFailureChild(t) {
		return
	}

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeDivergenceFile(t, root, "id: missing\nsource: manual\nowner: conformance\nstatus: accepted\nreason: upstream output differs\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")
	assertRunFixturesFailsBeforeCallbacks(t, root, "TestRunFixturesRejectsUnknownDivergenceBeforeCallbacks", errUnknownDivergence.Error())
}

func runFixturesFailureChild(t *testing.T) bool {
	t.Helper()

	root := os.Getenv(runFixturesFailureRoot)
	if root == "" {
		return false
	}
	RunFixtures(t, root, func(t *testing.T, _ Fixture) {
		marker := filepath.Join(root, "callback-ran")
		if err := os.WriteFile(marker, nil, 0o600); err != nil {
			t.Fatalf("write callback marker: %v", err)
		}
	})
	return true
}

func assertRunFixturesFailsBeforeCallbacks(t *testing.T, root, testName, want string) {
	t.Helper()

	cmd := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^"+testName+"$")
	cmd.Env = append(os.Environ(), runFixturesFailureRoot+"="+root)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("RunFixtures accepted invalid suite state and ran callbacks; child output:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(root, "callback-ran")); !os.IsNotExist(err) {
		t.Fatalf("fixture callback ran before suite failure: %v", err)
	}
	if !strings.Contains(string(output), want) {
		t.Fatalf("RunFixtures failure = %s, want %q", output, want)
	}
}

func TestValidateFixtureRootsUsesXFailExpiry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeXFailFile(t, root, `[
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2000-01-01","tracking_issue":"SPEC-70"}
	]`)

	err := ValidateFixtureRoots([]string{root}, conformanceAuditNow())
	if !errors.Is(err, errExpiredXFail) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want expired xfail", err)
	}
}

func TestValidateFixtureRootsRejectsUnknownDivergence(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeDivergenceFile(t, root, "id: missing\nsource: manual\nowner: conformance\nstatus: accepted\nreason: upstream output differs\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")

	err := ValidateFixtureRoots([]string{root}, conformanceAuditNow())
	if !errors.Is(err, errUnknownDivergence) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want unknown divergence", err)
	}
}

func TestValidateFixtureRootsRejectsMissingPackageRoot(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "missing")
	err := ValidateFixtureRoots([]string{root}, conformanceAuditNow())
	if !errors.Is(err, errMissingPackageRoot) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want missing package root", err)
	}
}

func TestValidateFixtureRootsRejectsUnknownOrDuplicateXFail(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeXFailFile(t, root, `[
		{"id":"nf-missing","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`)
	if err := ValidateFixtureRoots([]string{root}, conformanceAuditNow()); !errors.Is(err, errUnknownXFailID) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want unknown xfail id", err)
	}

	writeXFailFile(t, root, `[
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"},
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`)
	if err := ValidateFixtureRoots([]string{root}, conformanceAuditNow()); !errors.Is(err, errDuplicateXFailID) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want duplicate xfail id", err)
	}
}

func TestValidateFixtureRootsRejectsMalformedXFail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
		want error
	}{
		{
			name: "missing id",
			data: `[{"reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}]`,
			want: errMissingXFailField,
		},
		{
			name: "missing reason",
			data: `[{"id":"nf-basic","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}]`,
			want: errMissingXFailField,
		},
		{
			name: "missing expiry",
			data: `[{"id":"nf-basic","reason":"pending implementation","tracking_issue":"SPEC-70"}]`,
			want: errMissingXFailField,
		},
		{
			name: "missing tracking issue",
			data: `[{"id":"nf-basic","reason":"pending implementation","expires_at":"2999-01-01"}]`,
			want: errMissingXFailField,
		},
		{
			name: "invalid expiry",
			data: `[{"id":"nf-basic","reason":"pending implementation","expires_at":"soon","tracking_issue":"SPEC-70"}]`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeConformanceFixtureFile(t, root, `[
				{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
			]`)
			writeXFailFile(t, root, tc.data)
			err := ValidateFixtureRoots([]string{root}, conformanceAuditNow())
			if tc.want != nil {
				if !errors.Is(err, tc.want) {
					t.Fatalf("ValidateFixtureRoots() error = %v, want %v", err, tc.want)
				}
				return
			}
			if err == nil {
				t.Fatal("ValidateFixtureRoots() error = nil, want error")
			}
		})
	}
}

func TestValidateFixtureRootsRejectsInvalidDateTimeInput(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "datetimeformat")
	writeConformanceFixtureFile(t, root, `[
		{"id":"dtf-invalid","source":"manual","locale":"en-US","options":{},"input":"not-a-date","expected":"May 8, 2026"}
	]`)

	err := ValidateFixtureRoots([]string{root}, conformanceAuditNow())
	if !errors.Is(err, errInvalidDateTimeInput) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want invalid datetime input", err)
	}
}

func TestValidateFixtureRootsRejectsInvalidDateTimeRangeInput(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "datetimeformat")
	writeConformanceFixtureFile(t, root, `[
		{"id":"dtf-invalid-range","source":"manual","locale":"en-US","options":{},"input":{"start":"2026-05-08T12:00:00Z","end":"not-a-date"},"expectedRange":"May 8 - 10, 2026"}
	]`)

	err := ValidateFixtureRoots([]string{root}, conformanceAuditNow())
	if !errors.Is(err, errInvalidDateTimeInput) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want invalid datetime input", err)
	}
}

func TestValidateFixtureRootsAcceptsValidDateTimeRangeInput(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "datetimeformat")
	writeConformanceFixtureFile(t, root, `[
		{"id":"dtf-range","source":"manual","locale":"en-US","options":{},"input":{"start":"2026-05-08T12:00:00Z","end":"2026-05-10T12:00:00Z"},"expectedRange":"May 8 - 10, 2026"}
	]`)

	if err := ValidateFixtureRoots([]string{root}, conformanceAuditNow()); err != nil {
		t.Fatalf("ValidateFixtureRoots(valid datetime range) error = %v, want nil", err)
	}
}

func TestValidateFixtureRootsRejectsDuplicateFixtureID(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"duplicate","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"},
		{"id":"duplicate","source":"manual","locale":"en-US","options":{},"input":2,"expected":"2"}
	]`)

	err := ValidateFixtureRoots([]string{root}, conformanceAuditNow())
	if !errors.Is(err, errDuplicateFixtureID) {
		t.Fatalf("ValidateFixtureRoots(duplicate IDs) error = %v, want duplicate fixture id", err)
	}
}

func TestLoadFixturesRejectsShapeSourceAndErrorFileBoundaries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		data string
		want error
	}{
		{
			name: "malformed json",
			rel:  "manual/basic.json",
			data: `{`,
		},
		{
			name: "missing locale",
			rel:  "manual/basic.json",
			data: `[{"id":"missing-locale","source":"manual","options":{},"input":1,"expected":"1"}]`,
			want: errMissingFixtureField,
		},
		{
			name: "missing id",
			rel:  "manual/basic.json",
			data: `[{"source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
			want: errMissingFixtureField,
		},
		{
			name: "missing source",
			rel:  "manual/basic.json",
			data: `[{"id":"missing-source","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
			want: errMissingFixtureField,
		},
		{
			name: "missing options",
			rel:  "manual/basic.json",
			data: `[{"id":"missing-options","source":"manual","locale":"en-US","input":1,"expected":"1"}]`,
			want: errMissingFixtureField,
		},
		{
			name: "missing input",
			rel:  "manual/basic.json",
			data: `[{"id":"missing-input","source":"manual","locale":"en-US","options":{},"expected":"1"}]`,
			want: errMissingFixtureField,
		},
		{
			name: "mixed sources in one file",
			rel:  "manual/basic.json",
			data: `[
				{"id":"one","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"},
				{"id":"two","source":"manual:two","locale":"en-US","options":{},"input":2,"expected":"2"}
			]`,
			want: errMixedFixtureSources,
		},
		{
			name: "source directory mismatch",
			rel:  "formatjs/basic.json",
			data: `[{"id":"manual-in-formatjs","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
			want: errFixtureSourceDir,
		},
		{
			name: "node witness source version mismatch",
			rel:  "node-v26/basic.json",
			data: `[{"id":"numberformat-node-v26-smoke","source":"node:v25.0.0:numberformat","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
			want: errFixtureSourceDir,
		},
		{
			name: "node witness id version mismatch",
			rel:  "node-v26/basic.json",
			data: `[{"id":"numberformat-node-v25-smoke","source":"node:v26.0.0:numberformat","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
			want: errFixtureSourceDir,
		},
		{
			name: "error code outside errors file",
			rel:  "manual/basic.json",
			data: `[{"id":"error-outside","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1","errorCode":"RangeError"}]`,
			want: errInvalidErrorFixture,
		},
		{
			name: "errors file missing error code",
			rel:  "manual/errors.json",
			data: `[{"id":"missing-error-code","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
			want: errInvalidErrorFixture,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeCoverageFixtureFile(t, root, tc.rel, tc.data)
			_, err := LoadFixtures(root)
			if tc.want == nil {
				if err == nil {
					t.Fatal("LoadFixtures() error = nil, want malformed JSON error")
				}
				return
			}
			if !errors.Is(err, tc.want) {
				t.Fatalf("LoadFixtures() error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestLoadFixturesAcceptsErrorFixtureFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeCoverageFixtureFile(t, root, "manual/errors.json", `[
		{"id":"range-error","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1","errorCode":"RangeError"}
	]`)
	fixtures, err := LoadFixtures(root)
	if err != nil {
		t.Fatalf("LoadFixtures(errors.json) error = %v, want nil", err)
	}
	if len(fixtures) != 1 || fixtures[0].ErrorCode != "RangeError" {
		t.Fatalf("LoadFixtures(errors.json) = %+v, want RangeError fixture", fixtures)
	}
}

func TestLoadFixturesAcceptsSourceDirectoryOwnership(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rel  string
		data string
	}{
		{
			name: "manual source in manual directory",
			rel:  "manual/basic.json",
			data: `[{"id":"numberformat-manual-basic","source":"manual:numberformat","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
		},
		{
			name: "formatjs source in formatjs directory",
			rel:  "formatjs/basic.json",
			data: `[{"id":"numberformat-formatjs-basic","source":"formatjs:packages/intl-numberformat/tests/basic.test.ts","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
		},
		{
			name: "node source in exact node version directory",
			rel:  "node-v26/basic.json",
			data: `[{"id":"numberformat-node-v26-basic","source":"node:v26.0.0:numberformat","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
		},
		{
			name: "node source in legacy node directory",
			rel:  "node/basic.json",
			data: `[{"id":"numberformat-node-basic","source":"node:v26.0.0:numberformat","locale":"en-US","options":{},"input":1,"expected":"1"}]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			writeCoverageFixtureFile(t, root, tc.rel, tc.data)
			fixtures, err := LoadFixtures(root)
			if err != nil {
				t.Fatalf("LoadFixtures() error = %v, want nil", err)
			}
			if len(fixtures) != 1 {
				t.Fatalf("LoadFixtures() returned %d fixtures, want 1", len(fixtures))
			}
		})
	}
}

func TestFixtureSourceKindOfClassifiesFixtureSources(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		source string
		want   fixtureSourceKind
	}{
		{name: "manual root", source: "manual", want: fixtureSourceManual},
		{name: "manual scoped", source: "manual:numberformat", want: fixtureSourceManual},
		{name: "formatjs", source: "formatjs:packages/intl-numberformat/tests/basic.test.ts", want: fixtureSourceFormatJS},
		{name: "node colon source", source: "node:v26.0.0:numberformat", want: fixtureSourceNode},
		{name: "node dotted source", source: "node:v26.0.0.datetimeformat:edge", want: fixtureSourceNode},
		{name: "unknown prefix", source: "native:numberformat", want: fixtureSourceUnknown},
		{name: "empty source", source: "", want: fixtureSourceUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := fixtureSourceKindOf(tc.source); got != tc.want {
				t.Fatalf("fixtureSourceKindOf(%q) = %q, want %q", tc.source, got, tc.want)
			}
		})
	}
}

func TestFixtureHasNativeExpectationRecognizesObservableFields(t *testing.T) {
	t.Parallel()

	text := "ok"
	ok := true
	comparison := 0
	byteIndex := 0
	wordLike := true

	tests := []struct {
		name    string
		fixture Fixture
		want    bool
	}{
		{name: "empty fixture"},
		{name: "expected output", fixture: Fixture{Expected: &text}, want: true},
		{name: "expected ok", fixture: Fixture{ExpectedOK: &ok}, want: true},
		{name: "expected locales", fixture: Fixture{ExpectedLocales: []string{"en"}}, want: true},
		{name: "expected parts", fixture: Fixture{ExpectedParts: []Part{{Type: "integer", Value: "1"}}}, want: true},
		{name: "expected range", fixture: Fixture{ExpectedRange: &text}, want: true},
		{name: "expected range parts", fixture: Fixture{ExpectedRangeParts: []RangePart{{Type: "integer", Value: "1", Source: "shared"}}}, want: true},
		{name: "expected comparison", fixture: Fixture{ExpectedComparison: &comparison}, want: true},
		{name: "expected resolved options", fixture: Fixture{ExpectedResolved: json.RawMessage(`{"locale":"en"}`)}, want: true},
		{name: "expected segments", fixture: Fixture{ExpectedSegments: []SegmentRecord{{Segment: "a", CodeUnitIndex: 0, ByteIndex: &byteIndex, IsWordLike: &wordLike}}}, want: true},
		{name: "error code", fixture: Fixture{ErrorCode: "RangeError"}, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := fixtureHasNativeExpectation(tc.fixture); got != tc.want {
				t.Fatalf("fixtureHasNativeExpectation() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestLoadFixturesReturnsEmptyForMissingConformanceDirectory(t *testing.T) {
	t.Parallel()

	fixtures, err := LoadFixtures(t.TempDir())
	if err != nil {
		t.Fatalf("LoadFixtures() error = %v, want nil", err)
	}
	if len(fixtures) != 0 {
		t.Fatalf("LoadFixtures() returned %d fixtures, want none", len(fixtures))
	}
}

func TestXFailLoaderHandlesMissingInvalidAndValidFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	missingPath := filepath.Join(root, "missing.json")
	entries, err := loadXFails(missingPath)
	if err != nil {
		t.Fatalf("loadXFails(missing) error = %v, want nil", err)
	}
	if len(entries) != 0 {
		t.Fatalf("loadXFails(missing) returned %d entries, want none", len(entries))
	}

	path := filepath.Join(root, "xfail.json")
	writeXFailFileAt(t, path, `{`)
	if _, err := loadXFails(path); err == nil {
		t.Fatal("loadXFails(invalid) error = nil, want error")
	}

	writeXFailFileAt(t, path, `[
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`)
	entries, err = loadXFails(path)
	if err != nil {
		t.Fatalf("loadXFails(valid) error = %v, want nil", err)
	}
	if len(entries) != 1 || entries[0].ID != "nf-basic" || entries[0].TrackingIssue != "SPEC-70" {
		t.Fatalf("loadXFails(valid) = %+v, want nf-basic/SPEC-70", entries)
	}
}

func TestLoadRunSuiteCompilesDivergenceOrUnexpiredXFail(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"diverged","source":"manual:test","locale":"en-US","options":{},"input":1,"expected":"1"},
		{"id":"xfail","source":"manual:test","locale":"en-US","options":{},"input":2,"expected":"2"}
	]`)
	writeDivergenceFile(t, root, "id: diverged\nsource: manual:test\nowner: conformance\nstatus: accepted\nreason: upstream output differs\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")
	writeXFailFile(t, root, `[
		{"id":"xfail","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`)

	suite, err := loadRunSuite(root, conformanceAuditNow())
	if err != nil {
		t.Fatalf("loadRunSuite() error = %v, want nil", err)
	}
	if err := os.RemoveAll(conformanceFixturesPath(root)); err != nil {
		t.Fatalf("remove fixture files after suite construction: %v", err)
	}
	writeDivergenceFile(t, root, "malformed divergence line\n")
	writeXFailFile(t, root, `{`)
	if len(suite.fixtures) != 2 {
		t.Fatalf("compiled suite fixtures = %d, want 2", len(suite.fixtures))
	}
	divergenceReason, ok := suite.skipReasons["diverged"]
	if !ok || !strings.Contains(divergenceReason, "divergence") {
		t.Fatalf("suite divergence reason = %q, %v; want divergence skip", divergenceReason, ok)
	}
	xfailReason, ok := suite.skipReasons["xfail"]
	if !ok || !strings.Contains(xfailReason, "pending implementation") {
		t.Fatalf("suite XFAIL reason = %q, %v; want XFAIL skip", xfailReason, ok)
	}
}

func TestRunFixturesSkipsXFailAndRunsCases(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"run","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"},
		{"id":"xfail","source":"manual","locale":"en-US","options":{},"input":2,"expected":"2"}
	]`)
	writeXFailFile(t, root, `[
		{"id":"xfail","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`)

	runIDs := map[string]bool{}
	t.Cleanup(func() {
		if !runIDs["run"] || runIDs["xfail"] {
			t.Fatalf("RunFixtures executed IDs = %v", runIDs)
		}
	})
	RunFixtures(t, root, func(t *testing.T, fixture Fixture) {
		runIDs[fixture.ID] = true
	})
}

func writeConformanceFixtureFile(t *testing.T, root, data string) {
	t.Helper()

	writeCoverageFixtureFile(t, root, "manual/basic.json", data)
}

func writeXFailFile(t *testing.T, root, data string) {
	t.Helper()

	path := filepath.Join(root, "testdata", "xfail.json")
	writeXFailFileAt(t, path, data)
}

func writeXFailFileAt(t *testing.T, path, data string) {
	t.Helper()

	writeConformanceTestFile(t, path, data)
}
