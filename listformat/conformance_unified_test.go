package listformat

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
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

	conformance.RunFixtures(t, ".", runListConformanceFixture)
}

func runListConformanceFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	if fixture.IsSupportedLocalesOf() {
		runSupportedLocalesFixture(t, fixture)
		return
	}
	if fixture.ExpectedRange != nil || fixture.ExpectedRangeParts != nil || fixture.ExpectedOK != nil || fixture.ExpectedLocales != nil ||
		fixture.ErrorCode != "" && (fixture.Expected != nil || fixture.ExpectedParts != nil || fixture.ExpectedResolved != nil) {
		t.Fatalf("fixture %s: unsupported listformat observation", fixture.ID)
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
	if fixture.ExpectedResolved != nil {
		testcontract.AssertResolvedOptionsJSON(t, format.ResolvedOptions(), fixture.ExpectedResolved)
	}
	if fixture.Expected == nil && fixture.ExpectedParts == nil {
		return
	}
	input := conformanceStringListInput(t, fixture)
	if fixture.Expected != nil {
		if got := format.Format(input); got != *fixture.Expected {
			t.Fatalf("Format(%v) = %q, want %q", input, got, *fixture.Expected)
		}
	}
	if fixture.ExpectedParts != nil {
		parts := format.FormatToParts(input)
		testcontract.AssertParts(t, "FormatToParts", parts, fixture.ExpectedParts, conformanceListPart)
	}
}

func TestConformanceListOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceListOptions(t, conformance.Fixture{
		Options: jsontext.Value(`{"style":""}`),
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
