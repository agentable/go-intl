package conformance

import (
	"encoding/json/jsontext"
	"errors"
	"path/filepath"
	"testing"
)

func TestValidateNodeWitnessCoverageRequiresMatrixTopics(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := filepath.Join(root, "numberformat")
	writeCoverageFixtureFile(t, packageDir, "node-v26/smoke.json", `[
		{"id":"numberformat-node-v26-smoke","source":"node:v26.0.0:numberformat","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)

	err := ValidateNodeWitnessCoverage([]string{packageDir})
	if !errors.Is(err, errMissingNodeWitnessCoverage) {
		t.Fatalf("ValidateNodeWitnessCoverage() error = %v, want missing node witness coverage", err)
	}

	writeCoverageFixtureFile(t, packageDir, "node-v26/resolved-options.json", `[
		{"id":"numberformat-node-v26-resolved","source":"node:v26.0.0:numberformat:resolved-options","locale":"en-US","options":{},"input":1,"expected":"1","expectedResolvedOptions":{"locale":"en-US"}}
	]`)
	writeCoverageFixtureFile(t, packageDir, "node-v26/errors.json", `[
		{"id":"numberformat-node-v26-invalid-style","source":"node:v26.0.0:numberformat:errors","locale":"en-US","options":{"style":"invalid"},"input":1,"errorCode":"invalid_option"},
		{"id":"numberformat-node-v26-unit-casing-rejected","source":"node:v26.0.0:numberformat:errors","locale":"en-US","options":{"style":"unit","unit":"METER"},"input":1,"errorCode":"invalid_option"}
	]`)
	if err := ValidateNodeWitnessCoverage([]string{packageDir}); !errors.Is(err, errMissingNodeWitnessCoverage) {
		t.Fatalf("ValidateNodeWitnessCoverage() error = %v, want missing edge coverage", err)
	}

	writeCoverageFixtureFile(t, packageDir, "node-v26/edge.json", `[
		{"id":"numberformat-node-v26-negative-zero-sign","source":"node:v26.0.0:numberformat:edge","locale":"en-US","options":{"signDisplay":"auto"},"input":"-0","expected":"-0","expectedParts":[{"type":"minusSign","value":"-"},{"type":"integer","value":"0"}],"expectedResolvedOptions":{"locale":"en-US"}},
		{"id":"numberformat-node-v26-rounding-increment","source":"node:v26.0.0:numberformat:edge","locale":"en-US","options":{"minimumFractionDigits":2,"maximumFractionDigits":2,"roundingIncrement":5},"input":1.234,"expected":"1.25","expectedParts":[{"type":"integer","value":"1"}],"expectedResolvedOptions":{"locale":"en-US"}},
		{"id":"numberformat-node-v26-rounding-priority-more-precision","source":"node:v26.0.0:numberformat:edge","locale":"en-US","options":{"minimumSignificantDigits":2,"maximumFractionDigits":0,"roundingPriority":"morePrecision"},"input":1.234,"expected":"1.2","expectedParts":[{"type":"integer","value":"1"}],"expectedResolvedOptions":{"locale":"en-US"}},
		{"id":"numberformat-node-v26-compact-plural-few","source":"node:v26.0.0:numberformat:edge","locale":"ru","options":{"notation":"compact","compactDisplay":"long"},"input":2000,"expected":"2 тысячи","expectedParts":[{"type":"integer","value":"2"},{"type":"compact","value":"тысячи"}],"expectedResolvedOptions":{"locale":"ru"}},
		{"id":"numberformat-node-v26-range-collapse","source":"node:v26.0.0:numberformat:edge","locale":"en-US","options":{"maximumFractionDigits":0},"input":{"start":1.2,"end":1.4},"expectedRange":"1","expectedRangeParts":[{"type":"integer","value":"1","source":"shared"}],"expectedResolvedOptions":{"locale":"en-US"}},
		{"id":"numberformat-node-v26-czech-plural-range-unit","source":"node:v26.0.0:numberformat:edge","locale":"cs","options":{"style":"unit","unit":"meter","unitDisplay":"long"},"input":{"start":2,"end":1},"expectedRange":"2–1 metrů","expectedRangeParts":[{"type":"integer","value":"2","source":"startRange"},{"type":"literal","value":"–","source":"shared"},{"type":"integer","value":"1","source":"endRange"},{"type":"literal","value":" ","source":"shared"},{"type":"unit","value":"metrů","source":"shared"}],"expectedResolvedOptions":{"locale":"cs"}},
		{"id":"numberformat-node-v26-czech-plural-range-currency-name","source":"node:v26.0.0:numberformat:edge","locale":"cs","options":{"style":"currency","currency":"USD","currencyDisplay":"name","maximumFractionDigits":0},"input":{"start":2,"end":1},"expectedRange":"2–1 amerických dolarů","expectedRangeParts":[{"type":"integer","value":"2","source":"startRange"},{"type":"literal","value":"–","source":"shared"},{"type":"integer","value":"1","source":"endRange"},{"type":"literal","value":" ","source":"shared"},{"type":"currency","value":"amerických dolarů","source":"shared"}],"expectedResolvedOptions":{"locale":"cs"}},
		{"id":"numberformat-node-v26-negative-percent-range-affixes","source":"node:v26.0.0:numberformat:edge","locale":"en","options":{"style":"percent","maximumFractionDigits":0},"input":{"start":-0.01,"end":-0.02},"expectedRange":"-1–2%","expectedRangeParts":[{"type":"minusSign","value":"-","source":"shared"},{"type":"integer","value":"1","source":"startRange"},{"type":"literal","value":"–","source":"shared"},{"type":"integer","value":"2","source":"endRange"},{"type":"percentSign","value":"%","source":"shared"}],"expectedResolvedOptions":{"locale":"en"}}
	]`)
	if err := ValidateNodeWitnessCoverage([]string{packageDir}); err != nil {
		t.Fatalf("ValidateNodeWitnessCoverage() error = %v, want nil", err)
	}
}

func TestValidateNodeWitnessCoverageRequiresNumberFormatErrorWitness(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	packageDir := filepath.Join(root, "numberformat")
	writeCoverageFixtureFile(t, packageDir, "node-v26/smoke.json", `[
		{"id":"numberformat-node-v26-smoke","source":"node:v26.0.0:numberformat","locale":"en-US","options":{},"input":1,"expected":"1"}
	]`)
	writeCoverageFixtureFile(t, packageDir, "node-v26/resolved-options.json", `[
		{"id":"numberformat-node-v26-resolved","source":"node:v26.0.0:numberformat:resolved-options","locale":"en-US","options":{},"input":1,"expected":"1","expectedResolvedOptions":{"locale":"en-US"}}
	]`)

	err := ValidateNodeWitnessCoverage([]string{packageDir})
	if !errors.Is(err, errMissingNodeWitnessCoverage) {
		t.Fatalf("ValidateNodeWitnessCoverage() error = %v, want missing node witness coverage", err)
	}
}

func TestNodeWitnessTopicCoveredIgnoresNonNodeSources(t *testing.T) {
	t.Parallel()

	topic := requiredNodeTopic("numberformat", "unconstrained matcher", func(Fixture) bool { return true })
	if nodeWitnessTopicCovered([]Fixture{{Source: "manual"}}, topic) {
		t.Fatal("nodeWitnessTopicCovered(manual) = true, want false")
	}
	if !nodeWitnessTopicCovered([]Fixture{{Source: nodeWitnessSource(nodeSourceNumberFormat)}}, topic) {
		t.Fatal("nodeWitnessTopicCovered(node) = false, want true")
	}
}

func TestNodeWitnessSourceHelpersUseActiveVersion(t *testing.T) {
	t.Parallel()

	if got := nodeWitnessSource(nodeSourceNumberFormat); got != "node:v26.0.0:numberformat" {
		t.Fatalf("nodeWitnessSource(numberformat) = %q, want active Node v26 source", got)
	}
	if got := nodeWitnessFixtureID("numberformat", "negative-zero-sign"); got != "numberformat-node-v26-negative-zero-sign" {
		t.Fatalf("nodeWitnessFixtureID() = %q, want active Node v26 fixture ID", got)
	}
}

func TestNodeWitnessPredicateHelpersRequireObservableFields(t *testing.T) {
	t.Parallel()

	expected := "ok"
	expectedOK := false
	resolved := jsontext.Value(`{"locale":"en-US"}`)
	parts := []Part{{Type: "integer", Value: "1"}}
	rangeParts := []RangePart{{Type: "integer", Value: "1", Source: "shared"}}

	tests := []struct {
		name     string
		match    func(Fixture) bool
		valid    Fixture
		invalids []Fixture
	}{
		{
			name:  "expected",
			match: nodeWitnessHasExpected(nodeSourceNumberFormat),
			valid: Fixture{Source: nodeWitnessSource(nodeSourceNumberFormat), Expected: &expected},
			invalids: []Fixture{
				{Source: "formatjs:packages/intl-numberformat/tests/basic.test.ts", Expected: &expected},
				{Source: nodeWitnessSource(nodeSourceNumberFormat)},
			},
		},
		{
			name:  "expected feature",
			match: nodeWitnessHasExpectedForFeature(nodeSourceLocaleCanonicalization, nodeFeatureCanonicalize),
			valid: Fixture{Source: nodeWitnessSource(nodeSourceLocaleCanonicalization), Feature: string(nodeFeatureCanonicalize), Expected: &expected},
			invalids: []Fixture{
				{Source: "manual", Feature: string(nodeFeatureCanonicalize), Expected: &expected},
				{Source: nodeWitnessSource(nodeSourceLocaleCanonicalization), Feature: string(nodeFeatureSelect), Expected: &expected},
				{Source: nodeWitnessSource(nodeSourceLocaleCanonicalization), Feature: string(nodeFeatureCanonicalize)},
			},
		},
		{
			name:  "error",
			match: nodeWitnessHasError(nodeSourceNumberFormatErrors),
			valid: Fixture{Source: nodeWitnessSource(nodeSourceNumberFormatErrors), ErrorCode: "invalid_option"},
			invalids: []Fixture{
				{Source: "manual:errors", ErrorCode: "invalid_option"},
				{Source: nodeWitnessSource(nodeSourceNumberFormatErrors)},
			},
		},
		{
			name:  "error id",
			match: nodeWitnessHasErrorForID(nodeSourceNumberFormatErrors, "unit-casing"),
			valid: Fixture{ID: "numberformat-node-v26-unit-casing-rejected", Source: nodeWitnessSource(nodeSourceNumberFormatErrors), ErrorCode: "invalid_option"},
			invalids: []Fixture{
				{ID: "numberformat-node-v26-unit-casing-rejected", Source: "manual:errors", ErrorCode: "invalid_option"},
				{ID: "numberformat-node-v26-invalid-style", Source: nodeWitnessSource(nodeSourceNumberFormatErrors), ErrorCode: "invalid_option"},
				{ID: "numberformat-node-v26-unit-casing-rejected", Source: nodeWitnessSource(nodeSourceNumberFormatErrors)},
			},
		},
		{
			name:  "display name lookup",
			match: nodeWitnessHasDisplayNameLookup(nodeSourceDisplayNames),
			valid: Fixture{Source: nodeWitnessSource(nodeSourceDisplayNames), Expected: &expected, ExpectedOK: &expectedOK},
			invalids: []Fixture{
				{Source: "manual", Expected: &expected, ExpectedOK: &expectedOK},
				{Source: "NODE:V26.0.0:DISPLAYNAMES", Expected: &expected, ExpectedOK: &expectedOK},
				{Source: nodeWitnessSource(nodeSourceDisplayNames), ExpectedOK: &expectedOK},
				{Source: nodeWitnessSource(nodeSourceDisplayNames), Expected: &expected},
			},
		},
		{
			name:  "resolved",
			match: nodeWitnessHasResolved(nodeSourceNumberFormatResolvedOptions),
			valid: Fixture{Source: nodeWitnessSource(nodeSourceNumberFormatResolvedOptions), ExpectedResolved: resolved},
			invalids: []Fixture{
				{Source: "manual", ExpectedResolved: resolved},
				{Source: nodeWitnessSource(nodeSourceNumberFormatResolvedOptions)},
			},
		},
		{
			name:  "number edge",
			match: numberEdgeExpected("negative-zero"),
			valid: Fixture{ID: "numberformat-node-v26-negative-zero-sign", Source: nodeWitnessSource(nodeSourceNumberFormatEdge), Expected: &expected, ExpectedParts: parts, ExpectedResolved: resolved},
			invalids: []Fixture{
				{ID: "numberformat-node-v26-positive-zero-sign", Source: nodeWitnessSource(nodeSourceNumberFormatEdge), Expected: &expected, ExpectedParts: parts, ExpectedResolved: resolved},
				{ID: "numberformat-node-v26-negative-zero-sign", Source: nodeWitnessSource(nodeSourceNumberFormatEdge), Expected: &expected, ExpectedResolved: resolved},
				{ID: "numberformat-node-v26-negative-zero-sign", Source: nodeWitnessSource(nodeSourceNumberFormatEdge), Expected: &expected, ExpectedParts: parts},
			},
		},
		{
			name:  "locale info",
			match: localeInfoExpected("week-info-rg"),
			valid: Fixture{ID: "locale-node-v26-week-info-rg-us", Source: nodeWitnessSource(nodeSourceLocaleInfo), Feature: string(nodeFeatureWeekInfo), ExpectedResolved: resolved},
			invalids: []Fixture{
				{ID: "locale-node-v26-week-info-sd-us", Source: nodeWitnessSource(nodeSourceLocaleInfo), Feature: string(nodeFeatureWeekInfo), ExpectedResolved: resolved},
				{ID: "locale-node-v26-week-info-rg-us", Source: nodeWitnessSource(nodeSourceLocaleInfo), Feature: string(nodeFeatureCanonicalize), ExpectedResolved: resolved},
				{ID: "locale-node-v26-week-info-rg-us", Source: nodeWitnessSource(nodeSourceLocaleInfo), Feature: string(nodeFeatureWeekInfo)},
			},
		},
		{
			name:  "datetime edge",
			match: dateTimeEdgeExpected("offset-timezone"),
			valid: Fixture{ID: "datetimeformat-node-v26-offset-timezone", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), Expected: &expected, ExpectedParts: parts, ExpectedResolved: resolved},
			invalids: []Fixture{
				{ID: "datetimeformat-node-v26-time-zone-name", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), Expected: &expected, ExpectedParts: parts, ExpectedResolved: resolved},
				{ID: "datetimeformat-node-v26-offset-timezone", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), ExpectedParts: parts, ExpectedResolved: resolved},
				{ID: "datetimeformat-node-v26-offset-timezone", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), Expected: &expected, ExpectedResolved: resolved},
				{ID: "datetimeformat-node-v26-offset-timezone", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), Expected: &expected, ExpectedParts: parts},
			},
		},
		{
			name:  "number range edge",
			match: nodeWitnessTopicMatch(t, "numberformat", "range collapse edge"),
			valid: Fixture{ID: "numberformat-node-v26-range-collapse", Source: nodeWitnessSource(nodeSourceNumberFormatEdge), ExpectedRange: &expected, ExpectedRangeParts: rangeParts},
			invalids: []Fixture{
				{ID: "numberformat-node-v26-range-collapse", Source: "manual", ExpectedRange: &expected, ExpectedRangeParts: rangeParts},
				{ID: "numberformat-node-v26-range-collapse", Source: nodeWitnessSource(nodeSourceNumberFormatEdge), ExpectedRangeParts: rangeParts},
				{ID: "numberformat-node-v26-range-collapse", Source: nodeWitnessSource(nodeSourceNumberFormatEdge), ExpectedRange: &expected},
			},
		},
		{
			name:  "datetime range resolved edge",
			match: nodeWitnessTopicMatch(t, "datetimeformat", "dateStyle/timeStyle range parts"),
			valid: Fixture{ID: "datetimeformat-node-v26-date-time-style-range-parts", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), ExpectedRange: &expected, ExpectedRangeParts: rangeParts, ExpectedResolved: resolved},
			invalids: []Fixture{
				{ID: "datetimeformat-node-v26-date-time-style-range-parts", Source: "manual", ExpectedRange: &expected, ExpectedRangeParts: rangeParts, ExpectedResolved: resolved},
				{ID: "datetimeformat-node-v26-hour12-overrides-hour-cycle", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), ExpectedRange: &expected, ExpectedRangeParts: rangeParts, ExpectedResolved: resolved},
				{ID: "datetimeformat-node-v26-date-time-style-range-parts", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), ExpectedRangeParts: rangeParts, ExpectedResolved: resolved},
				{ID: "datetimeformat-node-v26-date-time-style-range-parts", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), ExpectedRange: &expected, ExpectedResolved: resolved},
				{ID: "datetimeformat-node-v26-date-time-style-range-parts", Source: nodeWitnessSource(nodeSourceDateTimeFormatEdge), ExpectedRange: &expected, ExpectedRangeParts: rangeParts},
			},
		},
		{
			name:  "parts resolved",
			match: nodeWitnessHasPartsResolved(nodeSourceDurationFormatDigital),
			valid: Fixture{Source: nodeWitnessSource(nodeSourceDurationFormatDigital), ExpectedParts: parts, ExpectedResolved: resolved},
			invalids: []Fixture{
				{Source: "manual", ExpectedParts: parts, ExpectedResolved: resolved},
				{Source: nodeWitnessSource(nodeSourceDurationFormatDigital), ExpectedResolved: resolved},
				{Source: nodeWitnessSource(nodeSourceDurationFormatDigital), ExpectedParts: parts},
			},
		},
		{
			name:  "parts resolved feature",
			match: nodeWitnessHasPartsResolvedForFeature(nodeSourceDateTimeFormatDeepContract, nodeFeatureTimeZoneName),
			valid: Fixture{Source: nodeWitnessSource(nodeSourceDateTimeFormatDeepContract), Feature: string(nodeFeatureTimeZoneName), ExpectedParts: parts, ExpectedResolved: resolved},
			invalids: []Fixture{
				{Source: nodeWitnessSource(nodeSourceDateTimeFormatDeepContract), Feature: string(nodeFeatureWeekInfo), ExpectedParts: parts, ExpectedResolved: resolved},
				{Source: nodeWitnessSource(nodeSourceDateTimeFormatDeepContract), Feature: string(nodeFeatureTimeZoneName), ExpectedResolved: resolved},
				{Source: nodeWitnessSource(nodeSourceDateTimeFormatDeepContract), Feature: string(nodeFeatureTimeZoneName), ExpectedParts: parts},
			},
		},
		{
			name:  "range parts resolved",
			match: nodeWitnessHasRangePartsResolved(nodeSourceDateTimeFormatDeepContract),
			valid: Fixture{Source: nodeWitnessSource(nodeSourceDateTimeFormatDeepContract), ExpectedRangeParts: rangeParts, ExpectedResolved: resolved},
			invalids: []Fixture{
				{Source: "manual", ExpectedRangeParts: rangeParts, ExpectedResolved: resolved},
				{Source: nodeWitnessSource(nodeSourceDateTimeFormatDeepContract), ExpectedResolved: resolved},
				{Source: nodeWitnessSource(nodeSourceDateTimeFormatDeepContract), ExpectedRangeParts: rangeParts},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if !tc.match(tc.valid) {
				t.Fatalf("%s matcher rejected valid fixture", tc.name)
			}
			for _, invalid := range tc.invalids {
				if tc.match(invalid) {
					t.Fatalf("%s matcher accepted invalid fixture: %+v", tc.name, invalid)
				}
			}
		})
	}
}

func nodeWitnessTopicMatch(t *testing.T, packageName, topicName string) func(Fixture) bool {
	t.Helper()

	for _, topic := range nodeWitnessCoverageMatrix() {
		if topic.Package == packageName && topic.Topic == topicName {
			return topic.Match
		}
	}
	t.Fatalf("node witness topic %s/%s not found", packageName, topicName)
	return nil
}

func TestNodeWitnessCoverageIntentionalGapsHaveReasons(t *testing.T) {
	t.Parallel()

	for _, topic := range nodeWitnessCoverageMatrix() {
		if topic.Status != nodeWitnessIntentionalGap {
			continue
		}
		if topic.Reason == "" {
			t.Fatalf("%s %s intentional gap missing reason", topic.Package, topic.Topic)
		}
	}
}
