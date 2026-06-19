package datetimeformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestDateTimeFormatResolvedOptionsPointerSnapshot(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en-US")}, Options{Hour: NumericFieldStyle, Hour12: boolPtr(true)})
	if err != nil {
		t.Fatal(err)
	}
	resolved := format.ResolvedOptions()
	if resolved.Hour12 == nil || !*resolved.Hour12 {
		t.Fatalf("ResolvedOptions().Hour12 = %v, want true", resolved.Hour12)
	}

	*resolved.Hour12 = false

	got := format.ResolvedOptions()
	if got.Hour12 == nil || !*got.Hour12 {
		t.Fatalf("ResolvedOptions().Hour12 after caller mutation = %v, want true", got.Hour12)
	}
}
