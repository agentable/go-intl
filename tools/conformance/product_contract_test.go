package conformance

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/segmentation"
)

func TestSegmenterNodeContractCoversSupportedLocales(t *testing.T) {
	t.Parallel()

	fixtures := loadPackageFixtures(t, "segmenter")
	coverage := map[string]map[string]bool{}
	localeContractSource := nodeWitnessSource(nodeSourceSegmenterLocaleContract)
	for _, fixture := range fixtures {
		if fixture.Source != localeContractSource {
			continue
		}
		var options struct {
			Granularity string `json:"granularity"`
		}
		if err := json.Unmarshal(fixture.Options, &options); err != nil {
			t.Fatal(err)
		}
		if options.Granularity != "word" && options.Granularity != "sentence" {
			continue
		}
		if coverage[fixture.Locale] == nil {
			coverage[fixture.Locale] = map[string]bool{}
		}
		coverage[fixture.Locale][options.Granularity] = true
	}
	for _, locale := range segmentation.SupportedLocales() {
		got := coverage[locale]
		if !got["word"] || !got["sentence"] {
			t.Fatalf("segmenter locale %q node contract coverage = %v, want word and sentence", locale, got)
		}
	}
}

func TestCollatorNodeOptionContractsExistBeforeBackendAcceptance(t *testing.T) {
	t.Parallel()

	fixtures := loadPackageFixtures(t, "collator")
	byID := fixturesByID(fixtures)
	for _, id := range []string{
		nodeWitnessFixtureID("collator", "search-usage-contract"),
		nodeWitnessFixtureID("collator", "numeric-locale-extension-contract"),
		nodeWitnessFixtureID("collator", "case-first-upper-contract"),
		nodeWitnessFixtureID("collator", "case-first-lower-contract"),
		nodeWitnessFixtureID("collator", "locale-case-first-upper-contract"),
		nodeWitnessFixtureID("collator", "german-phonebook-option-contract"),
		nodeWitnessFixtureID("collator", "german-phonebook-locale-contract"),
	} {
		fixture := requireFixtureByID(t, byID, id, "collator node option contract")
		if fixture.ExpectedComparison == nil {
			t.Fatalf("fixture %q missing expectedComparison", id)
		}
		requireExpectedResolvedOptions(t, fixture)
	}
}

func TestSegmenterTailoredLocaleContractsRemainWithheld(t *testing.T) {
	t.Parallel()

	root := packageRoot(t, "segmenter")
	suite, err := loadRunSuite(root, time.Now())
	if err != nil {
		t.Fatalf("loadRunSuite(segmenter) error = %v", err)
	}
	byID := fixturesByID(suite.fixtures)
	for _, id := range []string{
		nodeWitnessFixtureID("segmenter", "th-word-tailored-contract"),
		nodeWitnessFixtureID("segmenter", "ja-word-tailored-contract"),
		nodeWitnessFixtureID("segmenter", "zh-hant-word-tailored-contract"),
	} {
		fixture := requireFixtureByID(t, byID, id, "segmenter tailored-locale node contract")
		if len(fixture.ExpectedSegments) == 0 {
			t.Fatalf("fixture %q missing expectedSegments", id)
		}
		if reason, ok := suite.skipReasons[id]; !ok || reason == "" {
			t.Fatalf("fixture %q must stay xfailed until the backend supports tailored word boundaries", id)
		}
	}
}

func TestSegmenterManualCapabilityBoundaryCoversTailoredLocales(t *testing.T) {
	t.Parallel()

	fixtures := loadPackageFixtures(t, "segmenter")

	boundary := requireFixtureByID(
		t,
		fixturesByID(fixtures),
		"segmenter-manual-supported-locales-excludes-tailored-locales",
		"segmenter manual supported-locales capability boundary",
	)
	if got, want := boundary.Feature, FeatureSupportedLocalesOf; got != want {
		t.Fatalf("segmenter manual capability boundary feature = %q, want %q", got, want)
	}
	if len(boundary.ExpectedLocales) != 1 || boundary.ExpectedLocales[0] != "en" {
		t.Fatalf("segmenter manual capability boundary expectedLocales = %v, want [en]", boundary.ExpectedLocales)
	}

	var requested []string
	if err := json.Unmarshal(boundary.Input, &requested); err != nil {
		t.Fatal(err)
	}
	requestedSet := make(map[string]bool, len(requested))
	for _, tag := range requested {
		requestedSet[tag] = true
	}
	for _, tag := range []string{"ja", "ja-JP", "km", "lo", "my", "th", "zh", "zh-Hans", "zh-Hans-CN", "zh-Hant", "zh-Hant-TW"} {
		if !requestedSet[tag] {
			t.Fatalf("segmenter manual capability boundary missing withheld locale %q", tag)
		}
	}
}

func TestDateTimeFormatNodeDeepContractsExist(t *testing.T) {
	t.Parallel()

	fixtures := loadPackageFixtures(t, "datetimeformat")
	source := nodeWitnessSource(nodeSourceDateTimeFormatDeepContract)
	byID := map[string]Fixture{}
	timeZoneNameForms := map[string]bool{}
	for _, fixture := range fixtures {
		if fixture.Source != source {
			continue
		}
		byID[fixture.ID] = fixture
		requireExpectedResolvedOptions(t, fixture)
		var options struct {
			TimeZoneName string `json:"timeZoneName"`
		}
		if err := json.Unmarshal(fixture.Options, &options); err != nil {
			t.Fatal(err)
		}
		if options.TimeZoneName != "" {
			timeZoneNameForms[options.TimeZoneName] = true
		}
	}

	for _, id := range []string{
		nodeWitnessFixtureID("datetimeformat", "range-date-time-utc-h23"),
		nodeWitnessFixtureID("datetimeformat", "range-time-utc-h23"),
		nodeWitnessFixtureID("datetimeformat", "range-shared-month-prefix"),
		nodeWitnessFixtureID("datetimeformat", "tz-metazone-london-winter"),
		nodeWitnessFixtureID("datetimeformat", "tz-metazone-london-summer"),
		nodeWitnessFixtureID("datetimeformat", "tz-offset-string-kolkata"),
	} {
		fixture := requireFixtureByID(t, byID, id, "DateTimeFormat node deep contract")
		if fixture.ExpectedRange == nil && fixture.Expected == nil {
			t.Fatalf("fixture %q missing expected output", id)
		}
	}

	for _, form := range []string{"short", "long", "shortGeneric", "longGeneric", "shortOffset", "longOffset"} {
		if !timeZoneNameForms[form] {
			t.Fatalf("DateTimeFormat node deep contracts missing timeZoneName form %q", form)
		}
	}
}

func TestNumberFormatResolvedOptionContractsExist(t *testing.T) {
	t.Parallel()

	fixtures := loadPackageFixtures(t, "numberformat")
	byID := fixturesByID(fixtures)
	for _, id := range []string{
		"numberformat-manual-resolved-compact-defaults",
		"numberformat-manual-resolved-significant-digits-hide-fraction",
		"numberformat-manual-resolved-scientific-currency-defaults",
	} {
		fixture := requireFixtureByID(t, byID, id, "NumberFormat resolved-options contract")
		requireExpectedResolvedOptions(t, fixture)
	}
}

func TestDisplayNamesResolvedOptionContractsExist(t *testing.T) {
	t.Parallel()

	fixtures := loadPackageFixtures(t, "displaynames")
	byID := fixturesByID(fixtures)
	for _, id := range []string{
		"displaynames-manual-resolved-region-omits-language-display",
		"displaynames-manual-resolved-language-defaults",
		"displaynames-manual-resolved-language-standard-short",
	} {
		fixture := requireFixtureByID(t, byID, id, "DisplayNames resolved-options contract")
		requireExpectedResolvedOptions(t, fixture)
	}
}

func loadPackageFixtures(t *testing.T, packageName string) []Fixture {
	t.Helper()

	fixtures, err := LoadFixtures(packageRoot(t, packageName))
	if err != nil {
		t.Fatal(err)
	}
	return fixtures
}

func packageRoot(t *testing.T, packageName string) string {
	t.Helper()

	return filepath.Join(repositoryRoot(t), packageName)
}

func requireExpectedResolvedOptions(t *testing.T, fixture Fixture) {
	t.Helper()

	if len(fixture.ExpectedResolved) == 0 {
		t.Fatalf("fixture %q missing expectedResolvedOptions", fixture.ID)
	}
}

func requireFixtureByID(t *testing.T, byID map[string]Fixture, id, label string) Fixture {
	t.Helper()

	fixture, ok := byID[id]
	if !ok {
		t.Fatalf("missing %s fixture %q", label, id)
	}
	return fixture
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
