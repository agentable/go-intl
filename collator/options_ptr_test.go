package collator

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	numeric := true
	c, err := New(locale.List{intltest.Locale(t, "en")}, Options{Numeric: &numeric})
	if err != nil {
		t.Fatal(err)
	}
	numeric = false

	if got := c.ResolvedOptions().Numeric; !got {
		t.Fatal("ResolvedOptions().Numeric = false, want true")
	}
}
