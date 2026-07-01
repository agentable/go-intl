package locale

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		var input string
		if err := json.Unmarshal(fixture.Input, &input); err != nil {
			t.Fatal(err)
		}
		loc, err := Parse(input)
		if testcontract.AssertErrorCode(t, "Parse("+input+")", err, fixture.ErrorCode, func(code string) error {
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

func assertConformanceLocaleJSON(t *testing.T, got any, want json.RawMessage) {
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
	gotJSON, err := json.Marshal(got)
	if err != nil {
		return false
	}
	wantJSON, err := json.Marshal(want)
	if err != nil {
		return false
	}
	return bytes.Equal(gotJSON, wantJSON)
}

func conformanceLocaleError(t *testing.T, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "locale", code, "invalid_value")
}
