package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
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

	localeSmoke, err := os.ReadFile(filepath.Join(out, "locale", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read locale smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "locale-node-v26-canonicalize"`,
		`"source": "node:v26.0.0:locale"`,
		`"expected": "en-US-u-nu-latn"`,
	} {
		if !strings.Contains(string(localeSmoke), want) {
			t.Fatalf("locale smoke witness = %s, want %s", localeSmoke, want)
		}
	}
	localeCanonicalization, err := os.ReadFile(filepath.Join(out, "locale", "testdata", "conformance", "node-v26", "canonicalization.json"))
	if err != nil {
		t.Fatalf("read locale canonicalization: %v", err)
	}
	for _, want := range []string{
		`"id": "locale-node-v26-duplicate-calendar-first-wins"`,
		`"source": "node:v26.0.0:locale:canonicalization"`,
	} {
		if !strings.Contains(string(localeCanonicalization), want) {
			t.Fatalf("locale canonicalization witness = %s, want %s", localeCanonicalization, want)
		}
	}
	localeInfo, err := os.ReadFile(filepath.Join(out, "locale", "testdata", "conformance", "node-v26", "info.json"))
	if err != nil {
		t.Fatalf("read locale info: %v", err)
	}
	for _, want := range []string{
		`"id": "locale-node-v26-week-info-rg-override"`,
		`"source": "node:v26.0.0:locale:info"`,
		`"expectedResolvedOptions"`,
	} {
		if !strings.Contains(string(localeInfo), want) {
			t.Fatalf("locale info witness = %s, want %s", localeInfo, want)
		}
	}
	numberData, err := os.ReadFile(filepath.Join(out, "numberformat", "testdata", "conformance", "node-v26", "resolved-options.json"))
	if err != nil {
		t.Fatalf("read number resolved-options: %v", err)
	}
	for _, want := range []string{
		`"id": "numberformat-node-v26-resolved-decimal-default"`,
		`"source": "node:v26.0.0:numberformat:resolved-options"`,
		`"expectedResolvedOptions"`,
	} {
		if !strings.Contains(string(numberData), want) {
			t.Fatalf("number witness fixture = %s, want %s", numberData, want)
		}
	}
	numberSmoke, err := os.ReadFile(filepath.Join(out, "numberformat", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read number smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "numberformat-node-v26-currency-usd"`,
		`"source": "node:v26.0.0:numberformat"`,
		`"expected": "$1,234.50"`,
	} {
		if !strings.Contains(string(numberSmoke), want) {
			t.Fatalf("number smoke witness fixture = %s, want %s", numberSmoke, want)
		}
	}
	numberErrors, err := os.ReadFile(filepath.Join(out, "numberformat", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read number errors: %v", err)
	}
	for _, want := range []string{
		`"id": "numberformat-node-v26-invalid-style"`,
		`"source": "node:v26.0.0:numberformat:errors"`,
		`"errorCode": "invalid_option"`,
	} {
		if !strings.Contains(string(numberErrors), want) {
			t.Fatalf("number error witness fixture = %s, want %s", numberErrors, want)
		}
	}
	dateSmoke, err := os.ReadFile(filepath.Join(out, "datetimeformat", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read date time smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "datetimeformat-node-v26-utc-long-date"`,
		`"source": "node:v26.0.0:datetimeformat"`,
		`"expected": "January 2, 2020"`,
	} {
		if !strings.Contains(string(dateSmoke), want) {
			t.Fatalf("date time smoke witness fixture = %s, want %s", dateSmoke, want)
		}
	}
	dateErrors, err := os.ReadFile(filepath.Join(out, "datetimeformat", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read date time errors: %v", err)
	}
	for _, want := range []string{
		`"id": "datetimeformat-node-v26-invalid-date-style"`,
		`"source": "node:v26.0.0:datetimeformat:errors"`,
		`"errorCode": "invalid_option"`,
	} {
		if !strings.Contains(string(dateErrors), want) {
			t.Fatalf("date time error witness fixture = %s, want %s", dateErrors, want)
		}
	}
	dateEdge, err := os.ReadFile(filepath.Join(out, "datetimeformat", "testdata", "conformance", "node-v26", "edge.json"))
	if err != nil {
		t.Fatalf("read date time edge: %v", err)
	}
	for _, want := range []string{
		`"id": "datetimeformat-node-v26-offset-timezone-kolkata"`,
		`"source": "node:v26.0.0:datetimeformat:edge"`,
		`"expectedParts"`,
	} {
		if !strings.Contains(string(dateEdge), want) {
			t.Fatalf("date time edge witness fixture = %s, want %s", dateEdge, want)
		}
	}
	durationData, err := os.ReadFile(filepath.Join(out, "durationformat", "testdata", "conformance", "node-v26", "digital.json"))
	if err != nil {
		t.Fatalf("read duration digital: %v", err)
	}
	for _, want := range []string{
		`"id": "durationformat-node-v26-digital-hours-minutes-seconds"`,
		`"expectedParts"`,
		`"unit": "hour"`,
		`"expectedResolvedOptions"`,
	} {
		if !strings.Contains(string(durationData), want) {
			t.Fatalf("duration witness fixture = %s, want %s", durationData, want)
		}
	}
	durationSmoke, err := os.ReadFile(filepath.Join(out, "durationformat", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read duration smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "durationformat-node-v26-hours-minutes"`,
		`"source": "node:v26.0.0:durationformat"`,
		`"expected": "1 hr, 2 min"`,
	} {
		if !strings.Contains(string(durationSmoke), want) {
			t.Fatalf("duration smoke witness fixture = %s, want %s", durationSmoke, want)
		}
	}
	durationErrors, err := os.ReadFile(filepath.Join(out, "durationformat", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read duration errors: %v", err)
	}
	for _, want := range []string{
		`"id": "durationformat-node-v26-invalid-style"`,
		`"source": "node:v26.0.0:durationformat:errors"`,
		`"errorCode": "invalid_option"`,
	} {
		if !strings.Contains(string(durationErrors), want) {
			t.Fatalf("duration error witness fixture = %s, want %s", durationErrors, want)
		}
	}
	listSmoke, err := os.ReadFile(filepath.Join(out, "listformat", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read list smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "listformat-node-v26-conjunction-long"`,
		`"source": "node:v26.0.0:listformat"`,
		`"expected": "A, B, and C"`,
	} {
		if !strings.Contains(string(listSmoke), want) {
			t.Fatalf("list smoke witness fixture = %s, want %s", listSmoke, want)
		}
	}
	listErrors, err := os.ReadFile(filepath.Join(out, "listformat", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read list errors: %v", err)
	}
	for _, want := range []string{
		`"id": "listformat-node-v26-invalid-style"`,
		`"source": "node:v26.0.0:listformat:errors"`,
		`"errorCode": "invalid_option"`,
	} {
		if !strings.Contains(string(listErrors), want) {
			t.Fatalf("list error witness fixture = %s, want %s", listErrors, want)
		}
	}
	relativeSmoke, err := os.ReadFile(filepath.Join(out, "relativetimeformat", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read relative time smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "relativetimeformat-node-v26-day-auto"`,
		`"source": "node:v26.0.0:relativetimeformat"`,
		`"expected": "yesterday"`,
	} {
		if !strings.Contains(string(relativeSmoke), want) {
			t.Fatalf("relative time smoke witness fixture = %s, want %s", relativeSmoke, want)
		}
	}
	relativeErrors, err := os.ReadFile(filepath.Join(out, "relativetimeformat", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read relative time errors: %v", err)
	}
	for _, want := range []string{
		`"id": "relativetimeformat-node-v26-invalid-numeric"`,
		`"source": "node:v26.0.0:relativetimeformat:errors"`,
		`"errorCode": "invalid_option"`,
	} {
		if !strings.Contains(string(relativeErrors), want) {
			t.Fatalf("relative time error witness fixture = %s, want %s", relativeErrors, want)
		}
	}
	pluralSmoke, err := os.ReadFile(filepath.Join(out, "pluralrules", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read plural rules smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "pluralrules-node-v26-ordinal-two"`,
		`"source": "node:v26.0.0:pluralrules"`,
		`"expected": "two"`,
	} {
		if !strings.Contains(string(pluralSmoke), want) {
			t.Fatalf("plural rules smoke witness fixture = %s, want %s", pluralSmoke, want)
		}
	}
	pluralErrors, err := os.ReadFile(filepath.Join(out, "pluralrules", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read plural rules errors: %v", err)
	}
	for _, want := range []string{
		`"id": "pluralrules-node-v26-invalid-type"`,
		`"source": "node:v26.0.0:pluralrules:errors"`,
		`"errorCode": "invalid_option"`,
	} {
		if !strings.Contains(string(pluralErrors), want) {
			t.Fatalf("plural rules error witness fixture = %s, want %s", pluralErrors, want)
		}
	}
	displayNamesSmoke, err := os.ReadFile(filepath.Join(out, "displaynames", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read display names smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "displaynames-node-v26-region-us"`,
		`"source": "node:v26.0.0:displaynames"`,
		`"expected": "United States"`,
		`"expectedOk": true`,
		`"expectedResolvedOptions"`,
	} {
		if !strings.Contains(string(displayNamesSmoke), want) {
			t.Fatalf("display names smoke witness fixture = %s, want %s", displayNamesSmoke, want)
		}
	}
	displayNamesErrors, err := os.ReadFile(filepath.Join(out, "displaynames", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read display names errors: %v", err)
	}
	for _, want := range []string{
		`"id": "displaynames-node-v26-invalid-type"`,
		`"source": "node:v26.0.0:displaynames:errors"`,
		`"errorCode": "invalidOption"`,
	} {
		if !strings.Contains(string(displayNamesErrors), want) {
			t.Fatalf("display names error witness fixture = %s, want %s", displayNamesErrors, want)
		}
	}
	collatorSmoke, err := os.ReadFile(filepath.Join(out, "collator", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read collator smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "collator-node-v26-basic-order"`,
		`"source": "node:v26.0.0:collator"`,
		`"expectedComparison": -1`,
	} {
		if !strings.Contains(string(collatorSmoke), want) {
			t.Fatalf("collator smoke witness fixture = %s, want %s", collatorSmoke, want)
		}
	}
	collatorErrors, err := os.ReadFile(filepath.Join(out, "collator", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read collator errors: %v", err)
	}
	for _, want := range []string{
		`"id": "collator-node-v26-invalid-sensitivity"`,
		`"source": "node:v26.0.0:collator:errors"`,
		`"errorCode": "invalidOption"`,
	} {
		if !strings.Contains(string(collatorErrors), want) {
			t.Fatalf("collator error witness fixture = %s, want %s", collatorErrors, want)
		}
	}
	collatorOptions, err := os.ReadFile(filepath.Join(out, "collator", "testdata", "conformance", "node-v26", "options.json"))
	if err != nil {
		t.Fatalf("read collator options: %v", err)
	}
	for _, want := range []string{
		`"id": "collator-node-v26-numeric-locale-extension-contract"`,
		`"source": "node:v26.0.0:collator:option-contract"`,
		`"expectedResolvedOptions"`,
	} {
		if !strings.Contains(string(collatorOptions), want) {
			t.Fatalf("collator option witness fixture = %s, want %s", collatorOptions, want)
		}
	}
	collatorBackendProof, err := os.ReadFile(filepath.Join(out, "collator", "testdata", "conformance", "node-v26", "backend-proof.json"))
	if err != nil {
		t.Fatalf("read collator backend proof: %v", err)
	}
	for _, want := range []string{
		`"id": "collator-node-v26-swedish-z-before-a-ring"`,
		`"source": "node:v26.0.0:collator:backend-proof"`,
		`"expectedResolvedOptions"`,
	} {
		if !strings.Contains(string(collatorBackendProof), want) {
			t.Fatalf("collator backend proof witness fixture = %s, want %s", collatorBackendProof, want)
		}
	}
	segmenterSmoke, err := os.ReadFile(filepath.Join(out, "segmenter", "testdata", "conformance", "node-v26", "smoke.json"))
	if err != nil {
		t.Fatalf("read segmenter smoke: %v", err)
	}
	for _, want := range []string{
		`"id": "segmenter-node-v26-word-hello-world"`,
		`"source": "node:v26.0.0:segmenter"`,
		`"expectedSegments"`,
	} {
		if !strings.Contains(string(segmenterSmoke), want) {
			t.Fatalf("segmenter smoke witness fixture = %s, want %s", segmenterSmoke, want)
		}
	}
	segmenterErrors, err := os.ReadFile(filepath.Join(out, "segmenter", "testdata", "conformance", "node-v26", "errors.json"))
	if err != nil {
		t.Fatalf("read segmenter errors: %v", err)
	}
	for _, want := range []string{
		`"id": "segmenter-node-v26-invalid-granularity"`,
		`"source": "node:v26.0.0:segmenter:errors"`,
		`"errorCode": "invalid_option"`,
	} {
		if !strings.Contains(string(segmenterErrors), want) {
			t.Fatalf("segmenter error witness fixture = %s, want %s", segmenterErrors, want)
		}
	}
	segmenterLocale, err := os.ReadFile(filepath.Join(out, "segmenter", "testdata", "conformance", "node-v26", "locale-contract.json"))
	if err != nil {
		t.Fatalf("read segmenter locale contract: %v", err)
	}
	for _, want := range []string{
		`"id": "segmenter-node-v26-en-word-contract"`,
		`"source": "node:v26.0.0:segmenter:locale-contract"`,
		`"expectedSegments"`,
	} {
		if !strings.Contains(string(segmenterLocale), want) {
			t.Fatalf("segmenter locale contract fixture = %s, want %s", segmenterLocale, want)
		}
	}
	segmenterTailored, err := os.ReadFile(filepath.Join(out, "segmenter", "testdata", "conformance", "node-v26", "tailored-locale-contract.json"))
	if err != nil {
		t.Fatalf("read segmenter tailored contract: %v", err)
	}
	for _, want := range []string{
		`"id": "segmenter-node-v26-th-word-tailored-contract"`,
		`"source": "node:v26.0.0:segmenter:tailored-locale-contract"`,
		`"expectedSegments"`,
	} {
		if !strings.Contains(string(segmenterTailored), want) {
			t.Fatalf("segmenter tailored contract fixture = %s, want %s", segmenterTailored, want)
		}
	}
	supportedData, err := os.ReadFile(filepath.Join(out, "testdata", "native", "node-v26", "supported-values.json"))
	if err != nil {
		t.Fatalf("read supported-values witness: %v", err)
	}
	for _, want := range []string{
		`"source": "node:v26.0.0:intl:supportedValuesOf"`,
		`"calendar"`,
		`"collation"`,
		`"unit"`,
		`"icu": "78.1"`,
	} {
		if !strings.Contains(string(supportedData), want) {
			t.Fatalf("supported-values witness = %s, want %s", supportedData, want)
		}
	}
	if _, err := os.Stat(filepath.Join(out, ".skip-list.json")); !os.IsNotExist(err) {
		t.Fatalf("node-only import wrote .skip-list.json: %v", err)
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
	generated[filepath.Join("testdata", "native", nodeDir, "supported-values.json")] = true
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
		if d.IsDir() || filepath.Ext(path) != ".json" {
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
	numberTestsRoot := filepath.Join(formatJSRoot, "packages", "intl-numberformat", "tests")
	pluralTestsRoot := filepath.Join(formatJSRoot, "packages", "intl-pluralrules", "tests")
	dateTestsRoot := filepath.Join(formatJSRoot, "packages", "intl-datetimeformat", "tests")
	localeTestsRoot := filepath.Join(formatJSRoot, "packages", "intl-locale", "tests")
	canonicalTestsRoot := filepath.Join(formatJSRoot, "packages", "intl-getcanonicallocales", "tests")
	listTestsRoot := filepath.Join(formatJSRoot, "packages", "intl-listformat", "tests")
	relativeTimeTestsRoot := filepath.Join(formatJSRoot, "packages", "intl-relativetimeformat", "tests")
	durationTestsRoot := filepath.Join(formatJSRoot, "packages", "intl-durationformat", "tests")
	if err := os.MkdirAll(filepath.Join(numberTestsRoot, "decimal", "__snapshots__"), 0o777); err != nil {
		t.Fatalf("mkdir formatjs tree: %v", err)
	}
	for _, dir := range []string{pluralTestsRoot, dateTestsRoot, localeTestsRoot, canonicalTestsRoot, listTestsRoot, relativeTimeTestsRoot, durationTestsRoot} {
		if err := os.MkdirAll(dir, 0o777); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
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
	if err := os.WriteFile(filepath.Join(numberTestsRoot, "direct.test.ts"), []byte(directTest), 0o666); err != nil {
		t.Fatalf("write direct test: %v", err)
	}
	if err := os.WriteFile(filepath.Join(numberTestsRoot, "format_to_parts.test.ts"), []byte(`expect(parts).toEqual([{type: 'integer', value: '0'}])`), 0o666); err != nil {
		t.Fatalf("write unsupported test: %v", err)
	}
	if err := os.WriteFile(filepath.Join(numberTestsRoot, "decimal", "__snapshots__", "en.test.ts.snap"), []byte("exports[`snapshot 1`] = `\"1\"`;"), 0o666); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}
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
	if err := os.WriteFile(filepath.Join(pluralTestsRoot, "index.test.ts"), []byte(pluralTest), 0o666); err != nil {
		t.Fatalf("write plural test: %v", err)
	}
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
	if err := os.WriteFile(filepath.Join(dateTestsRoot, "index.test.ts"), []byte(dateTest), 0o666); err != nil {
		t.Fatalf("write date test: %v", err)
	}
	localeTest := `import {Locale} from '#packages/intl-locale/index.js'
import {describe, it, expect} from 'vitest'

describe('locale mechanical cases', () => {
  it('canonicalize and maximize', () => {
    expect(new Locale('en-u-foo-bar-nu-thai-ca-buddhist-kk-true').toString()).toBe('en-u-bar-foo-ca-buddhist-kk-nu-thai')
    expect(new Locale('en').maximize().toString()).toBe('en-Latn-US')
  })
})
`
	if err := os.WriteFile(filepath.Join(localeTestsRoot, "index.test.ts"), []byte(localeTest), 0o666); err != nil {
		t.Fatalf("write locale test: %v", err)
	}
	canonicalTest := `import {getCanonicalLocales} from '#packages/intl-getcanonicallocales/index.js'
import {describe, it, expect} from 'vitest'

describe('canonical locales mechanical cases', () => {
  it('regular', () => {
    expect(getCanonicalLocales('zh-hANs-sG')).toEqual(['zh-Hans-SG'])
  })
})
`
	if err := os.WriteFile(filepath.Join(canonicalTestsRoot, "index.test.ts"), []byte(canonicalTest), 0o666); err != nil {
		t.Fatalf("write canonical locale test: %v", err)
	}
	listTest := `import ListFormat from '#packages/intl-listformat/index.js'
import {describe, it, expect} from 'vitest'

describe('list mechanical cases', () => {
  it('format assertions', () => {
    expect(new ListFormat('zh-CN', {type: 'unit'}).format(['1', '2', '3'])).toBe('123')
    expect(new ListFormat('en-AI').format(['1', '2'])).toBe('1 and 2')
  })
})
`
	if err := os.WriteFile(filepath.Join(listTestsRoot, "index.test.ts"), []byte(listTest), 0o666); err != nil {
		t.Fatalf("write list test: %v", err)
	}
	relativeTimeTest := `import RelativeTimeFormat from '#packages/intl-relativetimeformat/index.js'
import {describe, it, expect} from 'vitest'

describe('relative time mechanical cases', () => {
  it('format assertions', () => {
    expect(new RelativeTimeFormat('zh-CN').format(-1, 'second')).toBe('1秒钟前')
    expect(new RelativeTimeFormat('zh-TW', {style: 'short', numeric: 'auto'}).format(-1, 'seconds')).toBe('1 秒前')
  })
})
`
	if err := os.WriteFile(filepath.Join(relativeTimeTestsRoot, "index.test.ts"), []byte(relativeTimeTest), 0o666); err != nil {
		t.Fatalf("write relative time test: %v", err)
	}
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
	if err := os.WriteFile(filepath.Join(durationTestsRoot, "index.test.ts"), []byte(durationTest), 0o666); err != nil {
		t.Fatalf("write duration test: %v", err)
	}
	out := filepath.Join(root, "out")

	if err := run([]string{"-formatjs", formatJSRoot, "-out", out}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	numberData, err := os.ReadFile(filepath.Join(out, "numberformat", "testdata", "conformance", "formatjs", "direct-test-ts.json"))
	if err != nil {
		t.Fatalf("read formatjs number fixtures: %v", err)
	}
	for _, want := range []string{
		`"source": "formatjs:packages/intl-numberformat/tests/direct.test.ts"`,
		`"minimumFractionDigits": 2`,
		`"style": "percent"`,
		`"currency": "USD"`,
		`"expectedParts"`,
		`"expectedRange": "1–2"`,
		`"expectedRangeParts"`,
		`"$42.00"`,
	} {
		if !strings.Contains(string(numberData), want) {
			t.Fatalf("formatjs number fixtures = %s, want %s", numberData, want)
		}
	}
	pluralData, err := os.ReadFile(filepath.Join(out, "pluralrules", "testdata", "conformance", "formatjs", "index-test-ts.json"))
	if err != nil {
		t.Fatalf("read formatjs plural fixtures: %v", err)
	}
	for _, want := range []string{
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
	} {
		if !strings.Contains(string(pluralData), want) {
			t.Fatalf("formatjs plural fixtures = %s, want %s", pluralData, want)
		}
	}
	dateData, err := os.ReadFile(filepath.Join(out, "datetimeformat", "testdata", "conformance", "formatjs", "index-test-ts.json"))
	if err != nil {
		t.Fatalf("read formatjs date fixtures: %v", err)
	}
	for _, want := range []string{
		`"source": "formatjs:packages/intl-datetimeformat/tests/index.test.ts"`,
		`"timeZone": "UTC"`,
		`"expected": "May 8, 2026"`,
		`"expectedRange": "May 8 – 10, 2026"`,
	} {
		if !strings.Contains(string(dateData), want) {
			t.Fatalf("formatjs date fixtures = %s, want %s", dateData, want)
		}
	}
	localeData, err := os.ReadFile(filepath.Join(out, "locale", "testdata", "conformance", "formatjs", "index-test-ts.json"))
	if err != nil {
		t.Fatalf("read formatjs locale fixtures: %v", err)
	}
	for _, want := range []string{
		`"feature": "canonicalize"`,
		`"feature": "maximize"`,
		`"expected": "en-Latn-US"`,
	} {
		if !strings.Contains(string(localeData), want) {
			t.Fatalf("formatjs locale fixtures = %s, want %s", localeData, want)
		}
	}
	canonicalData, err := os.ReadFile(filepath.Join(out, "locale", "testdata", "conformance", "formatjs", "intl-getcanonicallocales-index-test-ts.json"))
	if err != nil {
		t.Fatalf("read formatjs canonical locale fixtures: %v", err)
	}
	if !strings.Contains(string(canonicalData), `"expected": "zh-Hans-SG"`) {
		t.Fatalf("formatjs canonical locale fixtures = %s, want zh-Hans-SG", canonicalData)
	}
	listData, err := os.ReadFile(filepath.Join(out, "listformat", "testdata", "conformance", "formatjs", "index-test-ts.json"))
	if err != nil {
		t.Fatalf("read formatjs list fixtures: %v", err)
	}
	for _, want := range []string{
		`"source": "formatjs:packages/intl-listformat/tests/index.test.ts"`,
		`"type": "unit"`,
		`"input": [`,
		`"expected": "123"`,
		`"expected": "1 and 2"`,
	} {
		if !strings.Contains(string(listData), want) {
			t.Fatalf("formatjs list fixtures = %s, want %s", listData, want)
		}
	}
	relativeTimeData, err := os.ReadFile(filepath.Join(out, "relativetimeformat", "testdata", "conformance", "formatjs", "index-test-ts.json"))
	if err != nil {
		t.Fatalf("read formatjs relative time fixtures: %v", err)
	}
	for _, want := range []string{
		`"source": "formatjs:packages/intl-relativetimeformat/tests/index.test.ts"`,
		`"numeric": "auto"`,
		`"unit": "seconds"`,
		`"expected": "1秒钟前"`,
		`"expected": "1 秒前"`,
	} {
		if !strings.Contains(string(relativeTimeData), want) {
			t.Fatalf("formatjs relative time fixtures = %s, want %s", relativeTimeData, want)
		}
	}
	durationData, err := os.ReadFile(filepath.Join(out, "durationformat", "testdata", "conformance", "formatjs", "index-test-ts.json"))
	if err != nil {
		t.Fatalf("read formatjs duration fixtures: %v", err)
	}
	for _, want := range []string{
		`"source": "formatjs:packages/intl-durationformat/tests/index.test.ts"`,
		`"style": "digital"`,
		`"milliseconds": "numeric"`,
		`"months": 2`,
		`"expected": "1.473s"`,
		`"expected": "5:30:15"`,
	} {
		if !strings.Contains(string(durationData), want) {
			t.Fatalf("formatjs duration fixtures = %s, want %s", durationData, want)
		}
	}
	skipData, err := os.ReadFile(filepath.Join(out, ".skip-list.json"))
	if err != nil {
		t.Fatalf("read skip list: %v", err)
	}
	for _, want := range []string{
		`format_to_parts.test.ts`,
		`"category": "unsupported-extractor-shape"`,
		`unsupported Vitest assertion shape`,
	} {
		if !strings.Contains(string(skipData), want) {
			t.Fatalf("skip list = %s, want %s", skipData, want)
		}
	}
	for _, retired := range []string{
		`decimal/__snapshots__/en.test.ts.snap`,
		`"category": "snapshot-source"`,
	} {
		if strings.Contains(string(skipData), retired) {
			t.Fatalf("skip list = %s, want retired snapshot-only source %s absent", skipData, retired)
		}
	}
}
