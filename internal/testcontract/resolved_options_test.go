package testcontract

import (
	"encoding/json/jsontext"
	"testing"

	"github.com/agentable/go-intl/tools/conformance"
)

type resolvedString string

func TestExpectedResolvedOptions(t *testing.T) {
	t.Parallel()

	fixture := conformance.Fixture{
		ExpectedResolved: jsontext.Value(`{"locale":"en-US","fractionalDigits":null}`),
	}
	values := ExpectedResolvedOptions(t, fixture)
	AssertResolvedString(t, values, "locale", "en-US")
	AssertResolvedOptionalInt(t, values, "fractionalDigits", nil)
}

func TestAssertResolvedOptionalString(t *testing.T) {
	t.Parallel()

	values := ResolvedOptionsFixture{
		"display": jsontext.Value(`"short"`),
		"omit":    jsontext.Value(`null`),
	}
	got := resolvedString("short")
	AssertResolvedOptionalString(t, values, "display", &got)
	AssertResolvedOptionalString[resolvedString](t, values, "omit", nil)
}

func TestAssertResolvedInt(t *testing.T) {
	t.Parallel()

	values := ResolvedOptionsFixture{"digits": jsontext.Value(`2`)}
	AssertResolvedInt(t, values, "digits", 2)
}

func TestAssertResolvedBool(t *testing.T) {
	t.Parallel()

	values := ResolvedOptionsFixture{"numeric": jsontext.Value(`true`)}
	AssertResolvedBool(t, values, "numeric", true)
}
