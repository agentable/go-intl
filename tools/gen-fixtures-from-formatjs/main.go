package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/agentable/go-intl/tools/conformance"
)

const (
	fixtureTestdataDir    = "testdata"
	fixtureConformanceDir = "conformance"
	fixtureFormatJSDir    = "formatjs"
	fixtureNativeDir      = "native"

	nodeFixtureDirPrefix    = "node-v"
	nodeSupportedValuesFile = "supported-values.json"

	repositorySkipListFile = ".skip-list.json"
)

type fixture struct {
	ID                 string                      `json:"id"`
	Source             string                      `json:"source"`
	Locale             string                      `json:"locale"`
	Feature            string                      `json:"feature,omitempty"`
	Options            map[string]any              `json:"options"`
	Input              any                         `json:"input"`
	Expected           *string                     `json:"expected,omitempty"`
	ExpectedOK         *bool                       `json:"expectedOk,omitempty"`
	ExpectedLocales    []string                    `json:"expectedLocales,omitempty"`
	ExpectedParts      []conformance.Part          `json:"expectedParts,omitempty"`
	ExpectedRange      *string                     `json:"expectedRange,omitempty"`
	ExpectedRangeParts []conformance.RangePart     `json:"expectedRangeParts,omitempty"`
	ExpectedComparison *int                        `json:"expectedComparison,omitempty"`
	ExpectedResolved   any                         `json:"expectedResolvedOptions,omitempty"`
	ExpectedSegments   []conformance.SegmentRecord `json:"expectedSegments,omitempty"`
	ErrorCode          string                      `json:"errorCode,omitempty"`
}

const (
	fixtureFeatureCanonicalize       = "canonicalize"
	fixtureFeatureMaximize           = "maximize"
	fixtureFeatureMinimize           = "minimize"
	fixtureFeatureFormatToParts      = "formatToParts"
	fixtureFeatureFormatRange        = "formatRange"
	fixtureFeatureFormatRangeToParts = "formatRangeToParts"
	fixtureFeatureSelectRange        = "selectRange"
)

type skipEntry struct {
	Source       string `json:"source"`
	Category     string `json:"category"`
	Route        string `json:"route"`
	Reason       string `json:"reason"`
	DivergenceID string `json:"divergenceId,omitempty"`
}

const (
	skipCategoryPartialExtraction         = "partial-extraction"
	skipCategoryUnsupportedExtractorShape = "unsupported-extractor-shape"

	skipRouteExtractor = "extractor"

	formatJSTestsPathNotFound = "tests path not found"

	skipReasonFormatJSPartialExtraction       = "mechanical assertions outside current generated fixture gate"
	skipReasonUnsupportedVitestAssertionShape = "unsupported Vitest assertion shape"
)

func formatJSImportPartialExtractionSkip(source string) skipEntry {
	return skipEntry{Source: source, Category: skipCategoryPartialExtraction, Route: skipRouteExtractor, Reason: skipReasonFormatJSPartialExtraction}
}

func formatJSUnsupportedExtractorShapeSkip(source string) skipEntry {
	return skipEntry{Source: source, Category: skipCategoryUnsupportedExtractorShape, Route: skipRouteExtractor, Reason: skipReasonUnsupportedVitestAssertionShape}
}

type nodeWitness struct {
	NodeVersion            string              `json:"nodeVersion"`
	Versions               map[string]string   `json:"versions"`
	LocaleSmoke            []fixture           `json:"localeSmoke"`
	LocaleCanonicalization []fixture           `json:"localeCanonicalization"`
	LocaleInfo             []fixture           `json:"localeInfo"`
	NumberFormatSmoke      []fixture           `json:"numberFormatSmoke"`
	NumberFormatErrors     []fixture           `json:"numberFormatErrors"`
	NumberFormatEdge       []fixture           `json:"numberFormatEdge"`
	NumberFormatResolved   []fixture           `json:"numberFormatResolved"`
	DateTimeFormatSmoke    []fixture           `json:"dateTimeFormatSmoke"`
	DateTimeFormatErrors   []fixture           `json:"dateTimeFormatErrors"`
	DateTimeFormatEdge     []fixture           `json:"dateTimeFormatEdge"`
	DurationFormatSmoke    []fixture           `json:"durationFormatSmoke"`
	DurationFormatErrors   []fixture           `json:"durationFormatErrors"`
	DurationFormatDigital  []fixture           `json:"durationFormatDigital"`
	ListFormatSmoke        []fixture           `json:"listFormatSmoke"`
	ListFormatErrors       []fixture           `json:"listFormatErrors"`
	RelativeTimeSmoke      []fixture           `json:"relativeTimeSmoke"`
	RelativeTimeErrors     []fixture           `json:"relativeTimeErrors"`
	PluralRulesSmoke       []fixture           `json:"pluralRulesSmoke"`
	PluralRulesErrors      []fixture           `json:"pluralRulesErrors"`
	DisplayNamesSmoke      []fixture           `json:"displayNamesSmoke"`
	DisplayNamesErrors     []fixture           `json:"displayNamesErrors"`
	CollatorSmoke          []fixture           `json:"collatorSmoke"`
	CollatorErrors         []fixture           `json:"collatorErrors"`
	CollatorOptions        []fixture           `json:"collatorOptions"`
	CollatorBackendProof   []fixture           `json:"collatorBackendProof"`
	SegmenterSmoke         []fixture           `json:"segmenterSmoke"`
	SegmenterErrors        []fixture           `json:"segmenterErrors"`
	SegmenterLocale        []fixture           `json:"segmenterLocale"`
	SegmenterTailored      []fixture           `json:"segmenterTailored"`
	SupportedValues        nodeSupportedValues `json:"supportedValues"`
}

type nodeSupportedValues struct {
	Source   string              `json:"source"`
	Versions map[string]string   `json:"versions"`
	Values   map[string][]string `json:"values"`
}

const (
	formatJSNumberFormatPackageDir       = "intl-numberformat"
	formatJSPluralRulesPackageDir        = "intl-pluralrules"
	formatJSDateTimeFormatPackageDir     = "intl-datetimeformat"
	formatJSLocalePackageDir             = "intl-locale"
	formatJSCanonicalLocalesPackageDir   = "intl-getcanonicallocales"
	formatJSListFormatPackageDir         = "intl-listformat"
	formatJSRelativeTimeFormatPackageDir = "intl-relativetimeformat"
	formatJSDurationFormatPackageDir     = "intl-durationformat"
)

const (
	formatJSNumberFormatTargetPackage       = "numberformat"
	formatJSPluralRulesTargetPackage        = "pluralrules"
	formatJSDateTimeFormatTargetPackage     = "datetimeformat"
	formatJSListFormatTargetPackage         = "listformat"
	formatJSRelativeTimeFormatTargetPackage = "relativetimeformat"
	formatJSDurationFormatTargetPackage     = "durationformat"
)

const (
	formatJSNumberFormatTestSourcePrefix       = "formatjs:packages/" + formatJSNumberFormatPackageDir + "/tests/"
	formatJSPluralRulesTestSourcePrefix        = "formatjs:packages/" + formatJSPluralRulesPackageDir + "/tests/"
	formatJSDateTimeFormatTestSourcePrefix     = "formatjs:packages/" + formatJSDateTimeFormatPackageDir + "/tests/"
	formatJSLocaleTestSourcePrefix             = "formatjs:packages/" + formatJSLocalePackageDir + "/tests/"
	formatJSCanonicalLocalesTestSourcePrefix   = "formatjs:packages/" + formatJSCanonicalLocalesPackageDir + "/tests/"
	formatJSListFormatTestSourcePrefix         = "formatjs:packages/" + formatJSListFormatPackageDir + "/tests/"
	formatJSRelativeTimeFormatTestSourcePrefix = "formatjs:packages/" + formatJSRelativeTimeFormatPackageDir + "/tests/"
	formatJSDurationFormatTestSourcePrefix     = "formatjs:packages/" + formatJSDurationFormatPackageDir + "/tests/"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("gen-fixtures-from-formatjs", flag.ContinueOnError)
	formatjsPath := fs.String("formatjs", "", "FormatJS checkout path")
	nodePath := fs.String("node", "", "Node executable path for native Intl fixture generation")
	outDir := fs.String("out", "", "output repository root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *outDir == "" {
		return fmt.Errorf("missing -out")
	}
	if *nodePath != "" {
		if err := importNode(*nodePath, *outDir); err != nil {
			return err
		}
	}
	if *formatjsPath != "" {
		formatJSSkips, err := importFormatJS(*formatjsPath, *outDir)
		if err != nil {
			return err
		}
		return writeSkips(*outDir, formatJSSkips)
	}
	if *nodePath != "" {
		return nil
	}
	return fmt.Errorf("missing -formatjs or -node")
}

func writeSkips(outDir string, skips []skipEntry) error {
	slices.SortFunc(skips, func(a, b skipEntry) int {
		if a.Source != b.Source {
			return strings.Compare(a.Source, b.Source)
		}
		if a.Category != b.Category {
			return strings.Compare(a.Category, b.Category)
		}
		return strings.Compare(a.Reason, b.Reason)
	})
	return writeJSON(skipListPath(outDir), skips)
}

func skipListPath(outDir string) string {
	return filepath.Join(outDir, repositorySkipListFile)
}

func importNode(path, outDir string) error {
	witness, err := runNodeWitness(path)
	if err != nil {
		return err
	}
	nodeDir, err := nodeFixtureDir(witness.NodeVersion)
	if err != nil {
		return err
	}
	for _, file := range witnessFixtureFiles(witness, nodeDir) {
		if len(file.Fixtures) == 0 {
			continue
		}
		path := repositoryOutputPath(outDir, file.Path)
		if err := writeJSON(path, file.Fixtures); err != nil {
			return err
		}
	}
	if len(witness.SupportedValues.Values) != 0 {
		path := repositoryOutputPath(outDir, nodeSupportedValuesPath(nodeDir))
		if err := writeJSON(path, witness.SupportedValues); err != nil {
			return err
		}
	}
	return nil
}

func repositoryOutputPath(outDir string, rel []string) string {
	return filepath.Join(slices.Concat([]string{outDir}, rel)...)
}

type nodeWitnessFixtureFile struct {
	Path     []string
	Fixtures []fixture
}

func witnessFixtureFiles(witness nodeWitness, nodeDir string) []nodeWitnessFixtureFile {
	return []nodeWitnessFixtureFile{
		nodeConformanceFixtureFile("locale", nodeDir, "smoke.json", witness.LocaleSmoke),
		nodeConformanceFixtureFile("locale", nodeDir, "canonicalization.json", witness.LocaleCanonicalization),
		nodeConformanceFixtureFile("locale", nodeDir, "info.json", witness.LocaleInfo),
		nodeConformanceFixtureFile("numberformat", nodeDir, "smoke.json", witness.NumberFormatSmoke),
		nodeConformanceFixtureFile("numberformat", nodeDir, "errors.json", witness.NumberFormatErrors),
		nodeConformanceFixtureFile("numberformat", nodeDir, "edge.json", witness.NumberFormatEdge),
		nodeConformanceFixtureFile("numberformat", nodeDir, "resolved-options.json", witness.NumberFormatResolved),
		nodeConformanceFixtureFile("datetimeformat", nodeDir, "smoke.json", witness.DateTimeFormatSmoke),
		nodeConformanceFixtureFile("datetimeformat", nodeDir, "errors.json", witness.DateTimeFormatErrors),
		nodeConformanceFixtureFile("datetimeformat", nodeDir, "edge.json", witness.DateTimeFormatEdge),
		nodeConformanceFixtureFile("durationformat", nodeDir, "smoke.json", witness.DurationFormatSmoke),
		nodeConformanceFixtureFile("durationformat", nodeDir, "errors.json", witness.DurationFormatErrors),
		nodeConformanceFixtureFile("durationformat", nodeDir, "digital.json", witness.DurationFormatDigital),
		nodeConformanceFixtureFile("listformat", nodeDir, "smoke.json", witness.ListFormatSmoke),
		nodeConformanceFixtureFile("listformat", nodeDir, "errors.json", witness.ListFormatErrors),
		nodeConformanceFixtureFile("relativetimeformat", nodeDir, "smoke.json", witness.RelativeTimeSmoke),
		nodeConformanceFixtureFile("relativetimeformat", nodeDir, "errors.json", witness.RelativeTimeErrors),
		nodeConformanceFixtureFile("pluralrules", nodeDir, "smoke.json", witness.PluralRulesSmoke),
		nodeConformanceFixtureFile("pluralrules", nodeDir, "errors.json", witness.PluralRulesErrors),
		nodeConformanceFixtureFile("displaynames", nodeDir, "smoke.json", witness.DisplayNamesSmoke),
		nodeConformanceFixtureFile("displaynames", nodeDir, "errors.json", witness.DisplayNamesErrors),
		nodeConformanceFixtureFile("collator", nodeDir, "smoke.json", witness.CollatorSmoke),
		nodeConformanceFixtureFile("collator", nodeDir, "errors.json", witness.CollatorErrors),
		nodeConformanceFixtureFile("collator", nodeDir, "options.json", witness.CollatorOptions),
		nodeConformanceFixtureFile("collator", nodeDir, "backend-proof.json", witness.CollatorBackendProof),
		nodeConformanceFixtureFile("segmenter", nodeDir, "smoke.json", witness.SegmenterSmoke),
		nodeConformanceFixtureFile("segmenter", nodeDir, "errors.json", witness.SegmenterErrors),
		nodeConformanceFixtureFile("segmenter", nodeDir, "locale-contract.json", witness.SegmenterLocale),
		nodeConformanceFixtureFile("segmenter", nodeDir, "tailored-locale-contract.json", witness.SegmenterTailored),
	}
}

func nodeConformanceFixtureFile(packageName, nodeDir, fileName string, fixtures []fixture) nodeWitnessFixtureFile {
	return nodeWitnessFixtureFile{
		Path:     nodeConformanceFixturePath(packageName, nodeDir, fileName),
		Fixtures: fixtures,
	}
}

func nodeConformanceFixturePath(packageName, nodeDir, fileName string) []string {
	return []string{packageName, fixtureTestdataDir, fixtureConformanceDir, nodeDir, fileName}
}

func nodeSupportedValuesPath(nodeDir string) []string {
	return []string{fixtureTestdataDir, fixtureNativeDir, nodeDir, nodeSupportedValuesFile}
}

func runNodeWitness(nodePath string) (nodeWitness, error) {
	repoRoot, err := repoRoot()
	if err != nil {
		return nodeWitness{}, err
	}
	cmd := exec.Command("go", "run", "./tools/node-witness", "-node", nodePath)
	cmd.Dir = repoRoot
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nodeWitness{}, fmt.Errorf("run node witness %q: %v: %s", nodePath, err, strings.TrimSpace(stderr.String()))
	}
	var witness nodeWitness
	if err := json.Unmarshal(stdout.Bytes(), &witness); err != nil {
		return nodeWitness{}, fmt.Errorf("decode node witness output: %w", err)
	}
	if witness.NodeVersion == "" {
		return nodeWitness{}, fmt.Errorf("node witness output missing nodeVersion")
	}
	return witness, nil
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "tools", "node-witness")
		if stat, err := os.Stat(filepath.Join(candidate, "main.go")); err == nil && !stat.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("tools/node-witness not found from %s", dir)
		}
		dir = parent
	}
}

func nodeFixtureDir(version string) (string, error) {
	trimmed := strings.TrimPrefix(version, "v")
	major, _, ok := strings.Cut(trimmed, ".")
	if !ok || major == "" {
		return "", fmt.Errorf("invalid Node version %q", version)
	}
	if _, err := strconv.Atoi(major); err != nil {
		return "", fmt.Errorf("invalid Node major version %q: %w", version, err)
	}
	return nodeFixtureDirPrefix + major, nil
}

func importFormatJS(path, outDir string) ([]skipEntry, error) {
	skips := []skipEntry{}
	if err := appendFormatJSSurfaceSkips(&skips, path, outDir, formatJSPreLocaleSurfaceRoutes()); err != nil {
		return nil, err
	}
	localeSkips, err := importFormatJSLocales(path, outDir)
	if err != nil {
		return nil, err
	}
	skips = append(skips, localeSkips...)
	if err := appendFormatJSSurfaceSkips(&skips, path, outDir, formatJSPostLocaleSurfaceRoutes()); err != nil {
		return nil, err
	}
	return skips, nil
}

type formatJSImportSpec struct {
	packageDir string
	extract    func(string, string) []fixture
	supports   func(fixture) bool
	slug       func(string) string
}

func (s formatJSImportSpec) sourcePrefix() string {
	return formatJSTestSourcePrefix(s.packageDir)
}

type formatJSSurfaceRoute struct {
	targetPackage string
	spec          formatJSImportSpec
}

func formatJSSurfaceRouteFor(
	targetPackage, packageDir string,
	extract func(string, string) []fixture,
	supports func(fixture) bool,
) formatJSSurfaceRoute {
	return formatJSSurfaceRoute{
		targetPackage: targetPackage,
		spec: formatJSImportSpec{
			packageDir: packageDir,
			extract:    extract,
			supports:   supports,
			slug:       fixtureSlug,
		},
	}
}

func formatJSNumberFormatRoute() formatJSSurfaceRoute {
	return formatJSSurfaceRouteFor(
		formatJSNumberFormatTargetPackage,
		formatJSNumberFormatPackageDir,
		extractNumberFormatFixtures,
		supportsGeneratedNumberFormatFixture,
	)
}

func formatJSPluralRulesRoute() formatJSSurfaceRoute {
	return formatJSSurfaceRouteFor(
		formatJSPluralRulesTargetPackage,
		formatJSPluralRulesPackageDir,
		extractPluralRulesFixtures,
		supportsGeneratedPluralRulesFixture,
	)
}

func formatJSDateTimeFormatRoute() formatJSSurfaceRoute {
	return formatJSSurfaceRouteFor(
		formatJSDateTimeFormatTargetPackage,
		formatJSDateTimeFormatPackageDir,
		extractDateTimeFormatFixtures,
		supportsGeneratedDateTimeFormatFixture,
	)
}

func formatJSListFormatRoute() formatJSSurfaceRoute {
	return formatJSSurfaceRouteFor(
		formatJSListFormatTargetPackage,
		formatJSListFormatPackageDir,
		extractListFormatFixtures,
		supportsGeneratedListFormatFixture,
	)
}

func formatJSRelativeTimeFormatRoute() formatJSSurfaceRoute {
	return formatJSSurfaceRouteFor(
		formatJSRelativeTimeFormatTargetPackage,
		formatJSRelativeTimeFormatPackageDir,
		extractRelativeTimeFormatFixtures,
		supportsGeneratedRelativeTimeFormatFixture,
	)
}

func formatJSDurationFormatRoute() formatJSSurfaceRoute {
	return formatJSSurfaceRouteFor(
		formatJSDurationFormatTargetPackage,
		formatJSDurationFormatPackageDir,
		extractDurationFormatFixtures,
		supportsGeneratedDurationFormatFixture,
	)
}

func formatJSPreLocaleSurfaceRoutes() []formatJSSurfaceRoute {
	return []formatJSSurfaceRoute{
		formatJSNumberFormatRoute(),
		formatJSPluralRulesRoute(),
		formatJSDateTimeFormatRoute(),
	}
}

func formatJSPostLocaleSurfaceRoutes() []formatJSSurfaceRoute {
	return []formatJSSurfaceRoute{
		formatJSListFormatRoute(),
		formatJSRelativeTimeFormatRoute(),
		formatJSDurationFormatRoute(),
	}
}

func importFormatJSNumberFormat(path, outDir string) ([]skipEntry, error) {
	return importFormatJSSurface(path, outDir, formatJSNumberFormatRoute())
}

func importFormatJSPluralRules(path, outDir string) ([]skipEntry, error) {
	return importFormatJSSurface(path, outDir, formatJSPluralRulesRoute())
}

func importFormatJSDateTimeFormat(path, outDir string) ([]skipEntry, error) {
	return importFormatJSSurface(path, outDir, formatJSDateTimeFormatRoute())
}

func importFormatJSLocales(path, outDir string) ([]skipEntry, error) {
	targetRoot := formatJSConformanceRoot(outDir, "locale")
	if err := os.RemoveAll(targetRoot); err != nil {
		return nil, err
	}
	skips := []skipEntry{}
	localeSkips, err := importFormatJSLocalePackage(path, targetRoot)
	if err != nil {
		return nil, err
	}
	skips = append(skips, localeSkips...)
	canonicalSkips, err := importFormatJSCanonicalLocalesPackage(path, targetRoot)
	if err != nil {
		return nil, err
	}
	skips = append(skips, canonicalSkips...)
	return skips, nil
}

func importFormatJSLocalePackage(path, targetRoot string) ([]skipEntry, error) {
	return importFormatJSPackageTests(path, targetRoot, formatJSImportSpec{
		packageDir: formatJSLocalePackageDir,
		extract:    extractLocaleFixtures,
		supports:   supportsGeneratedLocaleFixture,
		slug:       fixtureSlug,
	})
}

func importFormatJSCanonicalLocalesPackage(path, targetRoot string) ([]skipEntry, error) {
	slug := func(rel string) string {
		return "intl-getcanonicallocales-" + fixtureSlug(rel)
	}
	return importFormatJSPackageTests(path, targetRoot, formatJSImportSpec{
		packageDir: formatJSCanonicalLocalesPackageDir,
		extract:    extractCanonicalLocalesFixtures,
		supports:   supportsGeneratedLocaleFixture,
		slug:       slug,
	})
}

func importFormatJSListFormat(path, outDir string) ([]skipEntry, error) {
	return importFormatJSSurface(path, outDir, formatJSListFormatRoute())
}

func importFormatJSRelativeTimeFormat(path, outDir string) ([]skipEntry, error) {
	return importFormatJSSurface(path, outDir, formatJSRelativeTimeFormatRoute())
}

func importFormatJSDurationFormat(path, outDir string) ([]skipEntry, error) {
	return importFormatJSSurface(path, outDir, formatJSDurationFormatRoute())
}

func appendFormatJSSurfaceSkips(skips *[]skipEntry, path, outDir string, routes []formatJSSurfaceRoute) error {
	for _, route := range routes {
		routeSkips, err := importFormatJSSurface(path, outDir, route)
		if err != nil {
			return err
		}
		*skips = append(*skips, routeSkips...)
	}
	return nil
}

func importFormatJSSurface(path, outDir string, route formatJSSurfaceRoute) ([]skipEntry, error) {
	targetRoot := formatJSConformanceRoot(outDir, route.targetPackage)
	if err := os.RemoveAll(targetRoot); err != nil {
		return nil, err
	}
	return importFormatJSPackageTests(path, targetRoot, route.spec)
}

func formatJSConformanceRoot(outDir, packageName string) string {
	return filepath.Join(outDir, packageName, fixtureTestdataDir, fixtureConformanceDir, fixtureFormatJSDir)
}

func importFormatJSPackageTests(path, targetRoot string, spec formatJSImportSpec) ([]skipEntry, error) {
	root := formatJSPackageTestsRoot(path, spec.packageDir)
	if stat, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%s: %s: %s", spec.sourcePrefix(), root, formatJSTestsPathNotFound)
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	return importFormatJSTestFixtures(root, targetRoot, spec)
}

func formatJSPackageTestsRoot(formatJSRoot, packageDir string) string {
	return filepath.Join(formatJSRoot, "packages", packageDir, "tests")
}

func formatJSTestSourcePrefix(packageDir string) string {
	return "formatjs:packages/" + packageDir + "/tests/"
}

func importFormatJSTestFixtures(root, targetRoot string, spec formatJSImportSpec) ([]skipEntry, error) {
	fixtures := []fixture{}
	skips := []skipEntry{}
	sourcePrefix := spec.sourcePrefix()
	if err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		switch filepath.Ext(path) {
		case ".snap":
			return nil
		case ".ts":
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			extracted := spec.extract(rel, string(data))
			supportedCount := 0
			for _, fixture := range extracted {
				if spec.supports(fixture) {
					fixtures = append(fixtures, fixture)
					supportedCount++
				}
			}
			if len(extracted) > 0 && (supportedCount < len(extracted) || strings.Count(string(data), "expect(") > len(extracted)) {
				skips = append(skips, formatJSImportPartialExtractionSkip(sourcePrefix+rel))
			}
			if len(extracted) == 0 && strings.Contains(string(data), "expect(") {
				skips = append(skips, formatJSUnsupportedExtractorShapeSkip(sourcePrefix+rel))
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := writeFixturesBySourceSlug(targetRoot, fixtures, sourcePrefix, spec.slug); err != nil {
		return nil, err
	}
	return skips, nil
}

func writeFixturesBySourceSlug(targetRoot string, fixtures []fixture, sourcePrefix string, slug func(string) string) error {
	if len(fixtures) == 0 {
		return nil
	}
	slices.SortFunc(fixtures, func(a, b fixture) int {
		return strings.Compare(a.ID, b.ID)
	})
	fixturesBySource := map[string][]fixture{}
	for _, fixture := range fixtures {
		fixturesBySource[fixture.Source] = append(fixturesBySource[fixture.Source], fixture)
	}
	sources := slices.Sorted(maps.Keys(fixturesBySource))
	for _, source := range sources {
		rel := strings.TrimPrefix(source, sourcePrefix)
		if err := writeJSON(formatJSFixtureFile(targetRoot, rel, slug), fixturesBySource[source]); err != nil {
			return err
		}
	}
	return nil
}

func formatJSFixtureFile(targetRoot, rel string, slug func(string) string) string {
	return filepath.Join(targetRoot, slug(rel)+".json")
}

func supportsGeneratedNumberFormatFixture(f fixture) bool {
	if f.Locale != "en" {
		return false
	}
	switch f.Feature {
	case "", fixtureFeatureFormatToParts, fixtureFeatureFormatRange, fixtureFeatureFormatRangeToParts:
	default:
		return false
	}
	style := "decimal"
	currency := ""
	for key, value := range f.Options {
		switch key {
		case "style":
			styleValue, ok := stringValueOneOf(value, "currency", "percent", "decimal")
			if !ok {
				return false
			}
			style = styleValue
		case "currency":
			currencyValue, ok := stringValueOneOf(value, "USD")
			if !ok {
				return false
			}
			currency = currencyValue
		case "minimumFractionDigits", "maximumFractionDigits":
			if _, ok := value.(int64); !ok {
				return false
			}
		case "useGrouping":
			switch value := value.(type) {
			case string:
				if value != "always" && value != "auto" && value != "min2" {
					return false
				}
			case bool:
				if value {
					return false
				}
			default:
				return false
			}
		default:
			return false
		}
	}
	if style == "currency" && currency != "USD" {
		return false
	}
	return true
}

type formatJSConstructorDeclaration struct {
	index   int
	locale  string
	options map[string]any
}

var (
	numberFormatDeclarationRE     = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*(?:new\s+)?(?:Intl\.)?NumberFormat\s*\(\s*(?:['"]([^'"]*)['"]\s*)?(?:,\s*(\{.*?\}))?\s*\)`)
	inlineFormatExpectationRE     = regexp.MustCompile(`(?s)expect\s*\(\s*(?:new\s+)?(?:Intl\.)?NumberFormat\s*\(\s*['"]([^'"]*)['"]\s*(?:,\s*(\{.*?\}))?\s*\)\.format\s*\(\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	varFormatExpectationRE        = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.format\s*\(\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	varFormatToPartsExpectationRE = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.formatToParts\s*\(\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Equal|StrictEqual)\s*\(\s*(\[[^\]]*\])\s*\)`)
	varFormatRangeExpectationRE   = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.formatRange\s*\(\s*([^,]+?)\s*,\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	varRangePartsExpectationRE    = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.formatRangeToParts\s*\(\s*([^,]+?)\s*,\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Equal|StrictEqual)\s*\(\s*(\[[^\]]*\])\s*\)`)
	partObjectRE                  = regexp.MustCompile(`\{([^{}]*)\}`)
	stringOptionRE                = regexp.MustCompile(`([A-Za-z][A-Za-z0-9]*)\s*:\s*['"]((?:\\.|[^\\'"])*?)['"]`)
	numberOptionRE                = regexp.MustCompile(`([A-Za-z][A-Za-z0-9]*)\s*:\s*(-?\d+(?:_\d+)*(?:\.\d+)?(?:e[+-]?\d+)?)`)
	boolOptionRE                  = regexp.MustCompile(`([A-Za-z][A-Za-z0-9]*)\s*:\s*(true|false)`)
)

func extractNumberFormatFixtures(rel, data string) []fixture {
	declarations := numberFormatDeclarations(data)
	fixtures := []fixture{}
	nextIndex := 0
	for _, match := range inlineFormatExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		locale := data[match[2]:match[3]]
		options := parseOptionsObject(matchString(data, match, 4))
		input, ok := parseNumberLiteral(data[match[6]:match[7]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newNumberFormatFixture(rel, nextIndex, locale, options, input, expected))
		nextIndex++
	}

	variableMatches := varFormatExpectationRE.FindAllStringSubmatchIndex(data, -1)
	for _, match := range variableMatches {
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		input, ok := parseNumberLiteral(data[match[4]:match[5]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[6]:match[7]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newNumberFormatFixture(rel, nextIndex, decl.locale, decl.options, input, expected))
		nextIndex++
	}
	for _, match := range varFormatToPartsExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		input, ok := parseNumberLiteral(data[match[4]:match[5]])
		if !ok {
			continue
		}
		parts, ok := parsePartArray(data[match[6]:match[7]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newNumberFormatPartsFixture(rel, nextIndex, decl.locale, decl.options, input, parts))
		nextIndex++
	}
	for _, match := range varFormatRangeExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		start, ok := parseNumberLiteral(data[match[4]:match[5]])
		if !ok {
			continue
		}
		end, ok := parseNumberLiteral(data[match[6]:match[7]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newNumberFormatRangeFixture(rel, nextIndex, decl.locale, decl.options, start, end, expected))
		nextIndex++
	}
	for _, match := range varRangePartsExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		start, ok := parseNumberLiteral(data[match[4]:match[5]])
		if !ok {
			continue
		}
		end, ok := parseNumberLiteral(data[match[6]:match[7]])
		if !ok {
			continue
		}
		parts, ok := parseRangePartArray(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newNumberFormatRangePartsFixture(rel, nextIndex, decl.locale, decl.options, start, end, parts))
		nextIndex++
	}
	return fixtures
}

func numberFormatDeclarations(data string) map[string][]formatJSConstructorDeclaration {
	return formatJSConstructorDeclarations(data, numberFormatDeclarationRE, 6, 4)
}

func newNumberFormatFixture(rel string, index int, locale string, options map[string]any, input any, expected string) fixture {
	return formatJSExpectedFixture(formatJSNumberFormatTargetPackage, formatJSNumberFormatTestSourcePrefix, rel, index, locale, options, input, expected)
}

func newNumberFormatPartsFixture(rel string, index int, locale string, options map[string]any, input any, parts []conformance.Part) fixture {
	f := formatJSSurfaceFixture(formatJSNumberFormatTargetPackage, formatJSNumberFormatTestSourcePrefix, rel, index, locale, options, input)
	f.Feature = fixtureFeatureFormatToParts
	f.Expected = ptr(joinPartValues(parts))
	f.ExpectedParts = parts
	return f
}

func newNumberFormatRangeFixture(rel string, index int, locale string, options map[string]any, start, end any, expected string) fixture {
	return formatJSRangeFixture(formatJSNumberFormatTargetPackage, formatJSNumberFormatTestSourcePrefix, rel, index, locale, options, start, end, expected)
}

func newNumberFormatRangePartsFixture(rel string, index int, locale string, options map[string]any, start, end any, parts []conformance.RangePart) fixture {
	return formatJSRangePartsFixture(formatJSNumberFormatTargetPackage, formatJSNumberFormatTestSourcePrefix, rel, index, locale, options, start, end, parts)
}

var (
	pluralRulesDeclarationRE     = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*(?:new\s+)?(?:Intl\.)?PluralRules\s*\(\s*(?:['"]([^'"]*)['"]\s*)?(?:,\s*(\{.*?\}))?\s*\)`)
	inlinePluralSelectRE         = regexp.MustCompile(`(?s)expect\s*\(\s*(?:new\s+)?(?:Intl\.)?PluralRules\s*\(\s*['"]([^'"]*)['"]\s*(?:,\s*(\{.*?\}))?\s*\)\.select\s*\(\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	varPluralSelectExpectationRE = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.select\s*\(\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	inlinePluralRangeRE          = regexp.MustCompile(`(?s)expect\s*\(\s*(?:new\s+)?(?:Intl\.)?PluralRules\s*\(\s*['"]([^'"]*)['"]\s*(?:,\s*(\{.*?\}))?\s*\)\.selectRange\s*\(\s*((?:BigInt\s*\([^)]*\)|[^,])+?)\s*,\s*((?:BigInt\s*\([^)]*\)|[^)])+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	varPluralRangeExpectationRE  = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.selectRange\s*\(\s*((?:BigInt\s*\([^)]*\)|[^,])+?)\s*,\s*((?:BigInt\s*\([^)]*\)|[^)])+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
)

func extractPluralRulesFixtures(rel, data string) []fixture {
	declarations := pluralRulesDeclarations(data)
	fixtures := []fixture{}
	nextIndex := 0
	for _, match := range inlinePluralSelectRE.FindAllStringSubmatchIndex(data, -1) {
		locale := data[match[2]:match[3]]
		options := parseOptionsObject(matchString(data, match, 4))
		input, ok := parsePluralInputLiteral(data[match[6]:match[7]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newPluralRulesFixture(rel, nextIndex, locale, options, input, expected))
		nextIndex++
	}
	for _, match := range inlinePluralRangeRE.FindAllStringSubmatchIndex(data, -1) {
		locale := data[match[2]:match[3]]
		options := parseOptionsObject(matchString(data, match, 4))
		start, ok := parsePluralInputLiteral(data[match[6]:match[7]])
		if !ok {
			continue
		}
		end, ok := parsePluralInputLiteral(data[match[8]:match[9]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[10]:match[11]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newPluralRulesRangeFixture(rel, nextIndex, locale, options, start, end, expected))
		nextIndex++
	}

	for _, match := range varPluralSelectExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		input, ok := parsePluralInputLiteral(data[match[4]:match[5]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[6]:match[7]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newPluralRulesFixture(rel, nextIndex, decl.locale, decl.options, input, expected))
		nextIndex++
	}
	for _, match := range varPluralRangeExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		start, ok := parsePluralInputLiteral(data[match[4]:match[5]])
		if !ok {
			continue
		}
		end, ok := parsePluralInputLiteral(data[match[6]:match[7]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newPluralRulesRangeFixture(rel, nextIndex, decl.locale, decl.options, start, end, expected))
		nextIndex++
	}
	return fixtures
}

func pluralRulesDeclarations(data string) map[string][]formatJSConstructorDeclaration {
	return formatJSConstructorDeclarations(data, pluralRulesDeclarationRE, 6, 4)
}

func newPluralRulesFixture(rel string, index int, locale string, options map[string]any, input any, expected string) fixture {
	return formatJSExpectedFixture(formatJSPluralRulesTargetPackage, formatJSPluralRulesTestSourcePrefix, rel, index, locale, options, input, expected)
}

func newPluralRulesRangeFixture(rel string, index int, locale string, options map[string]any, start, end any, expected string) fixture {
	f := formatJSSurfaceFixture(formatJSPluralRulesTargetPackage, formatJSPluralRulesTestSourcePrefix, rel, index, locale, options, rangeFixtureInput(start, end))
	f.Feature = fixtureFeatureSelectRange
	f.Expected = ptr(expected)
	return f
}

func supportsGeneratedPluralRulesFixture(f fixture) bool {
	switch f.Locale {
	case "en", "en-US", "en-XX", "fr":
	default:
		return false
	}
	if f.Expected == nil || !isPluralCategory(*f.Expected) {
		return false
	}
	for key, value := range f.Options {
		switch key {
		case "type":
			if _, ok := stringValueOneOf(value, "cardinal", "ordinal"); !ok {
				return false
			}
		case "notation":
			if _, ok := stringValueOneOf(value, "standard", "scientific", "engineering", "compact"); !ok {
				return false
			}
		case "compactDisplay":
			if _, ok := stringValueOneOf(value, "short", "long"); !ok {
				return false
			}
		case "roundingMode":
			if _, ok := stringValueOneOf(value, "ceil", "floor", "expand", "trunc", "halfCeil", "halfFloor", "halfExpand", "halfTrunc", "halfEven"); !ok {
				return false
			}
		case "roundingPriority":
			if _, ok := stringValueOneOf(value, "auto", "morePrecision", "lessPrecision"); !ok {
				return false
			}
		case "trailingZeroDisplay":
			if _, ok := stringValueOneOf(value, "auto", "stripIfInteger"); !ok {
				return false
			}
		case "minimumIntegerDigits", "minimumFractionDigits", "maximumFractionDigits", "minimumSignificantDigits", "maximumSignificantDigits", "roundingIncrement":
			if _, ok := value.(int64); !ok {
				return false
			}
		default:
			return false
		}
	}
	return true
}

type dateVariable struct {
	index int
	value string
}

var (
	dateTimeDeclarationRE      = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*(?:new\s+)?(?:Intl\.)?DateTimeFormat\s*\(\s*(?:['"]([^'"]*)['"]|\[['"]([^'"]*)['"]\]\s*)?(?:,\s*(\{.*?\}))?\s*\)`)
	dateVarStringRE            = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*new\s+Date\s*\(\s*['"]([^'"]*)['"]\s*\)`)
	dateVarNumberRE            = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*new\s+Date\s*\(\s*(-?\d+)\s*\)`)
	dateVarYMDRE               = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*new\s+Date\s*\(\s*(-?\d+)\s*,\s*(\d+)\s*,\s*(\d+)(?:\s*,\s*(\d+))?(?:\s*,\s*(\d+))?(?:\s*,\s*(\d+))?(?:\s*,\s*(\d+))?\s*\)`)
	dateVarUTCRE               = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*new\s+Date\s*\(\s*Date\.UTC\s*\(([^)]*)\)\s*\)`)
	varDateTimeFormatRE        = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.format\s*\(\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	varDateTimeFormatRangeRE   = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.formatRange\s*\(\s*([^,]+?)\s*,\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	varDateTimeRangePartsRE    = regexp.MustCompile(`(?s)expect\s*\(\s*([A-Za-z_]\w*)\.formatRangeToParts\s*\(\s*([^,]+?)\s*,\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Equal|StrictEqual)\s*\(\s*(\[[^\]]*\])\s*\)`)
	inlineDateTimeRangePartsRE = regexp.MustCompile(`(?s)expect\s*\(\s*(?:new\s+)?(?:Intl\.)?DateTimeFormat\s*\(\s*['"]([^'"]*)['"]\s*(?:,\s*(\{.*?\}))?\s*\)\.formatRangeToParts\s*\(\s*([^,]+?)\s*,\s*([^)]+?)\s*\)\s*\)\s*\.to(?:Equal|StrictEqual)\s*\(\s*(\[[^\]]*\])\s*\)`)
	localeToStringRE           = regexp.MustCompile(`(?s)expect\s*\(\s*new\s+Locale\s*\(\s*['"]([^'"]*)['"]\s*\)\.toString\s*\(\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	localeMaximizeRE           = regexp.MustCompile(`(?s)expect\s*\(\s*new\s+Locale\s*\(\s*['"]([^'"]*)['"]\s*\)\.maximize\s*\(\s*\)\.toString\s*\(\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	localeMinimizeRE           = regexp.MustCompile(`(?s)expect\s*\(\s*new\s+Locale\s*\(\s*['"]([^'"]*)['"]\s*\)\.minimize\s*\(\s*\)\.toString\s*\(\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	canonicalLocalesRE         = regexp.MustCompile(`(?s)expect\s*\(\s*getCanonicalLocales\s*\(\s*['"]([^'"]*)['"]\s*\)\s*\)\s*\.toEqual\s*\(\s*\[\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\]\s*\)`)
)

func extractDateTimeFormatFixtures(rel, data string) []fixture {
	declarations := dateTimeDeclarations(data)
	dates := dateVariables(data)
	fixtures := []fixture{}
	nextIndex := 0
	for _, match := range varDateTimeFormatRE.FindAllStringSubmatchIndex(data, -1) {
		if isSkippedVitestAssertion(data, match[0]) {
			continue
		}
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		input, ok := parseDateExpression(data[match[4]:match[5]], dates, match[0])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[6]:match[7]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newDateTimeFormatFixture(rel, nextIndex, decl.locale, decl.options, input, expected))
		nextIndex++
	}
	for _, match := range varDateTimeFormatRangeRE.FindAllStringSubmatchIndex(data, -1) {
		if isSkippedVitestAssertion(data, match[0]) {
			continue
		}
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		start, ok := parseDateExpression(data[match[4]:match[5]], dates, match[0])
		if !ok {
			continue
		}
		end, ok := parseDateExpression(data[match[6]:match[7]], dates, match[0])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newDateTimeFormatRangeFixture(rel, nextIndex, decl.locale, decl.options, start, end, expected))
		nextIndex++
	}
	for _, match := range varDateTimeRangePartsRE.FindAllStringSubmatchIndex(data, -1) {
		if isSkippedVitestAssertion(data, match[0]) {
			continue
		}
		name := data[match[2]:match[3]]
		decl, ok := latestExplicitLocaleConstructorDeclarationBefore(declarations, name, match[0])
		if !ok {
			continue
		}
		start, ok := parseDateExpression(data[match[4]:match[5]], dates, match[0])
		if !ok {
			continue
		}
		end, ok := parseDateExpression(data[match[6]:match[7]], dates, match[0])
		if !ok {
			continue
		}
		parts, ok := parseRangePartArray(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newDateTimeFormatRangePartsFixture(rel, nextIndex, decl.locale, decl.options, start, end, parts))
		nextIndex++
	}
	for _, match := range inlineDateTimeRangePartsRE.FindAllStringSubmatchIndex(data, -1) {
		if isSkippedVitestAssertion(data, match[0]) {
			continue
		}
		locale := data[match[2]:match[3]]
		options := parseOptionsObject(matchString(data, match, 4))
		start, ok := parseDateExpression(data[match[6]:match[7]], dates, match[0])
		if !ok {
			continue
		}
		end, ok := parseDateExpression(data[match[8]:match[9]], dates, match[0])
		if !ok {
			continue
		}
		parts, ok := parseRangePartArray(data[match[10]:match[11]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newDateTimeFormatRangePartsFixture(rel, nextIndex, locale, options, start, end, parts))
		nextIndex++
	}
	return fixtures
}

func dateTimeDeclarations(data string) map[string][]formatJSConstructorDeclaration {
	return formatJSConstructorDeclarations(data, dateTimeDeclarationRE, 8, 4, 6)
}

func formatJSConstructorDeclarations(data string, re *regexp.Regexp, optionsGroup int, localeGroups ...int) map[string][]formatJSConstructorDeclaration {
	declarations := map[string][]formatJSConstructorDeclaration{}
	for _, match := range re.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		declarations[name] = append(declarations[name], formatJSConstructorDeclaration{
			index:   match[0],
			locale:  firstMatchString(data, match, localeGroups...),
			options: parseOptionsObject(matchString(data, match, optionsGroup)),
		})
	}
	return declarations
}

func firstMatchString(data string, match []int, indexes ...int) string {
	for _, index := range indexes {
		if value := matchString(data, match, index); value != "" {
			return value
		}
	}
	return ""
}

func latestConstructorDeclarationBefore(declarations []formatJSConstructorDeclaration, index int) (formatJSConstructorDeclaration, bool) {
	for i := len(declarations) - 1; i >= 0; i-- {
		if declarations[i].index < index {
			return declarations[i], true
		}
	}
	return formatJSConstructorDeclaration{}, false
}

func latestExplicitLocaleConstructorDeclarationBefore(declarations map[string][]formatJSConstructorDeclaration, name string, index int) (formatJSConstructorDeclaration, bool) {
	decl, ok := latestConstructorDeclarationBefore(declarations[name], index)
	if !ok || decl.locale == "" {
		return formatJSConstructorDeclaration{}, false
	}
	return decl, true
}

func dateVariables(data string) map[string][]dateVariable {
	dates := map[string][]dateVariable{}
	for _, match := range dateVarStringRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		if parsed, ok := parseDateString(data[match[4]:match[5]]); ok {
			dates[name] = append(dates[name], dateVariable{index: match[0], value: parsed})
		}
	}
	for _, match := range dateVarNumberRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		ms, err := strconv.ParseInt(data[match[4]:match[5]], 10, 64)
		if err == nil {
			dates[name] = append(dates[name], dateVariable{index: match[0], value: formatFixtureInstant(time.UnixMilli(ms))})
		}
	}
	for _, match := range dateVarYMDRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		instant, ok := parseDateYMD(matchStrings(data, match, 4))
		if ok {
			dates[name] = append(dates[name], dateVariable{index: match[0], value: instant})
		}
	}
	for _, match := range dateVarUTCRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		instant, ok := parseDateUTCArgs(data[match[4]:match[5]])
		if ok {
			dates[name] = append(dates[name], dateVariable{index: match[0], value: instant})
		}
	}
	return dates
}

func parseDateExpression(raw string, variables map[string][]dateVariable, index int) (string, bool) {
	expr := strings.TrimSpace(raw)
	expr = strings.TrimSpace(strings.TrimSuffix(expr, "as any"))
	if value, ok := latestDateVariableBefore(variables[expr], index); ok {
		return value, true
	}
	if strings.HasPrefix(expr, "new Date(") && strings.HasSuffix(expr, ")") {
		return parseDateConstructor(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(expr, "new Date("), ")")))
	}
	if strings.HasPrefix(expr, "Date.UTC(") && strings.HasSuffix(expr, ")") {
		return parseDateUTCArgs(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(expr, "Date.UTC("), ")")))
	}
	if instant, ok := parseDateNumberLiteral(expr); ok {
		return instant, true
	}
	return "", false
}

func latestDateVariableBefore(variables []dateVariable, index int) (string, bool) {
	for i := len(variables) - 1; i >= 0; i-- {
		if variables[i].index < index {
			return variables[i].value, true
		}
	}
	return "", false
}

func parseDateConstructor(raw string) (string, bool) {
	if len(raw) >= 2 && (raw[0] == '\'' || raw[0] == '"') {
		value, ok := decodeJSString(strings.Trim(raw, `'"`))
		if !ok {
			return "", false
		}
		return parseDateString(value)
	}
	if strings.HasPrefix(raw, "Date.UTC(") && strings.HasSuffix(raw, ")") {
		return parseDateUTCArgs(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(raw, "Date.UTC("), ")")))
	}
	if strings.Contains(raw, ",") {
		return parseDateUTCArgs(raw)
	}
	return parseDateNumberLiteral(raw)
}

func parseDateNumberLiteral(raw string) (string, bool) {
	value, ok := parseNumberLiteral(raw)
	if !ok {
		return "", false
	}
	switch value := value.(type) {
	case int64:
		return formatFixtureInstant(time.UnixMilli(value)), true
	case uint64:
		if value <= uint64(1<<63-1) {
			return formatFixtureInstant(time.UnixMilli(int64(value))), true
		}
	}
	return "", false
}

func parseDateString(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return formatFixtureInstant(parsed), true
	}
	layouts := []string{"2006-1-2", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return formatFixtureInstant(parsed), true
		}
	}
	return "", false
}

func parseDateYMD(fields []string) (string, bool) {
	if len(fields) < 3 {
		return "", false
	}
	values := make([]int, 7)
	for i, field := range fields {
		if i >= len(values) || field == "" {
			continue
		}
		value, err := strconv.Atoi(field)
		if err != nil {
			return "", false
		}
		values[i] = value
	}
	if values[1] < 0 || values[1] > 11 || values[2] < 1 {
		return "", false
	}
	instant := time.Date(values[0], time.Month(values[1]+1), values[2], values[3], values[4], values[5], values[6]*int(time.Millisecond), time.UTC)
	return formatFixtureInstant(instant), true
}

func parseDateUTCArgs(raw string) (string, bool) {
	fields := strings.Split(raw, ",")
	args := make([]string, len(fields))
	for i, field := range fields {
		args[i] = strings.TrimSpace(field)
	}
	return parseDateYMD(args)
}

func formatFixtureInstant(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func newDateTimeFormatFixture(rel string, index int, locale string, options map[string]any, input string, expected string) fixture {
	return formatJSExpectedFixture(formatJSDateTimeFormatTargetPackage, formatJSDateTimeFormatTestSourcePrefix, rel, index, locale, options, input, expected)
}

func newDateTimeFormatRangeFixture(rel string, index int, locale string, options map[string]any, start, end, expected string) fixture {
	return formatJSRangeFixture(formatJSDateTimeFormatTargetPackage, formatJSDateTimeFormatTestSourcePrefix, rel, index, locale, options, start, end, expected)
}

func newDateTimeFormatRangePartsFixture(rel string, index int, locale string, options map[string]any, start, end string, parts []conformance.RangePart) fixture {
	return formatJSRangePartsFixture(formatJSDateTimeFormatTargetPackage, formatJSDateTimeFormatTestSourcePrefix, rel, index, locale, options, start, end, parts)
}

func supportsGeneratedDateTimeFormatFixture(f fixture) bool {
	switch f.Locale {
	case "en", "en-US":
	default:
		return false
	}
	for key, value := range f.Options {
		switch key {
		case "weekday", "era", "year", "day", "hour", "minute", "second":
			if _, ok := stringValueOneOf(value, "numeric", "2-digit", "narrow", "short", "long"); !ok {
				return false
			}
		case "month":
			if _, ok := stringValueOneOf(value, "numeric", "2-digit", "narrow", "short", "long"); !ok {
				return false
			}
		case "timeZoneName":
			if _, ok := stringValueOneOf(value, "short", "long", "shortOffset", "longOffset", "shortGeneric", "longGeneric"); !ok {
				return false
			}
		case "timeZone":
			str, ok := value.(string)
			if !ok || str == "" {
				return false
			}
		case "hour12":
			if _, ok := value.(bool); !ok {
				return false
			}
		case "hourCycle":
			if _, ok := stringValueOneOf(value, "h11", "h12", "h23", "h24"); !ok {
				return false
			}
		case "dateStyle", "timeStyle":
			if _, ok := stringValueOneOf(value, "full", "long", "medium", "short"); !ok {
				return false
			}
		case "fractionalSecondDigits":
			digits, ok := value.(int64)
			if !ok || digits < 1 || digits > 3 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func extractLocaleFixtures(rel, data string) []fixture {
	fixtures := []fixture{}
	nextIndex := 0
	for _, match := range localeToStringRE.FindAllStringSubmatchIndex(data, -1) {
		input := data[match[2]:match[3]]
		expected, ok := decodeJSString(data[match[4]:match[5]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newLocaleFixture(formatJSLocaleTestSourcePrefix, rel, nextIndex, fixtureFeatureCanonicalize, input, expected))
		nextIndex++
	}
	for _, match := range localeMaximizeRE.FindAllStringSubmatchIndex(data, -1) {
		input := data[match[2]:match[3]]
		expected, ok := decodeJSString(data[match[4]:match[5]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newLocaleFixture(formatJSLocaleTestSourcePrefix, rel, nextIndex, fixtureFeatureMaximize, input, expected))
		nextIndex++
	}
	for _, match := range localeMinimizeRE.FindAllStringSubmatchIndex(data, -1) {
		input := data[match[2]:match[3]]
		expected, ok := decodeJSString(data[match[4]:match[5]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newLocaleFixture(formatJSLocaleTestSourcePrefix, rel, nextIndex, fixtureFeatureMinimize, input, expected))
		nextIndex++
	}
	return fixtures
}

func extractCanonicalLocalesFixtures(rel, data string) []fixture {
	fixtures := []fixture{}
	nextIndex := 0
	for _, match := range canonicalLocalesRE.FindAllStringSubmatchIndex(data, -1) {
		input := data[match[2]:match[3]]
		expected, ok := decodeJSString(data[match[4]:match[5]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newLocaleFixture(formatJSCanonicalLocalesTestSourcePrefix, rel, nextIndex, fixtureFeatureCanonicalize, input, expected))
		nextIndex++
	}
	return fixtures
}

func newLocaleFixture(sourcePrefix, rel string, index int, feature, input, expected string) fixture {
	return fixture{
		ID:       formatJSFixtureID("locale", sourcePrefix+rel, index),
		Source:   sourcePrefix + rel,
		Locale:   input,
		Feature:  feature,
		Options:  map[string]any{},
		Input:    input,
		Expected: ptr(expected),
	}
}

func supportsGeneratedLocaleFixture(f fixture) bool {
	if f.Expected == nil || f.Locale == "" {
		return false
	}
	return oneOf(f.Feature, fixtureFeatureCanonicalize, fixtureFeatureMaximize, fixtureFeatureMinimize)
}

var (
	inlineListFormatExpectationRE         = regexp.MustCompile(`(?s)expect\s*\(\s*new\s+ListFormat\s*\(\s*['"]([^'"]*)['"]\s*(?:,\s*(\{.*?\}))?\s*\)\.format\s*\(\s*(\[[^\]]*\])\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	inlineRelativeTimeFormatExpectationRE = regexp.MustCompile(`(?s)expect\s*\(\s*new\s+RelativeTimeFormat\s*\(\s*['"]([^'"]*)['"]\s*(?:,\s*(\{.*?\}))?\s*\)\.format\s*\(\s*([^,]+?)\s*,\s*['"]([^'"]*)['"]\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
	inlineDurationFormatExpectationRE     = regexp.MustCompile(`(?s)expect\s*\(\s*(?:new\s+)?(?:Intl\.)?DurationFormat\s*\(\s*['"]([^'"]*)['"]\s*(?:,\s*(\{.*?\}))?\s*\)\.format\s*\(\s*(\{.*?\})\s*\)\s*\)\s*\.to(?:Be|Equal)\s*\(\s*['"]((?:\\.|[^\\'"])*?)['"]\s*\)`)
)

func extractListFormatFixtures(rel, data string) []fixture {
	fixtures := []fixture{}
	nextIndex := 0
	for _, match := range inlineListFormatExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		if !isExtractableComposedFormatAssertion(data, match[0]) {
			continue
		}
		locale := data[match[2]:match[3]]
		options := parseOptionsObject(matchString(data, match, 4))
		input, ok := parseStringArray(data[match[6]:match[7]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newListFormatFixture(rel, nextIndex, locale, options, input, expected))
		nextIndex++
	}
	return fixtures
}

func newListFormatFixture(rel string, index int, locale string, options map[string]any, input []string, expected string) fixture {
	return formatJSExpectedFixture(formatJSListFormatTargetPackage, formatJSListFormatTestSourcePrefix, rel, index, locale, options, input, expected)
}

func supportsGeneratedListFormatFixture(f fixture) bool {
	if f.Expected == nil {
		return false
	}
	switch f.Locale {
	case "en-AI", "zh-CN", "zh-TW":
	default:
		return false
	}
	for key, value := range f.Options {
		switch key {
		case "type":
			if _, ok := stringValueOneOf(value, "conjunction", "disjunction", "unit"); !ok {
				return false
			}
		case "style":
			if _, ok := stringValueOneOf(value, "long", "short", "narrow"); !ok {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func extractRelativeTimeFormatFixtures(rel, data string) []fixture {
	fixtures := []fixture{}
	nextIndex := 0
	for _, match := range inlineRelativeTimeFormatExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		if !isExtractableComposedFormatAssertion(data, match[0]) {
			continue
		}
		locale := data[match[2]:match[3]]
		options := parseOptionsObject(matchString(data, match, 4))
		value, ok := parseNumberLiteral(data[match[6]:match[7]])
		if !ok {
			continue
		}
		unit, ok := decodeJSString(data[match[8]:match[9]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[10]:match[11]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newRelativeTimeFormatFixture(rel, nextIndex, locale, options, value, unit, expected))
		nextIndex++
	}
	return fixtures
}

func newRelativeTimeFormatFixture(rel string, index int, locale string, options map[string]any, value any, unit, expected string) fixture {
	return formatJSExpectedFixture(formatJSRelativeTimeFormatTargetPackage, formatJSRelativeTimeFormatTestSourcePrefix, rel, index, locale, options, map[string]any{"value": value, "unit": unit}, expected)
}

func supportsGeneratedRelativeTimeFormatFixture(f fixture) bool {
	if f.Expected == nil {
		return false
	}
	switch f.Locale {
	case "en-AI", "zh-CN", "zh-TW":
	default:
		return false
	}
	input, ok := f.Input.(map[string]any)
	if !ok {
		return false
	}
	if _, ok := stringValueOneOf(input["unit"], "second", "seconds", "minute", "minutes", "hour", "hours", "day", "days", "week", "weeks", "month", "months", "quarter", "quarters", "year", "years"); !ok {
		return false
	}
	for key, value := range f.Options {
		switch key {
		case "style":
			if _, ok := stringValueOneOf(value, "long", "short", "narrow"); !ok {
				return false
			}
		case "numeric":
			if _, ok := stringValueOneOf(value, "always", "auto"); !ok {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func extractDurationFormatFixtures(rel, data string) []fixture {
	fixtures := []fixture{}
	nextIndex := 0
	for _, match := range inlineDurationFormatExpectationRE.FindAllStringSubmatchIndex(data, -1) {
		if !isExtractableComposedFormatAssertion(data, match[0]) {
			continue
		}
		locale := data[match[2]:match[3]]
		options := parseOptionsObject(matchString(data, match, 4))
		input, ok := parseDurationObject(data[match[6]:match[7]])
		if !ok {
			continue
		}
		expected, ok := decodeJSString(data[match[8]:match[9]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newDurationFormatFixture(rel, nextIndex, locale, options, input, expected))
		nextIndex++
	}
	return fixtures
}

func newDurationFormatFixture(rel string, index int, locale string, options map[string]any, input map[string]any, expected string) fixture {
	return formatJSExpectedFixture(formatJSDurationFormatTargetPackage, formatJSDurationFormatTestSourcePrefix, rel, index, locale, options, input, expected)
}

func supportsGeneratedDurationFormatFixture(f fixture) bool {
	if f.Expected == nil || f.Locale != "en" {
		return false
	}
	input, ok := f.Input.(map[string]any)
	if !ok {
		return false
	}
	for key, value := range input {
		if !isFormatJSDurationUnitKey(key) {
			return false
		}
		if _, ok := value.(int64); !ok {
			return false
		}
	}
	for key, value := range f.Options {
		switch {
		case key == "localeMatcher":
			if _, ok := stringValueOneOf(value, "lookup", "best fit"); !ok {
				return false
			}
		case key == "numberingSystem":
			if _, ok := value.(string); !ok {
				return false
			}
		case key == "style":
			if _, ok := stringValueOneOf(value, "long", "short", "narrow", "digital"); !ok {
				return false
			}
		case isFormatJSDurationUnitKey(key):
			if _, ok := stringValueOneOf(value, "long", "short", "narrow", "numeric", "2-digit"); !ok {
				return false
			}
		case isFormatJSDurationUnitDisplayKey(key):
			if _, ok := stringValueOneOf(value, "always", "auto"); !ok {
				return false
			}
		case key == "fractionalDigits":
			digits, ok := value.(int64)
			if !ok || digits < 0 || digits > 9 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func isExtractableComposedFormatAssertion(data string, index int) bool {
	return !isSkippedVitestAssertion(data, index) && !isConditionalVitestAssertion(data, index)
}

func isSkippedVitestAssertion(data string, index int) bool {
	prefix := data[:index]
	skip := maxLastIndex(prefix, "it.skip(", "test.skip(")
	if skip < 0 {
		return false
	}
	normal := maxLastIndex(prefix, "it(", "test(")
	return skip > normal
}

func isConditionalVitestAssertion(data string, index int) bool {
	prefix := data[:index]
	conditional := maxLastIndex(prefix, "if (", "if(")
	if conditional < 0 {
		return false
	}
	test := maxLastIndex(prefix, "it(", "test(", "it.skip(", "test.skip(", "it.skipIf(", "test.skipIf(")
	return conditional > test
}

func maxLastIndex(data string, needles ...string) int {
	out := -1
	for _, needle := range needles {
		out = max(out, strings.LastIndex(data, needle))
	}
	return out
}

func oneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

var formatJSDurationUnitKeys = [...]string{
	"years",
	"months",
	"weeks",
	"days",
	"hours",
	"minutes",
	"seconds",
	"milliseconds",
	"microseconds",
	"nanoseconds",
}

func isFormatJSDurationUnitKey(key string) bool {
	return oneOf(key, formatJSDurationUnitKeys[:]...)
}

func isFormatJSDurationUnitDisplayKey(key string) bool {
	unit, ok := strings.CutSuffix(key, "Display")
	return ok && isFormatJSDurationUnitKey(unit)
}

func stringValueOneOf(value any, allowed ...string) (string, bool) {
	str, ok := value.(string)
	if !ok || !oneOf(str, allowed...) {
		return "", false
	}
	return str, true
}

func isPluralCategory(category string) bool {
	switch category {
	case "zero", "one", "two", "few", "many", "other":
		return true
	default:
		return false
	}
}

func parsePartArray(raw string) ([]conformance.Part, bool) {
	objects, ok := parsePartObjects(raw)
	if !ok {
		return nil, false
	}
	parts := make([]conformance.Part, len(objects))
	for i, object := range objects {
		parts[i] = conformance.Part{Type: object.typ, Value: object.value}
	}
	return parts, true
}

func parseRangePartArray(raw string) ([]conformance.RangePart, bool) {
	objects, ok := parsePartObjects(raw)
	if !ok {
		return nil, false
	}
	parts := make([]conformance.RangePart, len(objects))
	for i, object := range objects {
		source, hasSource := object.fields["source"]
		if !hasSource {
			return nil, false
		}
		parts[i] = conformance.RangePart{Type: object.typ, Value: object.value, Source: source}
	}
	return parts, true
}

type parsedPartObject struct {
	typ    string
	value  string
	fields map[string]string
}

func parsePartObjects(raw string) ([]parsedPartObject, bool) {
	matches := partObjectRE.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil, false
	}
	objects := make([]parsedPartObject, len(matches))
	for i, match := range matches {
		values := objectStringFields(match[1])
		typ, hasType := values["type"]
		value, hasValue := values["value"]
		if !hasType || !hasValue {
			return nil, false
		}
		objects[i] = parsedPartObject{typ: typ, value: value, fields: values}
	}
	return objects, true
}

func objectStringFields(raw string) map[string]string {
	values := map[string]string{}
	for _, match := range stringOptionRE.FindAllStringSubmatch(raw, -1) {
		value, ok := decodeJSString(match[2])
		if ok {
			values[match[1]] = value
		}
	}
	return values
}

func joinPartValues(parts []conformance.Part) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func joinRangePartValues(parts []conformance.RangePart) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func parseOptionsObject(raw string) map[string]any {
	options := map[string]any{}
	for _, match := range stringOptionRE.FindAllStringSubmatch(raw, -1) {
		value, ok := decodeJSString(match[2])
		if ok {
			options[match[1]] = value
		}
	}
	for _, match := range numberOptionRE.FindAllStringSubmatch(raw, -1) {
		if _, exists := options[match[1]]; exists {
			continue
		}
		if value, ok := parseNumberLiteral(match[2]); ok {
			options[match[1]] = value
		}
	}
	for _, match := range boolOptionRE.FindAllStringSubmatch(raw, -1) {
		if _, exists := options[match[1]]; exists {
			continue
		}
		options[match[1]] = match[2] == "true"
	}
	return options
}

func parseStringArray(raw string) ([]string, bool) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") {
		return nil, false
	}
	inner := strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))
	if inner == "" {
		return []string{}, true
	}
	values := []string{}
	for _, field := range strings.Split(inner, ",") {
		literal := strings.TrimSpace(field)
		if len(literal) < 2 {
			return nil, false
		}
		quote := literal[0]
		if quote != '\'' && quote != '"' {
			return nil, false
		}
		if literal[len(literal)-1] != quote {
			return nil, false
		}
		value, ok := decodeJSString(literal[1 : len(literal)-1])
		if !ok {
			return nil, false
		}
		values = append(values, value)
	}
	return values, true
}

func parseDurationObject(raw string) (map[string]any, bool) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "{") || !strings.HasSuffix(trimmed, "}") {
		return nil, false
	}
	matches := numberOptionRE.FindAllStringSubmatch(trimmed, -1)
	if len(matches) == 0 {
		return nil, false
	}
	values := make(map[string]any, len(matches))
	for _, match := range matches {
		name := match[1]
		if !isFormatJSDurationUnitKey(name) {
			return nil, false
		}
		value, ok := parseNumberLiteral(match[2])
		if !ok {
			return nil, false
		}
		switch value := value.(type) {
		case int64:
			values[name] = value
		case uint64:
			if value > 1<<63-1 {
				return nil, false
			}
			values[name] = int64(value)
		default:
			return nil, false
		}
	}
	return values, true
}

func parseNumberLiteral(raw string) (any, bool) {
	literal := strings.ReplaceAll(strings.TrimSpace(raw), "_", "")
	if literal == "" || strings.ContainsAny(literal, "(){}[]") {
		return nil, false
	}
	switch literal {
	case "NaN", "Infinity", "+Infinity", "-Infinity":
		return nil, false
	}
	if strings.ContainsAny(literal, ".eE") {
		value, err := strconv.ParseFloat(literal, 64)
		return value, err == nil
	}
	value, err := strconv.ParseInt(literal, 10, 64)
	if err == nil {
		return value, true
	}
	unsigned, err := strconv.ParseUint(literal, 10, 64)
	if err != nil {
		return nil, false
	}
	return unsigned, true
}

func parsePluralInputLiteral(raw string) (any, bool) {
	literal := strings.TrimSpace(strings.ReplaceAll(raw, "_", ""))
	literal = strings.TrimSpace(strings.TrimSuffix(literal, "as any"))
	if literal == "" || strings.ContainsAny(literal, "{}[]") {
		return nil, false
	}
	if strings.HasPrefix(literal, "BigInt(") && strings.HasSuffix(literal, ")") {
		return parsePluralIntegerString(strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(literal, "BigInt("), ")")))
	}
	if strings.HasSuffix(literal, "n") {
		return parsePluralIntegerString(strings.TrimSuffix(literal, "n"))
	}
	if len(literal) >= 2 && (literal[0] == '\'' || literal[0] == '"') {
		value, ok := decodeJSString(strings.Trim(literal, `'"`))
		if !ok {
			return nil, false
		}
		return value, true
	}
	return parseNumberLiteral(literal)
}

func parsePluralIntegerString(raw string) (any, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || strings.ContainsAny(value, ".eE+-*/ ") {
		if strings.HasPrefix(value, "-") && !strings.ContainsAny(strings.TrimPrefix(value, "-"), ".eE+-*/ ") {
			return value, true
		}
		return nil, false
	}
	return value, true
}

func decodeJSString(raw string) (string, bool) {
	escaped := strings.ReplaceAll(raw, `\'`, `'`)
	quoted := `"` + strings.ReplaceAll(escaped, `"`, `\"`) + `"`
	value, err := strconv.Unquote(quoted)
	return value, err == nil
}

func matchString(data string, match []int, index int) string {
	if index+1 >= len(match) || match[index] < 0 {
		return ""
	}
	return data[match[index]:match[index+1]]
}

func matchStrings(data string, match []int, start int) []string {
	values := make([]string, 0, (len(match)-start)/2)
	for i := start; i+1 < len(match); i += 2 {
		if match[i] < 0 {
			values = append(values, "")
			continue
		}
		values = append(values, data[match[i]:match[i+1]])
	}
	return values
}

func fixtureSlug(path string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(path) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if b.Len() == 0 || lastHyphen {
				continue
			}
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func formatJSFixtureID(surface, rel string, index int) string {
	return fmt.Sprintf("%s-formatjs-%s-%03d", surface, fixtureSlug(rel), index)
}

func formatJSSurfaceFixture(surface, sourcePrefix, rel string, index int, locale string, options map[string]any, input any) fixture {
	return fixture{
		ID:      formatJSFixtureID(surface, rel, index),
		Source:  sourcePrefix + rel,
		Locale:  locale,
		Options: options,
		Input:   input,
	}
}

func formatJSExpectedFixture(surface, sourcePrefix, rel string, index int, locale string, options map[string]any, input any, expected string) fixture {
	f := formatJSSurfaceFixture(surface, sourcePrefix, rel, index, locale, options, input)
	f.Expected = ptr(expected)
	return f
}

func formatJSRangeFixture(surface, sourcePrefix, rel string, index int, locale string, options map[string]any, start, end any, expected string) fixture {
	f := formatJSSurfaceFixture(surface, sourcePrefix, rel, index, locale, options, rangeFixtureInput(start, end))
	f.Feature = fixtureFeatureFormatRange
	f.ExpectedRange = ptr(expected)
	return f
}

func formatJSRangePartsFixture(surface, sourcePrefix, rel string, index int, locale string, options map[string]any, start, end any, parts []conformance.RangePart) fixture {
	f := formatJSSurfaceFixture(surface, sourcePrefix, rel, index, locale, options, rangeFixtureInput(start, end))
	f.Feature = fixtureFeatureFormatRangeToParts
	f.ExpectedRange = ptr(joinRangePartValues(parts))
	f.ExpectedRangeParts = parts
	return f
}

func rangeFixtureInput(start, end any) map[string]any {
	return map[string]any{"start": start, "end": end}
}

func ptr(value string) *string {
	return &value
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o666)
}
