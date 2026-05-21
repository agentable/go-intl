package relativetimeformat

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		if fixture.Feature == "supportedLocalesOf" {
			runSupportedLocalesFixture(t, fixture)
			return
		}

		format, err := New(locale.List{locale.MustParse(fixture.Locale)}, conformanceRelativeOptions(t, fixture))
		if fixture.ErrorCode == "invalid_option" {
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}

		input := conformanceRelativeInput(t, fixture)
		got, parts, err := conformanceRelativeOutput(t, format, input)
		if fixture.ErrorCode != "" {
			want := conformanceRelativeError(t, fixture.ErrorCode)
			if !errors.Is(err, want) {
				t.Fatalf("Format(%v, %q) error = %v, want %v", input.Value, input.Unit, err, want)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		if fixture.Expected == nil {
			t.Fatal("fixture expected is required")
		}
		if got != *fixture.Expected {
			t.Fatalf("Format(%v, %q) = %q, want %q", input.Value, input.Unit, got, *fixture.Expected)
		}
		if len(fixture.ExpectedParts) > 0 {
			assertRelativeParts(t, input, parts, fixture.ExpectedParts)
		}
	})
}

func runSupportedLocalesFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	var tags []string
	if err := json.Unmarshal(fixture.Input, &tags); err != nil {
		t.Fatal(err)
	}
	locales := make(locale.List, 0, len(tags))
	for _, tag := range tags {
		locales = append(locales, locale.MustParse(tag))
	}
	got, err := SupportedLocalesOf(locales, conformanceRelativeOptions(t, fixture))
	if fixture.ErrorCode != "" {
		if !errors.Is(err, conformanceRelativeError(t, fixture.ErrorCode)) {
			t.Fatalf("SupportedLocalesOf() error = %v, want %s", err, fixture.ErrorCode)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(fixture.ExpectedLocales) {
		t.Fatalf("SupportedLocalesOf(%v) = %v, want %v", tags, got, fixture.ExpectedLocales)
	}
	for i, want := range fixture.ExpectedLocales {
		if got[i].String() != want {
			t.Fatalf("SupportedLocalesOf(%v)[%d] = %q, want %q", tags, i, got[i].String(), want)
		}
	}
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

	var decimal string
	if err := json.Unmarshal(input.Value, &decimal); err == nil {
		got, err := format.FormatDecimal(decimal, input.Unit)
		if err != nil {
			return "", nil, err
		}
		parts, err := format.FormatDecimalToParts(decimal, input.Unit)
		return got, parts, err
	}

	var value float64
	if err := json.Unmarshal(input.Value, &value); err != nil {
		t.Fatal(err)
	}
	got, err := format.FormatFloat64(value, input.Unit)
	if err != nil {
		return "", nil, err
	}
	parts, err := format.FormatFloat64ToParts(value, input.Unit)
	return got, parts, err
}

func conformanceRelativeOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher   string `json:"localeMatcher"`
		NumberingSystem string `json:"numberingSystem"`
		Style           string `json:"style"`
		Numeric         string `json:"numeric"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	var opts Options
	if options.LocaleMatcher != "" {
		opts.LocaleMatcher = LocaleMatcher(options.LocaleMatcher)
	}
	if options.NumberingSystem != "" {
		opts.NumberingSystem = options.NumberingSystem
	}
	if options.Style != "" {
		opts.Style = Style(options.Style)
	}
	if options.Numeric != "" {
		opts.Numeric = Numeric(options.Numeric)
	}
	return opts
}

func conformanceRelativeError(t *testing.T, code string) error {
	t.Helper()

	switch code {
	case "invalid_option":
		return intlerr.ErrInvalidOption
	case "invalid_value":
		return intlerr.ErrInvalidValue
	default:
		t.Fatalf("unsupported errorCode %q", code)
		return nil
	}
}

func assertRelativeParts(t *testing.T, input relativeFixtureInput, got []Part, want []conformance.Part) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("FormatToParts(%v, %q) length = %d, want %d", input.Value, input.Unit, len(got), len(want))
	}
	for i, part := range got {
		wantPart := want[i]
		if string(part.Type) != wantPart.Type || part.Value != wantPart.Value || string(part.Unit) != wantPart.Unit {
			t.Fatalf(
				"FormatToParts(%v, %q)[%d] = {%q, %q, %q}, want {%q, %q, %q}",
				input.Value,
				input.Unit,
				i,
				part.Type,
				part.Value,
				part.Unit,
				wantPart.Type,
				wantPart.Value,
				wantPart.Unit,
			)
		}
	}
}
