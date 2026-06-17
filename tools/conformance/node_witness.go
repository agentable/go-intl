package conformance

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var errMissingNodeWitnessCoverage = errors.New("missing node witness coverage")

type nodeWitnessCoverageStatus string

const (
	nodeWitnessRequired       nodeWitnessCoverageStatus = "required"
	nodeWitnessIntentionalGap nodeWitnessCoverageStatus = "intentional-gap"
)

type nodeWitnessCoverageTopic struct {
	Package string
	Topic   string
	Status  nodeWitnessCoverageStatus
	Reason  string
	Match   func(Fixture) bool
}

func ValidateNodeWitnessCoverage(packageRoots []string) error {
	rootsByPackage := make(map[string]string, len(packageRoots))
	for _, root := range packageRoots {
		rootsByPackage[filepath.Base(filepath.Clean(root))] = root
	}

	for _, topic := range nodeWitnessCoverageMatrix() {
		if topic.Status == nodeWitnessIntentionalGap {
			if topic.Reason == "" {
				return fmt.Errorf("%s %s: intentional node witness gap missing reason: %w", topic.Package, topic.Topic, errMissingNodeWitnessCoverage)
			}
			continue
		}
		root, ok := rootsByPackage[topic.Package]
		if !ok {
			continue
		}
		fixtures, err := LoadFixtures(root)
		if err != nil {
			return err
		}
		if nodeWitnessTopicCovered(fixtures, topic) {
			continue
		}
		return fmt.Errorf("%s: missing required node witness topic %q: %w", topic.Package, topic.Topic, errMissingNodeWitnessCoverage)
	}
	return nil
}

func nodeWitnessTopicCovered(fixtures []Fixture, topic nodeWitnessCoverageTopic) bool {
	for _, fixture := range fixtures {
		if fixtureSourceKind(fixture.Source) != "node" {
			continue
		}
		if topic.Match(fixture) {
			return true
		}
	}
	return false
}

func nodeWitnessCoverageMatrix() []nodeWitnessCoverageTopic {
	return []nodeWitnessCoverageTopic{
		requiredNodeTopic("locale", "smoke canonicalization/maximize", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:locale" && f.Expected != nil
		}),
		requiredNodeTopic("locale", "Unicode extension canonicalization", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:locale:canonicalization" && f.Feature == "canonicalize" && f.Expected != nil
		}),
		requiredNodeTopic("locale", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:locale:errors")),
		requiredNodeTopic("numberformat", "smoke format", nodeSourceHasExpected("node:v26.0.0:numberformat")),
		requiredNodeTopic("numberformat", "resolved-options branches", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:numberformat:resolved-options" && len(f.ExpectedResolved) > 0
		}),
		requiredNodeTopic("numberformat", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:numberformat:errors")),
		requiredNodeTopic("datetimeformat", "smoke format", nodeSourceHasExpected("node:v26.0.0:datetimeformat")),
		requiredNodeTopic("datetimeformat", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:datetimeformat:errors")),
		requiredNodeTopic("datetimeformat", "range parts", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:datetimeformat:p4-deep-contract" && len(f.ExpectedRangeParts) > 0 && len(f.ExpectedResolved) > 0
		}),
		requiredNodeTopic("datetimeformat", "time-zone name parts", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:datetimeformat:p4-deep-contract" && f.Feature == "timeZoneName" && len(f.ExpectedParts) > 0 && len(f.ExpectedResolved) > 0
		}),
		requiredNodeTopic("pluralrules", "smoke select", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:pluralrules" && f.Feature == "select" && f.Expected != nil
		}),
		requiredNodeTopic("pluralrules", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:pluralrules:errors")),
		requiredNodeTopic("listformat", "smoke format", nodeSourceHasExpected("node:v26.0.0:listformat")),
		requiredNodeTopic("listformat", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:listformat:errors")),
		requiredNodeTopic("relativetimeformat", "numeric auto literal", nodeSourceHasExpected("node:v26.0.0:relativetimeformat")),
		requiredNodeTopic("relativetimeformat", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:relativetimeformat:errors")),
		requiredNodeTopic("durationformat", "smoke format", nodeSourceHasExpected("node:v26.0.0:durationformat")),
		requiredNodeTopic("durationformat", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:durationformat:errors")),
		requiredNodeTopic("durationformat", "digital parts and resolved options", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:durationformat:digital" && len(f.ExpectedParts) > 0 && len(f.ExpectedResolved) > 0
		}),
		requiredNodeTopic("displaynames", "smoke display name lookup", func(f Fixture) bool {
			return strings.EqualFold(f.Source, "node:v26.0.0:displaynames") && f.Expected != nil && f.ExpectedOK != nil
		}),
		requiredNodeTopic("displaynames", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:displaynames:errors")),
		requiredNodeTopic("collator", "smoke compare", nodeSourceHasComparison("node:v26.0.0:collator")),
		requiredNodeTopic("collator", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:collator:errors")),
		requiredNodeTopic("collator", "option resolved contracts", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:collator:option-contract" && f.ExpectedComparison != nil && len(f.ExpectedResolved) > 0
		}),
		requiredNodeTopic("collator", "backend ordering proof", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:collator:backend-proof" && f.ExpectedComparison != nil && len(f.ExpectedResolved) > 0
		}),
		requiredNodeTopic("segmenter", "smoke segmentation", nodeSourceHasSegments("node:v26.0.0:segmenter")),
		requiredNodeTopic("segmenter", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:segmenter:errors")),
		requiredNodeTopic("segmenter", "advertised locale word/sentence contract", nodeSourceHasSegments("node:v26.0.0:segmenter:locale-contract")),
		requiredNodeTopic("segmenter", "tailored locale withheld contract", nodeSourceHasSegments("node:v26.0.0:segmenter:tailored-locale-contract")),
	}
}

func requiredNodeTopic(packageName, topic string, match func(Fixture) bool) nodeWitnessCoverageTopic {
	return nodeWitnessCoverageTopic{Package: packageName, Topic: topic, Status: nodeWitnessRequired, Match: match}
}

func intentionalNodeGap(packageName, topic, reason string) nodeWitnessCoverageTopic {
	return nodeWitnessCoverageTopic{Package: packageName, Topic: topic, Status: nodeWitnessIntentionalGap, Reason: reason}
}

func nodeSourceHasExpected(source string) func(Fixture) bool {
	return func(f Fixture) bool {
		return f.Source == source && f.Expected != nil
	}
}

func nodeSourceHasComparison(source string) func(Fixture) bool {
	return func(f Fixture) bool {
		return f.Source == source && f.ExpectedComparison != nil
	}
}

func nodeSourceHasSegments(source string) func(Fixture) bool {
	return func(f Fixture) bool {
		return f.Source == source && len(f.ExpectedSegments) > 0
	}
}

func nodeSourceHasError(source string) func(Fixture) bool {
	return func(f Fixture) bool {
		return f.Source == source && f.ErrorCode != ""
	}
}
