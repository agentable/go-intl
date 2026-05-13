package ecma402_test

import (
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestToIntlMathematicalValue(t *testing.T) {
	t.Parallel()

	got, err := ecma402.ToIntlMathematicalValue("-42")
	if err != nil {
		t.Fatalf("ToIntlMathematicalValue err = %v", err)
	}
	if got.IsNaN() {
		t.Fatal("IsNaN() = true, want false")
	}
	if got.IsInfinity() {
		t.Fatal("IsInfinity() = true, want false")
	}
	if !got.IsNegative() {
		t.Fatal("IsNegative() = false, want true")
	}
	if got.Sign() != -1 {
		t.Fatalf("Sign() = %d, want -1", got.Sign())
	}
}
