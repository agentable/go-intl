package testcontract

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/tools/conformance"
)

type testLocale string

func (l testLocale) String() string {
	return string(l)
}

func TestAssertErrorCode(t *testing.T) {
	t.Parallel()

	want := errors.New("invalid option")
	if handled := AssertErrorCode(t, "New", want, "invalid_option", func(code string) error {
		if code != "invalid_option" {
			t.Fatalf("code = %q, want invalid_option", code)
		}
		return want
	}); !handled {
		t.Fatal("AssertErrorCode() handled = false, want true")
	}
}

func TestAssertErrorCodeNoCode(t *testing.T) {
	t.Parallel()

	if handled := AssertErrorCode(t, "New", nil, "", func(string) error {
		t.Fatal("wantError called for empty code")
		return nil
	}); handled {
		t.Fatal("AssertErrorCode() handled = true, want false")
	}
}

func TestIntlErrorCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code string
		want error
	}{
		{code: "invalid_option", want: intlerr.ErrInvalidOption},
		{code: "invalidOption", want: intlerr.ErrInvalidOption},
		{code: "invalid-option", want: intlerr.ErrInvalidOption},
		{code: "unsupported_option", want: intlerr.ErrUnsupportedOption},
		{code: "unsupportedOption", want: intlerr.ErrUnsupportedOption},
		{code: "unsupported-option", want: intlerr.ErrUnsupportedOption},
		{code: "invalid_value", want: intlerr.ErrInvalidValue},
		{code: "invalidValue", want: intlerr.ErrInvalidValue},
		{code: "invalid-value", want: intlerr.ErrInvalidValue},
		{code: "invalid_code", want: intlerr.ErrInvalidCode},
		{code: "invalidCode", want: intlerr.ErrInvalidCode},
		{code: "invalid-code", want: intlerr.ErrInvalidCode},
	}
	for _, tc := range tests {
		t.Run(tc.code, func(t *testing.T) {
			t.Parallel()

			if got := IntlErrorCode(t, "test", tc.code, tc.code); !errors.Is(got, tc.want) {
				t.Fatalf("IntlErrorCode(%q) = %v, want %v", tc.code, got, tc.want)
			}
		})
	}
}

func TestAssertSupportedLocalesOfFixture(t *testing.T) {
	t.Parallel()

	fixture := conformance.Fixture{
		Input:           json.RawMessage(`["fr-FR","en-US"]`),
		ExpectedLocales: []string{"fr-FR", "en-US"},
	}
	AssertSupportedLocalesOfFixture(t, fixture, testLocaleList, func(locales []testLocale) ([]testLocale, error) {
		return locales, nil
	}, func(code string) error {
		t.Fatalf("wantError called for empty code %q", code)
		return nil
	})
}

func TestAssertSupportedLocalesOfFixtureErrorCode(t *testing.T) {
	t.Parallel()

	want := errors.New("invalid option")
	fixture := conformance.Fixture{
		Input:     json.RawMessage(`[""]`),
		ErrorCode: "invalid_option",
	}
	AssertSupportedLocalesOfFixture(t, fixture, testLocaleList, func([]testLocale) ([]testLocale, error) {
		return nil, want
	}, func(code string) error {
		if code != "invalid_option" {
			t.Fatalf("code = %q, want invalid_option", code)
		}
		return want
	})
}

func testLocaleList(t testing.TB, data []byte) []testLocale {
	t.Helper()

	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		t.Fatal(err)
	}
	locales := make([]testLocale, len(values))
	for i, value := range values {
		locales[i] = testLocale(value)
	}
	return locales
}
