package listformat

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestListConformanceObservations(t *testing.T) {
	if testcontract.FixtureRunnerChild(t, runListConformanceFixture) {
		return
	}
	t.Parallel()

	parts := []conformance.Part{
		{Type: "element", Value: "A"},
		{Type: "literal", Value: " and "},
		{Type: "element", Value: "B"},
	}
	tests := []struct {
		name    string
		fixture conformance.Fixture
		failure string
	}{
		{"resolved only", conformance.Fixture{Input: jsontext.Value(`null`), ExpectedResolved: jsontext.Value(`{"locale":"en","type":"conjunction","style":"long"}`)}, ""},
		{"wrong resolved", conformance.Fixture{ExpectedResolved: jsontext.Value(`{"locale":"wrong","type":"conjunction","style":"long"}`)}, "ResolvedOptions"},
		{"missing resolved field", conformance.Fixture{ExpectedResolved: jsontext.Value(`{"locale":"en","style":"long"}`)}, "ResolvedOptions"},
		{"extra resolved field", conformance.Fixture{ExpectedResolved: jsontext.Value(`{"locale":"en","type":"conjunction","style":"long","extra":true}`)}, "ResolvedOptions"},
		{"text and wrong resolved", conformance.Fixture{Expected: new("A and B"), ExpectedResolved: jsontext.Value(`{"locale":"wrong"}`)}, "ResolvedOptions"},
		{"unsupported range", conformance.Fixture{ExpectedRange: new("wrong")}, "unsupported listformat observation"},
		{"unsupported range parts", conformance.Fixture{ExpectedRangeParts: []conformance.RangePart{}}, "unsupported listformat observation"},
		{"parts only", conformance.Fixture{ExpectedParts: parts}, ""},
		{"empty parts only", conformance.Fixture{Input: jsontext.Value(`[]`), ExpectedParts: []conformance.Part{}}, ""},
		{"text and parts", conformance.Fixture{Expected: new("A and B"), ExpectedParts: parts}, ""},
		{"empty text and parts", conformance.Fixture{Input: jsontext.Value(`[]`), Expected: new(""), ExpectedParts: []conformance.Part{}}, ""},
		{"wrong text", conformance.Fixture{Expected: new("wrong"), ExpectedParts: parts}, "Format("},
		{"wrong element", conformance.Fixture{Expected: new("A and B"), ExpectedParts: []conformance.Part{{Type: "element", Value: "wrong"}}}, "FormatToParts"},
		{"wrong literal", conformance.Fixture{Expected: new("A and B"), ExpectedParts: []conformance.Part{{Type: "literal", Value: "A and B"}}}, "FormatToParts"},
		{"empty expected parts", conformance.Fixture{Expected: new("A and B"), ExpectedParts: []conformance.Part{}}, "FormatToParts"},
		{"constructor error and text", conformance.Fixture{Options: jsontext.Value(`{"style":"invalid"}`), ErrorCode: "invalid_option", Expected: new("IMPOSSIBLE")}, "unsupported listformat observation"},
		{"constructor error and resolved", conformance.Fixture{Options: jsontext.Value(`{"style":"invalid"}`), ErrorCode: "invalid_option", ExpectedResolved: jsontext.Value(`{}`)}, "unsupported listformat observation"},
		{"static and instance text", conformance.Fixture{Feature: "supportedLocalesOf", Input: jsontext.Value(`["en"]`), ExpectedLocales: []string{"en"}, Expected: new("IMPOSSIBLE")}, "unsupported supportedLocalesOf observation"},
		{"static error and locales", conformance.Fixture{Feature: "supportedLocalesOf", Options: jsontext.Value(`{"localeMatcher":"bad"}`), Input: jsontext.Value(`["en"]`), ErrorCode: "invalidOption", ExpectedLocales: []string{}}, "unsupported supportedLocalesOf observation"},
		{"static ignores instance options", conformance.Fixture{Feature: "supportedLocalesOf", Options: jsontext.Value(`{"style":"invalid"}`), Input: jsontext.Value(`["en"]`), ExpectedLocales: []string{"en"}}, ""},
		{"static error only", conformance.Fixture{Feature: "supportedLocalesOf", Options: jsontext.Value(`{"localeMatcher":"bad"}`), Input: jsontext.Value(`["en"]`), ErrorCode: "invalidOption"}, ""},
		{"supported empty", conformance.Fixture{Feature: "supportedLocalesOf", Input: jsontext.Value(`[]`), ExpectedLocales: []string{}}, ""},
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
				f.Input = jsontext.Value(`["A","B"]`)
			}
			testcontract.AssertFixtureRunner(t, "listformat", "TestListConformanceObservations", f, tc.failure)
		})
	}
}
