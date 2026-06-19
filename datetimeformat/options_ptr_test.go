package datetimeformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func intPtr(v int) *int {
	return &v
}

func boolPtr(v bool) *bool {
	return &v
}

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	hour12 := false
	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{
		Hour:   NumericFieldStyle,
		Hour12: &hour12,
	})
	if err != nil {
		t.Fatal(err)
	}
	hour12 = true

	if got := format.ResolvedOptions().Hour12; got == nil || *got {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want false", got)
	}
}
