package collator

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

func TestOptionsPointerValuesCopiedDuringConstruction(t *testing.T) {
	t.Parallel()

	usage := string(SortUsage)
	sensitivity := string(AccentSensitivity)
	caseFirst := string(FalseCaseFirst)
	collation := "phonebk"
	numeric := true
	c, err := New(locale.List{intltest.Locale(t, "de")}, Options{
		Usage:       &usage,
		Sensitivity: &sensitivity,
		CaseFirst:   &caseFirst,
		Collation:   &collation,
		Numeric:     &numeric,
	})
	if err != nil {
		t.Fatal(err)
	}
	usage = "search"
	sensitivity = "base"
	caseFirst = "upper"
	collation = "default"
	numeric = false

	got := c.ResolvedOptions()
	if got.Usage != SortUsage ||
		got.Sensitivity != AccentSensitivity ||
		got.CaseFirst != FalseCaseFirst ||
		got.Collation != "phonebk" ||
		!got.Numeric {
		t.Fatalf("ResolvedOptions() = %+v, want copied usage/sensitivity/caseFirst/collation/numeric values", got)
	}
}
