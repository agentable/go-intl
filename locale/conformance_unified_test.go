package locale

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
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
		if fixture.ErrorCode != "" {
			if !errors.Is(err, conformanceLocaleError(t, fixture.ErrorCode)) {
				t.Fatalf("Parse(%q) error = %v, want %q", input, err, fixture.ErrorCode)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if fixture.Expected == nil {
			t.Fatal("fixture expected is required")
		}
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
		if got != *fixture.Expected {
			t.Fatalf("%s(%q) = %q, want %q", fixture.Feature, input, got, *fixture.Expected)
		}
	})
}

func conformanceLocaleError(t *testing.T, code string) error {
	t.Helper()

	switch code {
	case "invalid_value":
		return intlerr.ErrInvalidValue
	default:
		t.Fatalf("unsupported locale errorCode %q", code)
		return nil
	}
}
