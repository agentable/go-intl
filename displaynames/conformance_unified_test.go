package displaynames

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		format, err := New(locale.List{intltest.Locale(t, fixture.Locale)}, conformanceDisplayNamesOptions(t, fixture))
		if err != nil && fixture.ErrorCode != "" {
			if !errors.Is(err, conformanceDisplayNamesError(t, fixture.ErrorCode)) {
				t.Fatalf("New() error = %v, want %q", err, fixture.ErrorCode)
			}
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
		if fixture.ErrorCode != "" {
			if !errors.Is(err, conformanceDisplayNamesError(t, fixture.ErrorCode)) {
				t.Fatalf("Of(%q) error = %v, want %q", input, err, fixture.ErrorCode)
			}
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
		if fixture.Expected == nil {
			t.Fatal("fixture expected is required")
		}
		if got != *fixture.Expected {
			t.Fatalf("Of(%q) = %q, want %q", input, got, *fixture.Expected)
		}
	})
}

func assertDisplayNamesResolvedOptions(t *testing.T, fixture conformance.Fixture, got ResolvedOptions) {
	t.Helper()

	var want map[string]json.RawMessage
	if err := json.Unmarshal(fixture.ExpectedResolved, &want); err != nil {
		t.Fatal(err)
	}
	assertDisplayNamesResolvedString(t, want, "locale", got.Locale.String())
	assertDisplayNamesResolvedString(t, want, "style", string(got.Style))
	assertDisplayNamesResolvedString(t, want, "type", string(got.Type))
	assertDisplayNamesResolvedString(t, want, "fallback", string(got.Fallback))
	assertDisplayNamesResolvedOptionalString(t, want, "languageDisplay", got.LanguageDisplay)
}

func assertDisplayNamesResolvedString(t *testing.T, values map[string]json.RawMessage, name, got string) {
	t.Helper()

	raw, ok := values[name]
	if !ok {
		return
	}
	var want string
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("expectedResolvedOptions.%s = %s: %v", name, raw, err)
	}
	if got != want {
		t.Fatalf("ResolvedOptions().%s = %q, want %q", name, got, want)
	}
}

func assertDisplayNamesResolvedOptionalString(t *testing.T, values map[string]json.RawMessage, name string, got *LanguageDisplay) {
	t.Helper()

	raw, ok := values[name]
	if !ok {
		return
	}
	if string(raw) == "null" {
		if got != nil {
			t.Fatalf("ResolvedOptions().%s = %q, want omitted", name, *got)
		}
		return
	}
	var want LanguageDisplay
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("expectedResolvedOptions.%s = %s: %v", name, raw, err)
	}
	if got == nil {
		t.Fatalf("ResolvedOptions().%s omitted, want %q", name, want)
	}
	if *got != want {
		t.Fatalf("ResolvedOptions().%s = %q, want %q", name, *got, want)
	}
}

func conformanceDisplayNamesOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher   string `json:"localeMatcher"`
		Type            string `json:"type"`
		Style           string `json:"style"`
		Fallback        string `json:"fallback"`
		LanguageDisplay string `json:"languageDisplay"`
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
	if options.Fallback != "" {
		opts.Fallback = Fallback(options.Fallback)
	}
	if options.LanguageDisplay != "" {
		opts.LanguageDisplay = LanguageDisplay(options.LanguageDisplay)
	}
	return opts
}

func conformanceDisplayNamesError(t *testing.T, code string) error {
	t.Helper()

	switch code {
	case "invalidOption":
		return intlerr.ErrInvalidOption
	case "invalidCode":
		return intlerr.ErrInvalidCode
	default:
		t.Fatalf("unsupported displaynames error code %q", code)
		return nil
	}
}
