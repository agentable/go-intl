package conformance

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

var errMissingNodeWitnessCoverage = errors.New("missing node witness coverage")

const (
	activeNodeWitnessVersion = "v26.0.0"
	nodeWitnessSourcePrefix  = "node:" + activeNodeWitnessVersion
)

type nodeWitnessCoverageStatus string

const (
	nodeWitnessRequired       nodeWitnessCoverageStatus = "required"
	nodeWitnessIntentionalGap nodeWitnessCoverageStatus = "intentional-gap"
)

type nodeWitnessSourceSuffix string

const (
	nodeSourceLocale                      nodeWitnessSourceSuffix = "locale"
	nodeSourceLocaleCanonicalization      nodeWitnessSourceSuffix = "locale:canonicalization"
	nodeSourceLocaleInfo                  nodeWitnessSourceSuffix = "locale:info"
	nodeSourceLocaleErrors                nodeWitnessSourceSuffix = "locale:errors"
	nodeSourceNumberFormat                nodeWitnessSourceSuffix = "numberformat"
	nodeSourceNumberFormatResolvedOptions nodeWitnessSourceSuffix = "numberformat:resolved-options"
	nodeSourceNumberFormatErrors          nodeWitnessSourceSuffix = "numberformat:errors"
	nodeSourceNumberFormatEdge            nodeWitnessSourceSuffix = "numberformat:edge"
	nodeSourceDateTimeFormat              nodeWitnessSourceSuffix = "datetimeformat"
	nodeSourceDateTimeFormatErrors        nodeWitnessSourceSuffix = "datetimeformat:errors"
	nodeSourceDateTimeFormatEdge          nodeWitnessSourceSuffix = "datetimeformat:edge"
	nodeSourceDateTimeFormatDeepContract  nodeWitnessSourceSuffix = "datetimeformat:p4-deep-contract"
	nodeSourcePluralRules                 nodeWitnessSourceSuffix = "pluralrules"
	nodeSourcePluralRulesErrors           nodeWitnessSourceSuffix = "pluralrules:errors"
	nodeSourceListFormat                  nodeWitnessSourceSuffix = "listformat"
	nodeSourceListFormatErrors            nodeWitnessSourceSuffix = "listformat:errors"
	nodeSourceRelativeTimeFormat          nodeWitnessSourceSuffix = "relativetimeformat"
	nodeSourceRelativeTimeFormatErrors    nodeWitnessSourceSuffix = "relativetimeformat:errors"
	nodeSourceDurationFormat              nodeWitnessSourceSuffix = "durationformat"
	nodeSourceDurationFormatErrors        nodeWitnessSourceSuffix = "durationformat:errors"
	nodeSourceDurationFormatDigital       nodeWitnessSourceSuffix = "durationformat:digital"
	nodeSourceDisplayNames                nodeWitnessSourceSuffix = "displaynames"
	nodeSourceDisplayNamesErrors          nodeWitnessSourceSuffix = "displaynames:errors"
)

type nodeFixtureFeature string

const (
	nodeFeatureCanonicalize nodeFixtureFeature = "canonicalize"
	nodeFeatureSelect       nodeFixtureFeature = "select"
	nodeFeatureTimeZoneName nodeFixtureFeature = "timeZoneName"
	nodeFeatureWeekInfo     nodeFixtureFeature = "weekInfo"
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
	fixturesByPackage := make(map[string][]Fixture, len(rootsByPackage))

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
		fixtures, ok := fixturesByPackage[topic.Package]
		if !ok {
			var err error
			fixtures, err = LoadFixtures(root)
			if err != nil {
				return err
			}
			fixturesByPackage[topic.Package] = fixtures
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
		if fixtureSourceKindOf(fixture.Source) != fixtureSourceNode {
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
		requiredNodeTopic("locale", "smoke canonicalization/maximize", nodeWitnessHasExpected(nodeSourceLocale)),
		requiredNodeTopic("locale", "Unicode extension canonicalization", nodeWitnessHasExpectedForFeature(nodeSourceLocaleCanonicalization, nodeFeatureCanonicalize)),
		requiredNodeTopic("locale", "rg week-info override", localeInfoExpected("week-info-rg")),
		requiredNodeTopic("locale", "sd week-info region", localeInfoExpected("week-info-sd")),
		requiredNodeTopic("locale", "fw week-info override", localeInfoExpected("week-info-fw")),
		requiredNodeTopic("locale", "constructor error/refusal", nodeWitnessHasError(nodeSourceLocaleErrors)),
		requiredNodeTopic("numberformat", "smoke format", nodeWitnessHasExpected(nodeSourceNumberFormat)),
		requiredNodeTopic("numberformat", "resolved-options branches", nodeWitnessHasResolved(nodeSourceNumberFormatResolvedOptions)),
		requiredNodeTopic("numberformat", "constructor error/refusal", nodeWitnessHasError(nodeSourceNumberFormatErrors)),
		requiredNodeTopic("numberformat", "unit casing rejection", nodeWitnessHasErrorForID(nodeSourceNumberFormatErrors, "unit-casing")),
		requiredNodeTopic("numberformat", "negative zero edge", numberEdgeExpected("negative-zero")),
		requiredNodeTopic("numberformat", "rounding increment edge", numberEdgeExpected("rounding-increment")),
		requiredNodeTopic("numberformat", "rounding priority edge", numberEdgeExpected("rounding-priority")),
		requiredNodeTopic("numberformat", "compact plural edge", numberEdgeExpected("compact-plural")),
		requiredNodeTopic("numberformat", "range collapse edge", nodeWitnessHasExpectedRangePartsForID(nodeSourceNumberFormatEdge, "range-collapse")),
		requiredNodeTopic("numberformat", "Czech plural-range unit", nodeWitnessHasExpectedRangePartsResolvedForID(nodeSourceNumberFormatEdge, "czech-plural-range-unit")),
		requiredNodeTopic("numberformat", "Czech plural-range currency name", nodeWitnessHasExpectedRangePartsResolvedForID(nodeSourceNumberFormatEdge, "czech-plural-range-currency-name")),
		requiredNodeTopic("numberformat", "negative percent range affixes", nodeWitnessHasExpectedRangePartsResolvedForID(nodeSourceNumberFormatEdge, "negative-percent-range-affixes")),
		requiredNodeTopic("datetimeformat", "smoke format", nodeWitnessHasExpected(nodeSourceDateTimeFormat)),
		requiredNodeTopic("datetimeformat", "constructor error/refusal", nodeWitnessHasError(nodeSourceDateTimeFormatErrors)),
		requiredNodeTopic("datetimeformat", "calendar option rejection", nodeWitnessHasErrorForID(nodeSourceDateTimeFormatErrors, "calendar")),
		requiredNodeTopic("datetimeformat", "offset timezone edge", dateTimeEdgeExpected("offset-timezone")),
		requiredNodeTopic("datetimeformat", "offset timezone boundary", dateTimeEdgeExpected("offset-timezone-boundary")),
		requiredNodeTopic("datetimeformat", "day-period time-zone-name spacing", dateTimeEdgeExpected("day-period-time-zone-name-spacing")),
		requiredNodeTopic("datetimeformat", "hour12/hourCycle precedence", dateTimeEdgeExpected("hour12-overrides-hour-cycle")),
		requiredNodeTopic("datetimeformat", "dateStyle/timeStyle range parts", nodeWitnessHasExpectedRangePartsResolvedForID(nodeSourceDateTimeFormatEdge, "date-time-style-range-parts")),
		requiredNodeTopic("datetimeformat", "range parts", nodeWitnessHasRangePartsResolved(nodeSourceDateTimeFormatDeepContract)),
		requiredNodeTopic("datetimeformat", "time-zone name parts", nodeWitnessHasPartsResolvedForFeature(nodeSourceDateTimeFormatDeepContract, nodeFeatureTimeZoneName)),
		requiredNodeTopic("pluralrules", "smoke select", nodeWitnessHasExpectedForFeature(nodeSourcePluralRules, nodeFeatureSelect)),
		requiredNodeTopic("pluralrules", "constructor error/refusal", nodeWitnessHasError(nodeSourcePluralRulesErrors)),
		requiredNodeTopic("listformat", "smoke format", nodeWitnessHasExpected(nodeSourceListFormat)),
		requiredNodeTopic("listformat", "constructor error/refusal", nodeWitnessHasError(nodeSourceListFormatErrors)),
		requiredNodeTopic("relativetimeformat", "numeric auto literal", nodeWitnessHasExpected(nodeSourceRelativeTimeFormat)),
		requiredNodeTopic("relativetimeformat", "constructor error/refusal", nodeWitnessHasError(nodeSourceRelativeTimeFormatErrors)),
		requiredNodeTopic("durationformat", "smoke format", nodeWitnessHasExpected(nodeSourceDurationFormat)),
		requiredNodeTopic("durationformat", "constructor error/refusal", nodeWitnessHasError(nodeSourceDurationFormatErrors)),
		requiredNodeTopic("durationformat", "digital parts and resolved options", nodeWitnessHasPartsResolved(nodeSourceDurationFormatDigital)),
		requiredNodeTopic("displaynames", "smoke display name lookup", nodeWitnessHasDisplayNameLookup(nodeSourceDisplayNames)),
		requiredNodeTopic("displaynames", "constructor error/refusal", nodeWitnessHasError(nodeSourceDisplayNamesErrors)),
	}
}

func requiredNodeTopic(packageName, topic string, match func(Fixture) bool) nodeWitnessCoverageTopic {
	return nodeWitnessCoverageTopic{Package: packageName, Topic: topic, Status: nodeWitnessRequired, Match: match}
}

func nodeWitnessSource(suffix nodeWitnessSourceSuffix) string {
	return nodeWitnessSourcePrefix + ":" + string(suffix)
}

func nodeWitnessFixtureID(surface, topic string) string {
	return surface + "-" + nodeFixtureDirPrefix + nodeWitnessMajorVersion() + "-" + topic
}

func nodeWitnessMajorVersion() string {
	version := strings.TrimPrefix(activeNodeWitnessVersion, "v")
	major, _, _ := strings.Cut(version, ".")
	return major
}

func fixtureHasNodeFeature(f Fixture, feature nodeFixtureFeature) bool {
	return f.Feature == string(feature)
}

type nodeWitnessFixturePredicate func(Fixture) bool

func nodeWitnessFixtureMatcher(suffix nodeWitnessSourceSuffix, match nodeWitnessFixturePredicate) func(Fixture) bool {
	source := nodeWitnessSource(suffix)
	return func(f Fixture) bool {
		return f.Source == source && match(f)
	}
}

func nodeWitnessFixtureMatcherForFeature(suffix nodeWitnessSourceSuffix, feature nodeFixtureFeature, match nodeWitnessFixturePredicate) func(Fixture) bool {
	source := nodeWitnessSource(suffix)
	return func(f Fixture) bool {
		return f.Source == source && fixtureHasNodeFeature(f, feature) && match(f)
	}
}

func nodeWitnessFixtureMatcherForID(suffix nodeWitnessSourceSuffix, idPart string, match nodeWitnessFixturePredicate) func(Fixture) bool {
	source := nodeWitnessSource(suffix)
	return func(f Fixture) bool {
		return f.Source == source && strings.Contains(f.ID, idPart) && match(f)
	}
}

func nodeWitnessHasExpected(suffix nodeWitnessSourceSuffix) func(Fixture) bool {
	return nodeWitnessFixtureMatcher(suffix, nodeWitnessFixtureHasExpected)
}

func nodeWitnessHasExpectedForFeature(suffix nodeWitnessSourceSuffix, feature nodeFixtureFeature) func(Fixture) bool {
	return nodeWitnessFixtureMatcherForFeature(suffix, feature, nodeWitnessFixtureHasExpected)
}

func nodeWitnessHasError(suffix nodeWitnessSourceSuffix) func(Fixture) bool {
	return nodeWitnessFixtureMatcher(suffix, nodeWitnessFixtureHasError)
}

func nodeWitnessHasErrorForID(suffix nodeWitnessSourceSuffix, idPart string) func(Fixture) bool {
	return nodeWitnessFixtureMatcherForID(suffix, idPart, nodeWitnessFixtureHasError)
}

func nodeWitnessHasDisplayNameLookup(suffix nodeWitnessSourceSuffix) func(Fixture) bool {
	return nodeWitnessFixtureMatcher(suffix, nodeWitnessFixtureHasDisplayNameLookup)
}

func nodeWitnessHasResolved(suffix nodeWitnessSourceSuffix) func(Fixture) bool {
	return nodeWitnessFixtureMatcher(suffix, nodeWitnessFixtureHasResolved)
}

func nodeWitnessHasExpectedRangePartsForID(suffix nodeWitnessSourceSuffix, idPart string) func(Fixture) bool {
	return nodeWitnessFixtureMatcherForID(suffix, idPart, nodeWitnessFixtureHasExpectedRangeParts)
}

func nodeWitnessHasExpectedRangePartsResolvedForID(suffix nodeWitnessSourceSuffix, idPart string) func(Fixture) bool {
	return nodeWitnessFixtureMatcherForID(suffix, idPart, nodeWitnessFixtureHasExpectedRangePartsResolved)
}

func nodeWitnessHasPartsResolved(suffix nodeWitnessSourceSuffix) func(Fixture) bool {
	return nodeWitnessFixtureMatcher(suffix, nodeWitnessFixtureHasPartsResolved)
}

func nodeWitnessHasPartsResolvedForFeature(suffix nodeWitnessSourceSuffix, feature nodeFixtureFeature) func(Fixture) bool {
	return nodeWitnessFixtureMatcherForFeature(suffix, feature, nodeWitnessFixtureHasPartsResolved)
}

func nodeWitnessHasRangePartsResolved(suffix nodeWitnessSourceSuffix) func(Fixture) bool {
	return nodeWitnessFixtureMatcher(suffix, nodeWitnessFixtureHasRangePartsResolved)
}

func numberEdgeExpected(idPart string) func(Fixture) bool {
	return nodeWitnessHasExpectedPartsResolved(nodeSourceNumberFormatEdge, idPart)
}

func localeInfoExpected(idPart string) func(Fixture) bool {
	return nodeWitnessFixtureMatcherForID(nodeSourceLocaleInfo, idPart, func(f Fixture) bool {
		return fixtureHasNodeFeature(f, nodeFeatureWeekInfo) && nodeWitnessFixtureHasResolved(f)
	})
}

func dateTimeEdgeExpected(idPart string) func(Fixture) bool {
	return nodeWitnessHasExpectedPartsResolved(nodeSourceDateTimeFormatEdge, idPart)
}

func nodeWitnessHasExpectedPartsResolved(suffix nodeWitnessSourceSuffix, idPart string) func(Fixture) bool {
	return nodeWitnessFixtureMatcherForID(suffix, idPart, nodeWitnessFixtureHasExpectedPartsResolved)
}

func nodeWitnessFixtureHasExpected(f Fixture) bool {
	return f.Expected != nil
}

func nodeWitnessFixtureHasError(f Fixture) bool {
	return f.ErrorCode != ""
}

func nodeWitnessFixtureHasDisplayNameLookup(f Fixture) bool {
	return f.Expected != nil && f.ExpectedOK != nil
}

func nodeWitnessFixtureHasPartsResolved(f Fixture) bool {
	return len(f.ExpectedParts) > 0 && nodeWitnessFixtureHasResolved(f)
}

func nodeWitnessFixtureHasExpectedPartsResolved(f Fixture) bool {
	return nodeWitnessFixtureHasExpected(f) &&
		len(f.ExpectedParts) > 0 &&
		nodeWitnessFixtureHasResolved(f)
}

func nodeWitnessFixtureHasRangePartsResolved(f Fixture) bool {
	return len(f.ExpectedRangeParts) > 0 && nodeWitnessFixtureHasResolved(f)
}

func nodeWitnessFixtureHasExpectedRangeParts(f Fixture) bool {
	return f.ExpectedRange != nil && len(f.ExpectedRangeParts) > 0
}

func nodeWitnessFixtureHasExpectedRangePartsResolved(f Fixture) bool {
	return nodeWitnessFixtureHasExpectedRangeParts(f) && nodeWitnessFixtureHasResolved(f)
}

func nodeWitnessFixtureHasResolved(f Fixture) bool {
	return len(f.ExpectedResolved) > 0
}
