package relativetimeformat

import (
	"errors"
	"math"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestRelativeTimeValuesUseECMAScriptNumberSemantics(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Numeric: stringPtr(NumericAlways)})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}

	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "largest safe integer", value: Int(9007199254740991), want: "in 9,007,199,254,740,991 days"},
		{name: "two to the fifty third", value: Int(9007199254740992), want: "in 9,007,199,254,740,992 days"},
		{name: "integer rounds through Number", value: Int(9007199254740993), want: "in 9,007,199,254,740,992 days"},
		{name: "maximum uint64 rounds through Number", value: Uint(math.MaxUint64), want: "in 18,446,744,073,709,552,000 days"},
		{name: "negative zero", value: Float(math.Copysign(0, -1)), want: "0 days ago"},
		{name: "positive zero", value: Float(0), want: "in 0 days"},
		{name: "fraction", value: Float(1.5), want: "in 1.5 days"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := format.Format(tc.value, Day)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("Format() = %q, want %q", got, tc.want)
			}
			parts, err := format.FormatToParts(tc.value, Day)
			if err != nil {
				t.Fatalf("FormatToParts() error = %v", err)
			}
			var joined strings.Builder
			for _, part := range parts {
				joined.WriteString(part.Value)
			}
			if joined.String() != got {
				t.Fatalf("concat(FormatToParts()) = %q, want Format() %q", joined.String(), got)
			}
		})
	}
}

func TestRelativeTimeValuesRejectNonFiniteNumbers(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{})
	if err != nil {
		t.Fatalf("New(en-US) error = %v", err)
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		_, err := format.Format(Float(value), Day)
		if !errors.Is(err, intlerr.ErrInvalidValue) {
			t.Fatalf("Format(Float(%v)) error = %v, want ErrInvalidValue", value, err)
		}
	}
}
