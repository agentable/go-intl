package durationformat

import (
	"testing"

	"github.com/agentable/go-intl/internal/ecma402"
)

func TestDurationUnitSpecsFollowResolutionOrder(t *testing.T) {
	t.Parallel()

	want := [...]struct {
		index    unitIndex
		unit     string
		styleSet durationUnitStyleSet
	}{
		{yearsIndex, "years", durationDateUnitStyleSet},
		{monthsIndex, "months", durationDateUnitStyleSet},
		{weeksIndex, "weeks", durationDateUnitStyleSet},
		{daysIndex, "days", durationDateUnitStyleSet},
		{hoursIndex, "hours", durationTimeUnitStyleSet},
		{minutesIndex, "minutes", durationTimeUnitStyleSet},
		{secondsIndex, "seconds", durationTimeUnitStyleSet},
		{millisecondsIndex, "milliseconds", durationSubsecondUnitStyleSet},
		{microsecondsIndex, "microseconds", durationSubsecondUnitStyleSet},
		{nanosecondsIndex, "nanoseconds", durationSubsecondUnitStyleSet},
	}
	if got := len(durationUnitSpecs); got != int(unitCount) {
		t.Fatalf("durationUnitSpecs length = %d, unitCount = %d", got, unitCount)
	}
	if got := len(durationUnitSpecs); got != len(want) {
		t.Fatalf("durationUnitSpecs length = %d, want %d", got, len(want))
	}
	for i, want := range want {
		spec := durationUnitSpecs[i]
		if spec.index != want.index || spec.unit != want.unit || spec.styleSet != want.styleSet {
			t.Fatalf("durationUnitSpecs[%d] = {%d %q %d}, want {%d %q %d}", i, spec.index, spec.unit, spec.styleSet, want.index, want.unit, want.styleSet)
		}
	}
}

func TestDurationUnitStyleOptionContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		set          durationUnitStyleSet
		valid        string
		invalid      string
		wantExpected string
	}{
		{
			name:         "date units",
			set:          durationDateUnitStyleSet,
			valid:        string(NarrowUnitStyle),
			invalid:      string(NumericUnitStyle),
			wantExpected: `one of "long", "short", "narrow"`,
		},
		{
			name:         "time units",
			set:          durationTimeUnitStyleSet,
			valid:        string(TwoDigitUnitStyle),
			invalid:      string(fractionalUnitStyle),
			wantExpected: `one of "long", "short", "narrow", "numeric", "2-digit"`,
		},
		{
			name:         "subsecond units",
			set:          durationSubsecondUnitStyleSet,
			valid:        string(NumericUnitStyle),
			invalid:      string(TwoDigitUnitStyle),
			wantExpected: `one of "long", "short", "narrow", "numeric"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if check, ok := ecma402.InvalidStringOption(durationUnitStyleOption("unit", tc.valid, tc.set)); ok {
				t.Fatalf("durationUnitStyleOption(%s valid) invalid = %+v, want valid", tc.name, check)
			}
			check, ok := ecma402.InvalidStringOption(durationUnitStyleOption("unit", tc.invalid, tc.set))
			if !ok {
				t.Fatalf("durationUnitStyleOption(%s invalid) ok = false, want invalid", tc.name)
			}
			if got := check.Expected(); got != tc.wantExpected {
				t.Fatalf("durationUnitStyleOption(%s).Expected() = %q, want %q", tc.name, got, tc.wantExpected)
			}
		})
	}
}
