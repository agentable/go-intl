package numberformat

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestNumberConformanceObservations(t *testing.T) {
	if testcontract.FixtureRunnerChild(t, runNumberConformanceFixture) {
		return
	}
	t.Parallel()

	resolved := jsontext.Value(`{"locale":"en","numberingSystem":"latn","style":"decimal","minimumIntegerDigits":1,"minimumFractionDigits":0,"maximumFractionDigits":3,"useGrouping":"auto","notation":"standard","signDisplay":"auto","roundingIncrement":1,"roundingMode":"halfExpand","roundingPriority":"auto","trailingZeroDisplay":"auto"}`)
	parts := []conformance.Part{{Type: "integer", Value: "1"}}
	rangeInput := jsontext.Value(`{"start":1,"end":2}`)
	rangeParts := []conformance.RangePart{
		{Type: "integer", Value: "1", Source: "startRange"},
		{Type: "literal", Value: "–", Source: "shared"},
		{Type: "integer", Value: "2", Source: "endRange"},
	}
	tests := []struct {
		name    string
		fixture conformance.Fixture
		failure string
	}{
		{"foreign ok", conformance.Fixture{Expected: new("1"), ExpectedOK: new(false)}, "unsupported numberformat observation"},
		{"foreign locales", conformance.Fixture{Expected: new("1"), ExpectedLocales: []string{}}, "unsupported numberformat observation"},
		{"constructor error and text", conformance.Fixture{Options: jsontext.Value(`{"style":"bad"}`), ErrorCode: "invalid_option", Expected: new("IMPOSSIBLE")}, "unsupported numberformat observation"},
		{"constructor error and resolved", conformance.Fixture{Options: jsontext.Value(`{"style":"bad"}`), ErrorCode: "invalid_option", ExpectedResolved: jsontext.Value(`{}`)}, "unsupported numberformat observation"},
		{"range and single text", conformance.Fixture{Input: rangeInput, ExpectedRange: new("1–2"), Expected: new("IMPOSSIBLE")}, "unsupported numberformat observation"},
		{"range and single parts", conformance.Fixture{Input: rangeInput, ExpectedRangeParts: rangeParts, ExpectedParts: []conformance.Part{}}, "unsupported numberformat observation"},
		{"parts only", conformance.Fixture{ExpectedParts: parts}, ""},
		{"range parts only", conformance.Fixture{Input: rangeInput, ExpectedRangeParts: rangeParts}, ""},
		{"resolved only", conformance.Fixture{Input: jsontext.Value(`{}`), ExpectedResolved: resolved}, ""},
		{"single observations", conformance.Fixture{Expected: new("1"), ExpectedParts: parts, ExpectedResolved: resolved}, ""},
		{"range observations", conformance.Fixture{Input: rangeInput, ExpectedRange: new("1–2"), ExpectedRangeParts: rangeParts, ExpectedResolved: resolved}, ""},
		{"wrong text", conformance.Fixture{Expected: new("2"), ExpectedParts: parts}, "Format()"},
		{"wrong parts", conformance.Fixture{Expected: new("1"), ExpectedParts: []conformance.Part{{Type: "integer", Value: "2"}}}, "FormatToParts"},
		{"empty parts", conformance.Fixture{Expected: new("1"), ExpectedParts: []conformance.Part{}}, "FormatToParts"},
		{"wrong range text", conformance.Fixture{Input: rangeInput, ExpectedRange: new("wrong"), ExpectedRangeParts: rangeParts}, "FormatRange"},
		{"wrong range parts", conformance.Fixture{Input: rangeInput, ExpectedRange: new("1–2"), ExpectedRangeParts: []conformance.RangePart{{Type: "integer", Value: "2", Source: "startRange"}}}, "FormatRangeToParts"},
		{"empty range parts", conformance.Fixture{Input: rangeInput, ExpectedRange: new("1–2"), ExpectedRangeParts: []conformance.RangePart{}}, "FormatRangeToParts"},
		{"wrong resolved", conformance.Fixture{Expected: new("1"), ExpectedResolved: jsontext.Value(`{"locale":"fr"}`)}, "ResolvedOptions"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := tc.fixture
			f.ID = "observation"
			f.Source = "manual:adapter"
			f.Locale = "en"
			if f.Options == nil {
				f.Options = jsontext.Value(`{}`)
			}
			if f.Input == nil {
				f.Input = jsontext.Value(`1`)
			}
			testcontract.AssertFixtureRunner(t, "numberformat", "TestNumberConformanceObservations", f, tc.failure)
		})
	}
}
