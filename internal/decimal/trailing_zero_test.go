package decimal

import "testing"

func TestApplyTrailingZeroDisplay(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		formatted string
		integer   bool
		display   TrailingZeroDisplay
		want      string
	}{
		{name: "auto preserves", formatted: "1.00", integer: true, display: TrailingZeroAuto, want: "1.00"},
		{name: "strip integer", formatted: "1.00", integer: true, display: TrailingZeroStripIfInteger, want: "1"},
		{name: "strip plain integer unchanged", formatted: "1", integer: true, display: TrailingZeroStripIfInteger, want: "1"},
		{name: "fraction preserves", formatted: "1.20", integer: false, display: TrailingZeroStripIfInteger, want: "1.20"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ApplyTrailingZeroDisplay(tc.formatted, tc.integer, tc.display)
			if got != tc.want {
				t.Fatalf("ApplyTrailingZeroDisplay() = %q, want %q", got, tc.want)
			}
		})
	}
}
