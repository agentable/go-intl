package locale

import (
	"encoding/json"
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		var input string
		if err := json.Unmarshal(fixture.Input, &input); err != nil {
			t.Fatal(err)
		}
		if fixture.Expected == nil {
			t.Fatal("fixture expected is required")
		}
		loc := MustParse(input)
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
