package testcontract

import (
	"encoding/json"
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

type resolvedString string

func TestExpectedResolvedOptions(t *testing.T) {
	t.Parallel()

	fixture := conformance.Fixture{
		ExpectedResolved: json.RawMessage(`{"locale":"en-US","fractionalDigits":null}`),
	}
	values := ExpectedResolvedOptions(t, fixture)
	AssertResolvedString(t, values, "locale", "en-US")
	AssertResolvedOptionalInt(t, values, "fractionalDigits", nil)
}

func TestAssertResolvedOptionalString(t *testing.T) {
	t.Parallel()

	values := ResolvedOptionsFixture{
		"display": json.RawMessage(`"short"`),
		"omit":    json.RawMessage(`null`),
	}
	got := resolvedString("short")
	AssertResolvedOptionalString(t, values, "display", &got)
	AssertResolvedOptionalString[resolvedString](t, values, "omit", nil)
}

func TestAssertResolvedInt(t *testing.T) {
	t.Parallel()

	values := ResolvedOptionsFixture{"digits": json.RawMessage(`2`)}
	AssertResolvedInt(t, values, "digits", 2)
}

func TestAssertResolvedBool(t *testing.T) {
	t.Parallel()

	values := ResolvedOptionsFixture{"numeric": json.RawMessage(`true`)}
	AssertResolvedBool(t, values, "numeric", true)
}
