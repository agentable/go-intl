package pluralrules

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestPluralConformanceObservations(t *testing.T) {
	if testcontract.FixtureRunnerChild(t, runPluralConformanceFixture) {
		return
	}
	t.Parallel()

	resolved := jsontext.Value(`{"locale":"en","type":"cardinal","notation":"standard","minimumIntegerDigits":1,"minimumFractionDigits":0,"maximumFractionDigits":3,"pluralCategories":["one","other"],"roundingIncrement":1,"roundingMode":"halfExpand","roundingPriority":"auto","trailingZeroDisplay":"auto"}`)
	tests := []struct {
		name    string
		fixture conformance.Fixture
		failure string
	}{
		{"constructor error and text", conformance.Fixture{Options: jsontext.Value(`{"type":"bad"}`), ErrorCode: "invalid_option", Expected: new("IMPOSSIBLE")}, "unsupported pluralrules observation"},
		{"constructor error and resolved", conformance.Fixture{Options: jsontext.Value(`{"type":"bad"}`), ErrorCode: "invalid_option", ExpectedResolved: resolved}, "unsupported pluralrules observation"},
		{"method error and category", conformance.Fixture{Feature: "selectRange", Input: jsontext.Value(`{"start":"NaN","end":1}`), ErrorCode: "invalid_value", Expected: new("IMPOSSIBLE")}, "unsupported pluralrules observation"},
		{"unknown operation and constructor error", conformance.Fixture{Feature: "unknown", Options: jsontext.Value(`{"type":"bad"}`), ErrorCode: "invalid_option"}, "unsupported pluralrules feature"},
		{"unknown operation and resolved", conformance.Fixture{Feature: "unknown", ExpectedResolved: resolved}, "unsupported pluralrules feature"},
		{"resolved only", conformance.Fixture{Input: jsontext.Value(`null`), ExpectedResolved: resolved}, ""},
		{"category and resolved", conformance.Fixture{Expected: new("one"), ExpectedResolved: resolved}, ""},
		{"wrong resolved", conformance.Fixture{Expected: new("one"), ExpectedResolved: jsontext.Value(`{"locale":"wrong"}`)}, "ResolvedOptions"},
		{"missing resolved", conformance.Fixture{Expected: new("one"), ExpectedResolved: jsontext.Value(`{}`)}, "ResolvedOptions"},
		{"extra resolved", conformance.Fixture{Expected: new("one"), ExpectedResolved: jsontext.Value(`{"locale":"en","type":"cardinal","notation":"standard","minimumIntegerDigits":1,"minimumFractionDigits":0,"maximumFractionDigits":3,"pluralCategories":["one","other"],"roundingIncrement":1,"roundingMode":"halfExpand","roundingPriority":"auto","trailingZeroDisplay":"auto","extra":true}`)}, "ResolvedOptions"},
		{"range start NaN", conformance.Fixture{Feature: "selectRange", Input: jsontext.Value(`{"start":"NaN","end":1}`), ErrorCode: "invalid_value"}, ""},
		{"range end NaN", conformance.Fixture{Feature: "selectRange", Input: jsontext.Value(`{"start":1,"end":"NaN"}`), ErrorCode: "invalid_value"}, ""},
		{"range start infinity", conformance.Fixture{Feature: "selectRange", Input: jsontext.Value(`{"start":"Infinity","end":1}`), Expected: new("other")}, ""},
		{"range end infinity", conformance.Fixture{Feature: "selectRange", Input: jsontext.Value(`{"start":1,"end":"-Infinity"}`), Expected: new("other")}, ""},
		{"range error and resolved", conformance.Fixture{Feature: "selectRange", Input: jsontext.Value(`{"start":"NaN","end":1}`), ErrorCode: "invalid_value", ExpectedResolved: resolved}, ""},
		{"range error and wrong resolved", conformance.Fixture{Feature: "selectRange", Input: jsontext.Value(`{"start":"NaN","end":1}`), ErrorCode: "invalid_value", ExpectedResolved: jsontext.Value(`{"locale":"wrong"}`)}, "ResolvedOptions"},
		{"select has no error lane", conformance.Fixture{Feature: "select", Input: jsontext.Value(`"NaN"`), ErrorCode: "invalid_value"}, "unsupported pluralrules runtime feature"},
		{"unsupported parts", conformance.Fixture{Expected: new("one"), ExpectedParts: []conformance.Part{}}, "unsupported pluralrules observation"},
		{"unsupported range parts", conformance.Fixture{Expected: new("one"), ExpectedRangeParts: []conformance.RangePart{}}, "unsupported pluralrules observation"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := tc.fixture
			f.ID, f.Source, f.Locale = "observation", "manual:adapter", "en"
			if f.Options == nil {
				f.Options = jsontext.Value(`{}`)
			}
			if f.Input == nil {
				f.Input = jsontext.Value(`1`)
			}
			testcontract.AssertFixtureRunner(t, "pluralrules", "TestPluralConformanceObservations", f, tc.failure)
		})
	}
}
