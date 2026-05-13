package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunImportsSyntheticNodeSmokeFixtures(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	nodePath := filepath.Join(root, "localizationData-v76.1.json")
	if err := os.WriteFile(nodePath, []byte(`{
		"en-US": {
			"numberFormats": [{"input": 1234.5, "expected": "1,234.5"}],
			"dateFormats": [{"input": "2026-05-08T12:00:00Z", "expected": "May 8, 2026"}]
		}
	}`), 0o666); err != nil {
		t.Fatalf("write node fixture: %v", err)
	}
	out := filepath.Join(root, "out")

	if err := run([]string{"-node", nodePath, "-out", out}); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	numberData, err := os.ReadFile(filepath.Join(out, "numberformat", "testdata", "conformance", "node-v76", "smoke.json"))
	if err != nil {
		t.Fatalf("read number smoke: %v", err)
	}
	if !strings.Contains(string(numberData), "numberformat-node-v76-en-us-0") {
		t.Fatalf("number smoke fixture = %s, want deterministic ID", numberData)
	}
	dateData, err := os.ReadFile(filepath.Join(out, "datetimeformat", "testdata", "conformance", "node-v76", "smoke.json"))
	if err != nil {
		t.Fatalf("read date smoke: %v", err)
	}
	if !strings.Contains(string(dateData), "datetimeformat-node-v76-en-us-0") {
		t.Fatalf("date smoke fixture = %s, want deterministic ID", dateData)
	}
	skipData, err := os.ReadFile(filepath.Join(out, ".skip-list.json"))
	if err != nil {
		t.Fatalf("read skip list: %v", err)
	}
	if !strings.Contains(string(skipData), `"source": "formatjs"`) || !strings.Contains(string(skipData), "formatjs path not provided") {
		t.Fatalf("skip list = %s, want missing FormatJS entry", skipData)
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
	if err := os.MkdirAll(filepath.Join(numberTestsRoot, "decimal", "__snapshots__"), 0o777); err != nil {
		t.Fatalf("mkdir formatjs tree: %v", err)
	}
	for _, dir := range []string{pluralTestsRoot, dateTestsRoot, localeTestsRoot, canonicalTestsRoot} {
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
	skipData, err := os.ReadFile(filepath.Join(out, ".skip-list.json"))
	if err != nil {
		t.Fatalf("read skip list: %v", err)
	}
	for _, want := range []string{
		`format_to_parts.test.ts`,
		`"category": "unsupported-extractor-shape"`,
		`unsupported Vitest assertion shape`,
		`decimal/__snapshots__/en.test.ts.snap`,
		`"category": "snapshot-source"`,
		`snapshot output requires source test input mapping`,
	} {
		if !strings.Contains(string(skipData), want) {
			t.Fatalf("skip list = %s, want %s", skipData, want)
		}
	}
}
