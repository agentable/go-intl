package displaynames

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

	conformance.RunFixtures(t, ".", runDisplayNamesConformanceFixture)
}

func runDisplayNamesConformanceFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	if fixture.ExpectedParts != nil || fixture.ExpectedRange != nil || fixture.ExpectedRangeParts != nil {
		t.Fatal("unsupported displaynames observation")
	}
	if fixture.ErrorCode != "" {
		if fixture.Expected != nil || fixture.ExpectedOK != nil || fixture.ExpectedLocales != nil || fixture.ErrorCode == "invalidOption" && fixture.ExpectedResolved != nil {
			t.Fatal("unsupported displaynames observation")
		}
	}
	if fixture.IsSupportedLocalesOf() {
		if fixture.Expected != nil || fixture.ExpectedOK != nil || fixture.ExpectedResolved != nil {
			t.Fatal("unsupported displaynames observation")
		}
		testcontract.AssertSupportedLocalesOfFixture(t, fixture, intltest.LocaleListJSON, func(locales locale.List) (locale.List, error) {
			return SupportedLocalesOf(locales, conformanceDisplayNamesOptions(t, fixture))
		}, func(code string) error { return conformanceDisplayNamesConstructorError(t, code) })
		return
	}
	if fixture.Feature != "" && fixture.Feature != "of" && fixture.Feature != "resolvedOptions" {
		t.Fatal("unsupported displaynames feature")
	}
	if fixture.ExpectedLocales != nil {
		t.Fatal("unsupported displaynames observation")
	}

	format, err := New(locale.List{intltest.Locale(t, fixture.Locale)}, conformanceDisplayNamesOptions(t, fixture))
	if fixture.ErrorCode == "invalidOption" {
		testcontract.AssertErrorCode(t, "New()", err, fixture.ErrorCode, func(code string) error {
			return conformanceDisplayNamesConstructorError(t, code)
		})
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if fixture.ExpectedResolved != nil {
		testcontract.AssertResolvedOptionsJSON(t, format.ResolvedOptions(), fixture.ExpectedResolved)
	}
	if fixture.Expected == nil && fixture.ExpectedOK == nil && fixture.ErrorCode == "" {
		return
	}
	var input string
	if err := json.Unmarshal(fixture.Input, &input); err != nil {
		t.Fatal(err)
	}
	got, ok, err := format.Of(input)
	if testcontract.AssertErrorCode(t, "Of()", err, fixture.ErrorCode, func(code string) error {
		return conformanceDisplayNamesOfError(t, code)
	}) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if fixture.ExpectedOK != nil && ok != *fixture.ExpectedOK {
		t.Fatalf("Of(%q) ok = %v, want %v", input, ok, *fixture.ExpectedOK)
	}
	if fixture.Expected != nil && got != *fixture.Expected {
		t.Fatalf("Of(%q) = %q, want %q", input, got, *fixture.Expected)
	}
}

func TestConformanceDisplayNamesOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceDisplayNamesOptions(t, conformance.Fixture{
		Options: jsontext.Value(`{"type":"language","style":""}`),
	}))
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want %v", err, intlerr.ErrInvalidOption)
	}
	testcontract.AssertOptionError(t, err, "displaynames", intlerr.InvalidOption, "style", "", "en")
	testcontract.AssertOptionExpected(t, err, `one of "long", "short", "narrow"`)
}

func conformanceDisplayNamesOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher   *string `json:"localeMatcher"`
		Type            *string `json:"type"`
		Style           *string `json:"style"`
		Fallback        *string `json:"fallback"`
		LanguageDisplay *string `json:"languageDisplay"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	return Options{
		LocaleMatcher:   options.LocaleMatcher,
		Type:            options.Type,
		Style:           options.Style,
		Fallback:        options.Fallback,
		LanguageDisplay: options.LanguageDisplay,
	}
}

func conformanceDisplayNamesConstructorError(t testing.TB, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "displaynames constructor", code, "invalidOption")
}

func conformanceDisplayNamesOfError(t testing.TB, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "displaynames Of", code, "invalidCode")
}
