package relativetimeformat

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

		format, err := New(locale.List{intltest.Locale(t, fixture.Locale)}, conformanceRelativeOptions(t, fixture))
		if fixture.ErrorCode == "invalid_option" {
			testcontract.AssertErrorCode(t, "New()", err, fixture.ErrorCode, func(code string) error {
				return conformanceRelativeOptionError(t, code)
			})
			return
		}
		if err != nil {
			t.Fatal(err)
		}

		input := conformanceRelativeInput(t, fixture)
		got, parts, err := conformanceRelativeOutput(t, format, input)
		if testcontract.AssertErrorCode(t, "Format()", err, fixture.ErrorCode, func(code string) error {
			return conformanceRelativeFormatError(t, code)
		}) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		want := fixture.RequiredExpected(t)
		if got != want {
			t.Fatalf("Format(%v, %q) = %q, want %q", input.Value, input.Unit, got, want)
		}
		if len(fixture.ExpectedParts) > 0 {
			testcontract.AssertParts(t, "FormatToParts", parts, fixture.ExpectedParts, conformanceRelativePart)
		}
	})
}

func TestConformanceRelativeOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceRelativeOptions(t, conformance.Fixture{
		Options: json.RawMessage(`{"style":""}`),
	}))
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want %v", err, intlerr.ErrInvalidOption)
	}
	testcontract.AssertOptionError(t, err, "relativetimeformat", intlerr.InvalidOption, "style", "", "en")
	testcontract.AssertOptionExpected(t, err, `one of "long", "short", "narrow"`)
}

func runSupportedLocalesFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	testcontract.AssertSupportedLocalesOfFixture(t, fixture, intltest.LocaleListJSON, func(locales locale.List) (locale.List, error) {
		return SupportedLocalesOf(locales, conformanceRelativeOptions(t, fixture))
	}, func(code string) error {
		return conformanceRelativeOptionError(t, code)
	})
}

type relativeFixtureInput struct {
	Value json.RawMessage `json:"value"`
	Unit  Unit            `json:"unit"`
}

func conformanceRelativeInput(t *testing.T, fixture conformance.Fixture) relativeFixtureInput {
	t.Helper()

	var input relativeFixtureInput
	if err := json.Unmarshal(fixture.Input, &input); err != nil {
		t.Fatal(err)
	}
	if input.Value == nil {
		t.Fatal("relative time fixture input value is required")
	}
	return input
}

func conformanceRelativeOutput(t *testing.T, format *RelativeTimeFormat, input relativeFixtureInput) (string, []Part, error) {
	t.Helper()

	var value float64
	if err := json.Unmarshal(input.Value, &value); err != nil {
		t.Fatal(err)
	}
	got, err := format.Format(Float(value), input.Unit)
	if err != nil {
		return "", nil, err
	}
	parts, err := format.FormatToParts(Float(value), input.Unit)
	return got, parts, err
}

func conformanceRelativeOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher   *string `json:"localeMatcher"`
		NumberingSystem *string `json:"numberingSystem"`
		Style           *string `json:"style"`
		Numeric         *string `json:"numeric"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	return Options{
		LocaleMatcher:   options.LocaleMatcher,
		NumberingSystem: options.NumberingSystem,
		Style:           options.Style,
		Numeric:         options.Numeric,
	}
}

func conformanceRelativeOptionError(t testing.TB, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "relativetimeformat option", code, "invalid_option")
}

func conformanceRelativeFormatError(t testing.TB, code string) error {
	t.Helper()

	return testcontract.IntlErrorCode(t, "relativetimeformat format", code, "invalid_value")
}

func conformanceRelativePart(part Part) conformance.Part {
	return conformance.Part{Type: string(part.Type), Value: part.Value, Unit: string(part.Unit)}
}
