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
		requiredNodeTopic("locale", "rg week-info override", localeInfoExpected("week-info-rg")),
		requiredNodeTopic("locale", "sd week-info region", localeInfoExpected("week-info-sd")),
		requiredNodeTopic("locale", "fw week-info override", localeInfoExpected("week-info-fw")),
		requiredNodeTopic("locale", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:locale:errors")),
		requiredNodeTopic("numberformat", "smoke format", nodeSourceHasExpected("node:v26.0.0:numberformat")),
		requiredNodeTopic("numberformat", "resolved-options branches", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:numberformat:resolved-options" && len(f.ExpectedResolved) > 0
		}),
		requiredNodeTopic("numberformat", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:numberformat:errors")),
		requiredNodeTopic("numberformat", "unit casing rejection", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:numberformat:errors" && strings.Contains(f.ID, "unit-casing") && f.ErrorCode != ""
		}),
		requiredNodeTopic("numberformat", "negative zero edge", numberEdgeExpected("negative-zero")),
		requiredNodeTopic("numberformat", "rounding increment edge", numberEdgeExpected("rounding-increment")),
		requiredNodeTopic("numberformat", "rounding priority edge", numberEdgeExpected("rounding-priority")),
		requiredNodeTopic("numberformat", "compact plural edge", numberEdgeExpected("compact-plural")),
		requiredNodeTopic("numberformat", "range collapse edge", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:numberformat:edge" && strings.Contains(f.ID, "range-collapse") && f.ExpectedRange != nil && len(f.ExpectedRangeParts) > 0
		}),
		requiredNodeTopic("datetimeformat", "smoke format", nodeSourceHasExpected("node:v26.0.0:datetimeformat")),
		requiredNodeTopic("datetimeformat", "constructor error/refusal", nodeSourceHasError("node:v26.0.0:datetimeformat:errors")),
		requiredNodeTopic("datetimeformat", "calendar option rejection", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:datetimeformat:errors" && strings.Contains(f.ID, "calendar") && f.ErrorCode != ""
		}),
		requiredNodeTopic("datetimeformat", "offset timezone edge", dateTimeEdgeExpected("offset-timezone")),
		requiredNodeTopic("datetimeformat", "offset timezone boundary", dateTimeEdgeExpected("offset-timezone-boundary")),
		requiredNodeTopic("datetimeformat", "day-period time-zone-name spacing", dateTimeEdgeExpected("day-period-time-zone-name-spacing")),
		requiredNodeTopic("datetimeformat", "hour12/hourCycle precedence", dateTimeEdgeExpected("hour12-overrides-hour-cycle")),
		requiredNodeTopic("datetimeformat", "dateStyle/timeStyle range parts", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:datetimeformat:edge" && strings.Contains(f.ID, "date-time-style-range-parts") && f.ExpectedRange != nil && len(f.ExpectedRangeParts) > 0 && len(f.ExpectedResolved) > 0
		}),
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
		requiredNodeTopic("collator", "numeric option overrides locale extension", func(f Fixture) bool {
			return f.Source == "node:v26.0.0:collator:option-contract" && strings.Contains(f.ID, "numeric-option-overrides-locale-extension") && f.ExpectedComparison != nil && len(f.ExpectedResolved) > 0
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

func numberEdgeExpected(idPart string) func(Fixture) bool {
	return func(f Fixture) bool {
		return f.Source == "node:v26.0.0:numberformat:edge" &&
			strings.Contains(f.ID, idPart) &&
			f.Expected != nil &&
			len(f.ExpectedParts) > 0 &&
			len(f.ExpectedResolved) > 0
	}
}

func localeInfoExpected(idPart string) func(Fixture) bool {
	return func(f Fixture) bool {
		return f.Source == "node:v26.0.0:locale:info" &&
			strings.Contains(f.ID, idPart) &&
			f.Feature == "weekInfo" &&
			len(f.ExpectedResolved) > 0
	}
}

func dateTimeEdgeExpected(idPart string) func(Fixture) bool {
	return func(f Fixture) bool {
		return f.Source == "node:v26.0.0:datetimeformat:edge" &&
			strings.Contains(f.ID, idPart) &&
			f.Expected != nil &&
			len(f.ExpectedParts) > 0 &&
			len(f.ExpectedResolved) > 0
	}
}
