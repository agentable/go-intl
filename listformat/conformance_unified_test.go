package listformat

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		if fixture.IsSupportedLocalesOf() {
			runSupportedLocalesFixture(t, fixture)
			return
		}

		format, err := New(locale.List{intltest.Locale(t, fixture.Locale)}, conformanceListOptions(t, fixture))
		if testcontract.AssertErrorCode(t, "New()", err, fixture.ErrorCode, func(code string) error {
			return conformanceListConstructorError(t, code)
		}) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		input := conformanceStringListInput(t, fixture)
		want := fixture.RequiredExpected(t)
		if got := format.Format(input); got != want {
			t.Fatalf("Format(%v) = %q, want %q", input, got, want)
		}
		if len(fixture.ExpectedParts) > 0 {
			parts := format.FormatToParts(input)
			testcontract.AssertParts(t, "FormatToParts", parts, fixture.ExpectedParts, conformanceListPart)
		}
	})
}

func TestConformanceListOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceListOptions(t, conformance.Fixture{
		Options: json.RawMessage(`{"style":""}`),
	}))
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want %v", err, intlerr.ErrInvalidOption)
	}
	testcontract.AssertOptionError(t, err, "listformat", intlerr.InvalidOption, "style", "", "en")
	testcontract.AssertOptionExpected(t, err, `one of "long", "short", "narrow"`)
}

func runSupportedLocalesFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	testcontract.AssertSupportedLocalesOfFixture(t, fixture, intltest.LocaleListJSON, func(locales locale.List) (locale.List, error) {
		return SupportedLocalesOf(locales, conformanceListOptions(t, fixture))
	}, func(code string) error {
		return conformanceListSupportedLocalesError(t, code)
	})
}

func conformanceListConstructorError(t testing.TB, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "listformat constructor", code, "invalid_option")
}

func conformanceListSupportedLocalesError(t testing.TB, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "listformat supportedLocalesOf", code, "invalidOption")
}

func conformanceStringListInput(t *testing.T, fixture conformance.Fixture) []string {
	t.Helper()

	var input []string
	if err := json.Unmarshal(fixture.Input, &input); err != nil {
		t.Fatal(err)
	}
	return input
}

func conformanceListPart(part Part) conformance.Part {
	return conformance.Part{Type: string(part.Type), Value: part.Value}
}

func conformanceListOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher *string `json:"localeMatcher"`
		Type          *string `json:"type"`
		Style         *string `json:"style"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	return Options{
		LocaleMatcher: options.LocaleMatcher,
		Type:          options.Type,
		Style:         options.Style,
	}
}
