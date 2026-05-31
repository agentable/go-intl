package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type fixture struct {
	ID                 string         `json:"id"`
	Source             string         `json:"source"`
	Locale             string         `json:"locale"`
	Feature            string         `json:"feature,omitempty"`
	Options            map[string]any `json:"options"`
	Input              any            `json:"input"`
	Expected           *string        `json:"expected,omitempty"`
	ExpectedLocales    []string       `json:"expectedLocales,omitempty"`
	ExpectedParts      []fixturePart  `json:"expectedParts,omitempty"`
	ExpectedRange      *string        `json:"expectedRange,omitempty"`
	ExpectedRangeParts []rangePart    `json:"expectedRangeParts,omitempty"`
	ExpectedResolved   any            `json:"expectedResolvedOptions,omitempty"`
}

type fixturePart struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	Unit  string `json:"unit,omitempty"`
}

type rangePart struct {
	Type   string `json:"type"`
	Value  string `json:"value"`
	Source string `json:"source"`
}

type skipEntry struct {
	Source       string `json:"source"`
	Category     string `json:"category"`
	Reason       string `json:"reason"`
	DivergenceID string `json:"divergenceId,omitempty"`
}

type nodeWitness struct {
	NodeVersion           string              `json:"nodeVersion"`
	Versions              map[string]string   `json:"versions"`
	NumberFormatResolved  []fixture           `json:"numberFormatResolved"`
	DurationFormatDigital []fixture           `json:"durationFormatDigital"`
	SupportedValues       nodeSupportedValues `json:"supportedValues"`
}

type nodeSupportedValues struct {
	Source   string              `json:"source"`
	Versions map[string]string   `json:"versions"`
	Values   map[string][]string `json:"values"`
}

const formatJSNumberFormatTestSourcePrefix = "formatjs:packages/intl-numberformat/tests/"
const formatJSPluralRulesTestSourcePrefix = "formatjs:packages/intl-pluralrules/tests/"
const formatJSDateTimeFormatTestSourcePrefix = "formatjs:packages/intl-datetimeformat/tests/"
const formatJSLocaleTestSourcePrefix = "formatjs:packages/intl-locale/tests/"
const formatJSCanonicalLocalesTestSourcePrefix = "formatjs:packages/intl-getcanonicallocales/tests/"
const formatJSListFormatTestSourcePrefix = "formatjs:packages/intl-listformat/tests/"
const formatJSRelativeTimeFormatTestSourcePrefix = "formatjs:packages/intl-relativetimeformat/tests/"
const formatJSDurationFormatTestSourcePrefix = "formatjs:packages/intl-durationformat/tests/"

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
	return writeSkips(*outDir, []skipEntry{{Source: "formatjs", Category: "missing-reference", Reason: "formatjs path not provided"}})
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
	return writeJSON(filepath.Join(outDir, ".skip-list.json"), skips)
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
	if len(witness.NumberFormatResolved) > 0 {
		path := filepath.Join(outDir, "numberformat", "testdata", "conformance", nodeDir, "resolved-options.json")
		if err := writeJSON(path, witness.NumberFormatResolved); err != nil {
			return err
		}
	}
	if len(witness.DurationFormatDigital) > 0 {
		path := filepath.Join(outDir, "durationformat", "testdata", "conformance", nodeDir, "digital.json")
		if err := writeJSON(path, witness.DurationFormatDigital); err != nil {
			return err
		}
	}
	if len(witness.SupportedValues.Values) > 0 {
		path := filepath.Join(outDir, "testdata", "native", nodeDir, "supported-values.json")
		if err := writeJSON(path, witness.SupportedValues); err != nil {
			return err
		}
	}
	return nil
}

func runNodeWitness(nodePath string) (nodeWitness, error) {
	cmd := exec.Command(nodePath, "-e", nodeWitnessScript)
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

func nodeFixtureDir(version string) (string, error) {
	trimmed := strings.TrimPrefix(version, "v")
	major, _, ok := strings.Cut(trimmed, ".")
	if !ok || major == "" {
		return "", fmt.Errorf("invalid Node version %q", version)
	}
	if _, err := strconv.Atoi(major); err != nil {
		return "", fmt.Errorf("invalid Node major version %q: %w", version, err)
	}
	return "node-v" + major, nil
}

const nodeWitnessScript = `
const selectedVersions = {};
for (const key of ['node', 'v8', 'icu', 'cldr', 'tz', 'unicode']) {
  if (process.versions[key]) {
    selectedVersions[key] = process.versions[key];
  }
}

const nodeMajor = process.versions.node.split('.')[0];
const nodeVersion = process.version;

function source(surface, topic) {
  return 'node:' + nodeVersion + ':' + surface + ':' + topic;
}

function id(surface, topic) {
  return surface + '-node-v' + nodeMajor + '-' + topic;
}

function numberFormatFixture(topic, locale, options, input) {
  const format = new Intl.NumberFormat(locale, options);
  return {
    id: id('numberformat', topic),
    source: source('numberformat', 'resolved-options'),
    locale,
    options,
    input,
    expected: format.format(input),
    expectedResolvedOptions: format.resolvedOptions(),
  };
}

function durationFormatFixture(topic, locale, options, input) {
  const format = new Intl.DurationFormat(locale, options);
  return {
    id: id('durationformat', topic),
    source: source('durationformat', 'digital'),
    locale,
    options,
    input,
    expected: format.format(input),
    expectedParts: format.formatToParts(input),
    expectedResolvedOptions: format.resolvedOptions(),
  };
}

const supportedValues = {};
for (const key of ['calendar', 'collation', 'currency', 'numberingSystem', 'timeZone', 'unit']) {
  supportedValues[key] = Intl.supportedValuesOf(key);
}

const witness = {
  nodeVersion,
  versions: selectedVersions,
  numberFormatResolved: [
    numberFormatFixture('resolved-decimal-default', 'en', {}, 12345.6),
    numberFormatFixture('resolved-significant-digits', 'en', {minimumSignificantDigits: 3}, 1.2),
    numberFormatFixture('resolved-compact-defaults', 'en', {notation: 'compact'}, 1200),
  ],
  durationFormatDigital: [
    durationFormatFixture('digital-hours-minutes-seconds', 'en', {style: 'digital'}, {hours: 5, minutes: 30, seconds: 15}),
    durationFormatFixture('digital-fractional-seconds', 'en', {style: 'digital', fractionalDigits: 3}, {hours: 5, minutes: 30, seconds: 15, milliseconds: 123}),
    durationFormatFixture('digital-zero-hours', 'en', {style: 'digital'}, {minutes: 30, seconds: 15}),
  ],
  supportedValues: {
    source: 'node:' + nodeVersion + ':intl:supportedValuesOf',
    versions: selectedVersions,
    values: supportedValues,
  },
};

console.log(JSON.stringify(witness));
`

func importFormatJS(path, outDir string) ([]skipEntry, error) {
	numberSkips, err := importFormatJSNumberFormat(path, outDir)
	if err != nil {
		return nil, err
	}
	pluralSkips, err := importFormatJSPluralRules(path, outDir)
	if err != nil {
		return nil, err
	}
	dateSkips, err := importFormatJSDateTimeFormat(path, outDir)
	if err != nil {
		return nil, err
	}
	localeSkips, err := importFormatJSLocales(path, outDir)
	if err != nil {
		return nil, err
	}
	listSkips, err := importFormatJSListFormat(path, outDir)
	if err != nil {
		return nil, err
	}
	relativeTimeSkips, err := importFormatJSRelativeTimeFormat(path, outDir)
	if err != nil {
		return nil, err
	}
	durationSkips, err := importFormatJSDurationFormat(path, outDir)
	if err != nil {
		return nil, err
	}
	skips := append(numberSkips, pluralSkips...)
	skips = append(skips, dateSkips...)
	skips = append(skips, localeSkips...)
	skips = append(skips, listSkips...)
	skips = append(skips, relativeTimeSkips...)
	skips = append(skips, durationSkips...)
	return skips, nil
}

func importFormatJSNumberFormat(path, outDir string) ([]skipEntry, error) {
	root := filepath.Join(path, "packages", "intl-numberformat", "tests")
	if stat, err := os.Stat(root); err != nil {
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	targetRoot := filepath.Join(outDir, "numberformat", "testdata", "conformance", "formatjs")
	if err := os.RemoveAll(targetRoot); err != nil {
		return nil, err
	}
	fixtures := []fixture{}
	skips := []skipEntry{}
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
			extracted := extractNumberFormatFixtures(rel, string(data))
			supportedCount := 0
			for _, fixture := range extracted {
				if supportsGeneratedNumberFormatFixture(fixture) {
					fixtures = append(fixtures, fixture)
					supportedCount++
				}
			}
			if len(extracted) > 0 && supportedCount < len(extracted) {
				skips = append(skips, skipEntry{Source: formatJSNumberFormatTestSourcePrefix + rel, Category: "partial-extraction", Reason: "mechanical assertions outside current generated fixture gate"})
			}
			if len(extracted) == 0 && strings.Contains(string(data), "expect(") {
				skips = append(skips, skipEntry{Source: formatJSNumberFormatTestSourcePrefix + rel, Category: "unsupported-extractor-shape", Reason: "unsupported Vitest assertion shape"})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := writeFixturesBySource(targetRoot, fixtures, formatJSNumberFormatTestSourcePrefix); err != nil {
		return nil, err
	}
	return skips, nil
}

func importFormatJSPluralRules(path, outDir string) ([]skipEntry, error) {
	root := filepath.Join(path, "packages", "intl-pluralrules", "tests")
	if stat, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []skipEntry{{Source: formatJSPluralRulesTestSourcePrefix, Category: "missing-reference", Reason: "tests path not found"}}, nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	targetRoot := filepath.Join(outDir, "pluralrules", "testdata", "conformance", "formatjs")
	if err := os.RemoveAll(targetRoot); err != nil {
		return nil, err
	}
	fixtures := []fixture{}
	skips := []skipEntry{}
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
			extracted := extractPluralRulesFixtures(rel, string(data))
			supportedCount := 0
			for _, fixture := range extracted {
				if supportsGeneratedPluralRulesFixture(fixture) {
					fixtures = append(fixtures, fixture)
					supportedCount++
				}
			}
			if len(extracted) > 0 && supportedCount < len(extracted) {
				skips = append(skips, skipEntry{Source: formatJSPluralRulesTestSourcePrefix + rel, Category: "partial-extraction", Reason: "mechanical assertions outside current generated fixture gate"})
			}
			if len(extracted) == 0 && strings.Contains(string(data), "expect(") {
				skips = append(skips, skipEntry{Source: formatJSPluralRulesTestSourcePrefix + rel, Category: "unsupported-extractor-shape", Reason: "unsupported Vitest assertion shape"})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := writeFixturesBySource(targetRoot, fixtures, formatJSPluralRulesTestSourcePrefix); err != nil {
		return nil, err
	}
	return skips, nil
}

func importFormatJSDateTimeFormat(path, outDir string) ([]skipEntry, error) {
	root := filepath.Join(path, "packages", "intl-datetimeformat", "tests")
	if stat, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []skipEntry{{Source: formatJSDateTimeFormatTestSourcePrefix, Category: "missing-reference", Reason: "tests path not found"}}, nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	targetRoot := filepath.Join(outDir, "datetimeformat", "testdata", "conformance", "formatjs")
	if err := os.RemoveAll(targetRoot); err != nil {
		return nil, err
	}
	fixtures := []fixture{}
	skips := []skipEntry{}
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
			extracted := extractDateTimeFormatFixtures(rel, string(data))
			supportedCount := 0
			for _, fixture := range extracted {
				if supportsGeneratedDateTimeFormatFixture(fixture) {
					fixtures = append(fixtures, fixture)
					supportedCount++
				}
			}
			if len(extracted) > 0 && supportedCount < len(extracted) {
				skips = append(skips, skipEntry{Source: formatJSDateTimeFormatTestSourcePrefix + rel, Category: "partial-extraction", Reason: "mechanical assertions outside current generated fixture gate"})
			}
			if len(extracted) == 0 && strings.Contains(string(data), "expect(") {
				skips = append(skips, skipEntry{Source: formatJSDateTimeFormatTestSourcePrefix + rel, Category: "unsupported-extractor-shape", Reason: "unsupported Vitest assertion shape"})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := writeFixturesBySource(targetRoot, fixtures, formatJSDateTimeFormatTestSourcePrefix); err != nil {
		return nil, err
	}
	return skips, nil
}

func importFormatJSLocales(path, outDir string) ([]skipEntry, error) {
	targetRoot := filepath.Join(outDir, "locale", "testdata", "conformance", "formatjs")
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
	root := filepath.Join(path, "packages", "intl-locale", "tests")
	if stat, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []skipEntry{{Source: formatJSLocaleTestSourcePrefix, Category: "missing-reference", Reason: "tests path not found"}}, nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	return importFormatJSLocaleFixtures(root, targetRoot, formatJSLocaleTestSourcePrefix, extractLocaleFixtures, fixtureSlug)
}

func importFormatJSCanonicalLocalesPackage(path, targetRoot string) ([]skipEntry, error) {
	root := filepath.Join(path, "packages", "intl-getcanonicallocales", "tests")
	if stat, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []skipEntry{{Source: formatJSCanonicalLocalesTestSourcePrefix, Category: "missing-reference", Reason: "tests path not found"}}, nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}
	slug := func(rel string) string {
		return "intl-getcanonicallocales-" + fixtureSlug(rel)
	}
	return importFormatJSLocaleFixtures(root, targetRoot, formatJSCanonicalLocalesTestSourcePrefix, extractCanonicalLocalesFixtures, slug)
}

func importFormatJSLocaleFixtures(root, targetRoot, sourcePrefix string, extract func(string, string) []fixture, slug func(string) string) ([]skipEntry, error) {
	fixtures := []fixture{}
	skips := []skipEntry{}
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
			extracted := extract(rel, string(data))
			supportedCount := 0
			for _, fixture := range extracted {
				if supportsGeneratedLocaleFixture(fixture) {
					fixtures = append(fixtures, fixture)
					supportedCount++
				}
			}
			if len(extracted) > 0 && supportedCount < len(extracted) {
				skips = append(skips, skipEntry{Source: sourcePrefix + rel, Category: "partial-extraction", Reason: "mechanical assertions outside current generated fixture gate"})
			}
			if len(extracted) == 0 && strings.Contains(string(data), "expect(") {
				skips = append(skips, skipEntry{Source: sourcePrefix + rel, Category: "unsupported-extractor-shape", Reason: "unsupported Vitest assertion shape"})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := writeLocaleFixturesBySource(targetRoot, fixtures, sourcePrefix, slug); err != nil {
		return nil, err
	}
	return skips, nil
}

func importFormatJSListFormat(path, outDir string) ([]skipEntry, error) {
	root := filepath.Join(path, "packages", "intl-listformat", "tests")
	if stat, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []skipEntry{{Source: formatJSListFormatTestSourcePrefix, Category: "missing-reference", Reason: "tests path not found"}}, nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	targetRoot := filepath.Join(outDir, "listformat", "testdata", "conformance", "formatjs")
	return importSimpleFormatJSFixtures(root, targetRoot, formatJSListFormatTestSourcePrefix, extractListFormatFixtures, supportsGeneratedListFormatFixture)
}

func importFormatJSRelativeTimeFormat(path, outDir string) ([]skipEntry, error) {
	root := filepath.Join(path, "packages", "intl-relativetimeformat", "tests")
	if stat, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []skipEntry{{Source: formatJSRelativeTimeFormatTestSourcePrefix, Category: "missing-reference", Reason: "tests path not found"}}, nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	targetRoot := filepath.Join(outDir, "relativetimeformat", "testdata", "conformance", "formatjs")
	return importSimpleFormatJSFixtures(root, targetRoot, formatJSRelativeTimeFormatTestSourcePrefix, extractRelativeTimeFormatFixtures, supportsGeneratedRelativeTimeFormatFixture)
}

func importFormatJSDurationFormat(path, outDir string) ([]skipEntry, error) {
	root := filepath.Join(path, "packages", "intl-durationformat", "tests")
	if stat, err := os.Stat(root); err != nil {
		if os.IsNotExist(err) {
			return []skipEntry{{Source: formatJSDurationFormatTestSourcePrefix, Category: "missing-reference", Reason: "tests path not found"}}, nil
		}
		return nil, err
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	targetRoot := filepath.Join(outDir, "durationformat", "testdata", "conformance", "formatjs")
	return importSimpleFormatJSFixtures(root, targetRoot, formatJSDurationFormatTestSourcePrefix, extractDurationFormatFixtures, supportsGeneratedDurationFormatFixture)
}

func importSimpleFormatJSFixtures(root, targetRoot, sourcePrefix string, extract func(string, string) []fixture, supports func(fixture) bool) ([]skipEntry, error) {
	if err := os.RemoveAll(targetRoot); err != nil {
		return nil, err
	}
	fixtures := []fixture{}
	skips := []skipEntry{}
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
			extracted := extract(rel, string(data))
			supportedCount := 0
			for _, fixture := range extracted {
				if supports(fixture) {
					fixtures = append(fixtures, fixture)
					supportedCount++
				}
			}
			if len(extracted) > 0 && (supportedCount < len(extracted) || strings.Count(string(data), "expect(") > len(extracted)) {
				skips = append(skips, skipEntry{Source: sourcePrefix + rel, Category: "partial-extraction", Reason: "mechanical assertions outside current generated fixture gate"})
			}
			if len(extracted) == 0 && strings.Contains(string(data), "expect(") {
				skips = append(skips, skipEntry{Source: sourcePrefix + rel, Category: "unsupported-extractor-shape", Reason: "unsupported Vitest assertion shape"})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	if err := writeFixturesBySource(targetRoot, fixtures, sourcePrefix); err != nil {
		return nil, err
	}
	return skips, nil
}

func writeFixturesBySource(targetRoot string, fixtures []fixture, sourcePrefix string) error {
	return writeFixturesBySourceSlug(targetRoot, fixtures, sourcePrefix, fixtureSlug)
}

func writeLocaleFixturesBySource(targetRoot string, fixtures []fixture, sourcePrefix string, slug func(string) string) error {
	return writeFixturesBySourceSlug(targetRoot, fixtures, sourcePrefix, slug)
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
	sources := make([]string, 0, len(fixturesBySource))
	for source := range fixturesBySource {
		sources = append(sources, source)
	}
	slices.Sort(sources)
	for _, source := range sources {
		rel := strings.TrimPrefix(source, sourcePrefix)
		if err := writeJSON(filepath.Join(targetRoot, slug(rel)+".json"), fixturesBySource[source]); err != nil {
			return err
		}
	}
	return nil
}

func supportsGeneratedNumberFormatFixture(f fixture) bool {
	if f.Locale != "en" {
		return false
	}
	switch f.Feature {
	case "", "formatToParts", "formatRange", "formatRangeToParts":
	default:
		return false
	}
	style := "decimal"
	currency := ""
	for key, value := range f.Options {
		switch key {
		case "style":
			styleValue, ok := value.(string)
			if !ok || (styleValue != "currency" && styleValue != "percent" && styleValue != "decimal") {
				return false
			}
			style = styleValue
		case "currency":
			currencyValue, ok := value.(string)
			if !ok || currencyValue != "USD" {
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

type numberFormatDeclaration struct {
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
		decl, ok := latestDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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
		decl, ok := latestDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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
		decl, ok := latestDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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
		decl, ok := latestDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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

func numberFormatDeclarations(data string) map[string][]numberFormatDeclaration {
	declarations := map[string][]numberFormatDeclaration{}
	for _, match := range numberFormatDeclarationRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		locale := ""
		if match[4] >= 0 {
			locale = data[match[4]:match[5]]
		}
		declarations[name] = append(declarations[name], numberFormatDeclaration{
			index:   match[0],
			locale:  locale,
			options: parseOptionsObject(matchString(data, match, 6)),
		})
	}
	return declarations
}

func latestDeclarationBefore(declarations []numberFormatDeclaration, index int) (numberFormatDeclaration, bool) {
	for i := len(declarations) - 1; i >= 0; i-- {
		if declarations[i].index < index {
			return declarations[i], true
		}
	}
	return numberFormatDeclaration{}, false
}

func newNumberFormatFixture(rel string, index int, locale string, options map[string]any, input any, expected string) fixture {
	source := formatJSNumberFormatTestSourcePrefix + rel
	return fixture{
		ID:       fmt.Sprintf("numberformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:   source,
		Locale:   locale,
		Options:  options,
		Input:    input,
		Expected: ptr(expected),
	}
}

func newNumberFormatPartsFixture(rel string, index int, locale string, options map[string]any, input any, parts []fixturePart) fixture {
	source := formatJSNumberFormatTestSourcePrefix + rel
	return fixture{
		ID:            fmt.Sprintf("numberformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:        source,
		Locale:        locale,
		Feature:       "formatToParts",
		Options:       options,
		Input:         input,
		Expected:      ptr(joinPartValues(parts)),
		ExpectedParts: parts,
	}
}

func newNumberFormatRangeFixture(rel string, index int, locale string, options map[string]any, start, end any, expected string) fixture {
	source := formatJSNumberFormatTestSourcePrefix + rel
	return fixture{
		ID:            fmt.Sprintf("numberformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:        source,
		Locale:        locale,
		Feature:       "formatRange",
		Options:       options,
		Input:         map[string]any{"start": start, "end": end},
		ExpectedRange: ptr(expected),
	}
}

func newNumberFormatRangePartsFixture(rel string, index int, locale string, options map[string]any, start, end any, parts []rangePart) fixture {
	source := formatJSNumberFormatTestSourcePrefix + rel
	return fixture{
		ID:                 fmt.Sprintf("numberformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:             source,
		Locale:             locale,
		Feature:            "formatRangeToParts",
		Options:            options,
		Input:              map[string]any{"start": start, "end": end},
		ExpectedRange:      ptr(joinRangePartValues(parts)),
		ExpectedRangeParts: parts,
	}
}

type pluralRulesDeclaration struct {
	index   int
	locale  string
	options map[string]any
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
		decl, ok := latestPluralRulesDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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
		decl, ok := latestPluralRulesDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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

func pluralRulesDeclarations(data string) map[string][]pluralRulesDeclaration {
	declarations := map[string][]pluralRulesDeclaration{}
	for _, match := range pluralRulesDeclarationRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		locale := ""
		if match[4] >= 0 {
			locale = data[match[4]:match[5]]
		}
		declarations[name] = append(declarations[name], pluralRulesDeclaration{
			index:   match[0],
			locale:  locale,
			options: parseOptionsObject(matchString(data, match, 6)),
		})
	}
	return declarations
}

func latestPluralRulesDeclarationBefore(declarations []pluralRulesDeclaration, index int) (pluralRulesDeclaration, bool) {
	for i := len(declarations) - 1; i >= 0; i-- {
		if declarations[i].index < index {
			return declarations[i], true
		}
	}
	return pluralRulesDeclaration{}, false
}

func newPluralRulesFixture(rel string, index int, locale string, options map[string]any, input any, expected string) fixture {
	source := formatJSPluralRulesTestSourcePrefix + rel
	return fixture{
		ID:       fmt.Sprintf("pluralrules-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:   source,
		Locale:   locale,
		Options:  options,
		Input:    input,
		Expected: ptr(expected),
	}
}

func newPluralRulesRangeFixture(rel string, index int, locale string, options map[string]any, start, end any, expected string) fixture {
	source := formatJSPluralRulesTestSourcePrefix + rel
	return fixture{
		ID:       fmt.Sprintf("pluralrules-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:   source,
		Locale:   locale,
		Feature:  "selectRange",
		Options:  options,
		Input:    map[string]any{"start": start, "end": end},
		Expected: ptr(expected),
	}
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
			typeValue, ok := value.(string)
			if !ok || (typeValue != "cardinal" && typeValue != "ordinal") {
				return false
			}
		case "notation":
			notation, ok := value.(string)
			if !ok || (notation != "standard" && notation != "scientific" && notation != "engineering" && notation != "compact") {
				return false
			}
		case "compactDisplay":
			display, ok := value.(string)
			if !ok || (display != "short" && display != "long") {
				return false
			}
		case "roundingMode":
			mode, ok := value.(string)
			if !ok || !oneOf(mode, "ceil", "floor", "expand", "trunc", "halfCeil", "halfFloor", "halfExpand", "halfTrunc", "halfEven") {
				return false
			}
		case "roundingPriority":
			priority, ok := value.(string)
			if !ok || !oneOf(priority, "auto", "morePrecision", "lessPrecision") {
				return false
			}
		case "trailingZeroDisplay":
			display, ok := value.(string)
			if !ok || !oneOf(display, "auto", "stripIfInteger") {
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

type dateTimeDeclaration struct {
	index   int
	locale  string
	options map[string]any
}

type dateVariable struct {
	index int
	value string
}

var (
	dateTimeDeclarationRE      = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*(?:new\s+)?(?:Intl\.)?DateTimeFormat\s*\(\s*(?:['"]([^'"]*)['"]|\[['"]([^'"]*)['"]\]\s*)?(?:,\s*(\{.*?\}))?\s*\)`)
	dateVarStringRE            = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*new\s+Date\s*\(\s*['"]([^'"]*)['"]\s*\)`)
	dateVarNumberRE            = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*new\s+Date\s*\(\s*(-?\d+)\s*\)`)
	dateVarYMDRE               = regexp.MustCompile(`(?s)(?:const|let|var)\s+([A-Za-z_]\w*)\s*=\s*new\s+Date\s*\(\s*(-?\d+)\s*,\s*(\d+)\s*,\s*(\d+)(?:\s*,\s*(\d+))?(?:\s*,\s*(\d+))?(?:\s*,\s*(\d+))?\s*\)`)
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
		decl, ok := latestDateTimeDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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
		decl, ok := latestDateTimeDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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
		decl, ok := latestDateTimeDeclarationBefore(declarations[name], match[0])
		if !ok || decl.locale == "" {
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

func dateTimeDeclarations(data string) map[string][]dateTimeDeclaration {
	declarations := map[string][]dateTimeDeclaration{}
	for _, match := range dateTimeDeclarationRE.FindAllStringSubmatchIndex(data, -1) {
		name := data[match[2]:match[3]]
		locale := matchString(data, match, 4)
		if locale == "" {
			locale = matchString(data, match, 6)
		}
		declarations[name] = append(declarations[name], dateTimeDeclaration{
			index:   match[0],
			locale:  locale,
			options: parseOptionsObject(matchString(data, match, 8)),
		})
	}
	return declarations
}

func latestDateTimeDeclarationBefore(declarations []dateTimeDeclaration, index int) (dateTimeDeclaration, bool) {
	for i := len(declarations) - 1; i >= 0; i-- {
		if declarations[i].index < index {
			return declarations[i], true
		}
	}
	return dateTimeDeclaration{}, false
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
			dates[name] = append(dates[name], dateVariable{index: match[0], value: time.UnixMilli(ms).UTC().Format(time.RFC3339)})
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
	if value, ok := parseNumberLiteral(expr); ok {
		switch value := value.(type) {
		case int64:
			return time.UnixMilli(value).UTC().Format(time.RFC3339), true
		case uint64:
			if value <= uint64(1<<63-1) {
				return time.UnixMilli(int64(value)).UTC().Format(time.RFC3339), true
			}
		}
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
	value, ok := parseNumberLiteral(raw)
	if !ok {
		return "", false
	}
	switch value := value.(type) {
	case int64:
		return time.UnixMilli(value).UTC().Format(time.RFC3339), true
	case uint64:
		if value <= uint64(1<<63-1) {
			return time.UnixMilli(int64(value)).UTC().Format(time.RFC3339), true
		}
	}
	return "", false
}

func parseDateString(raw string) (string, bool) {
	value := strings.TrimSpace(raw)
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC().Format(time.RFC3339), true
	}
	layouts := []string{"2006-1-2", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.UTC); err == nil {
			return parsed.UTC().Format(time.RFC3339), true
		}
	}
	return "", false
}

func parseDateYMD(fields []string) (string, bool) {
	if len(fields) < 3 {
		return "", false
	}
	values := make([]int, 6)
	for i := range values {
		values[i] = 0
	}
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
	instant := time.Date(values[0], time.Month(values[1]+1), values[2], values[3], values[4], values[5], 0, time.UTC)
	return instant.UTC().Format(time.RFC3339), true
}

func parseDateUTCArgs(raw string) (string, bool) {
	fields := strings.Split(raw, ",")
	args := make([]string, 0, len(fields))
	for _, field := range fields {
		args = append(args, strings.TrimSpace(field))
	}
	return parseDateYMD(args)
}

func newDateTimeFormatFixture(rel string, index int, locale string, options map[string]any, input string, expected string) fixture {
	source := formatJSDateTimeFormatTestSourcePrefix + rel
	return fixture{
		ID:       fmt.Sprintf("datetimeformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:   source,
		Locale:   locale,
		Options:  options,
		Input:    input,
		Expected: ptr(expected),
	}
}

func newDateTimeFormatRangeFixture(rel string, index int, locale string, options map[string]any, start, end, expected string) fixture {
	source := formatJSDateTimeFormatTestSourcePrefix + rel
	return fixture{
		ID:            fmt.Sprintf("datetimeformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:        source,
		Locale:        locale,
		Feature:       "formatRange",
		Options:       options,
		Input:         map[string]any{"start": start, "end": end},
		ExpectedRange: ptr(expected),
	}
}

func newDateTimeFormatRangePartsFixture(rel string, index int, locale string, options map[string]any, start, end string, parts []rangePart) fixture {
	source := formatJSDateTimeFormatTestSourcePrefix + rel
	return fixture{
		ID:                 fmt.Sprintf("datetimeformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:             source,
		Locale:             locale,
		Feature:            "formatRangeToParts",
		Options:            options,
		Input:              map[string]any{"start": start, "end": end},
		ExpectedRange:      ptr(joinRangePartValues(parts)),
		ExpectedRangeParts: parts,
	}
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
			str, ok := value.(string)
			if !ok || !oneOf(str, "numeric", "2-digit", "narrow", "short", "long") {
				return false
			}
		case "month":
			str, ok := value.(string)
			if !ok || !oneOf(str, "numeric", "2-digit", "narrow", "short", "long") {
				return false
			}
		case "timeZoneName":
			str, ok := value.(string)
			if !ok || !oneOf(str, "short", "long", "shortOffset", "longOffset", "shortGeneric", "longGeneric") {
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
			str, ok := value.(string)
			if !ok || !oneOf(str, "h11", "h12", "h23", "h24") {
				return false
			}
		case "dateStyle", "timeStyle":
			str, ok := value.(string)
			if !ok || !oneOf(str, "full", "long", "medium", "short") {
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
		fixtures = append(fixtures, newLocaleFixture(formatJSLocaleTestSourcePrefix, rel, nextIndex, "canonicalize", input, expected))
		nextIndex++
	}
	for _, match := range localeMaximizeRE.FindAllStringSubmatchIndex(data, -1) {
		input := data[match[2]:match[3]]
		expected, ok := decodeJSString(data[match[4]:match[5]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newLocaleFixture(formatJSLocaleTestSourcePrefix, rel, nextIndex, "maximize", input, expected))
		nextIndex++
	}
	for _, match := range localeMinimizeRE.FindAllStringSubmatchIndex(data, -1) {
		input := data[match[2]:match[3]]
		expected, ok := decodeJSString(data[match[4]:match[5]])
		if !ok {
			continue
		}
		fixtures = append(fixtures, newLocaleFixture(formatJSLocaleTestSourcePrefix, rel, nextIndex, "minimize", input, expected))
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
		fixtures = append(fixtures, newLocaleFixture(formatJSCanonicalLocalesTestSourcePrefix, rel, nextIndex, "canonicalize", input, expected))
		nextIndex++
	}
	return fixtures
}

func newLocaleFixture(sourcePrefix, rel string, index int, feature, input, expected string) fixture {
	return fixture{
		ID:       fmt.Sprintf("locale-formatjs-%s-%03d", fixtureSlug(sourcePrefix+rel), index),
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
	return oneOf(f.Feature, "canonicalize", "maximize", "minimize")
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
		if isSkippedVitestAssertion(data, match[0]) || isConditionalVitestAssertion(data, match[0]) {
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
	source := formatJSListFormatTestSourcePrefix + rel
	return fixture{
		ID:       fmt.Sprintf("listformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:   source,
		Locale:   locale,
		Options:  options,
		Input:    input,
		Expected: ptr(expected),
	}
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
		str, ok := value.(string)
		if !ok {
			return false
		}
		switch key {
		case "type":
			if !oneOf(str, "conjunction", "disjunction", "unit") {
				return false
			}
		case "style":
			if !oneOf(str, "long", "short", "narrow") {
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
		if isSkippedVitestAssertion(data, match[0]) || isConditionalVitestAssertion(data, match[0]) {
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
	source := formatJSRelativeTimeFormatTestSourcePrefix + rel
	return fixture{
		ID:       fmt.Sprintf("relativetimeformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:   source,
		Locale:   locale,
		Options:  options,
		Input:    map[string]any{"value": value, "unit": unit},
		Expected: ptr(expected),
	}
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
	unit, ok := input["unit"].(string)
	if !ok || !oneOf(unit, "second", "seconds", "minute", "minutes", "hour", "hours", "day", "days", "week", "weeks", "month", "months", "quarter", "quarters", "year", "years") {
		return false
	}
	for key, value := range f.Options {
		str, ok := value.(string)
		if !ok {
			return false
		}
		switch key {
		case "style":
			if !oneOf(str, "long", "short", "narrow") {
				return false
			}
		case "numeric":
			if !oneOf(str, "always", "auto") {
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
		if isSkippedVitestAssertion(data, match[0]) || isConditionalVitestAssertion(data, match[0]) {
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
	source := formatJSDurationFormatTestSourcePrefix + rel
	return fixture{
		ID:       fmt.Sprintf("durationformat-formatjs-%s-%03d", fixtureSlug(rel), index),
		Source:   source,
		Locale:   locale,
		Options:  options,
		Input:    input,
		Expected: ptr(expected),
	}
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
		if !oneOf(key, "years", "months", "weeks", "days", "hours", "minutes", "seconds", "milliseconds", "microseconds", "nanoseconds") {
			return false
		}
		if _, ok := value.(int64); !ok {
			return false
		}
	}
	for key, value := range f.Options {
		switch key {
		case "localeMatcher":
			str, ok := value.(string)
			if !ok || !oneOf(str, "lookup", "best fit") {
				return false
			}
		case "numberingSystem":
			if _, ok := value.(string); !ok {
				return false
			}
		case "style":
			str, ok := value.(string)
			if !ok || !oneOf(str, "long", "short", "narrow", "digital") {
				return false
			}
		case "years", "months", "weeks", "days", "hours", "minutes", "seconds", "milliseconds", "microseconds", "nanoseconds":
			str, ok := value.(string)
			if !ok || !oneOf(str, "long", "short", "narrow", "numeric", "2-digit") {
				return false
			}
		case "yearsDisplay", "monthsDisplay", "weeksDisplay", "daysDisplay", "hoursDisplay", "minutesDisplay", "secondsDisplay", "millisecondsDisplay", "microsecondsDisplay", "nanosecondsDisplay":
			str, ok := value.(string)
			if !ok || !oneOf(str, "always", "auto") {
				return false
			}
		case "fractionalDigits":
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

func isPluralCategory(category string) bool {
	switch category {
	case "zero", "one", "two", "few", "many", "other":
		return true
	default:
		return false
	}
}

func parsePartArray(raw string) ([]fixturePart, bool) {
	matches := partObjectRE.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil, false
	}
	parts := make([]fixturePart, 0, len(matches))
	for _, match := range matches {
		values := objectStringFields(match[1])
		typ, hasType := values["type"]
		value, hasValue := values["value"]
		if !hasType || !hasValue {
			return nil, false
		}
		parts = append(parts, fixturePart{Type: typ, Value: value})
	}
	return parts, true
}

func parseRangePartArray(raw string) ([]rangePart, bool) {
	matches := partObjectRE.FindAllStringSubmatch(raw, -1)
	if len(matches) == 0 {
		return nil, false
	}
	parts := make([]rangePart, 0, len(matches))
	for _, match := range matches {
		values := objectStringFields(match[1])
		typ, hasType := values["type"]
		value, hasValue := values["value"]
		source, hasSource := values["source"]
		if !hasType || !hasValue || !hasSource {
			return nil, false
		}
		parts = append(parts, rangePart{Type: typ, Value: value, Source: source})
	}
	return parts, true
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

func joinPartValues(parts []fixturePart) string {
	var b strings.Builder
	for _, part := range parts {
		b.WriteString(part.Value)
	}
	return b.String()
}

func joinRangePartValues(parts []rangePart) string {
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
		if !oneOf(name, "years", "months", "weeks", "days", "hours", "minutes", "seconds", "milliseconds", "microseconds", "nanoseconds") {
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
