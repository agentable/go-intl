package relativetimeformat

import (
	"math"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

// TestNumericAutoLiteralKeyCanonicalization pins the ECMA-402 rule that the
// numeric=auto literal key is ToString(value): -0 stringifies to "0" (so it
// matches CLDR's "today"), and trailing fraction zeros are dropped ("1.0" -> "1"
// -> "tomorrow"). Every numeric bridge must derive the same canonical key.
func TestNumericAutoLiteralKeyCanonicalization(t *testing.T) {
	t.Parallel()

	negZero := math.Copysign(0, -1)

	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "float negative zero", value: Float(negZero), want: "today"},
		{name: "float positive zero", value: Float(0), want: "today"},
		{name: "float one", value: Float(1), want: "tomorrow"},
		{name: "float negative one", value: Float(-1), want: "yesterday"},
		{name: "int zero", value: Int(0), want: "today"},
		{name: "int minus one", value: Int(-1), want: "yesterday"},
		{name: "int one", value: Int(1), want: "tomorrow"},
	}

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Numeric: stringPtr(NumericAuto)})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, err := format.Format(tc.value, Day)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("Format(%s, Day) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

// TestNumericAlwaysNegativeZeroTense pins that -0 is past (ECMA-402
// PartitionRelativeTimePattern step "If value is -0 or value < -0"), so with
// numeric=always it renders with the past tense.
func TestNumericAlwaysNegativeZeroTense(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Numeric: stringPtr(NumericAlways)})
	if err != nil {
		t.Fatal(err)
	}
	got, err := format.Format(Float(math.Copysign(0, -1)), Day)
	if err != nil {
		t.Fatal(err)
	}
	if got != "0 days ago" {
		t.Errorf("Format(-0, Day) numeric=always = %q, want %q", got, "0 days ago")
	}
}
