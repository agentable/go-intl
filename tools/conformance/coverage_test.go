package conformance

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestValidateSkipListAllowsOnlyExtractorLimitCategories(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skipListPath := filepath.Join(root, ".skip-list.json")
	if err := os.WriteFile(skipListPath, []byte(`[
		{"source":"formatjs:unsupported","category":"unsupported-extractor-shape","route":"extractor","reason":"unsupported shape"},
		{"source":"formatjs:partial","category":"partial-extraction","route":"extractor","reason":"remaining assertions"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSkipList(skipListPath, nil); err != nil {
		t.Fatalf("ValidateSkipList() error = %v, want nil", err)
	}

	if err := os.WriteFile(skipListPath, []byte(`[
		{"source":"formatjs:invalid","category":"typo","route":"extractor","reason":"invalid category"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	err := ValidateSkipList(skipListPath, nil)
	if !errors.Is(err, errInvalidSkipListCategory) {
		t.Fatalf("ValidateSkipList() error = %v, want invalid category", err)
	}

	if err := os.WriteFile(skipListPath, []byte(`[
		{"source":"formatjs:snapshot","category":"snapshot-source","route":"extractor","reason":"reference snapshot output has no source input mapping"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	err = ValidateSkipList(skipListPath, nil)
	if !errors.Is(err, errInvalidSkipListCategory) {
		t.Fatalf("ValidateSkipList() error = %v, want invalid category", err)
	}

	if err := os.WriteFile(skipListPath, []byte(`[
		{"source":"formatjs:accepted","category":"accepted-divergence","route":"extractor","reason":"accepted mismatch","divergenceId":"nf-accepted"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	err = ValidateSkipList(skipListPath, nil)
	if !errors.Is(err, errInvalidSkipListCategory) {
		t.Fatalf("ValidateSkipList() error = %v, want invalid category", err)
	}
}

func TestValidateSkipListRejectsMalformedEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	skipListPath := filepath.Join(root, ".skip-list.json")
	if err := os.WriteFile(skipListPath, []byte(`{`), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadSkipList(skipListPath); err == nil {
		t.Fatal("loadSkipList(invalid) error = nil, want error")
	}

	tests := []struct {
		name string
		data string
		want error
	}{
		{
			name: "missing source",
			data: `[{"category":"partial-extraction","reason":"pending"}]`,
			want: errMissingSkipListField,
		},
		{
			name: "missing category",
			data: `[{"source":"formatjs:one","reason":"pending"}]`,
			want: errMissingSkipListField,
		},
		{
			name: "missing reason",
			data: `[{"source":"formatjs:one","category":"partial-extraction"}]`,
			want: errMissingSkipListField,
		},
		{
			name: "missing route",
			data: `[{"source":"formatjs:one","category":"partial-extraction","reason":"pending"}]`,
			want: errMissingSkipListField,
		},
		{
			name: "duplicate source",
			data: `[
				{"source":"formatjs:one","category":"partial-extraction","route":"extractor","reason":"pending"},
				{"source":"formatjs:one","category":"partial-extraction","route":"extractor","reason":"still pending"}
			]`,
			want: errDuplicateSkipListSource,
		},
		{
			name: "unexpected divergence id",
			data: `[{"source":"formatjs:one","category":"partial-extraction","route":"extractor","reason":"pending","divergenceId":"nf-one"}]`,
			want: errInvalidSkipListCategory,
		},
		{
			name: "accepted divergence category",
			data: `[{"source":"formatjs:one","category":"accepted-divergence","route":"extractor","reason":"accepted mismatch"}]`,
			want: errInvalidSkipListCategory,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), ".skip-list.json")
			if err := os.WriteFile(path, []byte(tc.data), 0o666); err != nil {
				t.Fatal(err)
			}
			err := ValidateSkipList(path, nil)
			if !errors.Is(err, tc.want) {
				t.Fatalf("ValidateSkipList() error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestValidateSkipListAuditsNativeWitnessRoutes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := filepath.Join(root, "datetimeformat")
	writeCoverageFixtureFile(t, packageDir, "node-v26/range.json", `[
		{"id":"datetimeformat-node-v26-range","source":"node:v26.0.0:datetimeformat:p4-deep-contract","locale":"en","options":{},"input":{"start":"2021-01-10T00:00:00Z","end":"2021-01-20T00:00:00Z"},"expectedRange":"Jan 10 - Jan 20"}
	]`)
	skipListPath := filepath.Join(root, ".skip-list.json")
	if err := os.WriteFile(skipListPath, []byte(`[
		{"source":"formatjs:covered","category":"unsupported-extractor-shape","route":"native-witness","witness":"datetimeformat-node-v26-missing","reason":"native lane owns this observable case"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	err := ValidateSkipList(skipListPath, []string{packageDir})
	if !errors.Is(err, errUnknownSkipListWitness) {
		t.Fatalf("ValidateSkipList() error = %v, want unknown skip-list witness", err)
	}

	if err := os.WriteFile(skipListPath, []byte(`[
		{"source":"formatjs:covered","category":"unsupported-extractor-shape","route":"native-witness","witness":"datetimeformat-node-v26-range","reason":"native lane owns this observable case"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	if err := ValidateSkipList(skipListPath, []string{packageDir}); err != nil {
		t.Fatalf("ValidateSkipList() error = %v, want nil", err)
	}
}

func TestCoverageReportCountsFixtureAndSkipHealth(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := filepath.Join(root, "numberformat")
	writeCoverageFixtureFile(t, packageDir, "manual/basic.json", `[
		{"id":"nf-manual","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeCoverageFixtureFile(t, packageDir, "formatjs/basic.json", `[
		{"id":"nf-formatjs","source":"formatjs:packages/intl-numberformat/tests/basic.test.ts","locale":"en-US","options":{},"input":2,"expected":"2"}
	]`)
	writeCoverageFixtureFile(t, packageDir, "node/basic.json", `[
		{"id":"nf-node","source":"node:v76.1:numberFormats","locale":"en-US","options":{},"input":3,"expected":"3"}
	]`)
	writeDivergenceFile(t, packageDir, "id: nf-formatjs\nsource: formatjs:packages/intl-numberformat/tests/basic.test.ts\nowner: numberformat\nreason: accepted reference mismatch\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")
	if err := os.WriteFile(filepath.Join(packageDir, "testdata", "xfail.json"), []byte(`[
		{"id":"nf-node","reason":"pending implementation","expires_at":"2999-01-01","tracking_issue":"SPEC-70"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}
	skipListPath := filepath.Join(root, ".skip-list.json")
	if err := os.WriteFile(skipListPath, []byte(`[
		{"source":"formatjs:partial","category":"partial-extraction","route":"extractor","reason":"remaining assertions"},
		{"source":"formatjs:unsupported","category":"unsupported-extractor-shape","route":"not-applicable","reason":"unsupported shape"}
	]`), 0o666); err != nil {
		t.Fatal(err)
	}

	if err := ValidateSkipList(skipListPath, []string{packageDir}); err != nil {
		t.Fatalf("ValidateSkipList() error = %v, want nil", err)
	}
	report, err := buildCoverageReport([]string{packageDir}, skipListPath)
	if err != nil {
		t.Fatalf("buildCoverageReport() error = %v", err)
	}
	if report.Total.Fixtures != 3 || report.Total.Sources != 3 || report.Total.Manual != 1 || report.Total.FormatJS != 1 || report.Total.Node != 1 {
		t.Fatalf("coverage totals = %+v, want one manual, formatjs, and node fixture", report.Total)
	}
	if report.Total.Divergences != 1 || report.Total.XFails != 1 {
		t.Fatalf("coverage audit totals = %+v, want one divergence and one xfail", report.Total)
	}
	if report.SkipList.Entries != 2 || report.SkipList.Categories["partial-extraction"] != 1 || report.SkipList.Categories["unsupported-extractor-shape"] != 1 {
		t.Fatalf("skip coverage = %+v, want partial-extraction=1 and unsupported-extractor-shape=1", report.SkipList)
	}
	if report.SkipList.Routes["extractor"] != 1 || report.SkipList.Routes["not-applicable"] != 1 {
		t.Fatalf("skip routes = %+v, want extractor=1 and not-applicable=1", report.SkipList.Routes)
	}
	output := formatCoverageReport(report)
	for _, want := range []string{"numberformat:", "fixtures=3", "partial-extraction=1", "unsupported-extractor-shape=1", "routes extractor=1 not-applicable=1"} {
		if !strings.Contains(output, want) {
			t.Fatalf("formatCoverageReport() = %q, want %q", output, want)
		}
	}
	wrapperOutput, err := CoverageReport([]string{packageDir}, skipListPath)
	if err != nil {
		t.Fatalf("CoverageReport() error = %v", err)
	}
	if wrapperOutput != output {
		t.Fatalf("CoverageReport() = %q, want %q", wrapperOutput, output)
	}
}

func TestCoverageReportWithoutSkipListSortsPackages(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	zetaDir := filepath.Join(root, "zeta")
	alphaDir := filepath.Join(root, "alpha")
	writeCoverageFixtureFile(t, zetaDir, "manual/basic.json", `[
		{"id":"zeta-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeCoverageFixtureFile(t, alphaDir, "node/basic.json", `[
		{"id":"alpha-node","source":"node:v1","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	report, err := buildCoverageReport([]string{zetaDir, alphaDir}, "")
	if err != nil {
		t.Fatalf("buildCoverageReport() error = %v", err)
	}
	if len(report.Packages) != 2 || report.Packages[0].Package != "alpha" || report.Packages[1].Package != "zeta" {
		t.Fatalf("buildCoverageReport().Packages = %+v, want alpha then zeta", report.Packages)
	}
	if report.SkipList.Entries != 0 || report.Total.Fixtures != 2 || report.Total.Sources != 2 {
		t.Fatalf("buildCoverageReport() = %+v, want no skip list and two fixtures", report)
	}
}

func TestDivergenceIDLoaderRequiresAuditableActiveEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "divergences.md")
	if err := os.WriteFile(path, []byte("id: missing-source\nowner: numberformat\nreason: mismatch\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDivergenceIDs(path); !errors.Is(err, errMissingDivergenceField) {
		t.Fatalf("loadDivergenceIDs() error = %v, want missing field", err)
	}

	if err := os.WriteFile(path, []byte("id: missing-owner\nsource: manual:one\nreason: mismatch\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDivergenceIDs(path); !errors.Is(err, errMissingDivergenceField) || !strings.Contains(err.Error(), "owner") {
		t.Fatalf("loadDivergenceIDs() error = %v, want missing owner field", err)
	}

	if err := os.WriteFile(path, []byte("id: missing-removal\nsource: manual:one\nowner: numberformat\nreason: mismatch\nreview_after: 2026-11-01\n"), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDivergenceIDs(path); !errors.Is(err, errMissingDivergenceField) || !strings.Contains(err.Error(), "removal_path") {
		t.Fatalf("loadDivergenceIDs() error = %v, want missing removal_path field", err)
	}

	if err := os.WriteFile(path, []byte(`
id: duplicate
source: manual:one
owner: numberformat
reason: mismatch
review_after: 2026-11-01
removal_path: refresh the native reference

id: duplicate
source: manual:two
owner: numberformat
reason: mismatch
review_after: 2026-11-01
removal_path: refresh the native reference
`), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDivergenceIDs(path); !errors.Is(err, errDuplicateDivergenceID) {
		t.Fatalf("loadDivergenceIDs() error = %v, want duplicate id", err)
	}

	if err := os.WriteFile(path, []byte(`
id: invalid-status
source: manual:one
owner: numberformat
status: pending
reason: mismatch
review_after: 2026-11-01
removal_path: refresh the native reference
`), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDivergenceIDs(path); !errors.Is(err, errInvalidDivergenceStatus) {
		t.Fatalf("loadDivergenceIDs() error = %v, want invalid status", err)
	}

	if err := os.WriteFile(path, []byte(`
id: invalid-date
source: manual:one
owner: numberformat
reason: mismatch
review_after: soon
removal_path: refresh the native reference
`), 0o666); err != nil {
		t.Fatal(err)
	}
	if _, err := loadDivergenceIDs(path); !errors.Is(err, errInvalidDivergenceReviewDate) {
		t.Fatalf("loadDivergenceIDs() error = %v, want invalid review date", err)
	}

	if err := os.WriteFile(path, []byte(`
id: active
source: manual:one
owner: numberformat
reason: mismatch
review_after: 2026-11-01
removal_path: refresh the native reference

id: historical
status: resolved
reason: old mismatch
`), 0o666); err != nil {
		t.Fatal(err)
	}
	ids, err := loadDivergenceIDs(path)
	if err != nil {
		t.Fatalf("loadDivergenceIDs() error = %v", err)
	}
	if _, ok := ids["active"]; !ok {
		t.Fatalf("loadDivergenceIDs() = %v, want active id", ids)
	}
	if _, ok := ids["historical"]; ok {
		t.Fatalf("loadDivergenceIDs() = %v, want resolved id excluded", ids)
	}
}

func TestValidateDivergencesAuditsActiveEntries(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := filepath.Join(root, "numberformat")
	if err := os.MkdirAll(packageDir, 0o777); err != nil {
		t.Fatal(err)
	}
	if err := ValidateDivergences(packageDir); err != nil {
		t.Fatalf("ValidateDivergences(missing file) error = %v, want nil", err)
	}

	writeConformanceFixtureFile(t, packageDir, `[
		{"id":"nf-basic","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	missingRoot := filepath.Join(root, "missing")
	if err := ValidateDivergences(missingRoot); !errors.Is(err, errMissingPackageRoot) {
		t.Fatalf("ValidateDivergences(missing root) error = %v, want missing package root", err)
	}

	writeDivergenceFile(t, packageDir, "id: nf-basic\nsource: manual\nowner: numberformat\nreason: accepted reference mismatch\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")
	if err := ValidateDivergences(packageDir); err != nil {
		t.Fatalf("ValidateDivergences() error = %v, want nil", err)
	}

	writeDivergenceFile(t, packageDir, "id: nf-missing\nsource: manual\nowner: numberformat\nreason: accepted reference mismatch\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")
	err := ValidateDivergences(packageDir)
	if !errors.Is(err, errUnknownDivergence) {
		t.Fatalf("ValidateDivergences() error = %v, want unknown divergence", err)
	}

	writeDivergenceFile(t, packageDir, "id: nf-basic\nsource: formatjs:wrong\nowner: numberformat\nreason: accepted reference mismatch\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")
	err = ValidateDivergences(packageDir)
	if !errors.Is(err, errDivergenceSourceMismatch) {
		t.Fatalf("ValidateDivergences() error = %v, want source mismatch", err)
	}
}

func TestValidateDivergencesRequiresDateTimeFormatNativeWitness(t *testing.T) {
	t.Parallel()

	packageDir := filepath.Join(t.TempDir(), "datetimeformat")
	writeCoverageFixtureFile(t, packageDir, "formatjs/range.json", `[
		{"id":"dtf-formatjs-range","source":"formatjs:packages/intl-datetimeformat/tests/format-range.test.ts","locale":"en","options":{},"input":{"start":"2021-01-10T00:00:00Z","end":"2021-01-20T00:00:00Z"},"expectedRange":"Jan 10 - Jan 20"}
	]`)
	writeCoverageFixtureFile(t, packageDir, "node-v26/range.json", `[
		{"id":"datetimeformat-node-v26-range","source":"node:v26.0.0:datetimeformat:p4-deep-contract","locale":"en","options":{},"input":{"start":"2021-01-10T00:00:00Z","end":"2021-01-20T00:00:00Z"},"expectedRange":"Jan 10 - Jan 20","expectedRangeParts":[{"type":"month","value":"Jan","source":"shared"}]}
	]`)

	writeDivergenceFile(t, packageDir, "id: dtf-formatjs-range\nsource: formatjs:packages/intl-datetimeformat/tests/format-range.test.ts\nowner: datetimeformat\nreason: accepted DateTimeFormat range mismatch\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")
	err := ValidateDivergences(packageDir)
	if !errors.Is(err, errMissingDivergenceWitness) {
		t.Fatalf("ValidateDivergences() error = %v, want missing native_witness", err)
	}

	writeDivergenceFile(t, packageDir, "id: dtf-formatjs-range\nsource: formatjs:packages/intl-datetimeformat/tests/format-range.test.ts\nowner: datetimeformat\nreason: accepted DateTimeFormat range mismatch\nnative_witness: datetimeformat-node-v26-range\nreview_after: 2026-11-01\nremoval_path: refresh the native reference\n")
	if err := ValidateDivergences(packageDir); err != nil {
		t.Fatalf("ValidateDivergences() error = %v, want nil", err)
	}
}

func TestValidateFixtureRootsRejectsDuplicateIDsAcrossRoots(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	localeDir := filepath.Join(root, "locale")
	numberDir := filepath.Join(root, "numberformat")
	writeConformanceFixtureFile(t, localeDir, `[
		{"id":"shared-id","source":"manual","locale":"en-US","options":{},"input":"en-US","expected":"en-US"}
	]`)
	writeConformanceFixtureFile(t, numberDir, `[
		{"id":"shared-id","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	err := ValidateFixtureRoots([]string{localeDir, numberDir}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC))
	if !errors.Is(err, errDuplicateFixtureID) {
		t.Fatalf("ValidateFixtureRoots() error = %v, want duplicate fixture id", err)
	}
}

func TestValidateFixtureRootsAcceptsDistinctIDs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	first := filepath.Join(root, "first")
	second := filepath.Join(root, "second")
	writeConformanceFixtureFile(t, first, `[
		{"id":"first-id","source":"manual","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeConformanceFixtureFile(t, second, `[
		{"id":"second-id","source":"manual","locale":"en-US","options":{},"input":2,"expected":"2"}
	]`)

	if err := ValidateFixtureRoots([]string{first, second}, time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("ValidateFixtureRoots() error = %v, want nil", err)
	}
}

func writeCoverageFixtureFile(t *testing.T, packageDir, rel, data string) {
	t.Helper()

	path := filepath.Join(packageDir, "testdata", "conformance", rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatal(err)
	}
}

func writeDivergenceFile(t *testing.T, packageDir, data string) {
	t.Helper()

	path := filepath.Join(packageDir, "testdata", "divergences.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatal(err)
	}
}
