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
		{"id":"numberformat-node-v26-invalid-style","source":"node:v26.0.0:numberformat:errors","locale":"en-US","options":{"style":"invalid"},"input":1,"errorCode":"invalid_option"}
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
