package listformat

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

		format, err := New(locale.List{locale.MustParse(fixture.Locale)}, conformanceListOptions(t, fixture))
		if fixture.ErrorCode != "" {
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		input := conformanceStringListInput(t, fixture)
		if fixture.Expected == nil {
			t.Fatal("fixture expected is required")
		}
		if got := format.Format(input); got != *fixture.Expected {
			t.Fatalf("Format(%v) = %q, want %q", input, got, *fixture.Expected)
		}
		if len(fixture.ExpectedParts) > 0 {
			parts := format.FormatToParts(input)
			if len(parts) != len(fixture.ExpectedParts) {
				t.Fatalf("FormatToParts(%v) length = %d, want %d", input, len(parts), len(fixture.ExpectedParts))
			}
			for i, part := range parts {
				want := fixture.ExpectedParts[i]
				if string(part.Type) != want.Type || part.Value != want.Value {
					t.Fatalf("FormatToParts(%v)[%d] = {%q, %q}, want {%q, %q}", input, i, part.Type, part.Value, want.Type, want.Value)
				}
			}
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
	got, err := SupportedLocalesOf(locales, conformanceListOptions(t, fixture))
	if fixture.ErrorCode != "" {
		if !errors.Is(err, intlerr.ErrInvalidOption) {
			t.Fatalf("SupportedLocalesOf() error = %v, want intlerr.ErrInvalidOption", err)
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

func conformanceStringListInput(t *testing.T, fixture conformance.Fixture) []string {
	t.Helper()

	var input []string
	if err := json.Unmarshal(fixture.Input, &input); err != nil {
		t.Fatal(err)
	}
	return input
}

func conformanceListOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher string `json:"localeMatcher"`
		Type          string `json:"type"`
		Style         string `json:"style"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	var opts Options
	if options.LocaleMatcher != "" {
		opts.LocaleMatcher = LocaleMatcher(options.LocaleMatcher)
	}
	if options.Type != "" {
		opts.Type = Type(options.Type)
	}
	if options.Style != "" {
		opts.Style = Style(options.Style)
	}
	return opts
}
