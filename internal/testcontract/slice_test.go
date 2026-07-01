package testcontract

import (
	"slices"
	"testing"
)

func TestAssertStringSliceReturnsCopy(t *testing.T) {
	t.Parallel()

	values := []string{"latn", "arab"}
	AssertStringSliceReturnsCopy(t, "values", func() []string {
		return slices.Clone(values)
	})
}

func TestAssertOptionalStringSliceReturnsCopyEmpty(t *testing.T) {
	t.Parallel()

	calls := 0
	AssertOptionalStringSliceReturnsCopy(t, "values", func() []string {
		calls++
		return nil
	})
	if calls != 1 {
		t.Fatalf("values calls = %d, want 1", calls)
	}
}

func TestAssertOptionalStringSliceReturnsCopy(t *testing.T) {
	t.Parallel()

	values := []string{"gregory"}
	AssertOptionalStringSliceReturnsCopy(t, "values", func() []string {
		return slices.Clone(values)
	})
}

func TestAssertStringSliceSubset(t *testing.T) {
	t.Parallel()

	AssertStringSliceSubset(t, "values", []string{"en-US", "fr-FR"}, "profile", []string{"en-US", "fr-FR", "zh-Hans-CN"})
}
