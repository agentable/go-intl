package testcontract

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/tools/conformance"
)

// AssertSupportedLocalesOfFixture verifies the shared supportedLocalesOf
// conformance fixture contract while keeping formatter-owned option parsing at
// the call site.
func AssertSupportedLocalesOfFixture[L any, S ~[]E, E interface{ String() string }](
	t testing.TB,
	fixture conformance.Fixture,
	requested func(testing.TB, []byte) L,
	run func(L) (S, error),
	wantError func(string) error,
) {
	t.Helper()

	got, err := run(requested(t, fixture.Input))
	if AssertErrorCode(t, "SupportedLocalesOf()", err, fixture.ErrorCode, wantError) {
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	AssertLocaleListStrings(t, "SupportedLocalesOf", got, fixture.ExpectedLocales)
}

// AssertErrorCode verifies a fixture errorCode against a package-owned error
// mapping. It reports whether the fixture expected an error and was handled.
func AssertErrorCode(t testing.TB, operation string, err error, code string, wantError func(string) error) bool {
	t.Helper()

	if code == "" {
		return false
	}
	want := wantError(code)
	if !errors.Is(err, want) {
		t.Fatalf("%s error = %v, want %v", operation, err, want)
	}
	return true
}

// IntlErrorCode maps supported conformance fixture errorCode spellings to the
// package sentinel they expect while call sites keep their allowed-code boundary.
func IntlErrorCode(t testing.TB, owner, code string, allowed ...string) error {
	t.Helper()

	if len(allowed) > 0 && !errorCodeAllowed(code, allowed) {
		t.Fatalf("unsupported %s errorCode %q", owner, code)
		return nil
	}
	switch code {
	case "invalid_option", "invalidOption", "invalid-option":
		return intlerr.ErrInvalidOption
	case "unsupported_option", "unsupportedOption", "unsupported-option":
		return intlerr.ErrUnsupportedOption
	case "invalid_value", "invalidValue", "invalid-value":
		return intlerr.ErrInvalidValue
	case "invalid_code", "invalidCode", "invalid-code":
		return intlerr.ErrInvalidCode
	default:
		t.Fatalf("unsupported %s errorCode %q", owner, code)
		return nil
	}
}

func errorCodeAllowed(code string, allowed []string) bool {
	for _, value := range allowed {
		if code == value {
			return true
		}
	}
	return false
}
