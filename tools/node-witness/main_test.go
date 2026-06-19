package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPrintsValidatedWitness(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nodePath := filepath.Join(root, "node")
	if err := os.WriteFile(nodePath, []byte(`#!/bin/sh
cat <<'JSON'
{
  "nodeVersion": "v26.0.0",
  "versions": {"node": "26.0.0", "icu": "78.1"},
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
  "localeCanonicalization": [],
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
  "durationFormatDigital": [],
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
    "versions": {"node": "26.0.0", "icu": "78.1"},
    "values": {"calendar": ["gregory"]}
  }
}
JSON
`), 0o777); err != nil {
		t.Fatalf("write fake node executable: %v", err)
	}

	var out bytes.Buffer
	if err := run([]string{"-node", nodePath}, &out); err != nil {
		t.Fatalf("run() error = %v", err)
	}
	for _, want := range []string{
		`"nodeVersion": "v26.0.0"`,
		`"localeSmoke": [`,
		`"localeInfo": [`,
		`"numberFormatSmoke": [`,
		`"numberFormatErrors": [`,
		`"numberFormatResolved": [`,
		`"dateTimeFormatSmoke": [`,
		`"dateTimeFormatErrors": [`,
		`"dateTimeFormatEdge": [`,
		`"durationFormatSmoke": [`,
		`"durationFormatErrors": [`,
		`"listFormatSmoke": [`,
		`"listFormatErrors": [`,
		`"relativeTimeSmoke": [`,
		`"relativeTimeErrors": [`,
		`"pluralRulesSmoke": [`,
		`"pluralRulesErrors": [`,
		`"displayNamesSmoke": [`,
		`"displayNamesErrors": [`,
		`"collatorSmoke": [`,
		`"collatorErrors": [`,
		`"collatorOptions": [`,
		`"collatorBackendProof": [`,
		`"segmenterSmoke": [`,
		`"segmenterErrors": [`,
		`"segmenterLocale": [`,
		`"segmenterTailored": [`,
		`"supportedValues": {`,
		`"calendar": [`,
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("witness output = %s, want %s", out.String(), want)
		}
	}
}

func TestRunRejectsMalformedWitness(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nodePath := filepath.Join(root, "node")
	if err := os.WriteFile(nodePath, []byte(`#!/bin/sh
echo '{}'
`), 0o777); err != nil {
		t.Fatalf("write fake node executable: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"-node", nodePath}, &out)
	if err == nil || !strings.Contains(err.Error(), "missing nodeVersion") {
		t.Fatalf("run() error = %v, want missing nodeVersion", err)
	}
}
