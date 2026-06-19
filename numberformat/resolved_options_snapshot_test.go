package numberformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestNumberFormatResolvedOptionsPointerSnapshot(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{intltest.Locale(t, "en")}, Options{
		MinimumFractionDigits: intPtr(2),
		MaximumFractionDigits: intPtr(2),
	})
	if err != nil {
		t.Fatal(err)
	}
	resolved := format.ResolvedOptions()
	if resolved.MinimumFractionDigits == nil || *resolved.MinimumFractionDigits != 2 {
		t.Fatalf("ResolvedOptions().MinimumFractionDigits = %v, want 2", resolved.MinimumFractionDigits)
	}

	*resolved.MinimumFractionDigits = 0

	got := format.ResolvedOptions()
	if got.MinimumFractionDigits == nil || *got.MinimumFractionDigits != 2 {
		t.Fatalf("ResolvedOptions().MinimumFractionDigits after caller mutation = %v, want 2", got.MinimumFractionDigits)
	}
}
