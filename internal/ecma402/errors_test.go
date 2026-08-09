package ecma402_test

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestInvalidOptionErrorExpectedWrapsSentinelAndCarriesContext(t *testing.T) {
	t.Parallel()

	err := ecma402.InvalidOptionErrorExpected("numberformat", "currency", "US", "en-US", "a well-formed ISO 4217 currency code", nil)
	if !errors.Is(err, ecma402.ErrInvalidOption) {
		t.Fatalf("InvalidOptionErrorExpected() error = %v, want ErrInvalidOption", err)
	}
	optErr, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("InvalidOptionErrorExpected() error = %T, want OptionError", err)
	}
	if optErr.Owner != "numberformat" || optErr.Kind != "invalidOption" || optErr.Name != "currency" || optErr.Value != "US" || optErr.Locale != "en-US" || optErr.Expected != "a well-formed ISO 4217 currency code" {
		t.Fatalf("OptionError = %+v, want numberformat invalid currency US en-US with expected guidance", optErr)
	}
}

func TestInvalidOptionErrorExpectedWrapsCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("embedded constructor failed")
	err := ecma402.InvalidOptionErrorExpected("relativetimeformat", "numberingSystem", "arab", "en-US", "a supported numbering system", cause)
	if !errors.Is(err, ecma402.ErrInvalidOption) {
		t.Fatalf("InvalidOptionErrorExpected() error = %v, want ErrInvalidOption", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("InvalidOptionErrorExpected() error = %v, want cause", err)
	}
}

func TestUnsupportedStringOptionErrorWrapsSentinelAndCarriesContext(t *testing.T) {
	t.Parallel()

	err := ecma402.UnsupportedStringOptionError("formatter", ecma402.RequiredStringOption("mode", "search", "sort"), "en")
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("UnsupportedStringOptionError() error = %v, want errors.ErrUnsupported", err)
	}
	optErr, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("UnsupportedStringOptionError() error = %T, want OptionError", err)
	}
	if optErr.Owner != "formatter" || optErr.Kind != "unsupportedOption" || optErr.Name != "mode" || optErr.Value != "search" || optErr.Locale != "en" {
		t.Fatalf("OptionError = %+v, want formatter unsupported mode search en", optErr)
	}
	if optErr.Expected != `"sort"` {
		t.Fatalf("OptionError.Expected = %q, want allowed supported value", optErr.Expected)
	}
}

func TestUnsupportedOptionErrorExpectedCarriesGuidance(t *testing.T) {
	t.Parallel()

	err := ecma402.UnsupportedOptionErrorExpected("formatter", "mode", "search", "en", `"sort"`, nil)
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("UnsupportedOptionErrorExpected() error = %v, want errors.ErrUnsupported", err)
	}
	optErr, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("UnsupportedOptionErrorExpected() error = %T, want OptionError", err)
	}
	if optErr.Expected != `"sort"` {
		t.Fatalf("OptionError.Expected = %q, want caller guidance", optErr.Expected)
	}
}

func TestUnsupportedOptionErrorExpectedWrapsCause(t *testing.T) {
	t.Parallel()

	cause := errors.New("time zone backend failed")
	err := ecma402.UnsupportedOptionErrorExpected("datetimeformat", "timeZone", "Mars/Olympus", "en-US", "a supported time zone", cause)
	if !errors.Is(err, ecma402.ErrUnsupportedOption) {
		t.Fatalf("UnsupportedOptionErrorExpected() error = %v, want ErrUnsupportedOption", err)
	}
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("UnsupportedOptionErrorExpected() error = %v, want errors.ErrUnsupported", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("UnsupportedOptionErrorExpected() error = %v, want cause", err)
	}
}

func TestInvalidValueErrorExpectedCarriesGuidance(t *testing.T) {
	t.Parallel()

	cause := errors.New("bad decimal")
	err := ecma402.InvalidValueErrorExpected("relativetimeformat", "value", "Infinity", "", "a finite numeric value", cause)
	if !errors.Is(err, ecma402.ErrInvalidValue) {
		t.Fatalf("InvalidValueErrorExpected() error = %v, want ErrInvalidValue", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("InvalidValueErrorExpected() error = %v, want cause", err)
	}
	intlErr, ok := errors.AsType[*ecma402.Error](err)
	if !ok {
		t.Fatalf("InvalidValueErrorExpected() error = %T, want Error", err)
	}
	if intlErr.Expected != "a finite numeric value" {
		t.Fatalf("Error.Expected = %q, want caller guidance", intlErr.Expected)
	}
}

func TestDecimalValueErrorsCarrySharedGuidance(t *testing.T) {
	t.Parallel()

	cause := errors.New("bad decimal")
	err := ecma402.InvalidDecimalValueError("numberformat", "decimal", "NaN?", cause)
	if !errors.Is(err, ecma402.ErrInvalidValue) {
		t.Fatalf("InvalidDecimalValueError() error = %v, want ErrInvalidValue", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("InvalidDecimalValueError() error = %v, want cause", err)
	}
	intlErr, ok := errors.AsType[*ecma402.Error](err)
	if !ok {
		t.Fatalf("InvalidDecimalValueError() error = %T, want Error", err)
	}
	if intlErr.Owner != "numberformat" || intlErr.Name != "decimal" || intlErr.Value != "NaN?" || intlErr.Expected != ecma402.ExpectedDecimalValue {
		t.Fatalf("Error = %+v, want shared decimal guidance", intlErr)
	}
}

func TestFiniteNumericValueErrorsCarrySharedGuidance(t *testing.T) {
	t.Parallel()

	cause := errors.New("non-finite")
	err := ecma402.InvalidFiniteNumericValueError("pluralrules", "value", "Infinity", "en-US", cause)
	if !errors.Is(err, ecma402.ErrInvalidValue) {
		t.Fatalf("InvalidFiniteNumericValueError() error = %v, want ErrInvalidValue", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("InvalidFiniteNumericValueError() error = %v, want cause", err)
	}
	intlErr, ok := errors.AsType[*ecma402.Error](err)
	if !ok {
		t.Fatalf("InvalidFiniteNumericValueError() error = %T, want Error", err)
	}
	if intlErr.Owner != "pluralrules" || intlErr.Name != "value" || intlErr.Value != "Infinity" || intlErr.Locale != "en-US" || intlErr.Expected != ecma402.ExpectedFiniteNumericValue {
		t.Fatalf("Error = %+v, want shared finite numeric guidance", intlErr)
	}
}

func TestInvalidCodeErrorExpectedCarriesGuidance(t *testing.T) {
	t.Parallel()

	cause := errors.New("bad region")
	err := ecma402.InvalidCodeErrorExpected("displaynames", "region", "USA!", "", "a two-letter ASCII or three-digit region code", cause)
	if !errors.Is(err, ecma402.ErrInvalidCode) {
		t.Fatalf("InvalidCodeErrorExpected() error = %v, want ErrInvalidCode", err)
	}
	if !errors.Is(err, cause) {
		t.Fatalf("InvalidCodeErrorExpected() error = %v, want cause", err)
	}
	intlErr, ok := errors.AsType[*ecma402.Error](err)
	if !ok {
		t.Fatalf("InvalidCodeErrorExpected() error = %T, want Error", err)
	}
	if intlErr.Expected != "a two-letter ASCII or three-digit region code" {
		t.Fatalf("Error.Expected = %q, want caller guidance", intlErr.Expected)
	}
}
