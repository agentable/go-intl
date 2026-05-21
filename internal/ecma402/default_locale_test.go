package ecma402_test

import (
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestDefaultLocale(t *testing.T) {
	t.Parallel()
	if got := ecma402.DefaultLocale(); got != "en" {
		t.Fatalf("DefaultLocale() = %q, want en", got)
	}
}

func TestDefaultLocaleOverrideForTest(t *testing.T) {
	restore := ecma402.OverrideDefaultLocaleForTest("fr")
	t.Cleanup(restore)
	if got := ecma402.DefaultLocale(); got != "fr" {
		t.Fatalf("DefaultLocale() = %q, want fr", got)
	}
}
