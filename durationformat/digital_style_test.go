package durationformat_test

import (
	"testing"

	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
)

// ECMA-402 §18.2.3 InitializeDurationFormat / GetDurationUnitOptions:
// when style="digital" and the user supplies no per-unit overrides, the
// resolved unit options must follow:
//
//   - years / months / weeks / days  → unit style "short",  display "auto"
//   - hours / minutes / seconds      → unit style "numeric", display "always"
//   - milliseconds / microseconds / nanoseconds — numeric, also "always",
//     with fractional rollup engaging when the prior time unit is numeric.
func TestDurationFormatResolvedDigitalDefaults(t *testing.T) {
	t.Parallel()
	df, err := durationformat.New(locale.List{intltest.Locale(t, "en")}, durationformat.Options{Style: durationformat.DigitalStyle})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	got := df.ResolvedOptions()

	dateUnits := []struct {
		name    string
		style   durationformat.UnitStyle
		display durationformat.Display
	}{
		{"years", got.Years, got.YearsDisplay},
		{"months", got.Months, got.MonthsDisplay},
		{"weeks", got.Weeks, got.WeeksDisplay},
		{"days", got.Days, got.DaysDisplay},
	}
	for _, u := range dateUnits {
		if u.style != durationformat.ShortUnitStyle {
			t.Errorf("%s style = %q, want %q", u.name, u.style, durationformat.ShortUnitStyle)
		}
		if u.display != durationformat.AutoDisplay {
			t.Errorf("%s display = %q, want %q", u.name, u.display, durationformat.AutoDisplay)
		}
	}

	timeUnits := []struct {
		name    string
		style   durationformat.UnitStyle
		display durationformat.Display
	}{
		{"hours", got.Hours, got.HoursDisplay},
		{"minutes", got.Minutes, got.MinutesDisplay},
		{"seconds", got.Seconds, got.SecondsDisplay},
	}
	for _, u := range timeUnits {
		// First (hours) is "numeric"; the trackPrevious chain bumps the rest
		// to "2-digit" because the previous resolved style was numeric/2-digit.
		switch u.style {
		case durationformat.NumericUnitStyle, durationformat.TwoDigitUnitStyle:
		case durationformat.LongUnitStyle, durationformat.ShortUnitStyle, durationformat.NarrowUnitStyle:
			t.Errorf("%s style = %q, want numeric or 2-digit", u.name, u.style)
		}
		if u.display != durationformat.AlwaysDisplay {
			t.Errorf("%s display = %q, want %q", u.name, u.display, durationformat.AlwaysDisplay)
		}
	}
}

// User-supplied per-unit overrides must keep the rest of the digital
// defaults intact (display chain unchanged for fields the user did not set).
func TestDurationFormatDigitalRespectsExplicitDisplay(t *testing.T) {
	t.Parallel()
	df, err := durationformat.New(locale.List{intltest.Locale(t, "en")}, durationformat.Options{
		Style:        durationformat.DigitalStyle,
		Years:        durationformat.LongUnitStyle,
		YearsDisplay: durationformat.AlwaysDisplay,
	})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	got := df.ResolvedOptions()
	if got.Years != durationformat.LongUnitStyle {
		t.Errorf("Years = %q, want long", got.Years)
	}
	if got.YearsDisplay != durationformat.AlwaysDisplay {
		t.Errorf("YearsDisplay = %q, want always", got.YearsDisplay)
	}
	if got.HoursDisplay != durationformat.AlwaysDisplay {
		t.Errorf("HoursDisplay should remain always under digital style, got %q", got.HoursDisplay)
	}
}
