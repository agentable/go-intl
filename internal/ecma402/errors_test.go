package ecma402_test

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestOptionErrorWrapsSentinelAndCarriesContext(t *testing.T) {
	t.Parallel()

	err := ecma402.InvalidOptionError("numberformat", "currency", "US", "en-US", ecma402.ErrInvalidOption)
	if !errors.Is(err, ecma402.ErrInvalidOption) {
		t.Fatalf("InvalidOptionError() error = %v, want ErrInvalidOption", err)
	}
	optErr, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("InvalidOptionError() error = %T, want OptionError", err)
	}
	if optErr.Owner != "numberformat" || optErr.Kind != "invalidOption" || optErr.Name != "currency" || optErr.Value != "US" || optErr.Locale != "en-US" {
		t.Fatalf("OptionError = %+v, want numberformat invalid currency US en-US", optErr)
	}
}

func TestUnsupportedOptionErrorWrapsSentinelAndCarriesContext(t *testing.T) {
	t.Parallel()

	err := ecma402.UnsupportedOptionError("collator", "usage", "search", "en", errors.ErrUnsupported)
	if !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("UnsupportedOptionError() error = %v, want errors.ErrUnsupported", err)
	}
	optErr, ok := errors.AsType[*ecma402.OptionError](err)
	if !ok {
		t.Fatalf("UnsupportedOptionError() error = %T, want OptionError", err)
	}
	if optErr.Owner != "collator" || optErr.Kind != "unsupportedOption" || optErr.Name != "usage" || optErr.Value != "search" || optErr.Locale != "en" {
		t.Fatalf("OptionError = %+v, want collator unsupported usage search en", optErr)
	}
}
