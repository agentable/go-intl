package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateFixturesUsesXFailExpiry(t *testing.T) {
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

	err := ValidateFixtures(root, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("ValidateFixtures succeeded, want expired xfail error")
	}
	if !strings.Contains(err.Error(), "expired") || !strings.Contains(err.Error(), "nf-basic") {
		t.Fatalf("ValidateFixtures error = %v, want expired xfail ID", err)
	}
}

func TestValidateFixturesRejectsInvalidDateTimeInput(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "datetimeformat")
	writeConformanceFixtureFile(t, root, `[
		{"id":"dtf-invalid","source":"manual","locale":"en-US","options":{},"input":"not-a-date","expected":"May 8, 2026"}
	]`)

	err := ValidateFixtures(root, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("ValidateFixtures succeeded, want invalid datetime input error")
	}
	if !strings.Contains(err.Error(), "dtf-invalid") || !strings.Contains(err.Error(), "ISO-8601") {
		t.Fatalf("ValidateFixtures error = %v, want fixture ID and ISO-8601 guidance", err)
	}
}

func TestSkipReasonReportsDivergenceOrUnexpiredXFail(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "testdata"), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "testdata", "divergences.md"), []byte("id: diverged\nreason: upstream output differs\n"), 0o666); err != nil {
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
