package decimal

import "testing"

func TestApplyRoundingPriority(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hasSD    bool
		hasFD    bool
		priority RoundingPriority
		want     RoundingType
	}{
		{name: "auto defaults fraction", priority: PriorityAuto, want: RoundingTypeFractionDigits},
		{name: "auto significant digits", hasSD: true, hasFD: true, priority: PriorityAuto, want: RoundingTypeSignificantDigits},
		{name: "auto fraction digits", hasFD: true, priority: PriorityAuto, want: RoundingTypeFractionDigits},
		{name: "more precision", hasSD: true, hasFD: true, priority: PriorityMorePrecision, want: RoundingTypeMorePrecision},
		{name: "less precision", hasSD: true, hasFD: true, priority: PriorityLessPrecision, want: RoundingTypeLessPrecision},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ApplyRoundingPriority(tc.hasSD, tc.hasFD, tc.priority)
			if got != tc.want {
				t.Fatalf("ApplyRoundingPriority() = %v, want %v", got, tc.want)
			}
		})
	}
}
