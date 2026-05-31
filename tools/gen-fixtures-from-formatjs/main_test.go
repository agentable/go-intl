package main

import (
	"os"
	"path/filepath"
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
