package conformance

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateFixtureRootsUsesXFailExpiry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	if err := os.WriteFile(filepath.Join(root, "testdata", "xfail.json"), []byte(`[
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2000-01-01","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}

	err := ValidateFixtureRoots([]string{root}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("ValidateFixtureRoots succeeded, want expired xfail error")
	}
	if !strings.Contains(err.Error(), "expired") || !strings.Contains(err.Error(), "nf-basic") {
		t.Fatalf("ValidateFixtureRoots error = %v, want expired xfail ID", err)
	}
}

func TestValidateFixtureRootsRejectsUnknownOrDuplicateXFail(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	if err := os.WriteFile(filepath.Join(root, "testdata", "xfail.json"), []byte(`[
		{"id":"nf-missing","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFixtureRoots([]string{root}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)); !errors.Is(err, errUnknownXFailID) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want unknown xfail id", err)
	}

	if err := os.WriteFile(filepath.Join(root, "testdata", "xfail.json"), []byte(`[
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"},
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := ValidateFixtureRoots([]string{root}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)); !errors.Is(err, errDuplicateXFailID) {
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
			if err := os.WriteFile(filepath.Join(root, "testdata", "xfail.json"), []byte(tc.data), 0o666); err != nil {
				t.Fatal(err)
			}
			err := ValidateFixtureRoots([]string{root}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
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

	err := ValidateFixtureRoots([]string{root}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("ValidateFixtureRoots succeeded, want invalid datetime input error")
	}
	if !strings.Contains(err.Error(), "dtf-invalid") || !strings.Contains(err.Error(), "ISO-8601") {
		t.Fatalf("ValidateFixtureRoots error = %v, want fixture ID and ISO-8601 guidance", err)
	}
}

func TestValidateFixtureRootsRejectsInvalidDateTimeRangeInput(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "datetimeformat")
	writeConformanceFixtureFile(t, root, `[
		{"id":"dtf-invalid-range","source":"manual","locale":"en-US","options":{},"input":{"start":"2026-05-08T12:00:00Z","end":"not-a-date"},"expectedRange":"May 8 - 10, 2026"}
	]`)

	err := ValidateFixtureRoots([]string{root}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
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

	if err := ValidateFixtureRoots([]string{root}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)); err != nil {
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

	err := ValidateFixtureRoots([]string{root}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
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
	if err := os.WriteFile(path, []byte(`{`), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadXFails(path); err == nil {
		t.Fatal("loadXFails(invalid) error = nil, want error")
	}

	if err := os.WriteFile(path, []byte(`[
		{"id":"nf-basic","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	entries, err = loadXFails(path)
	if err != nil {
		t.Fatalf("loadXFails(valid) error = %v, want nil", err)
	}
	if len(entries) != 1 || entries[0].ID != "nf-basic" || entries[0].TrackingIssue != "SPEC-70" {
		t.Fatalf("loadXFails(valid) = %+v, want nf-basic/SPEC-70", entries)
	}
}

func TestSkipReasonReportsDivergenceOrUnexpiredXFail(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "testdata"), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "testdata", "divergences.md"), []byte("id: diverged\nsource: manual:test\nowner: conformance\nreason: upstream output differs\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "testdata", "xfail.json"), []byte(`[
		{"id":"xfail","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}

	divergenceReason, ok := SkipReason(root, "diverged", time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if !ok || !strings.Contains(divergenceReason, "divergence") {
		t.Fatalf("SkipReason(diverged) = %q, %v; want divergence skip", divergenceReason, ok)
	}
	xfailReason, ok := SkipReason(root, "xfail", time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if !ok || !strings.Contains(xfailReason, "pending implementation") {
		t.Fatalf("SkipReason(xfail) = %q, %v; want xfail skip", xfailReason, ok)
	}
}

func TestRunFixturesSkipsXFailAndRunsCases(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeConformanceFixtureFile(t, root, `[
		{"id":"run","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"},
		{"id":"xfail","source":"manual","locale":"en-US","options":{},"input":2,"expected":"2"}
	]`)
	if err := os.WriteFile(filepath.Join(root, "testdata", "xfail.json"), []byte(`[
		{"id":"xfail","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}

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

	path := filepath.Join(root, "testdata", "conformance", "manual", "basic.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatal(err)
	}
}
