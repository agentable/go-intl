package datetimeformat

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestDateTimeConformanceObservations(t *testing.T) {
	if testcontract.FixtureRunnerChild(t, runDateTimeConformanceFixture) {
		return
	}
	t.Parallel()

	// Node 26.10.0, en, Gregorian year-only UTC observations.
	resolved := jsontext.Value(`{"locale":"en","calendar":"gregory","numberingSystem":"latn","timeZone":"UTC","year":"numeric"}`)
	parts := []conformance.Part{{Type: "year", Value: "1970"}}
	rangeInput := jsontext.Value(`{"start":"1970-01-01T00:00:00Z","end":"1971-01-01T00:00:00Z"}`)
	rangeParts := []conformance.RangePart{
		{Type: "year", Value: "1970", Source: "startRange"},
		{Type: "literal", Value: " – ", Source: "shared"},
		{Type: "year", Value: "1971", Source: "endRange"},
	}
	tests := []struct {
		name    string
		fixture conformance.Fixture
		failure string
	}{
		{"foreign ok", conformance.Fixture{Expected: new("1970"), ExpectedOK: new(false)}, "unsupported datetimeformat observation"},
		{"foreign locales", conformance.Fixture{Expected: new("1970"), ExpectedLocales: []string{}}, "unsupported datetimeformat observation"},
		{"constructor error and text", conformance.Fixture{Options: jsontext.Value(`{"year":"bad"}`), ErrorCode: "invalid_option", Expected: new("IMPOSSIBLE")}, "unsupported datetimeformat observation"},
		{"constructor error and resolved", conformance.Fixture{Options: jsontext.Value(`{"year":"bad"}`), ErrorCode: "invalid_option", ExpectedResolved: jsontext.Value(`{}`)}, "unsupported datetimeformat observation"},
		{"range and single text", conformance.Fixture{Input: rangeInput, ExpectedRange: new("1970 – 1971"), Expected: new("IMPOSSIBLE")}, "unsupported datetimeformat observation"},
		{"range and single parts", conformance.Fixture{Input: rangeInput, ExpectedRangeParts: rangeParts, ExpectedParts: []conformance.Part{}}, "unsupported datetimeformat observation"},
		{"parts only", conformance.Fixture{ExpectedParts: parts}, ""},
		{"range parts only", conformance.Fixture{Input: rangeInput, ExpectedRangeParts: rangeParts}, ""},
		{"resolved only", conformance.Fixture{Input: rangeInput, ExpectedResolved: resolved}, ""},
		{"single observations", conformance.Fixture{Expected: new("1970"), ExpectedParts: parts, ExpectedResolved: resolved}, ""},
		{"range observations", conformance.Fixture{Input: rangeInput, ExpectedRange: new("1970 – 1971"), ExpectedRangeParts: rangeParts, ExpectedResolved: resolved}, ""},
		{"wrong text", conformance.Fixture{Expected: new("wrong"), ExpectedParts: parts}, "Format("},
		{"wrong parts", conformance.Fixture{Expected: new("1970"), ExpectedParts: []conformance.Part{{Type: "year", Value: "1971"}}}, "FormatToParts"},
		{"empty parts", conformance.Fixture{Expected: new("1970"), ExpectedParts: []conformance.Part{}}, "FormatToParts"},
		{"wrong range text", conformance.Fixture{Input: rangeInput, ExpectedRange: new("wrong"), ExpectedRangeParts: rangeParts}, "FormatRange"},
		{"wrong range parts", conformance.Fixture{Input: rangeInput, ExpectedRange: new("1970 – 1971"), ExpectedRangeParts: []conformance.RangePart{{Type: "year", Value: "1970", Source: "shared"}}}, "FormatRangeToParts"},
		{"empty range parts", conformance.Fixture{Input: rangeInput, ExpectedRange: new("1970 – 1971"), ExpectedRangeParts: []conformance.RangePart{}}, "FormatRangeToParts"},
		{"wrong resolved", conformance.Fixture{Expected: new("1970"), ExpectedResolved: jsontext.Value(`{"locale":"fr"}`)}, "ResolvedOptions"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := tc.fixture
			f.ID = "observation"
			f.Source = "manual:adapter"
			f.Locale = "en"
			if f.Options == nil {
				f.Options = jsontext.Value(`{"timeZone":"UTC","year":"numeric"}`)
			}
			if f.Input == nil {
				f.Input = jsontext.Value(`"1970-01-01T00:00:00Z"`)
			}
			testcontract.AssertFixtureRunner(t, "datetimeformat", "TestDateTimeConformanceObservations", f, tc.failure)
		})
	}
}
