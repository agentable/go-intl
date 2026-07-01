package main

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

func TestRunImportsSyntheticNodeWitnessFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nodePath := filepath.Join(root, "node")
	if err := os.WriteFile(nodePath, []byte(`#!/bin/sh
cat <<'JSON'
{
  "nodeVersion": "v26.0.0",
  "versions": {"node": "26.0.0", "v8": "14.0.365.4-node.2", "icu": "78.1", "cldr": "48.0", "tz": "2025b"},
  "localeSmoke": [
    {
      "id": "locale-node-v26-canonicalize",
      "source": "node:v26.0.0:locale",
      "locale": "EN-us-u-NU-LATN",
      "feature": "canonicalize",
      "options": {},
      "input": "EN-us-u-NU-LATN",
      "expected": "en-US-u-nu-latn"
    }
  ],
  "localeCanonicalization": [
    {
      "id": "locale-node-v26-duplicate-calendar-first-wins",
      "source": "node:v26.0.0:locale:canonicalization",
      "locale": "en-u-ca-buddhist-ca-gregory",
      "feature": "canonicalize",
      "options": {},
      "input": "en-u-ca-buddhist-ca-gregory",
      "expected": "en-u-ca-buddhist"
    }
  ],
  "localeInfo": [
    {
      "id": "locale-node-v26-week-info-rg-override",
      "source": "node:v26.0.0:locale:info",
      "locale": "en-US-u-rg-gbzzzz",
      "feature": "weekInfo",
      "options": {},
      "input": "en-US-u-rg-gbzzzz",
      "expectedResolvedOptions": {"firstDay": 1, "weekend": [6, 7]}
    }
  ],
  "numberFormatSmoke": [
    {
      "id": "numberformat-node-v26-currency-usd",
      "source": "node:v26.0.0:numberformat",
      "locale": "en-US",
      "options": {"style": "currency", "currency": "USD"},
      "input": 1234.5,
      "expected": "$1,234.50"
    }
  ],
  "numberFormatErrors": [
    {
      "id": "numberformat-node-v26-invalid-style",
      "source": "node:v26.0.0:numberformat:errors",
      "locale": "en-US",
      "options": {"style": "invalid"},
      "input": 1,
      "errorCode": "invalid_option"
    }
  ],
  "numberFormatResolved": [
    {
      "id": "numberformat-node-v26-resolved-decimal-default",
      "source": "node:v26.0.0:numberformat:resolved-options",
      "locale": "en",
      "options": {},
      "input": 12345.6,
      "expected": "12,345.6",
      "expectedResolvedOptions": {"locale": "en", "numberingSystem": "latn", "style": "decimal"}
    }
  ],
  "dateTimeFormatSmoke": [
    {
      "id": "datetimeformat-node-v26-utc-long-date",
      "source": "node:v26.0.0:datetimeformat",
      "locale": "en-US",
      "options": {"year": "numeric", "month": "long", "day": "numeric", "timeZone": "UTC"},
      "input": "2020-01-02T03:04:05Z",
      "expected": "January 2, 2020"
    }
  ],
  "dateTimeFormatErrors": [
    {
      "id": "datetimeformat-node-v26-invalid-date-style",
      "source": "node:v26.0.0:datetimeformat:errors",
      "locale": "en-US",
      "options": {"dateStyle": "bad"},
      "input": "2026-05-08T12:00:00Z",
      "errorCode": "invalid_option"
    }
  ],
  "dateTimeFormatEdge": [
    {
      "id": "datetimeformat-node-v26-offset-timezone-kolkata",
      "source": "node:v26.0.0:datetimeformat:edge",
      "locale": "en-US",
      "feature": "offsetTimeZone",
      "options": {"hour": "2-digit", "minute": "2-digit", "timeZone": "+05:30"},
      "input": "2021-01-10T12:00:00Z",
      "expected": "17:30",
      "expectedParts": [{"type": "hour", "value": "17"}],
      "expectedResolvedOptions": {"locale": "en-US"}
    }
  ],
  "durationFormatSmoke": [
    {
      "id": "durationformat-node-v26-hours-minutes",
      "source": "node:v26.0.0:durationformat",
      "locale": "en",
      "options": {},
      "input": {"hours": 1, "minutes": 2},
      "expected": "1 hr, 2 min"
    }
  ],
  "durationFormatErrors": [
    {
      "id": "durationformat-node-v26-invalid-style",
      "source": "node:v26.0.0:durationformat:errors",
      "locale": "en-US",
      "options": {"style": "bad"},
      "input": {},
      "errorCode": "invalid_option"
    }
  ],
  "durationFormatDigital": [
    {
      "id": "durationformat-node-v26-digital-hours-minutes-seconds",
      "source": "node:v26.0.0:durationformat:digital",
      "locale": "en",
      "options": {"style": "digital"},
      "input": {"hours": 5, "minutes": 30, "seconds": 15},
      "expected": "5:30:15",
      "expectedParts": [
        {"type": "integer", "value": "5", "unit": "hour"},
        {"type": "literal", "value": ":"},
        {"type": "integer", "value": "30", "unit": "minute"},
        {"type": "literal", "value": ":"},
        {"type": "integer", "value": "15", "unit": "second"}
      ],
      "expectedResolvedOptions": {"locale": "en", "numberingSystem": "latn", "style": "digital"}
    }
  ],
  "listFormatSmoke": [
    {
      "id": "listformat-node-v26-conjunction-long",
      "source": "node:v26.0.0:listformat",
      "locale": "en-US",
      "options": {},
      "input": ["A", "B", "C"],
      "expected": "A, B, and C"
    }
  ],
  "listFormatErrors": [
    {
      "id": "listformat-node-v26-invalid-style",
      "source": "node:v26.0.0:listformat:errors",
      "locale": "en-US",
      "options": {"style": "bad"},
      "input": [],
      "errorCode": "invalid_option"
    }
  ],
  "relativeTimeSmoke": [
    {
      "id": "relativetimeformat-node-v26-day-auto",
      "source": "node:v26.0.0:relativetimeformat",
      "locale": "en-US",
      "options": {"numeric": "auto"},
      "input": {"value": -1, "unit": "day"},
      "expected": "yesterday"
    }
  ],
  "relativeTimeErrors": [
    {
      "id": "relativetimeformat-node-v26-invalid-numeric",
      "source": "node:v26.0.0:relativetimeformat:errors",
      "locale": "en-US",
      "options": {"numeric": "bad"},
      "input": {"value": 1, "unit": "day"},
      "errorCode": "invalid_option"
    }
  ],
  "pluralRulesSmoke": [
    {
      "id": "pluralrules-node-v26-ordinal-two",
      "source": "node:v26.0.0:pluralrules",
      "locale": "en-US",
      "feature": "select",
      "options": {"type": "ordinal"},
      "input": 2,
      "expected": "two"
    }
  ],
  "pluralRulesErrors": [
    {
      "id": "pluralrules-node-v26-invalid-type",
      "source": "node:v26.0.0:pluralrules:errors",
      "locale": "en-US",
      "options": {"type": "bad"},
      "input": 1,
      "errorCode": "invalid_option"
    }
  ],
  "displayNamesSmoke": [
    {
      "id": "displaynames-node-v26-region-us",
      "source": "node:v26.0.0:displaynames",
      "locale": "en",
      "options": {"type": "region"},
      "input": "US",
      "expected": "United States",
      "expectedOk": true,
      "expectedResolvedOptions": {"locale": "en", "style": "long", "type": "region", "fallback": "code"}
    }
  ],
  "displayNamesErrors": [
    {
      "id": "displaynames-node-v26-invalid-type",
      "source": "node:v26.0.0:displaynames:errors",
      "locale": "en-US",
      "options": {"type": "bad"},
      "input": "en",
      "errorCode": "invalidOption"
    }
  ],
  "collatorSmoke": [
    {
      "id": "collator-node-v26-basic-order",
      "source": "node:v26.0.0:collator",
      "locale": "en",
      "options": {},
      "input": {"left": "a", "right": "b"},
      "expectedComparison": -1
    }
  ],
  "collatorErrors": [
    {
      "id": "collator-node-v26-invalid-sensitivity",
      "source": "node:v26.0.0:collator:errors",
      "locale": "en-US",
      "options": {"sensitivity": "bad"},
      "input": {"left": "a", "right": "b"},
      "errorCode": "invalidOption"
    }
  ],
  "collatorOptions": [
    {
      "id": "collator-node-v26-numeric-locale-extension-contract",
      "source": "node:v26.0.0:collator:option-contract",
      "locale": "en-u-kn-true",
      "options": {},
      "input": {"left": "item2", "right": "item10"},
      "expectedComparison": -1,
      "expectedResolvedOptions": {"locale": "en-u-kn", "usage": "sort", "sensitivity": "variant", "ignorePunctuation": false, "collation": "default", "numeric": true, "caseFirst": "false"}
    }
  ],
  "collatorBackendProof": [
    {
      "id": "collator-node-v26-swedish-z-before-a-ring",
      "source": "node:v26.0.0:collator:backend-proof",
      "locale": "sv",
      "options": {},
      "input": {"left": "z", "right": "å"},
      "expectedComparison": -1,
      "expectedResolvedOptions": {"locale": "sv", "usage": "sort", "sensitivity": "variant", "ignorePunctuation": false, "collation": "default", "numeric": false, "caseFirst": "false"}
    }
  ],
  "segmenterSmoke": [
    {
      "id": "segmenter-node-v26-word-hello-world",
      "source": "node:v26.0.0:segmenter",
      "locale": "en",
      "options": {"granularity": "word"},
      "input": "Hello",
      "expectedSegments": [{"segment": "Hello", "codeUnitIndex": 0, "isWordLike": true}]
    }
  ],
  "segmenterErrors": [
    {
      "id": "segmenter-node-v26-invalid-granularity",
      "source": "node:v26.0.0:segmenter:errors",
      "locale": "en-US",
      "options": {"granularity": "bad"},
      "input": "hello",
      "errorCode": "invalid_option"
    }
  ],
  "segmenterLocale": [
    {
      "id": "segmenter-node-v26-en-word-contract",
      "source": "node:v26.0.0:segmenter:locale-contract",
      "locale": "en",
      "options": {"granularity": "word"},
      "input": "Hello",
      "expectedSegments": [{"segment": "Hello", "codeUnitIndex": 0, "isWordLike": true}]
    }
  ],
  "segmenterTailored": [
    {
      "id": "segmenter-node-v26-th-word-tailored-contract",
      "source": "node:v26.0.0:segmenter:tailored-locale-contract",
      "locale": "th",
      "options": {"granularity": "word"},
      "input": "ภาษาไทย",
      "expectedSegments": [{"segment": "ภาษา", "codeUnitIndex": 0, "isWordLike": true}]
    }
  ],
  "supportedValues": {
    "source": "node:v26.0.0:intl:supportedValuesOf",
    "versions": {"node": "26.0.0", "icu": "78.1", "cldr": "48.0", "tz": "2025b"},
    "values": {
      "calendar": ["gregory", "iso8601"],
      "collation": ["emoji"],
      "unit": ["acre"]
    }
  }
}
JSON
`), 0o777); err != nil {
		t.Fatalf("write fake node executable: %v", err)
	}
	out := filepath.Join(root, "out")

	if err := run([]string{"-node", nodePath, "-out", out}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	assertFileContainsAll(t, "locale smoke witness", filepath.Join(out, "locale", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "locale-node-v26-canonicalize"`,
		`"source": "node:v26.0.0:locale"`,
		`"expected": "en-US-u-nu-latn"`,
	)
	assertFileContainsAll(t, "locale canonicalization witness", filepath.Join(out, "locale", "testdata", "conformance", "node-v26", "canonicalization.json"),
		`"id": "locale-node-v26-duplicate-calendar-first-wins"`,
		`"source": "node:v26.0.0:locale:canonicalization"`,
	)
	assertFileContainsAll(t, "locale info witness", filepath.Join(out, "locale", "testdata", "conformance", "node-v26", "info.json"),
		`"id": "locale-node-v26-week-info-rg-override"`,
		`"source": "node:v26.0.0:locale:info"`,
		`"expectedResolvedOptions"`,
	)
	assertFileContainsAll(t, "number witness fixture", filepath.Join(out, "numberformat", "testdata", "conformance", "node-v26", "resolved-options.json"),
		`"id": "numberformat-node-v26-resolved-decimal-default"`,
		`"source": "node:v26.0.0:numberformat:resolved-options"`,
		`"expectedResolvedOptions"`,
	)
	assertFileContainsAll(t, "number smoke witness fixture", filepath.Join(out, "numberformat", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "numberformat-node-v26-currency-usd"`,
		`"source": "node:v26.0.0:numberformat"`,
		`"expected": "$1,234.50"`,
	)
	assertFileContainsAll(t, "number error witness fixture", filepath.Join(out, "numberformat", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "numberformat-node-v26-invalid-style"`,
		`"source": "node:v26.0.0:numberformat:errors"`,
		`"errorCode": "invalid_option"`,
	)
	assertFileContainsAll(t, "date time smoke witness fixture", filepath.Join(out, "datetimeformat", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "datetimeformat-node-v26-utc-long-date"`,
		`"source": "node:v26.0.0:datetimeformat"`,
		`"expected": "January 2, 2020"`,
	)
	assertFileContainsAll(t, "date time error witness fixture", filepath.Join(out, "datetimeformat", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "datetimeformat-node-v26-invalid-date-style"`,
		`"source": "node:v26.0.0:datetimeformat:errors"`,
		`"errorCode": "invalid_option"`,
	)
	assertFileContainsAll(t, "date time edge witness fixture", filepath.Join(out, "datetimeformat", "testdata", "conformance", "node-v26", "edge.json"),
		`"id": "datetimeformat-node-v26-offset-timezone-kolkata"`,
		`"source": "node:v26.0.0:datetimeformat:edge"`,
		`"expectedParts"`,
	)
	assertFileContainsAll(t, "duration witness fixture", filepath.Join(out, "durationformat", "testdata", "conformance", "node-v26", "digital.json"),
		`"id": "durationformat-node-v26-digital-hours-minutes-seconds"`,
		`"expectedParts"`,
		`"unit": "hour"`,
		`"expectedResolvedOptions"`,
	)
	assertFileContainsAll(t, "duration smoke witness fixture", filepath.Join(out, "durationformat", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "durationformat-node-v26-hours-minutes"`,
		`"source": "node:v26.0.0:durationformat"`,
		`"expected": "1 hr, 2 min"`,
	)
	assertFileContainsAll(t, "duration error witness fixture", filepath.Join(out, "durationformat", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "durationformat-node-v26-invalid-style"`,
		`"source": "node:v26.0.0:durationformat:errors"`,
		`"errorCode": "invalid_option"`,
	)
	assertFileContainsAll(t, "list smoke witness fixture", filepath.Join(out, "listformat", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "listformat-node-v26-conjunction-long"`,
		`"source": "node:v26.0.0:listformat"`,
		`"expected": "A, B, and C"`,
	)
	assertFileContainsAll(t, "list error witness fixture", filepath.Join(out, "listformat", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "listformat-node-v26-invalid-style"`,
		`"source": "node:v26.0.0:listformat:errors"`,
		`"errorCode": "invalid_option"`,
	)
	assertFileContainsAll(t, "relative time smoke witness fixture", filepath.Join(out, "relativetimeformat", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "relativetimeformat-node-v26-day-auto"`,
		`"source": "node:v26.0.0:relativetimeformat"`,
		`"expected": "yesterday"`,
	)
	assertFileContainsAll(t, "relative time error witness fixture", filepath.Join(out, "relativetimeformat", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "relativetimeformat-node-v26-invalid-numeric"`,
		`"source": "node:v26.0.0:relativetimeformat:errors"`,
		`"errorCode": "invalid_option"`,
	)
	assertFileContainsAll(t, "plural rules smoke witness fixture", filepath.Join(out, "pluralrules", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "pluralrules-node-v26-ordinal-two"`,
		`"source": "node:v26.0.0:pluralrules"`,
		`"expected": "two"`,
	)
	assertFileContainsAll(t, "plural rules error witness fixture", filepath.Join(out, "pluralrules", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "pluralrules-node-v26-invalid-type"`,
		`"source": "node:v26.0.0:pluralrules:errors"`,
		`"errorCode": "invalid_option"`,
	)
	assertFileContainsAll(t, "display names smoke witness fixture", filepath.Join(out, "displaynames", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "displaynames-node-v26-region-us"`,
		`"source": "node:v26.0.0:displaynames"`,
		`"expected": "United States"`,
		`"expectedOk": true`,
		`"expectedResolvedOptions"`,
	)
	assertFileContainsAll(t, "display names error witness fixture", filepath.Join(out, "displaynames", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "displaynames-node-v26-invalid-type"`,
		`"source": "node:v26.0.0:displaynames:errors"`,
		`"errorCode": "invalidOption"`,
	)
	assertFileContainsAll(t, "collator smoke witness fixture", filepath.Join(out, "collator", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "collator-node-v26-basic-order"`,
		`"source": "node:v26.0.0:collator"`,
		`"expectedComparison": -1`,
	)
	assertFileContainsAll(t, "collator error witness fixture", filepath.Join(out, "collator", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "collator-node-v26-invalid-sensitivity"`,
		`"source": "node:v26.0.0:collator:errors"`,
		`"errorCode": "invalidOption"`,
	)
	assertFileContainsAll(t, "collator option witness fixture", filepath.Join(out, "collator", "testdata", "conformance", "node-v26", "options.json"),
		`"id": "collator-node-v26-numeric-locale-extension-contract"`,
		`"source": "node:v26.0.0:collator:option-contract"`,
		`"expectedResolvedOptions"`,
	)
	assertFileContainsAll(t, "collator backend proof witness fixture", filepath.Join(out, "collator", "testdata", "conformance", "node-v26", "backend-proof.json"),
		`"id": "collator-node-v26-swedish-z-before-a-ring"`,
		`"source": "node:v26.0.0:collator:backend-proof"`,
		`"expectedResolvedOptions"`,
	)
	assertFileContainsAll(t, "segmenter smoke witness fixture", filepath.Join(out, "segmenter", "testdata", "conformance", "node-v26", "smoke.json"),
		`"id": "segmenter-node-v26-word-hello-world"`,
		`"source": "node:v26.0.0:segmenter"`,
		`"expectedSegments"`,
	)
	assertFileContainsAll(t, "segmenter error witness fixture", filepath.Join(out, "segmenter", "testdata", "conformance", "node-v26", "errors.json"),
		`"id": "segmenter-node-v26-invalid-granularity"`,
		`"source": "node:v26.0.0:segmenter:errors"`,
		`"errorCode": "invalid_option"`,
	)
	assertFileContainsAll(t, "segmenter locale contract fixture", filepath.Join(out, "segmenter", "testdata", "conformance", "node-v26", "locale-contract.json"),
		`"id": "segmenter-node-v26-en-word-contract"`,
		`"source": "node:v26.0.0:segmenter:locale-contract"`,
		`"expectedSegments"`,
	)
	assertFileContainsAll(t, "segmenter tailored contract fixture", filepath.Join(out, "segmenter", "testdata", "conformance", "node-v26", "tailored-locale-contract.json"),
		`"id": "segmenter-node-v26-th-word-tailored-contract"`,
		`"source": "node:v26.0.0:segmenter:tailored-locale-contract"`,
		`"expectedSegments"`,
	)
	assertFileContainsAll(t, "supported-values witness", filepath.Join(out, "testdata", "native", "node-v26", "supported-values.json"),
		`"source": "node:v26.0.0:intl:supportedValuesOf"`,
		`"calendar"`,
		`"collation"`,
		`"unit"`,
		`"icu": "78.1"`,
	)
	assertPathAbsent(t, "node-only import .skip-list.json stat error", filepath.Join(out, ".skip-list.json"))
}

func TestRunWithoutReferencesWritesMissingFormatJSSkip(t *testing.T) {
	t.Parallel()

	out := t.TempDir()
	if err := run([]string{"-out", out}); err != nil {
		t.Fatalf("run(-out) error = %v", err)
	}

	got := mustReadSkips(t, filepath.Join(out, ".skip-list.json"))
	want := []skipEntry{missingReferenceSkip("formatjs", skipReasonFormatJSPathNotProvided)}
	assertSkipsEqual(t, "run(-out)", got, want)
}

func TestWriteSkipsSortsAndWritesStableJSON(t *testing.T) {
	t.Parallel()

	out := t.TempDir()
	path := filepath.Join(out, ".skip-list.json")
	if err := writeSkips(out, []skipEntry{
		{Source: "formatjs", Category: "z-last", Reason: "later"},
		{Source: "node", Category: "a-native", Reason: "native"},
		{Source: "formatjs", Category: "a-first", Reason: "second reason", DivergenceID: "D2"},
		{Source: "formatjs", Category: "a-first", Reason: "first reason", DivergenceID: "D1"},
	}); err != nil {
		t.Fatalf("writeSkips() error = %v", err)
	}

	raw := mustReadFile(t, path)
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		t.Fatalf("writeSkips() output = %q, want trailing newline", raw)
	}

	got := mustReadSkips(t, path)
	want := []skipEntry{
		{Source: "formatjs", Category: "a-first", Reason: "first reason", DivergenceID: "D1"},
		{Source: "formatjs", Category: "a-first", Reason: "second reason", DivergenceID: "D2"},
		{Source: "formatjs", Category: "z-last", Reason: "later"},
		{Source: "node", Category: "a-native", Reason: "native"},
	}
	assertSkipsEqual(t, "writeSkips()", got, want)
}

func TestWriteFixturesBySourceRoutesStableFiles(t *testing.T) {
	t.Parallel()

	out := t.TempDir()
	fixtures := []fixture{
		{ID: "list-b", Source: formatJSListFormatTestSourcePrefix + "src/lists/basic.test.ts", Locale: "en", Options: map[string]any{}, Input: []string{"b"}, Expected: ptr("B")},
		{ID: "other-a", Source: formatJSListFormatTestSourcePrefix + "other.case.ts", Locale: "en", Options: map[string]any{}, Input: []string{"a"}, Expected: ptr("A")},
		{ID: "list-a", Source: formatJSListFormatTestSourcePrefix + "src/lists/basic.test.ts", Locale: "en", Options: map[string]any{}, Input: []string{"a"}, Expected: ptr("A")},
	}

	if err := writeFixturesBySourceSlug(out, fixtures, formatJSListFormatTestSourcePrefix, fixtureSlug); err != nil {
		t.Fatalf("writeFixturesBySourceSlug() error = %v", err)
	}

	basicPath := filepath.Join(out, "src-lists-basic-test-ts.json")
	basic := mustReadFixtures(t, basicPath)
	assertFixtureIDs(t, "writeFixturesBySourceSlug() "+filepath.Base(basicPath), basic, "list-a", "list-b")
	otherPath := filepath.Join(out, "other-case-ts.json")
	other := mustReadFixtures(t, otherPath)
	assertFixtureIDs(t, "writeFixturesBySourceSlug() "+filepath.Base(otherPath), other, "other-a")
}

func TestImportFormatJSLocaleFixturesFiltersGeneratedAndRecordsSkips(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	out := t.TempDir()
	files := map[string]string{
		"supported.test.ts":       "expect(locale.toString()).toBe('en-US')",
		"unsupported.test.ts":     "expect(locale.weekInfo).toEqual({})",
		"ignored.snap":            "expect(snapshot).toMatchInlineSnapshot()",
		"nested/also.test.ts":     "no unsupported expectation here",
		"nested/ignored.fixture":  "expect(nonTypescript).toBe(true)",
		"nested/ignored.test.jsx": "expect(nonTypescript).toBe(true)",
	}
	for rel, data := range files {
		path := filepath.Join(root, filepath.FromSlash(rel))
		mustWriteFile(t, path, data)
	}

	extract := func(rel, data string) []fixture {
		switch rel {
		case "supported.test.ts":
			return []fixture{
				{ID: "locale-supported-b", Source: formatJSLocaleTestSourcePrefix + rel, Locale: "en-US", Feature: fixtureFeatureMaximize, Options: map[string]any{}, Input: "en-US", Expected: ptr("en-Latn-US")},
				{ID: "locale-unsupported-week-info", Source: formatJSLocaleTestSourcePrefix + rel, Locale: "en-US", Feature: "weekInfo", Options: map[string]any{}, Input: "en-US", Expected: ptr("ignored")},
				{ID: "locale-supported-a", Source: formatJSLocaleTestSourcePrefix + rel, Locale: "en-US", Feature: fixtureFeatureCanonicalize, Options: map[string]any{}, Input: "EN-us", Expected: ptr("en-US")},
			}
		case "nested/also.test.ts":
			return []fixture{{
				ID:       "locale-nested-minimize",
				Source:   formatJSLocaleTestSourcePrefix + rel,
				Locale:   "fr",
				Feature:  fixtureFeatureMinimize,
				Options:  map[string]any{},
				Input:    "fr-Latn-FR",
				Expected: ptr("fr"),
			}}
		default:
			if data == "" {
				t.Fatalf("extract(%q) data unexpectedly empty", rel)
			}
			return nil
		}
	}
	slug := func(rel string) string {
		return "locale-" + fixtureSlug(rel)
	}

	skips, err := importFormatJSTestFixtures(root, out, formatJSImportSpec{
		packageDir: formatJSLocalePackageDir,
		extract:    extract,
		supports:   supportsGeneratedLocaleFixture,
		slug:       slug,
	})
	if err != nil {
		t.Fatalf("importFormatJSTestFixtures() error = %v", err)
	}

	wantSkips := []skipEntry{
		formatJSImportPartialExtractionSkip(formatJSLocaleTestSourcePrefix + "supported.test.ts"),
		formatJSUnsupportedExtractorShapeSkip(formatJSLocaleTestSourcePrefix + "unsupported.test.ts"),
	}
	assertSkipsMatch(t, "importFormatJSTestFixtures()", skips, wantSkips)

	supported := mustReadFixtures(t, filepath.Join(out, "locale-supported-test-ts.json"))
	assertFixtureIDs(t, "importFormatJSTestFixtures() supported", supported, "locale-supported-a", "locale-supported-b")
	nested := mustReadFixtures(t, filepath.Join(out, "locale-nested-also-test-ts.json"))
	assertFixtureIDs(t, "importFormatJSTestFixtures() nested", nested, "locale-nested-minimize")
}

func TestImportFormatJSLocalesWritesLocaleAndCanonicalFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	out := t.TempDir()
	formatJSRoot := filepath.Join(root, "formatjs")
	files := map[string]string{
		"packages/intl-locale/tests/locale.test.ts": `import {Locale} from '#packages/intl-locale'
import {expect, it} from 'vitest'

it('extracts Locale canonicalization', () => {
  expect(new Locale('EN-us-u-NU-LATN').toString()).toBe('en-US-u-nu-latn')
})
`,
		"packages/intl-getcanonicallocales/tests/canonical.test.ts": `import {getCanonicalLocales} from '#packages/intl-getcanonicallocales'
import {expect, it} from 'vitest'

it('extracts Intl.getCanonicalLocales', () => {
  expect(getCanonicalLocales('FR-fr')).toEqual(['fr-FR'])
})
`,
	}
	for rel, data := range files {
		path := filepath.Join(formatJSRoot, filepath.FromSlash(rel))
		mustWriteFile(t, path, data)
	}

	targetRoot := filepath.Join(out, "locale", "testdata", "conformance", "formatjs")
	stalePath := filepath.Join(targetRoot, "stale.json")
	mustWriteFile(t, stalePath, "[]\n")

	skips, err := importFormatJSLocales(formatJSRoot, out)
	if err != nil {
		t.Fatalf("importFormatJSLocales() error = %v", err)
	}
	if len(skips) != 0 {
		t.Fatalf("importFormatJSLocales() skips = %+v, want none", skips)
	}
	assertPathAbsent(t, "importFormatJSLocales() stale file stat error", stalePath)

	localePath := filepath.Join(targetRoot, "locale-test-ts.json")
	localeFixtures := mustReadFixtures(t, localePath)
	wantLocaleID := formatJSFixtureID("locale", formatJSLocaleTestSourcePrefix+"locale.test.ts", 0)
	assertFixtureIDs(t, "importFormatJSLocales() "+filepath.Base(localePath), localeFixtures, wantLocaleID)

	canonicalPath := filepath.Join(targetRoot, "intl-getcanonicallocales-canonical-test-ts.json")
	canonicalFixtures := mustReadFixtures(t, canonicalPath)
	wantCanonicalID := formatJSFixtureID("locale", formatJSCanonicalLocalesTestSourcePrefix+"canonical.test.ts", 0)
	assertFixtureIDs(t, "importFormatJSLocales() "+filepath.Base(canonicalPath), canonicalFixtures, wantCanonicalID)
}

func TestImportFormatJSRecordsMissingPackageReferences(t *testing.T) {
	t.Parallel()

	formatJSRoot := t.TempDir()
	out := t.TempDir()
	stalePath := filepath.Join(out, "numberformat", "testdata", "conformance", "formatjs", "stale.json")
	mustWriteFile(t, stalePath, "[]\n")

	skips, err := importFormatJS(formatJSRoot, out)
	if err != nil {
		t.Fatalf("importFormatJS() error = %v", err)
	}
	want := []skipEntry{
		missingReferenceSkip(formatJSNumberFormatTestSourcePrefix, skipReasonFormatJSTestsPathNotFound),
		missingReferenceSkip(formatJSPluralRulesTestSourcePrefix, skipReasonFormatJSTestsPathNotFound),
		missingReferenceSkip(formatJSDateTimeFormatTestSourcePrefix, skipReasonFormatJSTestsPathNotFound),
		missingReferenceSkip(formatJSLocaleTestSourcePrefix, skipReasonFormatJSTestsPathNotFound),
		missingReferenceSkip(formatJSCanonicalLocalesTestSourcePrefix, skipReasonFormatJSTestsPathNotFound),
		missingReferenceSkip(formatJSListFormatTestSourcePrefix, skipReasonFormatJSTestsPathNotFound),
		missingReferenceSkip(formatJSRelativeTimeFormatTestSourcePrefix, skipReasonFormatJSTestsPathNotFound),
		missingReferenceSkip(formatJSDurationFormatTestSourcePrefix, skipReasonFormatJSTestsPathNotFound),
	}
	assertSkipsEqual(t, "importFormatJS()", skips, want)
	assertPathAbsent(t, "importFormatJS() stale file stat error", stalePath)
}

type formatJSImportCase struct {
	packageDir    string
	targetPackage string
	sourcePrefix  string
	mixedSource   string
	wantFixtureID string
}

type formatJSImportFunc func(formatJSRoot, out string) ([]skipEntry, error)

func TestImportFormatJSNumberFormatRecordsSkips(t *testing.T) {
	t.Parallel()

	tc := formatJSImportCase{
		packageDir:    formatJSNumberFormatPackageDir,
		targetPackage: formatJSNumberFormatTargetPackage,
		sourcePrefix:  formatJSNumberFormatTestSourcePrefix,
		mixedSource: `import {describe, expect, it} from 'vitest'

describe('NumberFormat import gate', () => {
  it('keeps only generated fixture support', () => {
    expect(new Intl.NumberFormat('en', {style: 'currency', currency: 'USD'}).format(42)).toBe('$42.00')
    expect(new Intl.NumberFormat('fr', {style: 'currency', currency: 'EUR'}).format(42)).toBe('42,00 €')
  })
})
`,
		wantFixtureID: formatJSFixtureID(formatJSNumberFormatTargetPackage, "mixed.test.ts", 0),
	}
	runFormatJSImportCase(t, tc, "importFormatJSNumberFormat", importFormatJSNumberFormat)
}

func TestImportFormatJSPluralRulesRecordsSkips(t *testing.T) {
	t.Parallel()

	tc := formatJSImportCase{
		packageDir:    formatJSPluralRulesPackageDir,
		targetPackage: formatJSPluralRulesTargetPackage,
		sourcePrefix:  formatJSPluralRulesTestSourcePrefix,
		mixedSource: `import {PluralRules} from '#packages/intl-pluralrules'
import {describe, expect, it} from 'vitest'

describe('PluralRules import gate', () => {
  it('keeps only generated fixture support', () => {
    expect(new PluralRules('en', {type: 'ordinal'}).select(2)).toBe('two')
    expect(new PluralRules('de').select(1)).toBe('one')
  })
})
`,
		wantFixtureID: formatJSFixtureID(formatJSPluralRulesTargetPackage, "mixed.test.ts", 0),
	}
	runFormatJSImportCase(t, tc, "importFormatJSPluralRules", importFormatJSPluralRules)
}

func TestImportFormatJSDateTimeFormatRecordsSkips(t *testing.T) {
	t.Parallel()

	tc := formatJSImportCase{
		packageDir:    formatJSDateTimeFormatPackageDir,
		targetPackage: formatJSDateTimeFormatTargetPackage,
		sourcePrefix:  formatJSDateTimeFormatTestSourcePrefix,
		mixedSource: `import {DateTimeFormat} from '#packages/intl-datetimeformat/core'
import {describe, expect, it} from 'vitest'

describe('DateTimeFormat import gate', () => {
  const start = new Date('2026-05-08T00:00:00Z')
  const en = new DateTimeFormat('en-US', {year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC'})
  const fr = new DateTimeFormat('fr', {year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC'})

  it('keeps only generated fixture support', () => {
    expect(en.format(start)).toBe('May 8, 2026')
    expect(fr.format(start)).toBe('8 mai 2026')
  })
})
`,
		wantFixtureID: formatJSFixtureID(formatJSDateTimeFormatTargetPackage, "mixed.test.ts", 0),
	}
	runFormatJSImportCase(t, tc, "importFormatJSDateTimeFormat", importFormatJSDateTimeFormat)
}

func TestImportFormatJSListFormatRecordsSkips(t *testing.T) {
	t.Parallel()

	tc := formatJSImportCase{
		packageDir:    formatJSListFormatPackageDir,
		targetPackage: formatJSListFormatTargetPackage,
		sourcePrefix:  formatJSListFormatTestSourcePrefix,
		mixedSource: `import {ListFormat} from '#packages/intl-listformat'
import {describe, expect, it} from 'vitest'

describe('ListFormat import gate', () => {
  it('keeps only generated fixture support', () => {
    expect(new ListFormat('en-AI', {type: 'unit'}).format(['A', 'B'])).toBe('A B')
    expect(new ListFormat('fr').format(['A', 'B'])).toBe('A et B')
  })
})
`,
		wantFixtureID: formatJSFixtureID(formatJSListFormatTargetPackage, "mixed.test.ts", 0),
	}
	runFormatJSImportCase(t, tc, "importFormatJSListFormat", importFormatJSListFormat)
}

func TestImportFormatJSRelativeTimeFormatRecordsSkips(t *testing.T) {
	t.Parallel()

	tc := formatJSImportCase{
		packageDir:    formatJSRelativeTimeFormatPackageDir,
		targetPackage: formatJSRelativeTimeFormatTargetPackage,
		sourcePrefix:  formatJSRelativeTimeFormatTestSourcePrefix,
		mixedSource: `import {RelativeTimeFormat} from '#packages/intl-relativetimeformat'
import {describe, expect, it} from 'vitest'

describe('RelativeTimeFormat import gate', () => {
  it('keeps only generated fixture support', () => {
    expect(new RelativeTimeFormat('en-AI', {numeric: 'auto'}).format(-1, 'day')).toBe('yesterday')
    expect(new RelativeTimeFormat('fr').format(-1, 'day')).toBe('hier')
  })
})
`,
		wantFixtureID: formatJSFixtureID(formatJSRelativeTimeFormatTargetPackage, "mixed.test.ts", 0),
	}
	runFormatJSImportCase(t, tc, "importFormatJSRelativeTimeFormat", importFormatJSRelativeTimeFormat)
}

func TestImportFormatJSDurationFormatRecordsSkips(t *testing.T) {
	t.Parallel()

	tc := formatJSImportCase{
		packageDir:    formatJSDurationFormatPackageDir,
		targetPackage: formatJSDurationFormatTargetPackage,
		sourcePrefix:  formatJSDurationFormatTestSourcePrefix,
		mixedSource: `import {DurationFormat} from '#packages/intl-durationformat'
import {describe, expect, it} from 'vitest'

describe('DurationFormat import gate', () => {
  it('keeps only generated fixture support', () => {
    expect(new DurationFormat('en', {style: 'digital'}).format({hours: 1, minutes: 2, seconds: 3})).toBe('1:02:03')
    expect(new DurationFormat('fr').format({hours: 1})).toBe('1 h')
  })
})
`,
		wantFixtureID: formatJSFixtureID(formatJSDurationFormatTargetPackage, "mixed.test.ts", 0),
	}
	runFormatJSImportCase(t, tc, "importFormatJSDurationFormat", importFormatJSDurationFormat)
}

func runFormatJSImportCase(t *testing.T, tc formatJSImportCase, importerName string, importer formatJSImportFunc) {
	t.Helper()

	formatJSRoot, out := writeFormatJSImportCase(t, tc)
	skips, err := importer(formatJSRoot, out)
	if err != nil {
		t.Fatalf("%s() error = %v", importerName, err)
	}
	assertFormatJSImport(t, tc, out, skips)
}

func writeFormatJSImportCase(t *testing.T, tc formatJSImportCase) (formatJSRoot, out string) {
	t.Helper()

	root := t.TempDir()
	out = t.TempDir()
	formatJSRoot = filepath.Join(root, "formatjs")
	testsRoot := formatJSPackageTestsRoot(formatJSRoot, tc.packageDir)
	files := map[string]string{
		"mixed.test.ts":       tc.mixedSource,
		"unsupported.test.ts": "import {expect, it} from 'vitest'\nit('unsupported', () => expect(render()).toBe('ignored'))\n",
	}
	for rel, data := range files {
		mustWriteFile(t, filepath.Join(testsRoot, rel), data)
	}
	return formatJSRoot, out
}

func assertFormatJSImport(t *testing.T, tc formatJSImportCase, out string, skips []skipEntry) {
	t.Helper()

	wantSkips := []skipEntry{
		formatJSImportPartialExtractionSkip(tc.sourcePrefix + "mixed.test.ts"),
		formatJSUnsupportedExtractorShapeSkip(tc.sourcePrefix + "unsupported.test.ts"),
	}
	assertSkipsMatch(t, "importFixtures()", skips, wantSkips)

	fixturePath := filepath.Join(out, tc.targetPackage, "testdata", "conformance", "formatjs", "mixed-test-ts.json")
	fixtures := mustReadFixtures(t, fixturePath)
	assertFixtureIDs(t, "importFixtures()", fixtures, tc.wantFixtureID)
	if fixtures[0].Source != tc.sourcePrefix+"mixed.test.ts" {
		t.Fatalf("importFixtures() source = %q, want %q", fixtures[0].Source, tc.sourcePrefix+"mixed.test.ts")
	}
	unsupportedPath := filepath.Join(out, tc.targetPackage, "testdata", "conformance", "formatjs", "unsupported-test-ts.json")
	assertPathAbsent(t, "importFixtures() unsupported file stat error", unsupportedPath)
}

func TestExtractLocaleFixturesMapsCoreLocaleOperations(t *testing.T) {
	t.Parallel()

	rel := "core.test.ts"
	fixtures := extractLocaleFixtures(rel, `import {Locale} from '#packages/intl-locale'
import {describe, expect, it} from 'vitest'

describe('Locale core operations', () => {
  it('extracts stable generated fixture operations', () => {
    expect(new Locale('EN-us-u-NU-LATN').toString()).toBe('en-US-u-nu-latn')
    expect(new Locale('zh').maximize().toString()).toEqual('zh-Hans-CN')
    expect(new Locale('zh-Hans-CN').minimize().toString()).toBe('zh')
  })
})
`)
	want := []localeFixtureSummary{
		{
			ID:       formatJSFixtureID("locale", formatJSLocaleTestSourcePrefix+rel, 0),
			Source:   formatJSLocaleTestSourcePrefix + rel,
			Locale:   "EN-us-u-NU-LATN",
			Feature:  fixtureFeatureCanonicalize,
			Input:    "EN-us-u-NU-LATN",
			Expected: "en-US-u-nu-latn",
		},
		{
			ID:       formatJSFixtureID("locale", formatJSLocaleTestSourcePrefix+rel, 1),
			Source:   formatJSLocaleTestSourcePrefix + rel,
			Locale:   "zh",
			Feature:  fixtureFeatureMaximize,
			Input:    "zh",
			Expected: "zh-Hans-CN",
		},
		{
			ID:       formatJSFixtureID("locale", formatJSLocaleTestSourcePrefix+rel, 2),
			Source:   formatJSLocaleTestSourcePrefix + rel,
			Locale:   "zh-Hans-CN",
			Feature:  fixtureFeatureMinimize,
			Input:    "zh-Hans-CN",
			Expected: "zh",
		},
	}
	assertLocaleFixtureSummaries(t, "extractLocaleFixtures()", fixtures, want)
}

func TestExtractCanonicalLocalesFixturesMapsGetCanonicalLocales(t *testing.T) {
	t.Parallel()

	rel := "canonical.test.ts"
	fixtures := extractCanonicalLocalesFixtures(rel, `import {getCanonicalLocales} from '#packages/intl-getcanonicallocales'
import {describe, expect, it} from 'vitest'

describe('Intl.getCanonicalLocales', () => {
  it('extracts string-array canonicalization', () => {
    expect(getCanonicalLocales('EN-us-u-NU-LATN')).toEqual(['en-US-u-nu-latn'])
  })
})
`)
	want := []localeFixtureSummary{{
		ID:       formatJSFixtureID("locale", formatJSCanonicalLocalesTestSourcePrefix+rel, 0),
		Source:   formatJSCanonicalLocalesTestSourcePrefix + rel,
		Locale:   "EN-us-u-NU-LATN",
		Feature:  fixtureFeatureCanonicalize,
		Input:    "EN-us-u-NU-LATN",
		Expected: "en-US-u-nu-latn",
	}}
	assertLocaleFixtureSummaries(t, "extractCanonicalLocalesFixtures()", fixtures, want)
}

func TestFixtureSlugNormalizesPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "path and extension", path: "src/lists/basic.test.ts", want: "src-lists-basic-test-ts"},
		{name: "collapses punctuation", path: "Intl.Locale - Canonicalize!.test.ts", want: "intl-locale-canonicalize-test-ts"},
		{name: "trims separators", path: "__cases__/range--parts??", want: "cases-range-parts"},
		{name: "keeps ASCII digits", path: "Node v26/ICU78", want: "node-v26-icu78"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := fixtureSlug(tc.path); got != tc.want {
				t.Fatalf("fixtureSlug(%q) = %q, want %q", tc.path, got, tc.want)
			}
		})
	}
}

func TestFormatJSFixtureIDUsesSharedWireShape(t *testing.T) {
	t.Parallel()

	got := formatJSFixtureID("numberformat", "ranges/parts.test.ts", 7)
	want := "numberformat-formatjs-ranges-parts-test-ts-007"
	if got != want {
		t.Fatalf("formatJSFixtureID() = %q, want %q", got, want)
	}
}

func TestRangeFixtureInputUsesSharedWireShape(t *testing.T) {
	t.Parallel()

	got := rangeFixtureInput("1970-01-01T00:00:00Z", "1970-01-02T00:00:00Z")
	want := `{"end":"1970-01-02T00:00:00Z","start":"1970-01-01T00:00:00Z"}`
	assertJSONEqual(t, "rangeFixtureInput()", got, want)
}

func TestFixtureFeaturesUseSharedWireShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "canonicalize", got: fixtureFeatureCanonicalize, want: "canonicalize"},
		{name: "maximize", got: fixtureFeatureMaximize, want: "maximize"},
		{name: "minimize", got: fixtureFeatureMinimize, want: "minimize"},
		{name: "format to parts", got: fixtureFeatureFormatToParts, want: "formatToParts"},
		{name: "format range", got: fixtureFeatureFormatRange, want: "formatRange"},
		{name: "format range to parts", got: fixtureFeatureFormatRangeToParts, want: "formatRangeToParts"},
		{name: "select range", got: fixtureFeatureSelectRange, want: "selectRange"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Fatalf("fixture feature %s = %q, want %q", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestFormatJSReferencePathsBuildPackageTestContracts(t *testing.T) {
	t.Parallel()

	if got, want := formatJSPackageTestsRoot("/refs/formatjs", "intl-numberformat"), filepath.Join("/refs/formatjs", "packages", "intl-numberformat", "tests"); got != want {
		t.Fatalf("formatJSPackageTestsRoot() = %q, want %q", got, want)
	}
	if got, want := formatJSTestSourcePrefix("intl-numberformat"), "formatjs:packages/intl-numberformat/tests/"; got != want {
		t.Fatalf("formatJSTestSourcePrefix() = %q, want %q", got, want)
	}
}

func TestFormatJSSurfaceRoutesPreserveImportOrder(t *testing.T) {
	t.Parallel()

	assertFormatJSSurfaceRoutes(t, "pre-locale", formatJSPreLocaleSurfaceRoutes(), []formatJSSurfaceRoute{
		formatJSNumberFormatRoute(),
		formatJSPluralRulesRoute(),
		formatJSDateTimeFormatRoute(),
	})
	assertFormatJSSurfaceRoutes(t, "post-locale", formatJSPostLocaleSurfaceRoutes(), []formatJSSurfaceRoute{
		formatJSListFormatRoute(),
		formatJSRelativeTimeFormatRoute(),
		formatJSDurationFormatRoute(),
	})

	routes := formatJSPreLocaleSurfaceRoutes()
	routes[0].targetPackage = "mutated"
	if got, want := formatJSPreLocaleSurfaceRoutes()[0].targetPackage, formatJSNumberFormatTargetPackage; got != want {
		t.Fatalf("formatJSPreLocaleSurfaceRoutes()[0].targetPackage after caller mutation = %q, want %q", got, want)
	}
}

func assertFormatJSSurfaceRoutes(t *testing.T, label string, got, want []formatJSSurfaceRoute) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s route count = %d, want %d", label, len(got), len(want))
	}
	for i := range got {
		if got[i].targetPackage != want[i].targetPackage {
			t.Fatalf("%s route %d targetPackage = %q, want %q", label, i, got[i].targetPackage, want[i].targetPackage)
		}
		if got[i].spec.packageDir != want[i].spec.packageDir {
			t.Fatalf("%s route %d packageDir = %q, want %q", label, i, got[i].spec.packageDir, want[i].spec.packageDir)
		}
		if got[i].spec.sourcePrefix() != want[i].spec.sourcePrefix() {
			t.Fatalf("%s route %d sourcePrefix = %q, want %q", label, i, got[i].spec.sourcePrefix(), want[i].spec.sourcePrefix())
		}
	}
}

func TestFormatJSSkipEntriesUseSharedWireShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  skipEntry
		want skipEntry
	}{
		{
			name: "missing reference",
			got:  missingReferenceSkip("formatjs", skipReasonFormatJSPathNotProvided),
			want: skipEntry{Source: "formatjs", Category: "missing-reference", Reason: "formatjs path not provided"},
		},
		{
			name: "partial extraction",
			got:  formatJSImportPartialExtractionSkip("formatjs:packages/intl-numberformat/tests/basic.test.ts"),
			want: skipEntry{Source: "formatjs:packages/intl-numberformat/tests/basic.test.ts", Category: "partial-extraction", Reason: "mechanical assertions outside current generated fixture gate"},
		},
		{
			name: "unsupported extractor shape",
			got:  formatJSUnsupportedExtractorShapeSkip("formatjs:packages/intl-numberformat/tests/basic.test.ts"),
			want: skipEntry{Source: "formatjs:packages/intl-numberformat/tests/basic.test.ts", Category: "unsupported-extractor-shape", Reason: "unsupported Vitest assertion shape"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.got != tc.want {
				t.Fatalf("%s skip = %+v, want %+v", tc.name, tc.got, tc.want)
			}
		})
	}
}

func TestNodeFixtureDir(t *testing.T) {
	t.Parallel()

	got, err := nodeFixtureDir("v26.0.0")
	if err != nil {
		t.Fatalf("nodeFixtureDir(v26.0.0) error = %v", err)
	}
	if got != "node-v26" {
		t.Fatalf("nodeFixtureDir(v26.0.0) = %q, want node-v26", got)
	}
	if _, err := nodeFixtureDir("not-node"); err == nil {
		t.Fatal("nodeFixtureDir(not-node) error = nil, want error")
	}
}

func mustReadSkips(t *testing.T, path string) []skipEntry {
	t.Helper()

	return mustReadJSON[[]skipEntry](t, path)
}

func hasSkip(skips []skipEntry, want skipEntry) bool {
	return slices.ContainsFunc(skips, func(got skipEntry) bool {
		return got == want
	})
}

func assertPathAbsent(t *testing.T, name, path string) {
	t.Helper()

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("%s = %v, want not exist", name, err)
	}
}

func assertSkipsEqual(t *testing.T, name string, got, want []skipEntry) {
	t.Helper()

	if !slices.Equal(got, want) {
		t.Fatalf("%s skips = %+v, want %+v", name, got, want)
	}
}

func assertSkipsMatch(t *testing.T, name string, skips, wants []skipEntry) {
	t.Helper()

	if len(skips) != len(wants) {
		t.Fatalf("%s skips = %+v, want %+v", name, skips, wants)
	}
	for _, want := range wants {
		if !hasSkip(skips, want) {
			t.Fatalf("%s skips = %+v, missing %+v", name, skips, want)
		}
	}
}

type localeFixtureSummary struct {
	ID       string
	Source   string
	Locale   string
	Feature  string
	Input    string
	Expected string
}

func localeFixtureSummaries(t *testing.T, fixtures []fixture) []localeFixtureSummary {
	t.Helper()

	summaries := make([]localeFixtureSummary, len(fixtures))
	for i, fixture := range fixtures {
		if fixture.Expected == nil {
			t.Fatalf("fixture %q Expected = nil", fixture.ID)
		}
		input, ok := fixture.Input.(string)
		if !ok {
			t.Fatalf("fixture %q Input = %T, want string", fixture.ID, fixture.Input)
		}
		if len(fixture.Options) != 0 {
			t.Fatalf("fixture %q Options = %v, want empty", fixture.ID, fixture.Options)
		}
		summaries[i] = localeFixtureSummary{
			ID:       fixture.ID,
			Source:   fixture.Source,
			Locale:   fixture.Locale,
			Feature:  fixture.Feature,
			Input:    input,
			Expected: *fixture.Expected,
		}
	}
	return summaries
}

func assertLocaleFixtureSummaries(t *testing.T, name string, fixtures []fixture, want []localeFixtureSummary) {
	t.Helper()

	if got := localeFixtureSummaries(t, fixtures); !slices.Equal(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}

type fixtureJSONSummary struct {
	ID          string
	Source      string
	Locale      string
	Feature     string
	OptionsJSON string
	InputJSON   string
	Expected    string
}

func fixtureJSONSummaries(t *testing.T, fixtures []fixture) []fixtureJSONSummary {
	t.Helper()

	summaries := make([]fixtureJSONSummary, len(fixtures))
	for i, fixture := range fixtures {
		if fixture.Expected == nil {
			t.Fatalf("fixture %q Expected = nil", fixture.ID)
		}
		summaries[i] = fixtureJSONSummary{
			ID:          fixture.ID,
			Source:      fixture.Source,
			Locale:      fixture.Locale,
			Feature:     fixture.Feature,
			OptionsJSON: mustJSON(t, fixture.Options),
			InputJSON:   mustJSON(t, fixture.Input),
			Expected:    *fixture.Expected,
		}
	}
	return summaries
}

func assertFixtureJSONSummaries(t *testing.T, name string, fixtures []fixture, want []fixtureJSONSummary) {
	t.Helper()

	if got := fixtureJSONSummaries(t, fixtures); !slices.Equal(got, want) {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal %T: %v", value, err)
	}
	return string(data)
}

func assertJSONEqual(t *testing.T, name string, got any, want string) {
	t.Helper()

	if gotJSON := mustJSON(t, got); gotJSON != want {
		t.Fatalf("%s = %s, want %s", name, gotJSON, want)
	}
}

func assertSliceParseResult[T comparable](t *testing.T, name, raw string, got []T, ok bool, want []T, wantOK bool) {
	t.Helper()

	if ok != wantOK {
		t.Fatalf("%s(%q) ok = %v, want %v", name, raw, ok, wantOK)
	}
	if ok && !slices.Equal(got, want) {
		t.Fatalf("%s(%q) = %+v, want %+v", name, raw, got, want)
	}
}

func assertAnyParseResult(t *testing.T, name, raw string, got any, ok bool, want any, wantOK bool) {
	t.Helper()

	if ok != wantOK {
		t.Fatalf("%s(%q) ok = %v, want %v", name, raw, ok, wantOK)
	}
	if ok && got != want {
		t.Fatalf("%s(%q) = %[3]v (%[3]T), want %[4]v (%[4]T)", name, raw, got, want)
	}
}

func assertStringParseResult(t *testing.T, name, raw, got string, ok bool, want string, wantOK bool) {
	t.Helper()

	if ok != wantOK {
		t.Fatalf("%s(%q) ok = %v, want %v", name, raw, ok, wantOK)
	}
	if got != want {
		t.Fatalf("%s(%q) = %q, want %q", name, raw, got, want)
	}
}

func mustReadFixtures(t *testing.T, path string) []fixture {
	t.Helper()

	return mustReadJSON[[]fixture](t, path)
}

func mustReadJSON[T any](t *testing.T, path string) T {
	t.Helper()

	raw := mustReadFile(t, path)
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("decode %s: %v\n%s", path, err, raw)
	}
	return value
}

func mustReadFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return data
}

func mustWriteFile(t *testing.T, path, data string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o777); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(data), 0o666); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func assertFileContainsAll(t *testing.T, name, path string, wants ...string) {
	t.Helper()

	assertContainsAll(t, name, mustReadFile(t, path), wants...)
}

func assertContainsAll(t *testing.T, name string, data []byte, wants ...string) {
	t.Helper()

	text := string(data)
	for _, want := range wants {
		if !strings.Contains(text, want) {
			t.Fatalf("%s = %s, want %s", name, data, want)
		}
	}
}

func assertFixtureIDs(t *testing.T, name string, fixtures []fixture, wants ...string) {
	t.Helper()

	if got := fixtureIDs(fixtures); !slices.Equal(got, wants) {
		t.Fatalf("%s IDs = %v, want %v", name, got, wants)
	}
}

func fixtureIDs(fixtures []fixture) []string {
	ids := make([]string, len(fixtures))
	for i, fixture := range fixtures {
		ids[i] = fixture.ID
	}
	return ids
}

func TestParseNumberLiteralMapsFormatJSNumericLiterals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want any
		ok   bool
	}{
		{name: "integer", raw: "42", want: int64(42), ok: true},
		{name: "signed integer", raw: "-42", want: int64(-42), ok: true},
		{name: "numeric separators", raw: "1_234", want: int64(1234), ok: true},
		{name: "float", raw: "12.5", want: float64(12.5), ok: true},
		{name: "exponent", raw: "1e3", want: float64(1000), ok: true},
		{name: "uint64 overflow bridge", raw: "18446744073709551615", want: uint64(18446744073709551615), ok: true},
		{name: "empty", raw: ""},
		{name: "NaN", raw: "NaN"},
		{name: "Infinity", raw: "Infinity"},
		{name: "call expression", raw: "Number(1)"},
		{name: "object expression", raw: "{value: 1}"},
		{name: "hex literal", raw: "0x10"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseNumberLiteral(tc.raw)
			assertAnyParseResult(t, "parseNumberLiteral", tc.raw, got, ok, tc.want, tc.ok)
		})
	}
}

func TestParseDateNumberLiteralMapsFormatJSDateInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{name: "epoch milliseconds", raw: "789", want: "1970-01-01T00:00:00.789Z", ok: true},
		{name: "negative epoch milliseconds", raw: "-1", want: "1969-12-31T23:59:59.999Z", ok: true},
		{name: "numeric separators", raw: "1_000", want: "1970-01-01T00:00:01Z", ok: true},
		{name: "float milliseconds rejected", raw: "12.5"},
		{name: "uint64 overflow rejected", raw: "9223372036854775808"},
		{name: "NaN rejected", raw: "NaN"},
		{name: "call expression rejected", raw: "Date.now()"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseDateNumberLiteral(tc.raw)
			assertStringParseResult(t, "parseDateNumberLiteral", tc.raw, got, ok, tc.want, tc.ok)
		})
	}
}

func TestParsePluralInputLiteralMapsFormatJSPluralInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want any
		ok   bool
	}{
		{name: "integer", raw: "2", want: int64(2), ok: true},
		{name: "float", raw: "2.5", want: float64(2.5), ok: true},
		{name: "numeric separators", raw: "1_000", want: int64(1000), ok: true},
		{name: "bigint suffix", raw: "9007199254740993n", want: "9007199254740993", ok: true},
		{name: "bigint call", raw: "BigInt(9007199254740993)", want: "9007199254740993", ok: true},
		{name: "negative bigint", raw: "BigInt(-1)", want: "-1", ok: true},
		{name: "quoted decimal", raw: "'1.0'", want: "1.0", ok: true},
		{name: "as any suffix", raw: "2n as any", want: "2", ok: true},
		{name: "empty", raw: ""},
		{name: "array expression", raw: "[1]"},
		{name: "object expression", raw: "{value: 1}"},
		{name: "decimal bigint", raw: "1.5n"},
		{name: "signed plus bigint", raw: "+1n"},
		{name: "math expression", raw: "1 + 2"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parsePluralInputLiteral(tc.raw)
			assertAnyParseResult(t, "parsePluralInputLiteral", tc.raw, got, ok, tc.want, tc.ok)
		})
	}
}

func TestParsePartArrayMapsFormatJSPartRecords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want []conformance.Part
		ok   bool
	}{
		{
			name: "part objects",
			raw:  `[{type: 'integer', value: '12'}, {value: "line\nfeed", type: 'literal'}]`,
			want: []conformance.Part{
				{Type: "integer", Value: "12"},
				{Type: "literal", Value: "line\nfeed"},
			},
			ok: true,
		},
		{name: "empty array", raw: `[]`},
		{name: "missing type", raw: `[{value: '12'}]`},
		{name: "missing value", raw: `[{type: 'integer'}]`},
		{name: "non-string value", raw: `[{type: 'integer', value: 12}]`},
		{name: "invalid escaped value", raw: `[{type: 'literal', value: 'bad\z'}]`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parsePartArray(tc.raw)
			assertSliceParseResult(t, "parsePartArray", tc.raw, got, ok, tc.want, tc.ok)
		})
	}
}

func TestParseRangePartArrayMapsFormatJSRangePartRecords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want []conformance.RangePart
		ok   bool
	}{
		{
			name: "range part objects",
			raw:  `[{type: 'integer', value: '1', source: 'startRange'}, {source: 'shared', value: "-", type: 'literal'}]`,
			want: []conformance.RangePart{
				{Type: "integer", Value: "1", Source: "startRange"},
				{Type: "literal", Value: "-", Source: "shared"},
			},
			ok: true,
		},
		{name: "empty array", raw: `[]`},
		{name: "missing type", raw: `[{value: '1', source: 'startRange'}]`},
		{name: "missing value", raw: `[{type: 'integer', source: 'startRange'}]`},
		{name: "missing source", raw: `[{type: 'integer', value: '1'}]`},
		{name: "non-string source", raw: `[{type: 'integer', value: '1', source: startRange}]`},
		{name: "invalid escaped source", raw: `[{type: 'integer', value: '1', source: 'bad\z'}]`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseRangePartArray(tc.raw)
			assertSliceParseResult(t, "parseRangePartArray", tc.raw, got, ok, tc.want, tc.ok)
		})
	}
}

func TestParseOptionsObjectMapsFormatJSOptionBag(t *testing.T) {
	t.Parallel()

	got := parseOptionsObject(`{
  style: 'digital',
  fractionalDigits: 3,
  useGrouping: false,
  minimumIntegerDigits: 1_000,
  numeric: 'auto',
	numeric: false,
}`)
	want := `{"fractionalDigits":3,"minimumIntegerDigits":1000,"numeric":"auto","style":"digital","useGrouping":false}`
	assertJSONEqual(t, "parseOptionsObject()", got, want)
}

func TestParseStringArrayMapsFormatJSArrayLiterals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want []string
		ok   bool
	}{
		{name: "empty array", raw: `[]`, want: []string{}, ok: true},
		{name: "mixed quotes and escapes", raw: `['alpha', "beta", 'it\'s', "line\nfeed"]`, want: []string{"alpha", "beta", "it's", "line\nfeed"}, ok: true},
		{name: "not an array", raw: `'alpha'`},
		{name: "unquoted value", raw: `[alpha]`},
		{name: "mismatched quote", raw: `['alpha"]`},
		{name: "invalid escape", raw: `['bad\z']`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseStringArray(tc.raw)
			assertSliceParseResult(t, "parseStringArray", tc.raw, got, ok, tc.want, tc.ok)
		})
	}
}

func TestParseDurationObjectMapsFormatJSDurationRecords(t *testing.T) {
	t.Parallel()

	got, ok := parseDurationObject(`{
  years: 1,
  months: -2,
  weeks: 3_000,
  days: 4,
  hours: 5,
  minutes: 6,
  seconds: 7,
  milliseconds: 8,
  microseconds: 9,
  nanoseconds: 10,
}`)
	if !ok {
		t.Fatalf("parseDurationObject() ok = false, want true")
	}
	want := `{"days":4,"hours":5,"microseconds":9,"milliseconds":8,"minutes":6,"months":-2,"nanoseconds":10,"seconds":7,"weeks":3000,"years":1}`
	assertJSONEqual(t, "parseDurationObject()", got, want)

	rejected := []struct {
		name string
		raw  string
	}{
		{name: "empty object", raw: `{}`},
		{name: "not object", raw: `[1]`},
		{name: "unknown field", raw: `{centuries: 1}`},
		{name: "float field", raw: `{seconds: 1.5}`},
		{name: "uint64 overflow", raw: `{seconds: 9223372036854775808}`},
		{name: "expression field", raw: `{seconds: Number(1)}`},
	}
	for _, tc := range rejected {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got, ok := parseDurationObject(tc.raw); ok {
				t.Fatalf("parseDurationObject(%q) = %v, want rejection", tc.raw, got)
			}
		})
	}
}

func TestParseDateConstructorPreservesDatePrecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
		ok   bool
	}{
		{
			name: "quoted RFC3339 offset with milliseconds",
			raw:  `'2026-05-08T12:00:00.789+02:00'`,
			want: "2026-05-08T10:00:00.789Z",
			ok:   true,
		},
		{
			name: "epoch milliseconds",
			raw:  "789",
			want: "1970-01-01T00:00:00.789Z",
			ok:   true,
		},
		{
			name: "Date.UTC milliseconds",
			raw:  "Date.UTC(2026, 4, 8, 12, 0, 0, 789)",
			want: "2026-05-08T12:00:00.789Z",
			ok:   true,
		},
		{
			name: "comma constructor milliseconds",
			raw:  "2026, 4, 8, 12, 0, 0, 789",
			want: "2026-05-08T12:00:00.789Z",
			ok:   true,
		},
		{
			name: "invalid month",
			raw:  "2026, 12, 8",
		},
		{
			name: "too few fields",
			raw:  "2026, 4",
		},
		{
			name: "not a date expression",
			raw:  "window.now()",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := parseDateConstructor(tc.raw)
			assertStringParseResult(t, "parseDateConstructor", tc.raw, got, ok, tc.want, tc.ok)
		})
	}
}

func TestExtractDateTimeFormatFixturesPreservesDatePrecision(t *testing.T) {
	t.Parallel()

	fixtures := extractDateTimeFormatFixtures("precision.test.ts", `import {DateTimeFormat} from '#packages/intl-datetimeformat/core'
import {describe, it, expect} from 'vitest'

describe('datetime precision cases', () => {
  const dtf = new DateTimeFormat('en-US', {hour: 'numeric', minute: 'numeric', second: 'numeric', timeZone: 'UTC'})
  const fromNumber = new Date(789)
  const fromFields = new Date(2026, 4, 8, 12, 0, 0, 789)
  const fromUTC = new Date(Date.UTC(2026, 4, 8, 12, 0, 0, 789))

  it('preserves millisecond inputs', () => {
    expect(dtf.format(fromNumber)).toBe('number variable')
    expect(dtf.format(fromFields)).toBe('field variable')
    expect(dtf.format(fromUTC)).toBe('utc variable')
    expect(dtf.format(789)).toBe('number literal')
  })
})
`)

	got := map[string]string{}
	for _, fixture := range fixtures {
		if fixture.Expected == nil {
			t.Fatalf("extractDateTimeFormatFixtures() fixture %q expected = nil", fixture.ID)
		}
		input, ok := fixture.Input.(string)
		if !ok {
			t.Fatalf("extractDateTimeFormatFixtures() fixture %q input = %T, want string", fixture.ID, fixture.Input)
		}
		got[*fixture.Expected] = input
	}
	want := map[string]string{
		"number variable": "1970-01-01T00:00:00.789Z",
		"field variable":  "2026-05-08T12:00:00.789Z",
		"utc variable":    "2026-05-08T12:00:00.789Z",
		"number literal":  "1970-01-01T00:00:00.789Z",
	}
	if len(got) != len(want) {
		t.Fatalf("extractDateTimeFormatFixtures() inputs = %v, want %v", got, want)
	}
	for label, wantInput := range want {
		if got[label] != wantInput {
			t.Fatalf("extractDateTimeFormatFixtures() input for %q = %q, want %q", label, got[label], wantInput)
		}
	}
}

func TestExtractDateTimeFormatFixturesUsesLatestDateVariable(t *testing.T) {
	t.Parallel()

	fixtures := extractDateTimeFormatFixtures("shadowed-date.test.ts", `import {DateTimeFormat} from '#packages/intl-datetimeformat/core'
import {describe, it, expect} from 'vitest'

describe('datetime variable position cases', () => {
  const dtf = new DateTimeFormat('en-US', {year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC'})
  const instant = new Date('2026-05-08T00:00:00Z')

  it('uses the variable value visible before each assertion', () => {
    expect(dtf.format(instant)).toBe('first')
    const instant = new Date('2026-05-09T00:00:00Z')
    expect(dtf.format(instant)).toBe('second')
  })
})
`)

	got := map[string]string{}
	for _, fixture := range fixtures {
		if fixture.Expected == nil {
			t.Fatalf("extractDateTimeFormatFixtures() fixture %q expected = nil", fixture.ID)
		}
		input, ok := fixture.Input.(string)
		if !ok {
			t.Fatalf("extractDateTimeFormatFixtures() fixture %q input = %T, want string", fixture.ID, fixture.Input)
		}
		got[*fixture.Expected] = input
	}
	want := map[string]string{
		"first":  "2026-05-08T00:00:00Z",
		"second": "2026-05-09T00:00:00Z",
	}
	if len(got) != len(want) {
		t.Fatalf("extractDateTimeFormatFixtures() inputs = %v, want %v", got, want)
	}
	for label, wantInput := range want {
		if got[label] != wantInput {
			t.Fatalf("extractDateTimeFormatFixtures() input for %q = %q, want %q", label, got[label], wantInput)
		}
	}
}

func TestFormatJSConstructorDeclarationsCaptureSharedShape(t *testing.T) {
	t.Parallel()

	data := `
const nf = new NumberFormat('en-US', {style: 'currency', currency: 'USD'})
const nf = new NumberFormat('fr', {style: 'percent'})
const nfDefault = new NumberFormat()
const pr = new Intl.PluralRules('en', {type: 'ordinal'})
const dtf = new DateTimeFormat(['en-US'], {timeZone: 'UTC'})
`
	allNumberDeclarations := numberFormatDeclarations(data)
	numberDeclarations := allNumberDeclarations["nf"]
	if len(numberDeclarations) != 2 {
		t.Fatalf("numberFormatDeclarations()[nf] length = %d, want 2", len(numberDeclarations))
	}
	latest, ok := latestConstructorDeclarationBefore(numberDeclarations, strings.Index(data, "const pr"))
	if !ok {
		t.Fatal("latestConstructorDeclarationBefore() ok = false, want true")
	}
	if latest.locale != "fr" || latest.options["style"] != "percent" {
		t.Fatalf("latest number declaration = %#v, want locale fr style percent", latest)
	}
	explicitLatest, ok := latestExplicitLocaleConstructorDeclarationBefore(allNumberDeclarations, "nf", strings.Index(data, "const pr"))
	if !ok {
		t.Fatal("latestExplicitLocaleConstructorDeclarationBefore(nf) ok = false, want true")
	}
	if explicitLatest.locale != "fr" || explicitLatest.options["style"] != "percent" {
		t.Fatalf("latest explicit number declaration = %#v, want locale fr style percent", explicitLatest)
	}
	if _, ok := latestExplicitLocaleConstructorDeclarationBefore(allNumberDeclarations, "nfDefault", strings.Index(data, "const pr")); ok {
		t.Fatal("latestExplicitLocaleConstructorDeclarationBefore(nfDefault) ok = true, want false")
	}

	pluralDeclarations := pluralRulesDeclarations(data)["pr"]
	if len(pluralDeclarations) != 1 {
		t.Fatalf("pluralRulesDeclarations()[pr] length = %d, want 1", len(pluralDeclarations))
	}
	plural := pluralDeclarations[0]
	if plural.locale != "en" || plural.options["type"] != "ordinal" {
		t.Fatalf("plural declaration = %#v, want locale en type ordinal", plural)
	}

	dateTimeDeclarations := dateTimeDeclarations(data)["dtf"]
	if len(dateTimeDeclarations) != 1 {
		t.Fatalf("dateTimeDeclarations()[dtf] length = %d, want 1", len(dateTimeDeclarations))
	}
	dateTime := dateTimeDeclarations[0]
	if dateTime.locale != "en-US" || dateTime.options["timeZone"] != "UTC" {
		t.Fatalf("datetime declaration = %#v, want locale en-US timeZone UTC", dateTime)
	}
}

func TestFormatJSFixtureWireShapes(t *testing.T) {
	t.Parallel()

	options := map[string]any{"style": "currency"}
	expected := formatJSExpectedFixture("numberformat", formatJSNumberFormatTestSourcePrefix, "number/basic.test.ts", 7, "en-US", options, 12, "$12")
	if expected.ID != "numberformat-formatjs-number-basic-test-ts-007" {
		t.Fatalf("formatJSExpectedFixture() id = %q, want numberformat-formatjs-number-basic-test-ts-007", expected.ID)
	}
	if expected.Source != formatJSNumberFormatTestSourcePrefix+"number/basic.test.ts" {
		t.Fatalf("formatJSExpectedFixture() source = %q", expected.Source)
	}
	if expected.Locale != "en-US" || expected.Options["style"] != "currency" || expected.Input != 12 {
		t.Fatalf("formatJSExpectedFixture() identity = %#v, want en-US currency input 12", expected)
	}
	if expected.Expected == nil || *expected.Expected != "$12" {
		t.Fatalf("formatJSExpectedFixture() expected = %v, want $12", expected.Expected)
	}

	rangeFixture := formatJSRangeFixture("numberformat", formatJSNumberFormatTestSourcePrefix, "range.test.ts", 3, "en", map[string]any{}, 1, 2, "1-2")
	if rangeFixture.Feature != fixtureFeatureFormatRange {
		t.Fatalf("formatJSRangeFixture() feature = %q, want %q", rangeFixture.Feature, fixtureFeatureFormatRange)
	}
	rangeInput, ok := rangeFixture.Input.(map[string]any)
	if !ok || rangeInput["start"] != 1 || rangeInput["end"] != 2 {
		t.Fatalf("formatJSRangeFixture() input = %#v, want start/end map", rangeFixture.Input)
	}
	if rangeFixture.ExpectedRange == nil || *rangeFixture.ExpectedRange != "1-2" {
		t.Fatalf("formatJSRangeFixture() expectedRange = %v, want 1-2", rangeFixture.ExpectedRange)
	}

	parts := []conformance.RangePart{
		{Type: "integer", Value: "1", Source: "startRange"},
		{Type: "literal", Value: "-", Source: "shared"},
		{Type: "integer", Value: "2", Source: "endRange"},
	}
	rangePartsFixture := formatJSRangePartsFixture("datetimeformat", formatJSDateTimeFormatTestSourcePrefix, "range-parts.test.ts", 4, "en-US", map[string]any{"timeZone": "UTC"}, "2026-05-08T00:00:00Z", "2026-05-09T00:00:00Z", parts)
	if rangePartsFixture.Feature != fixtureFeatureFormatRangeToParts {
		t.Fatalf("formatJSRangePartsFixture() feature = %q, want %q", rangePartsFixture.Feature, fixtureFeatureFormatRangeToParts)
	}
	if rangePartsFixture.ExpectedRange == nil || *rangePartsFixture.ExpectedRange != "1-2" {
		t.Fatalf("formatJSRangePartsFixture() expectedRange = %v, want 1-2", rangePartsFixture.ExpectedRange)
	}
	if !slices.Equal(rangePartsFixture.ExpectedRangeParts, parts) {
		t.Fatalf("formatJSRangePartsFixture() parts = %#v, want %#v", rangePartsFixture.ExpectedRangeParts, parts)
	}
	rangePartsInput, ok := rangePartsFixture.Input.(map[string]any)
	if !ok || rangePartsInput["start"] != "2026-05-08T00:00:00Z" || rangePartsInput["end"] != "2026-05-09T00:00:00Z" {
		t.Fatalf("formatJSRangePartsFixture() input = %#v, want start/end map", rangePartsFixture.Input)
	}
}

func TestStringValueOneOf(t *testing.T) {
	t.Parallel()

	value, ok := stringValueOneOf("short", "long", "short", "narrow")
	if !ok || value != "short" {
		t.Fatalf("stringValueOneOf(allowed) = %q, %t; want short, true", value, ok)
	}
	if value, ok := stringValueOneOf(12, "12"); ok || value != "" {
		t.Fatalf("stringValueOneOf(non-string) = %q, %t; want empty, false", value, ok)
	}
	if value, ok := stringValueOneOf("wide", "long", "short", "narrow"); ok || value != "" {
		t.Fatalf("stringValueOneOf(disallowed) = %q, %t; want empty, false", value, ok)
	}
}

func TestFormatJSDurationUnitKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		key     string
		unit    bool
		display bool
	}{
		{name: "plain unit", key: "seconds", unit: true},
		{name: "subsecond unit", key: "microseconds", unit: true},
		{name: "display unit", key: "secondsDisplay", display: true},
		{name: "unknown unit", key: "centuries"},
		{name: "unknown display unit", key: "centuriesDisplay"},
		{name: "bare display suffix", key: "Display"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := isFormatJSDurationUnitKey(tc.key); got != tc.unit {
				t.Fatalf("isFormatJSDurationUnitKey(%q) = %t, want %t", tc.key, got, tc.unit)
			}
			if got := isFormatJSDurationUnitDisplayKey(tc.key); got != tc.display {
				t.Fatalf("isFormatJSDurationUnitDisplayKey(%q) = %t, want %t", tc.key, got, tc.display)
			}
		})
	}
}

func TestExtractNumberFormatFixturesPreservesPartRecords(t *testing.T) {
	t.Parallel()

	fixtures := extractNumberFormatFixtures("parts.test.ts", `import {NumberFormat} from '#packages/intl-numberformat/core'
import {describe, it, expect} from 'vitest'

describe('number parts cases', () => {
  const nf = new NumberFormat('en-US', {style: 'currency', currency: 'USD'})

  it('preserves part records', () => {
    expect(nf.formatToParts(12)).toEqual([
      {type: 'currency', value: '$'},
      {type: 'integer', value: '12'},
    ])
    expect(nf.formatRangeToParts(1, 2)).toEqual([
      {type: 'integer', value: '1', source: 'startRange'},
      {type: 'literal', value: '-', source: 'shared'},
      {type: 'integer', value: '2', source: 'endRange'},
    ])
  })
})
`)

	byFeature := map[string]fixture{}
	for _, fixture := range fixtures {
		byFeature[fixture.Feature] = fixture
	}

	partsFixture, ok := byFeature[fixtureFeatureFormatToParts]
	if !ok {
		t.Fatalf("extractNumberFormatFixtures() missing formatToParts fixture: %v", fixtures)
	}
	wantParts := []conformance.Part{
		{Type: "currency", Value: "$"},
		{Type: "integer", Value: "12"},
	}
	if !slices.Equal(partsFixture.ExpectedParts, wantParts) {
		t.Fatalf("extractNumberFormatFixtures() formatToParts = %v, want %v", partsFixture.ExpectedParts, wantParts)
	}
	if partsFixture.Expected == nil || *partsFixture.Expected != "$12" {
		t.Fatalf("extractNumberFormatFixtures() formatToParts expected = %v, want $12", partsFixture.Expected)
	}

	rangeFixture, ok := byFeature[fixtureFeatureFormatRangeToParts]
	if !ok {
		t.Fatalf("extractNumberFormatFixtures() missing formatRangeToParts fixture: %v", fixtures)
	}
	wantRangeParts := []conformance.RangePart{
		{Type: "integer", Value: "1", Source: "startRange"},
		{Type: "literal", Value: "-", Source: "shared"},
		{Type: "integer", Value: "2", Source: "endRange"},
	}
	if !slices.Equal(rangeFixture.ExpectedRangeParts, wantRangeParts) {
		t.Fatalf("extractNumberFormatFixtures() formatRangeToParts = %v, want %v", rangeFixture.ExpectedRangeParts, wantRangeParts)
	}
	if rangeFixture.ExpectedRange == nil || *rangeFixture.ExpectedRange != "1-2" {
		t.Fatalf("extractNumberFormatFixtures() formatRangeToParts expectedRange = %v, want 1-2", rangeFixture.ExpectedRange)
	}
}

func TestExtractNumberFormatFixturesUsesLatestConstructor(t *testing.T) {
	t.Parallel()

	fixtures := extractNumberFormatFixtures("shadowed-numberformat.test.ts", `import {NumberFormat} from '#packages/intl-numberformat/core'
import {describe, it, expect} from 'vitest'

describe('number constructor position cases', () => {
  const nf = new NumberFormat('en-US', {style: 'currency', currency: 'USD'})

  it('uses the outer formatter before the local declaration', () => {
    expect(nf.format(1)).toBe('outer')
  })

  it('uses the local formatter after the local declaration', () => {
    const nf = new NumberFormat('fr', {style: 'percent'})
    expect(nf.format(0.25)).toBe('local')
  })
})
`)

	got := map[string]fixture{}
	for _, fixture := range fixtures {
		if fixture.Expected == nil {
			t.Fatalf("extractNumberFormatFixtures() fixture %q expected = nil", fixture.ID)
		}
		got[*fixture.Expected] = fixture
	}
	want := map[string]struct {
		locale  string
		options map[string]any
	}{
		"outer": {
			locale: "en-US",
			options: map[string]any{
				"style":    "currency",
				"currency": "USD",
			},
		},
		"local": {
			locale: "fr",
			options: map[string]any{
				"style": "percent",
			},
		},
	}
	if len(got) != len(want) {
		t.Fatalf("extractNumberFormatFixtures() fixtures = %v, want %v", got, want)
	}
	for label, wantFixture := range want {
		fixture, ok := got[label]
		if !ok {
			t.Fatalf("extractNumberFormatFixtures() missing fixture %q in %v", label, got)
		}
		if fixture.Locale != wantFixture.locale {
			t.Fatalf("extractNumberFormatFixtures() fixture %q locale = %q, want %q", label, fixture.Locale, wantFixture.locale)
		}
		if !maps.Equal(fixture.Options, wantFixture.options) {
			t.Fatalf("extractNumberFormatFixtures() fixture %q options = %v, want %v", label, fixture.Options, wantFixture.options)
		}
	}
}

func TestExtractPluralRulesFixturesMapsSelectAndRange(t *testing.T) {
	t.Parallel()

	rel := "plurals.test.ts"
	fixtures := extractPluralRulesFixtures(rel, `import {PluralRules} from '#packages/intl-pluralrules'
import {describe, expect, it} from 'vitest'

describe('PluralRules generated fixture shapes', () => {
  const cardinal = new PluralRules('en-US', {type: 'cardinal', minimumFractionDigits: 1})
  const ordinal = new Intl.PluralRules('en', {type: 'ordinal'})

  it('extracts select and selectRange assertions', () => {
    expect(new PluralRules('fr').select(0)).toBe('one')
    expect(new Intl.PluralRules('en', {type: 'cardinal'}).selectRange(1, 2)).toEqual('other')
    expect(cardinal.select(1n)).toBe('one')
    expect(ordinal.selectRange(BigInt(1), BigInt(2))).toEqual('other')
  })
})
`)
	source := formatJSPluralRulesTestSourcePrefix + rel
	want := []fixtureJSONSummary{
		{
			ID:          formatJSFixtureID("pluralrules", rel, 0),
			Source:      source,
			Locale:      "fr",
			OptionsJSON: "{}",
			InputJSON:   "0",
			Expected:    "one",
		},
		{
			ID:          formatJSFixtureID("pluralrules", rel, 1),
			Source:      source,
			Locale:      "en",
			Feature:     fixtureFeatureSelectRange,
			OptionsJSON: `{"type":"cardinal"}`,
			InputJSON:   `{"end":2,"start":1}`,
			Expected:    "other",
		},
		{
			ID:          formatJSFixtureID("pluralrules", rel, 2),
			Source:      source,
			Locale:      "en-US",
			OptionsJSON: `{"minimumFractionDigits":1,"type":"cardinal"}`,
			InputJSON:   `"1"`,
			Expected:    "one",
		},
		{
			ID:          formatJSFixtureID("pluralrules", rel, 3),
			Source:      source,
			Locale:      "en",
			Feature:     fixtureFeatureSelectRange,
			OptionsJSON: `{"type":"ordinal"}`,
			InputJSON:   `{"end":"2","start":"1"}`,
			Expected:    "other",
		},
	}
	assertFixtureJSONSummaries(t, "extractPluralRulesFixtures()", fixtures, want)
}

func TestExtractListFormatFixturesMapsFormatAssertions(t *testing.T) {
	t.Parallel()

	rel := "lists.test.ts"
	fixtures := extractListFormatFixtures(rel, `import {ListFormat} from '#packages/intl-listformat'
import {describe, expect, it} from 'vitest'

describe('ListFormat generated fixture shapes', () => {
  it('extracts supported list format assertions', () => {
    expect(new ListFormat('en-AI', {type: 'unit', style: 'short'}).format(['A', 'B', 'C'])).toBe('A, B, C')
    if (process.env.CI) {
      expect(new ListFormat('en-AI').format(['ignored'])).toBe('ignored')
    }
  })

  it.skip('keeps skipped assertions out of generated fixtures', () => {
    expect(new ListFormat('en-AI').format(['skip'])).toBe('skip')
  })
})
`)
	source := formatJSListFormatTestSourcePrefix + rel
	want := []fixtureJSONSummary{{
		ID:          formatJSFixtureID("listformat", rel, 0),
		Source:      source,
		Locale:      "en-AI",
		OptionsJSON: `{"style":"short","type":"unit"}`,
		InputJSON:   `["A","B","C"]`,
		Expected:    "A, B, C",
	}}
	assertFixtureJSONSummaries(t, "extractListFormatFixtures()", fixtures, want)
}

func TestExtractRelativeTimeFormatFixturesMapsFormatAssertions(t *testing.T) {
	t.Parallel()

	rel := "relative.test.ts"
	fixtures := extractRelativeTimeFormatFixtures(rel, `import {RelativeTimeFormat} from '#packages/intl-relativetimeformat'
import {describe, expect, it} from 'vitest'

describe('RelativeTimeFormat generated fixture shapes', () => {
  it('extracts supported relative time assertions', () => {
    expect(new RelativeTimeFormat('en-AI', {numeric: 'auto', style: 'long'}).format(-1, 'day')).toEqual('yesterday')
    if (process.env.CI) {
      expect(new RelativeTimeFormat('en-AI').format(1, 'day')).toBe('tomorrow')
    }
  })

  it.skip('keeps skipped relative time assertions out of generated fixtures', () => {
    expect(new RelativeTimeFormat('en-AI').format(2, 'day')).toBe('in 2 days')
  })
})
`)
	source := formatJSRelativeTimeFormatTestSourcePrefix + rel
	want := []fixtureJSONSummary{{
		ID:          formatJSFixtureID("relativetimeformat", rel, 0),
		Source:      source,
		Locale:      "en-AI",
		OptionsJSON: `{"numeric":"auto","style":"long"}`,
		InputJSON:   `{"unit":"day","value":-1}`,
		Expected:    "yesterday",
	}}
	assertFixtureJSONSummaries(t, "extractRelativeTimeFormatFixtures()", fixtures, want)
}

func TestExtractDurationFormatFixturesMapsFormatAssertions(t *testing.T) {
	t.Parallel()

	rel := "duration.test.ts"
	fixtures := extractDurationFormatFixtures(rel, `import {DurationFormat} from '#packages/intl-durationformat'
import {describe, expect, it} from 'vitest'

describe('DurationFormat generated fixture shapes', () => {
  it('extracts supported duration assertions', () => {
    expect(new Intl.DurationFormat('en', {style: 'digital', seconds: '2-digit', fractionalDigits: 3}).format({hours: 1, minutes: 2, seconds: 3, milliseconds: 4})).toBe('1:02:03.004')
    if (process.env.CI) {
      expect(new Intl.DurationFormat('en').format({seconds: 1})).toBe('1 sec')
    }
  })

  it.skip('keeps skipped duration assertions out of generated fixtures', () => {
    expect(new Intl.DurationFormat('en').format({seconds: 2})).toBe('2 sec')
  })
})
`)
	source := formatJSDurationFormatTestSourcePrefix + rel
	want := []fixtureJSONSummary{{
		ID:          formatJSFixtureID("durationformat", rel, 0),
		Source:      source,
		Locale:      "en",
		OptionsJSON: `{"fractionalDigits":3,"seconds":"2-digit","style":"digital"}`,
		InputJSON:   `{"hours":1,"milliseconds":4,"minutes":2,"seconds":3}`,
		Expected:    "1:02:03.004",
	}}
	assertFixtureJSONSummaries(t, "extractDurationFormatFixtures()", fixtures, want)
}

func TestWitnessFixtureFilesMapsGeneratedGroups(t *testing.T) {
	t.Parallel()

	witness := nodeWitness{
		LocaleSmoke:       []fixture{{ID: "locale-smoke"}},
		NumberFormatSmoke: []fixture{{ID: "number-smoke"}},
		DateTimeFormatSmoke: []fixture{{
			ID: "date-smoke",
		}},
		DurationFormatDigital: []fixture{{
			ID: "duration-digital",
		}},
		ListFormatErrors:   []fixture{{ID: "list-errors"}},
		RelativeTimeSmoke:  []fixture{{ID: "relative-smoke"}},
		PluralRulesErrors:  []fixture{{ID: "plural-errors"}},
		DisplayNamesSmoke:  []fixture{{ID: "displaynames-smoke"}},
		CollatorOptions:    []fixture{{ID: "collator-options"}},
		SegmenterLocale:    []fixture{{ID: "segmenter-locale"}},
		NumberFormatErrors: nil,
	}
	files := witnessFixtureFiles(witness, "node-v26")

	for _, want := range []struct {
		path string
		id   string
	}{
		{path: "locale/testdata/conformance/node-v26/smoke.json", id: "locale-smoke"},
		{path: "numberformat/testdata/conformance/node-v26/smoke.json", id: "number-smoke"},
		{path: "datetimeformat/testdata/conformance/node-v26/smoke.json", id: "date-smoke"},
		{path: "durationformat/testdata/conformance/node-v26/digital.json", id: "duration-digital"},
		{path: "listformat/testdata/conformance/node-v26/errors.json", id: "list-errors"},
		{path: "relativetimeformat/testdata/conformance/node-v26/smoke.json", id: "relative-smoke"},
		{path: "pluralrules/testdata/conformance/node-v26/errors.json", id: "plural-errors"},
		{path: "displaynames/testdata/conformance/node-v26/smoke.json", id: "displaynames-smoke"},
		{path: "collator/testdata/conformance/node-v26/options.json", id: "collator-options"},
		{path: "segmenter/testdata/conformance/node-v26/locale-contract.json", id: "segmenter-locale"},
	} {
		if !witnessFixtureFilesContains(files, want.path, want.id) {
			t.Fatalf("witnessFixtureFiles() missing %s for %s", want.id, want.path)
		}
	}
}

func TestCommittedNodeFixturesAreGeneratedOrExplicitlyManual(t *testing.T) {
	t.Parallel()

	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	nodeDir := "node-v26"
	generated := map[string]bool{}
	for _, file := range witnessFixtureFiles(nodeWitness{}, nodeDir) {
		generated[filepath.Join(file.Path...)] = true
	}
	generated[filepath.Join(nodeSupportedValuesPath(nodeDir)...)] = true
	manual := map[string]string{
		filepath.Join("datetimeformat", "testdata", "conformance", nodeDir, "deep-contract.json"): "deep DateTimeFormat range/parts contracts are still hand-curated",
		filepath.Join("locale", "testdata", "conformance", nodeDir, "errors.json"):                "Locale constructor errors are still hand-curated",
	}

	var unexpected []string
	for _, rel := range committedNodeFixtureFiles(t, root, nodeDir) {
		if generated[rel] {
			continue
		}
		if manual[rel] != "" {
			continue
		}
		unexpected = append(unexpected, rel)
	}
	if len(unexpected) != 0 {
		t.Fatalf("node fixture files are neither generated nor explicitly manual: %v", unexpected)
	}
}

func committedNodeFixtureFiles(t *testing.T, root, nodeDir string) []string {
	t.Helper()

	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch filepath.Base(path) {
			case ".git", ".tmp":
				return filepath.SkipDir
			default:
				return nil
			}
		}
		if filepath.Ext(path) != ".json" {
			return nil
		}
		if filepath.Base(filepath.Dir(path)) != nodeDir {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(files)
	return files
}

func witnessFixtureFilesContains(files []nodeWitnessFixtureFile, path, id string) bool {
	for _, file := range files {
		if filepath.Join(file.Path...) != path {
			continue
		}
		for _, fixture := range file.Fixtures {
			if fixture.ID == id {
				return true
			}
		}
	}
	return false
}

func TestRunImportsMechanicalFormatJSNumberFormatFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	formatJSRoot := filepath.Join(root, "formatjs")
	numberTestsRoot := formatJSPackageTestsRoot(formatJSRoot, formatJSNumberFormatPackageDir)
	pluralTestsRoot := formatJSPackageTestsRoot(formatJSRoot, formatJSPluralRulesPackageDir)
	dateTestsRoot := formatJSPackageTestsRoot(formatJSRoot, formatJSDateTimeFormatPackageDir)
	localeTestsRoot := formatJSPackageTestsRoot(formatJSRoot, formatJSLocalePackageDir)
	canonicalTestsRoot := formatJSPackageTestsRoot(formatJSRoot, formatJSCanonicalLocalesPackageDir)
	listTestsRoot := formatJSPackageTestsRoot(formatJSRoot, formatJSListFormatPackageDir)
	relativeTimeTestsRoot := formatJSPackageTestsRoot(formatJSRoot, formatJSRelativeTimeFormatPackageDir)
	durationTestsRoot := formatJSPackageTestsRoot(formatJSRoot, formatJSDurationFormatPackageDir)
	directTest := `import {describe, it, expect} from 'vitest'
import {NumberFormat} from '#packages/intl-numberformat/core'

describe('direct mechanical cases', () => {
  const nf = new NumberFormat('en', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
  const nfPercent = new NumberFormat("en", {style: "percent"})

  it('variable format assertions', () => {
    expect(nf.format(42)).toBe('42.00')
    expect(nfPercent.format(0.42)).toEqual("42%")
  })

  it('inline format assertion', () => {
    expect(new Intl.NumberFormat('en', {style: 'currency', currency: 'USD'}).format(42)).toBe('$42.00')
  })

  it('parts and range assertions', () => {
    expect(nf.formatToParts(0)).toEqual([{type: 'integer', value: '0'}])
    expect(nf.formatRange(1, 2)).toBe('1–2')
    expect(nf.formatRangeToParts(1, 2)).toEqual([
      {type: 'integer', value: '1', source: 'startRange'},
      {type: 'literal', value: '–', source: 'shared'},
      {type: 'integer', value: '2', source: 'endRange'},
    ])
  })
})
`
	mustWriteFile(t, filepath.Join(numberTestsRoot, "direct.test.ts"), directTest)
	mustWriteFile(t, filepath.Join(numberTestsRoot, "format_to_parts.test.ts"), `expect(parts).toEqual([{type: 'integer', value: '0'}])`)
	mustWriteFile(t, filepath.Join(numberTestsRoot, "decimal", "__snapshots__", "en.test.ts.snap"), "exports[`snapshot 1`] = `\"1\"`;")
	pluralTest := `import {describe, it, expect} from 'vitest'
import {PluralRules} from '#packages/intl-pluralrules'

describe('plural mechanical cases', () => {
  const cardinal = new PluralRules('en')
  const ordinal = new PluralRules('en', {type: 'ordinal'})
  const withFraction = new PluralRules('en', {minimumFractionDigits: 2})
  const compactFrench = new PluralRules('fr', {notation: 'compact'})
  const rounded = new PluralRules('en', {maximumFractionDigits: 0, roundingMode: 'trunc'})
  const significant = new PluralRules('en', {maximumSignificantDigits: 1})
  const stripped = new PluralRules('en', {minimumFractionDigits: 2, trailingZeroDisplay: 'stripIfInteger'})

  it('variable select assertions', () => {
    expect(cardinal.select(1)).toBe('one')
    expect(ordinal.select(2n)).toBe('two')
    expect(withFraction.select(1)).toBe('other')
    expect(compactFrench.select(1200000)).toBe('other')
    expect(rounded.select(1.9)).toBe('one')
    expect(significant.select(1.4)).toBe('one')
    expect(stripped.select(1)).toBe('one')
  })

  it('inline select assertion', () => {
    expect(new PluralRules('fr').select(1000000n)).toBe('many')
  })

  it('selectRange assertions', () => {
    expect(cardinal.selectRange(1, 2)).toBe('other')
    expect(cardinal.selectRange(BigInt(1), BigInt(1))).toBe('one')
    expect(ordinal.selectRange(2, 2)).toBe('two')
  })
})
`
	mustWriteFile(t, filepath.Join(pluralTestsRoot, "index.test.ts"), pluralTest)
	dateTest := `import {DateTimeFormat} from '#packages/intl-datetimeformat/core'
import {describe, it, expect} from 'vitest'

describe('datetime mechanical cases', () => {
  const dtf = new DateTimeFormat('en-US', {year: 'numeric', month: 'short', day: 'numeric', timeZone: 'UTC'})
  const start = new Date('2026-05-08T12:00:00Z')
  const end = new Date('2026-05-10T12:00:00Z')

  it('format and range assertions', () => {
    expect(dtf.format(start)).toBe('May 8, 2026')
    expect(dtf.formatRange(start, end)).toBe('May 8 – 10, 2026')
  })
})
`
	mustWriteFile(t, filepath.Join(dateTestsRoot, "index.test.ts"), dateTest)
	localeTest := `import {Locale} from '#packages/intl-locale/index.js'
import {describe, it, expect} from 'vitest'

describe('locale mechanical cases', () => {
  it('canonicalize and maximize', () => {
    expect(new Locale('en-u-foo-bar-nu-thai-ca-buddhist-kk-true').toString()).toBe('en-u-bar-foo-ca-buddhist-kk-nu-thai')
    expect(new Locale('en').maximize().toString()).toBe('en-Latn-US')
  })
})
`
	mustWriteFile(t, filepath.Join(localeTestsRoot, "index.test.ts"), localeTest)
	canonicalTest := `import {getCanonicalLocales} from '#packages/intl-getcanonicallocales/index.js'
import {describe, it, expect} from 'vitest'

describe('canonical locales mechanical cases', () => {
  it('regular', () => {
    expect(getCanonicalLocales('zh-hANs-sG')).toEqual(['zh-Hans-SG'])
  })
})
`
	mustWriteFile(t, filepath.Join(canonicalTestsRoot, "index.test.ts"), canonicalTest)
	listTest := `import ListFormat from '#packages/intl-listformat/index.js'
import {describe, it, expect} from 'vitest'

describe('list mechanical cases', () => {
  it('format assertions', () => {
    expect(new ListFormat('zh-CN', {type: 'unit'}).format(['1', '2', '3'])).toBe('123')
    expect(new ListFormat('en-AI').format(['1', '2'])).toBe('1 and 2')
  })
})
`
	mustWriteFile(t, filepath.Join(listTestsRoot, "index.test.ts"), listTest)
	relativeTimeTest := `import RelativeTimeFormat from '#packages/intl-relativetimeformat/index.js'
import {describe, it, expect} from 'vitest'

describe('relative time mechanical cases', () => {
  it('format assertions', () => {
    expect(new RelativeTimeFormat('zh-CN').format(-1, 'second')).toBe('1秒钟前')
    expect(new RelativeTimeFormat('zh-TW', {style: 'short', numeric: 'auto'}).format(-1, 'seconds')).toBe('1 秒前')
  })
})
`
	mustWriteFile(t, filepath.Join(relativeTimeTestsRoot, "index.test.ts"), relativeTimeTest)
	durationTest := `import {DurationFormat} from '#packages/intl-durationformat/core.js'
import {describe, it, expect} from 'vitest'

describe('duration mechanical cases', () => {
  it('format assertions', () => {
    expect(new DurationFormat('en').format({years: 1, months: 2})).toBe('1 yr, 2 mths')
    expect(new DurationFormat('en', {style: 'digital'}).format({hours: 5, minutes: 30, seconds: 15})).toBe('5:30:15')
    expect(new DurationFormat('en', {style: 'narrow', milliseconds: 'numeric'}).format({seconds: 1, milliseconds: 473})).toBe('1.473s')
  })
})
`
	mustWriteFile(t, filepath.Join(durationTestsRoot, "index.test.ts"), durationTest)
	out := filepath.Join(root, "out")

	if err := run([]string{"-formatjs", formatJSRoot, "-out", out}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	assertFileContainsAll(t, "formatjs number fixtures", filepath.Join(out, "numberformat", "testdata", "conformance", "formatjs", "direct-test-ts.json"),
		`"source": "formatjs:packages/intl-numberformat/tests/direct.test.ts"`,
		`"minimumFractionDigits": 2`,
		`"style": "percent"`,
		`"currency": "USD"`,
		`"expectedParts"`,
		`"expectedRange": "1–2"`,
		`"expectedRangeParts"`,
		`"$42.00"`,
	)
	assertFileContainsAll(t, "formatjs plural fixtures", filepath.Join(out, "pluralrules", "testdata", "conformance", "formatjs", "index-test-ts.json"),
		`"source": "formatjs:packages/intl-pluralrules/tests/index.test.ts"`,
		`"type": "ordinal"`,
		`"feature": "selectRange"`,
		`"notation": "compact"`,
		`"roundingMode": "trunc"`,
		`"maximumSignificantDigits": 1`,
		`"trailingZeroDisplay": "stripIfInteger"`,
		`"minimumFractionDigits": 2`,
		`"start": "1"`,
		`"end": "1"`,
		`"input": "2"`,
		`"many"`,
	)
	assertFileContainsAll(t, "formatjs date fixtures", filepath.Join(out, "datetimeformat", "testdata", "conformance", "formatjs", "index-test-ts.json"),
		`"source": "formatjs:packages/intl-datetimeformat/tests/index.test.ts"`,
		`"timeZone": "UTC"`,
		`"expected": "May 8, 2026"`,
		`"expectedRange": "May 8 – 10, 2026"`,
	)
	assertFileContainsAll(t, "formatjs locale fixtures", filepath.Join(out, "locale", "testdata", "conformance", "formatjs", "index-test-ts.json"),
		`"feature": "canonicalize"`,
		`"feature": "maximize"`,
		`"expected": "en-Latn-US"`,
	)
	assertFileContainsAll(t, "formatjs canonical locale fixtures",
		filepath.Join(out, "locale", "testdata", "conformance", "formatjs", "intl-getcanonicallocales-index-test-ts.json"),
		`"expected": "zh-Hans-SG"`,
	)
	assertFileContainsAll(t, "formatjs list fixtures", filepath.Join(out, "listformat", "testdata", "conformance", "formatjs", "index-test-ts.json"),
		`"source": "formatjs:packages/intl-listformat/tests/index.test.ts"`,
		`"type": "unit"`,
		`"input": [`,
		`"expected": "123"`,
		`"expected": "1 and 2"`,
	)
	assertFileContainsAll(t, "formatjs relative time fixtures", filepath.Join(out, "relativetimeformat", "testdata", "conformance", "formatjs", "index-test-ts.json"),
		`"source": "formatjs:packages/intl-relativetimeformat/tests/index.test.ts"`,
		`"numeric": "auto"`,
		`"unit": "seconds"`,
		`"expected": "1秒钟前"`,
		`"expected": "1 秒前"`,
	)
	assertFileContainsAll(t, "formatjs duration fixtures", filepath.Join(out, "durationformat", "testdata", "conformance", "formatjs", "index-test-ts.json"),
		`"source": "formatjs:packages/intl-durationformat/tests/index.test.ts"`,
		`"style": "digital"`,
		`"milliseconds": "numeric"`,
		`"months": 2`,
		`"expected": "1.473s"`,
		`"expected": "5:30:15"`,
	)
	skipData := mustReadFile(t, filepath.Join(out, ".skip-list.json"))
	assertContainsAll(t, "skip list", skipData,
		`format_to_parts.test.ts`,
		`"category": "`+skipCategoryUnsupportedExtractorShape+`"`,
		skipReasonUnsupportedVitestAssertionShape,
	)
	for _, retired := range []string{
		`decimal/__snapshots__/en.test.ts.snap`,
		`"category": "snapshot-source"`,
	} {
		if strings.Contains(string(skipData), retired) {
			t.Fatalf("skip list = %s, want retired snapshot-only source %s absent", skipData, retired)
		}
	}
}
