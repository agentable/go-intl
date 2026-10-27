package relativetimeformat

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestRelativeConformanceObservations(t *testing.T) {
	if testcontract.FixtureRunnerChild(t, runRelativeConformanceFixture) {
		return
	}
	t.Parallel()

	parts := []conformance.Part{
		{Type: "literal", Value: "in "},
		{Type: "integer", Value: "1", Unit: "day"},
		{Type: "literal", Value: " day"},
	}
	literal := []conformance.Part{{Type: "literal", Value: "tomorrow"}}
	tests := []struct {
		name    string
		fixture conformance.Fixture
		failure string
	}{
		{"constructor error and text", conformance.Fixture{Options: jsontext.Value(`{"style":"bad"}`), ErrorCode: "invalid_option", Expected: new("IMPOSSIBLE")}, "unsupported relativetimeformat observation"},
		{"constructor error and resolved", conformance.Fixture{Options: jsontext.Value(`{"style":"bad"}`), ErrorCode: "invalid_option", ExpectedResolved: jsontext.Value(`{}`)}, "unsupported relativetimeformat observation"},
		{"format error and text", conformance.Fixture{Feature: "format", Input: jsontext.Value(`{"value":1,"unit":"invalid"}`), ErrorCode: "invalid_value", Expected: new("IMPOSSIBLE")}, "unsupported relativetimeformat observation"},
		{"parts error and parts", conformance.Fixture{Feature: "formatToParts", Input: jsontext.Value(`{"value":1,"unit":"invalid"}`), ErrorCode: "invalid_value", ExpectedParts: []conformance.Part{}}, "unsupported relativetimeformat observation"},
		{"format error and resolved", conformance.Fixture{Feature: "format", Input: jsontext.Value(`{"value":1,"unit":"invalid"}`), ErrorCode: "invalid_value", ExpectedResolved: jsontext.Value(`{"locale":"en","style":"long","numeric":"always","numberingSystem":"latn"}`)}, ""},
		{"parts error and resolved", conformance.Fixture{Feature: "formatToParts", Input: jsontext.Value(`{"value":1,"unit":"invalid"}`), ErrorCode: "invalid_value", ExpectedResolved: jsontext.Value(`{"locale":"en","style":"long","numeric":"always","numberingSystem":"latn"}`)}, ""},
		{"format error and wrong resolved", conformance.Fixture{Feature: "format", Input: jsontext.Value(`{"value":1,"unit":"invalid"}`), ErrorCode: "invalid_value", ExpectedResolved: jsontext.Value(`{"locale":"wrong"}`)}, "ResolvedOptions"},
		{"resolved only", conformance.Fixture{Input: jsontext.Value(`null`), ExpectedResolved: jsontext.Value(`{"locale":"en","style":"long","numeric":"always","numberingSystem":"latn"}`)}, ""},
		{"wrong resolved", conformance.Fixture{ExpectedResolved: jsontext.Value(`{"locale":"wrong","style":"long","numeric":"always","numberingSystem":"latn"}`)}, "ResolvedOptions"},
		{"missing resolved field", conformance.Fixture{ExpectedResolved: jsontext.Value(`{"locale":"en","style":"long","numeric":"always"}`)}, "ResolvedOptions"},
		{"extra resolved field", conformance.Fixture{ExpectedResolved: jsontext.Value(`{"locale":"en","style":"long","numeric":"always","numberingSystem":"latn","extra":true}`)}, "ResolvedOptions"},
		{"text and wrong resolved", conformance.Fixture{Expected: new("in 1 day"), ExpectedResolved: jsontext.Value(`{"locale":"wrong"}`)}, "ResolvedOptions"},
		{"format invalid unit", conformance.Fixture{Feature: "format", Input: jsontext.Value(`{"value":1,"unit":"invalid"}`), ErrorCode: "invalid_value"}, ""},
		{"parts invalid unit", conformance.Fixture{Feature: "formatToParts", Input: jsontext.Value(`{"value":1,"unit":"invalid"}`), ErrorCode: "invalid_value"}, ""},
		{"unsupported error method", conformance.Fixture{Feature: "select", Input: jsontext.Value(`{"value":1,"unit":"invalid"}`), ErrorCode: "invalid_value"}, "unsupported relativetimeformat feature"},
		{"unsupported range", conformance.Fixture{ExpectedRange: new("wrong")}, "unsupported relativetimeformat observation"},
		{"numeric parts only", conformance.Fixture{ExpectedParts: parts}, ""},
		{"literal parts only", conformance.Fixture{Options: jsontext.Value(`{"numeric":"auto"}`), ExpectedParts: literal}, ""},
		{"numeric observations", conformance.Fixture{Expected: new("in 1 day"), ExpectedParts: parts}, ""},
		{"literal observations", conformance.Fixture{Options: jsontext.Value(`{"numeric":"auto"}`), Expected: new("tomorrow"), ExpectedParts: literal}, ""},
		{"wrong text", conformance.Fixture{Expected: new("wrong"), ExpectedParts: parts}, "Format("},
		{"wrong number parts", conformance.Fixture{Expected: new("in 1 day"), ExpectedParts: []conformance.Part{{Type: "integer", Value: "2", Unit: "day"}}}, "FormatToParts"},
		{"wrong literal parts", conformance.Fixture{Options: jsontext.Value(`{"numeric":"auto"}`), Expected: new("tomorrow"), ExpectedParts: []conformance.Part{{Type: "literal", Value: "yesterday"}}}, "FormatToParts"},
		{"empty expected parts", conformance.Fixture{Expected: new("in 1 day"), ExpectedParts: []conformance.Part{}}, "FormatToParts"},
		{"negative zero", conformance.Fixture{Input: jsontext.Value(`{"value":-0,"unit":"day"}`), ExpectedParts: []conformance.Part{{Type: "integer", Value: "0", Unit: "day"}, {Type: "literal", Value: " days ago"}}}, ""},
		{"positive zero", conformance.Fixture{Input: jsontext.Value(`{"value":0,"unit":"day"}`), ExpectedParts: []conformance.Part{{Type: "literal", Value: "in "}, {Type: "integer", Value: "0", Unit: "day"}, {Type: "literal", Value: " days"}}}, ""},
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
				f.Input = jsontext.Value(`{"value":1,"unit":"day"}`)
			}
			testcontract.AssertFixtureRunner(t, "relativetimeformat", "TestRelativeConformanceObservations", f, tc.failure)
		})
	}
}
