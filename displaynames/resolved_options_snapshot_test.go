package displaynames_test

import (
	"testing"

	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestDisplayNamesResolvedOptionsPointerSnapshot(t *testing.T) {
	t.Parallel()

	dn, err := displaynames.New(locale.List{intltest.Locale(t, "en")}, displaynames.Options{Type: stringPtr(displaynames.Language), LanguageDisplay: stringPtr(displaynames.StandardLanguageDisplay)})
	if err != nil {
		t.Fatal(err)
	}
	resolved := dn.ResolvedOptions()
	if resolved.LanguageDisplay == nil || *resolved.LanguageDisplay != displaynames.StandardLanguageDisplay {
		t.Fatalf("ResolvedOptions().LanguageDisplay = %v, want %q", resolved.LanguageDisplay, displaynames.StandardLanguageDisplay)
	}

	*resolved.LanguageDisplay = displaynames.DialectLanguageDisplay

	got := dn.ResolvedOptions()
	if got.LanguageDisplay == nil || *got.LanguageDisplay != displaynames.StandardLanguageDisplay {
		t.Fatalf("ResolvedOptions().LanguageDisplay after caller mutation = %v, want %q", got.LanguageDisplay, displaynames.StandardLanguageDisplay)
	}
}
