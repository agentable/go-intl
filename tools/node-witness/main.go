package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/agentable/go-intl/tools/conformance"
)

type witness struct {
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
	SupportedValues        nodeSupportedValues `json:"supportedValues"`
}

type fixture struct {
	ID                 string                  `json:"id"`
	Source             string                  `json:"source"`
	Locale             string                  `json:"locale"`
	Feature            string                  `json:"feature,omitempty"`
	Options            map[string]any          `json:"options"`
	Input              any                     `json:"input"`
	Expected           *string                 `json:"expected,omitempty"`
	ExpectedOK         *bool                   `json:"expectedOk,omitempty"`
	ExpectedLocales    []string                `json:"expectedLocales,omitempty"`
	ExpectedParts      []conformance.Part      `json:"expectedParts,omitempty"`
	ExpectedRange      *string                 `json:"expectedRange,omitempty"`
	ExpectedRangeParts []conformance.RangePart `json:"expectedRangeParts,omitempty"`
	ExpectedResolved   any                     `json:"expectedResolvedOptions,omitempty"`
	ErrorCode          string                  `json:"errorCode,omitempty"`
}

type nodeSupportedValues struct {
	Source   string              `json:"source"`
	Versions map[string]string   `json:"versions"`
	Values   map[string][]string `json:"values"`
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("node-witness", flag.ContinueOnError)
	nodePath := fs.String("node", "node", "Node executable path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	data, err := runNode(*nodePath)
	if err != nil {
		return err
	}
	var witness witness
	if err := json.Unmarshal(data, &witness); err != nil {
		return fmt.Errorf("decode node witness output: %w", err)
	}
	if witness.NodeVersion == "" {
		return fmt.Errorf("node witness output missing nodeVersion")
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(witness)
}

func runNode(nodePath string) ([]byte, error) {
	cmd := exec.CommandContext(context.Background(), nodePath, "-e", nodeWitnessScript)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run node witness %q: %w: %s", nodePath, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
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
const nodeFixtureIDPrefix = 'node-v' + nodeMajor;
const nodeSourcePrefix = 'node:' + nodeVersion;

function sourceRoot(surface) {
  return nodeSourcePrefix + ':' + surface;
}

function source(surface, topic) {
  return sourceRoot(surface) + ':' + topic;
}

function id(surface, topic) {
  return surface + '-' + nodeFixtureIDPrefix + '-' + topic;
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

function numberFormatSmokeFixture(topic, locale, options, input) {
  const format = new Intl.NumberFormat(locale, options);
  return expectedRootFixture('numberformat', topic, locale, options, input, format.format(input));
}

function numberFormatEdgeFixture(topic, locale, options, input) {
  const format = new Intl.NumberFormat(locale, options);
  return {
    id: id('numberformat', topic),
    source: source('numberformat', 'edge'),
    locale,
    options,
    input,
    expected: format.format(input),
    expectedParts: format.formatToParts(input),
    expectedResolvedOptions: format.resolvedOptions(),
  };
}

function numberFormatRangeEdgeFixture(topic, locale, options, start, end) {
  const format = new Intl.NumberFormat(locale, options);
  return {
    id: id('numberformat', topic),
    source: source('numberformat', 'edge'),
    locale,
    options,
    input: {start, end},
    expectedRange: format.formatRange(start, end),
    expectedRangeParts: format.formatRangeToParts(start, end),
    expectedResolvedOptions: format.resolvedOptions(),
  };
}

function numberFormatErrorFixture(topic, locale, options, input) {
  return constructorErrorFixture('numberformat', topic, locale, options, input, Intl.NumberFormat);
}

function localeFixture(topic, sourceTopic, feature, input) {
  const locale = new Intl.Locale(input);
  let expected;
  switch (feature) {
    case 'canonicalize':
      expected = locale.toString();
      break;
    case 'maximize':
      expected = locale.maximize().toString();
      break;
    case 'minimize':
      expected = locale.minimize().toString();
      break;
    default:
      throw new Error('unsupported locale feature ' + feature);
  }
  return {
    id: id('locale', topic),
    source: sourceTopic ? source('locale', sourceTopic) : sourceRoot('locale'),
    locale: input,
    feature,
    options: {},
    input,
    expected,
  };
}

function localeInfoFixture(topic, feature, input) {
  const locale = new Intl.Locale(input);
  let expectedResolvedOptions;
  switch (feature) {
    case 'weekInfo':
      expectedResolvedOptions = typeof locale.getWeekInfo === 'function' ? locale.getWeekInfo() : locale.weekInfo;
      break;
    default:
      throw new Error('unsupported locale info feature ' + feature);
  }
  return {
    id: id('locale', topic),
    source: source('locale', 'info'),
    locale: input,
    feature,
    options: {},
    input,
    expectedResolvedOptions,
  };
}

function dateTimeFormatSmokeFixture(topic, locale, options, input) {
  const format = new Intl.DateTimeFormat(locale, options);
  return expectedRootFixture('datetimeformat', topic, locale, options, input, format.format(new Date(input)));
}

function dateTimeFormatPartsFixture(topic, sourceTopic, feature, locale, options, input) {
  const format = new Intl.DateTimeFormat(locale, options);
  const date = new Date(input);
  return {
    id: id('datetimeformat', topic),
    source: source('datetimeformat', sourceTopic),
    locale,
    feature,
    options,
    input,
    expected: format.format(date),
    expectedParts: format.formatToParts(date),
    expectedResolvedOptions: format.resolvedOptions(),
  };
}

function dateTimeFormatRangeFixture(topic, sourceTopic, locale, options, start, end) {
  const format = new Intl.DateTimeFormat(locale, options);
  const startDate = new Date(start);
  const endDate = new Date(end);
  return {
    id: id('datetimeformat', topic),
    source: source('datetimeformat', sourceTopic),
    locale,
    feature: 'formatRange',
    options,
    input: {start, end},
    expectedRange: format.formatRange(startDate, endDate),
    expectedRangeParts: format.formatRangeToParts(startDate, endDate),
    expectedResolvedOptions: format.resolvedOptions(),
  };
}

function dateTimeFormatErrorFixture(topic, locale, options, input) {
  return constructorErrorFixture('datetimeformat', topic, locale, options, input, Intl.DateTimeFormat);
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

function durationFormatSmokeFixture(topic, locale, options, input) {
  const format = new Intl.DurationFormat(locale, options);
  return expectedRootFixture('durationformat', topic, locale, options, input, format.format(input));
}

function durationFormatErrorFixture(topic, locale, options, input) {
  return constructorErrorFixture('durationformat', topic, locale, options, input, Intl.DurationFormat);
}

function listFormatSmokeFixture(topic, locale, options, input) {
  const format = new Intl.ListFormat(locale, options);
  return expectedRootFixture('listformat', topic, locale, options, input, format.format(input));
}

function listFormatErrorFixture(topic, locale, options, input) {
  return constructorErrorFixture('listformat', topic, locale, options, input, Intl.ListFormat);
}

function relativeTimeFormatSmokeFixture(topic, locale, options, input) {
  const format = new Intl.RelativeTimeFormat(locale, options);
  return expectedRootFixture('relativetimeformat', topic, locale, options, input, format.format(input.value, input.unit));
}

function relativeTimeFormatErrorFixture(topic, locale, options, input) {
  return constructorErrorFixture('relativetimeformat', topic, locale, options, input, Intl.RelativeTimeFormat);
}

function pluralRulesSmokeFixture(topic, locale, options, input) {
  const rules = new Intl.PluralRules(locale, options);
  return {
    id: id('pluralrules', topic),
    source: sourceRoot('pluralrules'),
    locale,
    feature: 'select',
    options,
    input,
    expected: rules.select(input),
  };
}

function pluralRulesErrorFixture(topic, locale, options, input) {
  return constructorErrorFixture('pluralrules', topic, locale, options, input, Intl.PluralRules);
}

function displayNamesSmokeFixture(topic, locale, options, input) {
  const names = new Intl.DisplayNames(locale, options);
  const value = names.of(input);
  return {
    id: id('displaynames', topic),
    source: sourceRoot('displaynames'),
    locale,
    options,
    input,
    expected: value === undefined ? '' : value,
    expectedOk: value !== undefined,
    expectedResolvedOptions: names.resolvedOptions(),
  };
}

function displayNamesErrorFixture(topic, locale, options, input) {
  return constructorErrorFixture('displaynames', topic, locale, options, input, Intl.DisplayNames, 'invalidOption');
}

function constructorErrorFixture(surface, topic, locale, options, input, constructor, errorCode = 'invalid_option') {
  try {
    new constructor(locale, options);
  } catch (error) {
    return {
      id: id(surface, topic),
      source: source(surface, 'errors'),
      locale,
      options,
      input,
      errorCode,
    };
  }
  throw new Error('expected Intl.' + surface + ' to reject ' + topic);
}

function expectedRootFixture(surface, topic, locale, options, input, expected) {
  return {
    id: id(surface, topic),
    source: sourceRoot(surface),
    locale,
    options,
    input,
    expected,
  };
}

const supportedValues = {};
for (const key of ['calendar', 'currency', 'numberingSystem', 'timeZone', 'unit']) {
  supportedValues[key] = Intl.supportedValuesOf(key);
}

const witness = {
  nodeVersion,
  versions: selectedVersions,
  localeSmoke: [
    localeFixture('canonicalize', '', 'canonicalize', 'EN-us-u-NU-LATN'),
    localeFixture('maximize-en', '', 'maximize', 'en'),
  ],
  localeCanonicalization: [
    localeFixture('duplicate-calendar-first-wins', 'canonicalization', 'canonicalize', 'en-u-ca-buddhist-ca-gregory'),
    localeFixture('duplicate-case-first-first-wins', 'canonicalization', 'canonicalize', 'en-u-kf-upper-kf-lower'),
    localeFixture('private-use-after-unicode-extension', 'canonicalization', 'canonicalize', 'en-u-ca-gregory-x-private'),
  ],
  localeInfo: [
    localeInfoFixture('week-info-rg-override', 'weekInfo', 'en-US-u-rg-gbzzzz'),
    localeInfoFixture('week-info-sd-region', 'weekInfo', 'und-u-sd-gbeng'),
    localeInfoFixture('week-info-fw-override', 'weekInfo', 'en-US-u-fw-mon'),
  ],
  numberFormatSmoke: [
    numberFormatSmokeFixture('currency-usd', 'en-US', {style: 'currency', currency: 'USD'}, 1234.5),
  ],
  numberFormatErrors: [
    numberFormatErrorFixture('invalid-style', 'en-US', {style: 'invalid'}, 1),
    numberFormatErrorFixture('unit-casing-rejected', 'en', {style: 'unit', unit: 'METER'}, 1),
  ],
  numberFormatEdge: [
    numberFormatEdgeFixture('negative-zero-sign', 'en', {signDisplay: 'auto'}, '-0'),
    numberFormatEdgeFixture('rounding-increment', 'en', {minimumFractionDigits: 2, maximumFractionDigits: 2, roundingIncrement: 5}, 1.234),
    numberFormatEdgeFixture('rounding-priority-more-precision', 'en', {minimumSignificantDigits: 2, maximumFractionDigits: 0, roundingPriority: 'morePrecision'}, 1.234),
    numberFormatEdgeFixture('compact-plural-few', 'ru', {notation: 'compact', compactDisplay: 'long'}, 2000),
    numberFormatRangeEdgeFixture('range-collapse', 'en', {maximumFractionDigits: 0}, 1.2, 1.4),
    numberFormatRangeEdgeFixture('czech-plural-range-unit', 'cs', {style: 'unit', unit: 'meter', unitDisplay: 'long'}, 2, 1),
    numberFormatRangeEdgeFixture('czech-plural-range-currency-name', 'cs', {style: 'currency', currency: 'USD', currencyDisplay: 'name', maximumFractionDigits: 0}, 2, 1),
    numberFormatRangeEdgeFixture('negative-percent-range-affixes', 'en', {style: 'percent', maximumFractionDigits: 0}, -0.01, -0.02),
  ],
  numberFormatResolved: [
    numberFormatFixture('resolved-decimal-default', 'en', {}, 12345.6),
    numberFormatFixture('resolved-significant-digits', 'en', {minimumSignificantDigits: 3}, 1.2),
    numberFormatFixture('resolved-compact-defaults', 'en', {notation: 'compact'}, 1200),
  ],
  dateTimeFormatSmoke: [
    dateTimeFormatSmokeFixture('utc-long-date', 'en-US', {year: 'numeric', month: 'long', day: 'numeric', timeZone: 'UTC'}, '2020-01-02T03:04:05Z'),
  ],
  dateTimeFormatErrors: [
    dateTimeFormatErrorFixture('invalid-date-style', 'en-US', {dateStyle: 'bad'}, '2026-05-08T12:00:00Z'),
    dateTimeFormatErrorFixture('invalid-calendar', 'en-US', {calendar: 'bad!'}, '2026-05-08T12:00:00Z'),
  ],
  dateTimeFormatEdge: [
    dateTimeFormatPartsFixture('offset-timezone-kolkata', 'edge', 'offsetTimeZone', 'en-US', {hour: '2-digit', minute: '2-digit', hour12: false, timeZone: '+05:30', timeZoneName: 'longOffset'}, '2021-01-10T12:00:00Z'),
    dateTimeFormatPartsFixture('offset-timezone-boundary', 'edge', 'offsetTimeZoneBoundary', 'en-US', {hour: '2-digit', minute: '2-digit', hour12: false, timeZone: '+23:59', timeZoneName: 'longOffset'}, '2021-01-10T00:00:00Z'),
    dateTimeFormatPartsFixture('day-period-time-zone-name-spacing', 'edge', 'dayPeriodTimeZoneNameSpacing', 'en-US', {hour: '2-digit', minute: '2-digit', timeZone: 'Europe/Amsterdam', timeZoneName: 'short'}, '2020-09-16T09:55:32Z'),
    dateTimeFormatPartsFixture('hour12-overrides-hour-cycle', 'edge', 'hourCyclePrecedence', 'en-US', {hour: 'numeric', hourCycle: 'h11', hour12: true, timeZone: 'UTC'}, '2021-01-10T00:00:00Z'),
    dateTimeFormatRangeFixture('date-time-style-range-parts', 'edge', 'en-US', {dateStyle: 'medium', timeStyle: 'short', timeZone: 'UTC'}, '2021-01-10T00:00:00Z', '2021-01-10T03:30:00Z'),
  ],
  durationFormatSmoke: [
    durationFormatSmokeFixture('hours-minutes', 'en', {}, {hours: 1, minutes: 2}),
  ],
  durationFormatErrors: [
    durationFormatErrorFixture('invalid-style', 'en-US', {style: 'bad'}, {}),
  ],
  durationFormatDigital: [
    durationFormatFixture('digital-hours-minutes-seconds', 'en', {style: 'digital'}, {hours: 5, minutes: 30, seconds: 15}),
    durationFormatFixture('digital-fractional-seconds', 'en', {style: 'digital', fractionalDigits: 3}, {hours: 5, minutes: 30, seconds: 15, milliseconds: 123}),
    durationFormatFixture('digital-zero-hours', 'en', {style: 'digital'}, {minutes: 30, seconds: 15}),
  ],
  listFormatSmoke: [
    listFormatSmokeFixture('conjunction-long', 'en-US', {}, ['A', 'B', 'C']),
  ],
  listFormatErrors: [
    listFormatErrorFixture('invalid-style', 'en-US', {style: 'bad'}, []),
  ],
  relativeTimeSmoke: [
    relativeTimeFormatSmokeFixture('day-auto', 'en-US', {numeric: 'auto'}, {value: -1, unit: 'day'}),
  ],
  relativeTimeErrors: [
    relativeTimeFormatErrorFixture('invalid-numeric', 'en-US', {numeric: 'bad'}, {value: 1, unit: 'day'}),
  ],
  pluralRulesSmoke: [
    pluralRulesSmokeFixture('ordinal-two', 'en-US', {type: 'ordinal'}, 2),
  ],
  pluralRulesErrors: [
    pluralRulesErrorFixture('invalid-type', 'en-US', {type: 'bad'}, 1),
  ],
  displayNamesSmoke: [
    displayNamesSmokeFixture('region-us', 'en', {type: 'region'}, 'US'),
    displayNamesSmokeFixture('language-zh-hans', 'en', {type: 'language'}, 'zh-Hans'),
    displayNamesSmokeFixture('currency-eur-narrow', 'en', {type: 'currency', style: 'narrow'}, 'EUR'),
    displayNamesSmokeFixture('fallback-none-missing-region', 'en', {type: 'region', fallback: 'none'}, 'ZZ'),
  ],
  displayNamesErrors: [
    displayNamesErrorFixture('invalid-type', 'en-US', {type: 'bad'}, 'en'),
  ],
  supportedValues: {
    source: source('intl', 'supportedValuesOf'),
    versions: selectedVersions,
    values: supportedValues,
  },
};

console.log(JSON.stringify(witness));
`
