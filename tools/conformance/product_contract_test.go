package conformance

import (
	"encoding/json/v2"
	"path/filepath"
	"runtime"
	"testing"
)

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

func TestNumberFormatRangeNodeContractsExist(t *testing.T) {
	t.Parallel()

	fixtures := loadPackageFixtures(t, "numberformat")
	byID := fixturesByID(fixtures)
	for id, want := range map[string]string{
		nodeWitnessFixtureID("numberformat", "czech-plural-range-unit"):          "2–1 metrů",
		nodeWitnessFixtureID("numberformat", "czech-plural-range-currency-name"): "2–1 amerických dolarů",
		nodeWitnessFixtureID("numberformat", "negative-percent-range-affixes"):   "-1–2%",
	} {
		fixture := requireFixtureByID(t, byID, id, "NumberFormat range node contract")
		if fixture.ExpectedRange == nil || *fixture.ExpectedRange != want {
			t.Fatalf("fixture %q expectedRange = %v, want %q", id, fixture.ExpectedRange, want)
		}
		if len(fixture.ExpectedRangeParts) == 0 {
			t.Fatalf("fixture %q missing expectedRangeParts", id)
		}
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
