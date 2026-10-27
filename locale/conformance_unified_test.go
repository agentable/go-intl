package locale

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"reflect"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, testcontract.FixtureRunnerRoot("."), func(t *testing.T, fixture conformance.Fixture) {
		switch fixture.Feature {
		case "canonicalize", "maximize", "minimize", "weekInfo":
		default:
			t.Fatalf("fixture %s: unsupported locale feature %q", fixture.ID, fixture.Feature)
		}
		if fixture.ExpectedOK != nil || fixture.ExpectedLocales != nil || fixture.ExpectedParts != nil || fixture.ExpectedRange != nil || fixture.ExpectedRangeParts != nil ||
			fixture.ErrorCode != "" && (fixture.Expected != nil || fixture.ExpectedResolved != nil) ||
			fixture.Feature == "weekInfo" && fixture.Expected != nil ||
			fixture.Feature != "weekInfo" && fixture.ExpectedResolved != nil {
			t.Fatalf("fixture %s: unsupported locale observation", fixture.ID)
		}
		var input string
		if err := json.Unmarshal(fixture.Input, &input); err != nil {
			t.Fatal(err)
		}
		var opts Options
		if err := json.Unmarshal(fixture.Options, &opts, json.MatchCaseInsensitiveNames(true)); err != nil {
			t.Fatal(err)
		}
		loc, err := New(input, opts)
		if testcontract.AssertErrorCode(t, "New("+input+")", err, fixture.ErrorCode, func(code string) error {
			return conformanceLocaleError(t, code)
		}) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if fixture.Feature == "weekInfo" {
			assertConformanceLocaleJSON(t, loc.GetWeekInfo(), fixture.ExpectedResolved)
			return
		}
		want := fixture.RequiredExpected(t)
		got := ""
		switch fixture.Feature {
		case "canonicalize":
			got = loc.String()
		case "maximize":
			got = loc.Maximize().String()
		case "minimize":
			got = loc.Minimize().String()
		default:
			t.Fatalf("unsupported locale feature %q", fixture.Feature)
		}
		if got != want {
			t.Fatalf("%s(%q) = %q, want %q", fixture.Feature, input, got, want)
		}
	})
}

func assertConformanceLocaleJSON(t *testing.T, got any, want jsontext.Value) {
	t.Helper()

	if len(want) == 0 {
		t.Fatal("fixture expectedResolvedOptions is required")
	}
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(gotJSON, want) {
		t.Fatalf("JSON = %s, want %s", gotJSON, want)
	}
}

func jsonEqual(a, b []byte) bool {
	var got, want any
	if err := json.Unmarshal(a, &got); err != nil {
		return false
	}
	if err := json.Unmarshal(b, &want); err != nil {
		return false
	}
	return reflect.DeepEqual(got, want)
}

func conformanceLocaleError(t *testing.T, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "locale", code, "invalid_value", "invalid_option")
}

func TestLocaleConformanceObservations(t *testing.T) {
	t.Parallel()
	week := jsontext.Value(`{"firstDay":7,"weekend":[6,7]}`)
	tests := []struct {
		name    string
		fixture conformance.Fixture
		failure string
	}{
		{"canonicalize", conformance.Fixture{Expected: new("en-US")}, ""},
		{"foreign ok", conformance.Fixture{Expected: new("en-US"), ExpectedOK: new(false)}, "unsupported locale observation"},
		{"canonicalize and resolved", conformance.Fixture{Expected: new("en-US"), ExpectedResolved: week}, "unsupported locale observation"},
		{"week info", conformance.Fixture{Feature: "weekInfo", ExpectedResolved: week}, ""},
		{"week info and text", conformance.Fixture{Feature: "weekInfo", ExpectedResolved: week, Expected: new("IMPOSSIBLE")}, "unsupported locale observation"},
		{"constructor error", conformance.Fixture{Input: jsontext.Value(`"en_US"`), ErrorCode: "invalid_value"}, ""},
		{"constructor error and text", conformance.Fixture{Input: jsontext.Value(`"en_US"`), ErrorCode: "invalid_value", Expected: new("IMPOSSIBLE")}, "unsupported locale observation"},
		{"unknown operation and constructor error", conformance.Fixture{Feature: "unknown", Input: jsontext.Value(`"en_US"`), ErrorCode: "invalid_value"}, "unsupported locale feature"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := tc.fixture
			f.ID, f.Source, f.Locale = "observation", "manual:adapter", "en-US"
			f.Options = jsontext.Value(`{}`)
			if f.Input == nil {
				f.Input = jsontext.Value(`"en-US"`)
			}
			if f.Feature == "" {
				f.Feature = "canonicalize"
			}
			testcontract.AssertFixtureRunner(t, "locale", "TestUnifiedConformanceFixtures", f, tc.failure)
		})
	}
}
