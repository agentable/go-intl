package conformance

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var (
	errMissingSkipListField    = errors.New("missing skip-list field")
	errInvalidSkipListCategory = errors.New("invalid skip-list category")
	errInvalidSkipListRoute    = errors.New("invalid skip-list route")
	errDuplicateSkipListSource = errors.New("duplicate skip-list source")
	errMissingSkipListWitness  = errors.New("missing skip-list witness")
	errUnknownSkipListWitness  = errors.New("unknown skip-list witness")
	errInvalidSkipListWitness  = errors.New("invalid skip-list witness")
)

type skipListEntry struct {
	Source       string `json:"source"`
	Category     string `json:"category"`
	Route        string `json:"route"`
	Reason       string `json:"reason"`
	Witness      string `json:"witness,omitempty"`
	DivergenceID string `json:"divergenceId,omitempty"`
}

type coverageReport struct {
	Packages []packageCoverage
	SkipList skipListCoverageReport
	Total    packageCoverage
}

type packageCoverage struct {
	Package     string
	Fixtures    int
	Sources     int
	Manual      int
	FormatJS    int
	Node        int
	Divergences int
	XFails      int
}

type skipListCoverageReport struct {
	Entries    int
	Categories map[string]int
	Routes     map[string]int
}

func loadSkipList(path string) ([]skipListEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []skipListEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return entries, nil
}

func ValidateSkipList(path string, packageRoots []string) error {
	entries, err := loadSkipList(path)
	if err != nil {
		return err
	}
	return validateSkipListEntries(path, entries, packageRoots)
}

func validateSkipListEntries(path string, entries []skipListEntry, packageRoots []string) error {
	witnesses, err := loadSkipListWitnesses(packageRoots)
	if err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(entries))
	for _, entry := range entries {
		if err := validateSkipListRequiredFields(path, entry); err != nil {
			return err
		}
		if _, ok := seen[entry.Source]; ok {
			return fmt.Errorf("%s: duplicate source %q: %w", path, entry.Source, errDuplicateSkipListSource)
		}
		seen[entry.Source] = struct{}{}
		if err := validateSkipListCategory(path, entry); err != nil {
			return err
		}
		if err := validateSkipListRoute(path, entry, witnesses); err != nil {
			return err
		}
	}
	return nil
}

func validateSkipListRequiredFields(path string, entry skipListEntry) error {
	if entry.Source == "" {
		return fmt.Errorf("%s: missing required field %q: %w", path, "source", errMissingSkipListField)
	}
	switch {
	case entry.Category == "":
		return fmt.Errorf("%s: source %q missing required field %q: %w", path, entry.Source, "category", errMissingSkipListField)
	case entry.Route == "":
		return fmt.Errorf("%s: source %q missing required field %q: %w", path, entry.Source, "route", errMissingSkipListField)
	case entry.Reason == "":
		return fmt.Errorf("%s: source %q missing required field %q: %w", path, entry.Source, "reason", errMissingSkipListField)
	default:
		return nil
	}
}

func validateSkipListCategory(path string, entry skipListEntry) error {
	if !isSkipListCategory(entry.Category) {
		return fmt.Errorf("%s: source %q category %q: %w", path, entry.Source, entry.Category, errInvalidSkipListCategory)
	}
	if entry.DivergenceID != "" {
		return fmt.Errorf("%s: source %q category %q has divergenceId %q: %w", path, entry.Source, entry.Category, entry.DivergenceID, errInvalidSkipListCategory)
	}
	return nil
}

func validateSkipListRoute(path string, entry skipListEntry, witnesses map[string]Fixture) error {
	if !isSkipListRoute(entry.Route) {
		return fmt.Errorf("%s: source %q route %q: %w", path, entry.Source, entry.Route, errInvalidSkipListRoute)
	}
	if entry.Route != "native-witness" {
		if entry.Witness != "" {
			return fmt.Errorf("%s: source %q route %q has witness %q: %w", path, entry.Source, entry.Route, entry.Witness, errInvalidSkipListWitness)
		}
		return nil
	}
	if entry.Witness == "" {
		return fmt.Errorf("%s: source %q route %q missing witness: %w", path, entry.Source, entry.Route, errMissingSkipListWitness)
	}
	witness, ok := witnesses[entry.Witness]
	if !ok {
		return fmt.Errorf("%s: source %q route %q witness %q does not match any fixture: %w", path, entry.Source, entry.Route, entry.Witness, errUnknownSkipListWitness)
	}
	if fixtureSourceKind(witness.Source) != "node" || !fixtureHasNativeExpectation(witness) {
		return fmt.Errorf("%s: source %q route %q witness %q is not an observable native fixture: %w", path, entry.Source, entry.Route, entry.Witness, errInvalidSkipListWitness)
	}
	return nil
}

func loadSkipListWitnesses(packageRoots []string) (map[string]Fixture, error) {
	witnesses := map[string]Fixture{}
	for _, root := range packageRoots {
		fixtures, err := LoadFixtures(root)
		if err != nil {
			return nil, err
		}
		for _, fixture := range fixtures {
			if fixtureSourceKind(fixture.Source) == "node" {
				witnesses[fixture.ID] = fixture
			}
		}
	}
	return witnesses, nil
}

func CoverageReport(packageRoots []string, skipListPath string) (string, error) {
	report, err := buildCoverageReport(packageRoots, skipListPath)
	if err != nil {
		return "", err
	}
	return formatCoverageReport(report), nil
}

func buildCoverageReport(packageRoots []string, skipListPath string) (coverageReport, error) {
	var report coverageReport
	for _, root := range packageRoots {
		coverage, err := buildPackageCoverage(root)
		if err != nil {
			return coverageReport{}, err
		}
		report.Packages = append(report.Packages, coverage)
		report.Total.add(coverage)
	}
	slices.SortFunc(report.Packages, func(a, b packageCoverage) int {
		return strings.Compare(a.Package, b.Package)
	})
	if skipListPath != "" {
		entries, err := loadSkipList(skipListPath)
		if err != nil {
			return coverageReport{}, err
		}
		report.SkipList = skipListCoverage(entries)
	}
	return report, nil
}

func formatCoverageReport(report coverageReport) string {
	var b strings.Builder
	fmt.Fprintf(
		&b,
		"conformance coverage: fixtures=%d sources=%d manual=%d formatjs=%d node=%d divergences=%d xfails=%d skipped=%d\n",
		report.Total.Fixtures,
		report.Total.Sources,
		report.Total.Manual,
		report.Total.FormatJS,
		report.Total.Node,
		report.Total.Divergences,
		report.Total.XFails,
		report.SkipList.Entries,
	)
	for _, coverage := range report.Packages {
		fmt.Fprintf(
			&b,
			"  %s: fixtures=%d sources=%d manual=%d formatjs=%d node=%d divergences=%d xfails=%d\n",
			coverage.Package,
			coverage.Fixtures,
			coverage.Sources,
			coverage.Manual,
			coverage.FormatJS,
			coverage.Node,
			coverage.Divergences,
			coverage.XFails,
		)
	}
	if report.SkipList.Entries == 0 {
		return b.String()
	}
	categories := slices.Sorted(maps.Keys(report.SkipList.Categories))
	b.WriteString("  skip-list:")
	for _, category := range categories {
		fmt.Fprintf(&b, " %s=%d", category, report.SkipList.Categories[category])
	}
	routes := slices.Sorted(maps.Keys(report.SkipList.Routes))
	b.WriteString(" routes")
	for _, route := range routes {
		fmt.Fprintf(&b, " %s=%d", route, report.SkipList.Routes[route])
	}
	b.WriteByte('\n')
	return b.String()
}

func buildPackageCoverage(root string) (packageCoverage, error) {
	fixtures, err := LoadFixtures(root)
	if err != nil {
		return packageCoverage{}, err
	}
	coverage := packageCoverage{Package: filepath.Base(filepath.Clean(root)), Fixtures: len(fixtures)}
	sources := make(map[string]struct{}, len(fixtures))
	for _, fixture := range fixtures {
		sources[fixture.Source] = struct{}{}
		switch fixtureSourceKind(fixture.Source) {
		case "manual":
			coverage.Manual++
		case "formatjs":
			coverage.FormatJS++
		case "node":
			coverage.Node++
		}
	}
	coverage.Sources = len(sources)
	divergences, err := loadDivergenceIDs(filepath.Join(root, "testdata", "divergences.md"))
	if err != nil {
		return packageCoverage{}, err
	}
	coverage.Divergences = len(divergences)
	xfails, err := loadXFails(filepath.Join(root, "testdata", "xfail.json"))
	if err != nil {
		return packageCoverage{}, err
	}
	coverage.XFails = len(xfails)
	return coverage, nil
}

func (p *packageCoverage) add(other packageCoverage) {
	p.Fixtures += other.Fixtures
	p.Sources += other.Sources
	p.Manual += other.Manual
	p.FormatJS += other.FormatJS
	p.Node += other.Node
	p.Divergences += other.Divergences
	p.XFails += other.XFails
}

func skipListCoverage(entries []skipListEntry) skipListCoverageReport {
	coverage := skipListCoverageReport{
		Entries:    len(entries),
		Categories: make(map[string]int, len(entries)),
		Routes:     make(map[string]int, len(entries)),
	}
	for _, entry := range entries {
		coverage.Categories[entry.Category]++
		coverage.Routes[entry.Route]++
	}
	return coverage
}

func isSkipListCategory(category string) bool {
	switch category {
	case "unsupported-extractor-shape",
		"partial-extraction":
		return true
	default:
		return false
	}
}

func isSkipListRoute(route string) bool {
	switch route {
	case "extractor",
		"native-witness",
		"not-applicable":
		return true
	default:
		return false
	}
}
