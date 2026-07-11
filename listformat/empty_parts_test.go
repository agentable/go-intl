package listformat

import (
	"encoding/json"
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

// TestFormatToPartsEmptyIsSlice pins that FormatToParts([]) returns an empty
// slice, not nil, so it marshals to "[]" (matching FormatJS/native), never "null".
func TestFormatToPartsEmptyIsSlice(t *testing.T) {
	t.Parallel()

	f, err := New(locale.List{intltest.Locale(t, "en")}, Options{})
	if err != nil {
		t.Fatal(err)
	}

	parts := f.FormatToParts(nil)
	if parts == nil {
		t.Error("FormatToParts([]) = nil, want empty non-nil slice")
	}
	data, err := json.Marshal(parts)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "[]" {
		t.Errorf("json.Marshal(FormatToParts([])) = %s, want []", data)
	}
}
