package ecma402

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/decimal"
)

func TestParseDecimalInputAllowsSpecialValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"NaN", "Infinity", "-Infinity", "1.25"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			if _, err := ParseDecimalInput(value); err != nil {
				t.Fatalf("ParseDecimalInput(%q) error = %v, want nil", value, err)
			}
		})
	}
}

func TestParseFiniteDecimalInputRejectsMalformedAndNonFinite(t *testing.T) {
	t.Parallel()

	for _, value := range []string{"bad", "NaN", "Infinity", "-Infinity"} {
		t.Run(value, func(t *testing.T) {
			t.Parallel()

			_, err := ParseFiniteDecimalInput(value)
			if !errors.Is(err, decimal.ErrInvalidDecimal) {
				t.Fatalf("ParseFiniteDecimalInput(%q) error = %v, want ErrInvalidDecimal", value, err)
			}
		})
	}
}

func TestRequireFiniteDecimalInput(t *testing.T) {
	t.Parallel()

	if err := RequireFiniteDecimalInput(decimal.FromInt64(1)); err != nil {
		t.Fatalf("RequireFiniteDecimalInput(finite) error = %v, want nil", err)
	}
	if err := RequireFiniteDecimalInput(decimal.NaNValue); !errors.Is(err, decimal.ErrInvalidDecimal) {
		t.Fatalf("RequireFiniteDecimalInput(NaN) error = %v, want ErrInvalidDecimal", err)
	}
}
