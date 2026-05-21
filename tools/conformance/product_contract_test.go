package conformance

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/agentable/go-intl/internal/segmentation"
)

func TestSegmenterNodeContractCoversSupportedLocales(t *testing.T) {
	t.Parallel()

	fixtures, err := LoadFixtures(filepath.Join(repositoryRoot(t), "segmenter"))
	if err != nil {
		t.Fatal(err)
	}
	coverage := map[string]map[string]bool{}
	for _, fixture := range fixtures {
		if fixture.Source != "node:v26.0.0:segmenter:locale-contract" {
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

	fixtures, err := LoadFixtures(filepath.Join(repositoryRoot(t), "collator"))
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]Fixture, len(fixtures))
	for _, fixture := range fixtures {
		byID[fixture.ID] = fixture
	}
	for _, id := range []string{
		"collator-node-v26-search-usage-contract",
		"collator-node-v26-case-first-upper-contract",
		"collator-node-v26-case-first-lower-contract",
		"collator-node-v26-locale-case-first-upper-contract",
		"collator-node-v26-german-phonebook-option-contract",
		"collator-node-v26-german-phonebook-locale-contract",
	} {
		fixture, ok := byID[id]
		if !ok {
			t.Fatalf("missing collator node option contract fixture %q", id)
		}
		if fixture.ExpectedComparison == nil {
			t.Fatalf("fixture %q missing expectedComparison", id)
		}
		if len(fixture.ExpectedResolved) == 0 {
			t.Fatalf("fixture %q missing expectedResolvedOptions", id)
		}
	}
}

func TestDateTimeFormatNodeDeepContractsExist(t *testing.T) {
	t.Parallel()

	fixtures, err := LoadFixtures(filepath.Join(repositoryRoot(t), "datetimeformat"))
	if err != nil {
		t.Fatal(err)
	}
	const source = "node:v26.0.0:datetimeformat:p4-deep-contract"
	byID := map[string]Fixture{}
	timeZoneNameForms := map[string]bool{}
	for _, fixture := range fixtures {
		if fixture.Source != source {
			continue
		}
		byID[fixture.ID] = fixture
		if len(fixture.ExpectedResolved) == 0 {
			t.Fatalf("fixture %q missing expectedResolvedOptions", fixture.ID)
		}
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
		"datetimeformat-node-v26-range-date-time-utc-h23",
		"datetimeformat-node-v26-range-time-utc-h23",
		"datetimeformat-node-v26-range-shared-month-prefix",
		"datetimeformat-node-v26-tz-metazone-london-winter",
		"datetimeformat-node-v26-tz-metazone-london-summer",
		"datetimeformat-node-v26-tz-offset-string-kolkata",
	} {
		fixture, ok := byID[id]
		if !ok {
			t.Fatalf("missing DateTimeFormat node deep contract fixture %q", id)
		}
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

	fixtures, err := LoadFixtures(filepath.Join(repositoryRoot(t), "numberformat"))
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]Fixture, len(fixtures))
	for _, fixture := range fixtures {
		byID[fixture.ID] = fixture
	}
	for _, id := range []string{
		"numberformat-manual-resolved-compact-defaults",
		"numberformat-manual-resolved-significant-digits-hide-fraction",
		"numberformat-manual-resolved-scientific-currency-defaults",
	} {
		fixture, ok := byID[id]
		if !ok {
			t.Fatalf("missing NumberFormat resolved-options contract fixture %q", id)
		}
		if len(fixture.ExpectedResolved) == 0 {
			t.Fatalf("fixture %q missing expectedResolvedOptions", id)
		}
	}
}

func TestDisplayNamesResolvedOptionContractsExist(t *testing.T) {
	t.Parallel()

	fixtures, err := LoadFixtures(filepath.Join(repositoryRoot(t), "displaynames"))
	if err != nil {
		t.Fatal(err)
	}
	byID := make(map[string]Fixture, len(fixtures))
	for _, fixture := range fixtures {
		byID[fixture.ID] = fixture
	}
	for _, id := range []string{
		"displaynames-manual-resolved-region-omits-language-display",
		"displaynames-manual-resolved-language-defaults",
		"displaynames-manual-resolved-language-standard-short",
	} {
		fixture, ok := byID[id]
		if !ok {
			t.Fatalf("missing DisplayNames resolved-options contract fixture %q", id)
		}
		if len(fixture.ExpectedResolved) == 0 {
			t.Fatalf("fixture %q missing expectedResolvedOptions", id)
		}
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
