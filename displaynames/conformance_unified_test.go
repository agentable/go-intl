package displaynames

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
		if len(fixture.ExpectedResolved) != 0 {
			assertDisplayNamesResolvedOptions(t, fixture, format.ResolvedOptions())
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
		wantOK := true
		if fixture.ExpectedOK != nil {
			wantOK = *fixture.ExpectedOK
		}
		if ok != wantOK {
			t.Fatalf("Of(%q) ok = %v, want %v", input, ok, wantOK)
		}
		want := fixture.RequiredExpected(t)
		if got != want {
			t.Fatalf("Of(%q) = %q, want %q", input, got, want)
		}
	})
}

func TestConformanceDisplayNamesOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceDisplayNamesOptions(t, conformance.Fixture{
		Options: json.RawMessage(`{"type":"language","style":""}`),
	}))
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want %v", err, intlerr.ErrInvalidOption)
	}
	testcontract.AssertOptionError(t, err, "displaynames", intlerr.InvalidOption, "style", "", "en")
	testcontract.AssertOptionExpected(t, err, `one of "long", "short", "narrow"`)
}

func assertDisplayNamesResolvedOptions(t *testing.T, fixture conformance.Fixture, got ResolvedOptions) {
	t.Helper()

	want := testcontract.ExpectedResolvedOptions(t, fixture)
	testcontract.AssertResolvedString(t, want, "locale", got.Locale.String())
	testcontract.AssertResolvedString(t, want, "style", string(got.Style))
	testcontract.AssertResolvedString(t, want, "type", string(got.Type))
	testcontract.AssertResolvedString(t, want, "fallback", string(got.Fallback))
	testcontract.AssertResolvedOptionalString(t, want, "languageDisplay", got.LanguageDisplay)
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
