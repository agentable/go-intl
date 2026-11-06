package displaynames

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestDisplayNamesConformanceObservations(t *testing.T) {
	if testcontract.FixtureRunnerChild(t, runDisplayNamesConformanceFixture) {
		return
	}
	t.Parallel()

	resolved := jsontext.Value(`{"locale":"en","style":"long","type":"language","fallback":"code","languageDisplay":"dialect"}`)
	tests := []struct {
		name    string
		fixture conformance.Fixture
		failure string
	}{
		{"empty parts", conformance.Fixture{ExpectedParts: []conformance.Part{}}, "unsupported displaynames observation"},
		{"range", conformance.Fixture{ExpectedRange: new("English")}, "unsupported displaynames observation"},
		{"range parts", conformance.Fixture{ExpectedRangeParts: []conformance.RangePart{}}, "unsupported displaynames observation"},
		{"text with parts", conformance.Fixture{Expected: new("English"), ExpectedParts: []conformance.Part{}}, "unsupported displaynames observation"},
		{"unknown feature", conformance.Fixture{Feature: "bogus", Expected: new("English")}, "unsupported displaynames feature"},
		{"supported locales", conformance.Fixture{Feature: "supportedLocalesOf", Options: jsontext.Value(`{}`), Input: jsontext.Value(`["en"]`), ExpectedLocales: []string{"en"}}, ""},
		{"wrong supported locales", conformance.Fixture{Feature: "supportedLocalesOf", Options: jsontext.Value(`{}`), Input: jsontext.Value(`["en"]`), ExpectedLocales: []string{"fr"}}, "SupportedLocalesOf"},
		{"empty supported locales", conformance.Fixture{Feature: "supportedLocalesOf", Options: jsontext.Value(`{}`), Input: jsontext.Value(`[]`), ExpectedLocales: []string{}}, ""},
		{"invalid supported matcher", conformance.Fixture{Feature: "supportedLocalesOf", Options: jsontext.Value(`{"localeMatcher":"bad"}`), Input: jsontext.Value(`["en"]`), ErrorCode: "invalidOption"}, ""},
		{"resolved only", conformance.Fixture{Input: jsontext.Value(`{}`), ExpectedResolved: resolved}, ""},
		{"ok only", conformance.Fixture{ExpectedOK: new(true)}, ""},
		{"missing ok only", conformance.Fixture{Options: jsontext.Value(`{"type":"language","fallback":"none"}`), Input: jsontext.Value(`"zz"`), ExpectedOK: new(false)}, ""},
		{"text only", conformance.Fixture{Expected: new("English")}, ""},
		{"missing text only", conformance.Fixture{Options: jsontext.Value(`{"type":"language","fallback":"none"}`), Input: jsontext.Value(`"zz"`), Expected: new("")}, ""},
		{"observations", conformance.Fixture{Expected: new("English"), ExpectedOK: new(true), ExpectedResolved: resolved}, ""},
		{"wrong ok", conformance.Fixture{Expected: new("English"), ExpectedOK: new(false)}, "ok ="},
		{"wrong text", conformance.Fixture{Expected: new("wrong"), ExpectedOK: new(true)}, "Of("},
		{"wrong resolved", conformance.Fixture{Expected: new("English"), ExpectedResolved: jsontext.Value(`{"locale":"fr"}`)}, "ResolvedOptions"},

		{"constructor error with text", conformance.Fixture{Options: jsontext.Value(`{"type":""}`), ErrorCode: "invalidOption", Expected: new("ignored")}, "unsupported displaynames observation"},
		{"constructor error with ok", conformance.Fixture{Options: jsontext.Value(`{"type":""}`), ErrorCode: "invalidOption", ExpectedOK: new(false)}, "unsupported displaynames observation"},
		{"constructor error with resolved", conformance.Fixture{Options: jsontext.Value(`{"type":""}`), ErrorCode: "invalidOption", ExpectedResolved: resolved}, "unsupported displaynames observation"},
		{"of error with text", conformance.Fixture{Input: jsontext.Value(`"not@language"`), ErrorCode: "invalidCode", Expected: new("ignored")}, "unsupported displaynames observation"},
		{"of error with ok", conformance.Fixture{Input: jsontext.Value(`"not@language"`), ErrorCode: "invalidCode", ExpectedOK: new(false)}, "unsupported displaynames observation"},
		{"of error with resolved", conformance.Fixture{Input: jsontext.Value(`"not@language"`), ErrorCode: "invalidCode", ExpectedResolved: resolved}, ""},
		{"of error with wrong resolved", conformance.Fixture{Input: jsontext.Value(`"not@language"`), ErrorCode: "invalidCode", ExpectedResolved: jsontext.Value(`{"locale":"wrong"}`)}, "ResolvedOptions"},
		{"supported error with locales", conformance.Fixture{Feature: "supportedLocalesOf", Options: jsontext.Value(`{"localeMatcher":"bad"}`), Input: jsontext.Value(`["en"]`), ErrorCode: "invalidOption", ExpectedLocales: []string{}}, "unsupported displaynames observation"},
		{"constructor error", conformance.Fixture{Options: jsontext.Value(`{"type":""}`), ErrorCode: "invalidOption"}, ""},
		{"of error", conformance.Fixture{Input: jsontext.Value(`"not@language"`), ErrorCode: "invalidCode"}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			f := tc.fixture
			f.ID = "observation"
			f.Source = "manual:adapter"
			f.Locale = "en"
			if f.Options == nil {
				f.Options = jsontext.Value(`{"type":"language"}`)
			}
			if f.Input == nil {
				f.Input = jsontext.Value(`"en"`)
			}
			testcontract.AssertFixtureRunner(t, "displaynames", "TestDisplayNamesConformanceObservations", f, tc.failure)
		})
	}
}
