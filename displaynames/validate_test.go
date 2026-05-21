package displaynames_test

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/locale"
)

func TestDisplayNames_OfRejectsInvalidShape(t *testing.T) {
	t.Parallel()
	en := locale.MustParse("en")
	tests := []struct {
		name string
		opts displaynames.Options
		code string
	}{
		{"region-too-long", displaynames.Options{Type: displaynames.Region}, "USA"},
		{"region-with-digit", displaynames.Options{Type: displaynames.Region}, "U1"},
		{"region-empty", displaynames.Options{Type: displaynames.Region}, ""},
		{"script-three-letter", displaynames.Options{Type: displaynames.Script}, "lat"},
		{"script-five-letter", displaynames.Options{Type: displaynames.Script}, "latnn"},
		{"script-with-digit", displaynames.Options{Type: displaynames.Script}, "lat1"},
		{"currency-two-letter", displaynames.Options{Type: displaynames.Currency}, "US"},
		{"currency-with-digit", displaynames.Options{Type: displaynames.Currency}, "US1"},
		{"calendar-too-short", displaynames.Options{Type: displaynames.Calendar}, "gr"},
		{"calendar-underscore", displaynames.Options{Type: displaynames.Calendar}, "invalid_calendar"},
		{"dateTimeField-unknown", displaynames.Options{Type: displaynames.DateTimeField}, "century"},
		{"language-empty", displaynames.Options{Type: displaynames.Language}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dn, err := displaynames.New(locale.List{en}, tc.opts)
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			got, ok, err := dn.Of(tc.code)
			if !errors.Is(err, intlerr.ErrInvalidCode) {
				t.Fatalf("Of(%q) error = %v, want intlerr.ErrInvalidCode", tc.code, err)
			}
			if ok || got != "" {
				t.Fatalf("Of(%q) = (%q, %v), want (\"\", false)", tc.code, got, ok)
			}
		})
	}
}
