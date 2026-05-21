package collator_test

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/locale"
)

func TestSearchUsageRequiresRealTailoring(t *testing.T) {
	t.Parallel()
	_, err := collator.New(locale.List{locale.MustParse("en")}, collator.Options{Usage: collator.SearchUsage})
	if !errors.Is(err, intlerr.ErrUnsupportedOption) {
		t.Fatalf("New(search) error = %v, want intlerr.ErrUnsupportedOption", err)
	}
}

func TestSortUsageKeepsVariantDefault(t *testing.T) {
	t.Parallel()
	c, err := collator.New(locale.List{locale.MustParse("en")}, collator.Options{})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got := c.ResolvedOptions().Sensitivity; got != collator.VariantSensitivity {
		t.Fatalf("sort default Sensitivity = %q, want %q", got, collator.VariantSensitivity)
	}
}
