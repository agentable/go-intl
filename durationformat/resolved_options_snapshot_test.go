package durationformat

import (
	"testing"

	"github.com/agentable/go-intl/locale"
)

func TestDurationFormatResolvedOptionsPointerSnapshot(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en")}, Options{FractionalDigits: intPtr(3)})
	if err != nil {
		t.Fatal(err)
	}
	resolved := format.ResolvedOptions()
	if resolved.FractionalDigits == nil || *resolved.FractionalDigits != 3 {
		t.Fatalf("ResolvedOptions().FractionalDigits = %v, want 3", resolved.FractionalDigits)
	}

	*resolved.FractionalDigits = 0

	got := format.ResolvedOptions()
	if got.FractionalDigits == nil || *got.FractionalDigits != 3 {
		t.Fatalf("ResolvedOptions().FractionalDigits after caller mutation = %v, want 3", got.FractionalDigits)
	}
}
