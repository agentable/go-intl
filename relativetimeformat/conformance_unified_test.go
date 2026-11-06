package relativetimeformat

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

	conformance.RunFixtures(t, ".", runRelativeConformanceFixture)
}

func runRelativeConformanceFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	if fixture.IsSupportedLocalesOf() {
		runSupportedLocalesFixture(t, fixture)
		return
	}

	constructorError := fixture.ErrorCode == "invalid_option"
	if fixture.ExpectedRange != nil || fixture.ExpectedRangeParts != nil || fixture.ExpectedOK != nil || fixture.ExpectedLocales != nil ||
		fixture.ErrorCode != "" && (fixture.Expected != nil || fixture.ExpectedParts != nil) ||
		constructorError && fixture.ExpectedResolved != nil {
		t.Fatalf("fixture %s: unsupported relativetimeformat observation", fixture.ID)
	}
	format, err := New(locale.List{intltest.Locale(t, fixture.Locale)}, conformanceRelativeOptions(t, fixture))
	if constructorError {
		testcontract.AssertErrorCode(t, "New()", err, fixture.ErrorCode, func(code string) error {
			return conformanceRelativeOptionError(t, code)
		})
		return
	}
	if err != nil {
		t.Fatal(err)
	}

	if fixture.ExpectedResolved != nil {
		testcontract.AssertResolvedOptionsJSON(t, format.ResolvedOptions(), fixture.ExpectedResolved)
	}
	if fixture.ErrorCode == "" && fixture.Expected == nil && fixture.ExpectedParts == nil {
		return
	}
	input := conformanceRelativeInput(t, fixture)
	var value float64
	if err := json.Unmarshal(input.Value, &value); err != nil {
		t.Fatal(err)
	}
	if fixture.ErrorCode != "" {
		operation := "Format()"
		switch fixture.Feature {
		case "", "format":
			_, err = format.Format(Float(value), input.Unit)
		case "formatToParts":
			operation = "FormatToParts()"
			_, err = format.FormatToParts(Float(value), input.Unit)
		default:
			t.Fatalf("unsupported relativetimeformat feature %q", fixture.Feature)
		}
		testcontract.AssertErrorCode(t, operation, err, fixture.ErrorCode, func(code string) error {
			return conformanceRelativeFormatError(t, code)
		})
		return
	}
	if fixture.Expected != nil {
		got, err := format.Format(Float(value), input.Unit)
		if err != nil {
			t.Fatal(err)
		}
		if got != *fixture.Expected {
			t.Fatalf("Format(%v, %q) = %q, want %q", input.Value, input.Unit, got, *fixture.Expected)
		}
	}
	if fixture.ExpectedParts != nil {
		parts, err := format.FormatToParts(Float(value), input.Unit)
		if err != nil {
			t.Fatal(err)
		}
		testcontract.AssertParts(t, "FormatToParts", parts, fixture.ExpectedParts, conformanceRelativePart)
	}
}

func TestConformanceRelativeOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceRelativeOptions(t, conformance.Fixture{
		Options: jsontext.Value(`{"style":""}`),
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
	Value jsontext.Value `json:"value"`
	Unit  Unit           `json:"unit"`
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
