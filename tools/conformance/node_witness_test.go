package conformance

import (
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
		{"id":"numberformat-node-v26-range-collapse","source":"node:v26.0.0:numberformat:edge","locale":"en-US","options":{"maximumFractionDigits":0},"input":{"start":1.2,"end":1.4},"expectedRange":"1","expectedRangeParts":[{"type":"integer","value":"1","source":"shared"}],"expectedResolvedOptions":{"locale":"en-US"}}
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
