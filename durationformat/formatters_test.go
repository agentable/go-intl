package durationformat

import (
	"testing"

	"github.com/agentable/go-intl/numberformat"
)

func TestDurationFractionalNumberOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		resolved ResolvedOptions
		wantMin  int
		wantMax  int
	}{
		{name: "default precision", wantMin: 0, wantMax: 9},
		{name: "explicit fractional digits", resolved: ResolvedOptions{FractionalDigits: intPtr(3)}, wantMin: 3, wantMax: 3},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := durationFractionalNumberOptions(numberformat.Options{}, tc.resolved)
			requireIntOption(t, "MinimumFractionDigits", got.MinimumFractionDigits, tc.wantMin)
			requireIntOption(t, "MaximumFractionDigits", got.MaximumFractionDigits, tc.wantMax)
			if got.RoundingMode == nil || *got.RoundingMode != string(numberformat.TruncRoundingMode) {
				t.Fatalf("RoundingMode = %v, want %q", got.RoundingMode, numberformat.TruncRoundingMode)
			}
		})
	}
}

func requireIntOption(t *testing.T, name string, got *int, want int) {
	t.Helper()

	if got == nil || *got != want {
		t.Fatalf("%s = %v, want %d", name, got, want)
	}
}
