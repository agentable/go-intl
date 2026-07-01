package datetimeformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestDateTimeFormatResolvedOptionsPointerSnapshot(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Hour:                   stringPtr(NumericFieldStyle),
		Hour12:                 boolPtr(true),
		FractionalSecondDigits: intPtr(3),
	})
	if err != nil {
		t.Fatal(err)
	}
	resolved := format.ResolvedOptions()
	if resolved.Hour12 == nil || !*resolved.Hour12 {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want true", resolved.Hour12)
	}
	if resolved.FractionalSecondDigits == nil || *resolved.FractionalSecondDigits != 3 {
		t.Fatalf("ResolvedOptions().FractionalSecondDigits = %v, want 3", resolved.FractionalSecondDigits)
	}

	*resolved.Hour12 = false
	*resolved.FractionalSecondDigits = 1

	got := format.ResolvedOptions()
	if got.Hour12 == nil || !*got.Hour12 {
		t.Fatalf("ResolvedOptions().Hour12 after caller mutation = %v, want true", got.Hour12)
	}
	if got.FractionalSecondDigits == nil || *got.FractionalSecondDigits != 3 {
		t.Fatalf("ResolvedOptions().FractionalSecondDigits after caller mutation = %v, want 3", got.FractionalSecondDigits)
	}
}
