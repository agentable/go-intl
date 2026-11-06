package datetimeformat

import (
	"strconv"
	"time"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
)

func weekdayName(gregorian *cldrdate.Gregorian, weekday time.Weekday, width int) string {
	if width == 5 {
		return gregorian.Weekdays.Narrow[int(weekday)]
	}
	if width == 4 {
		return gregorian.Weekdays.Wide[int(weekday)]
	}
	return gregorian.Weekdays.Abbr[int(weekday)]
}

func monthName(gregorian *cldrdate.Gregorian, month time.Month, width int, numberingSystem string, standalone bool) string {
	idx := int(month) - 1
	names := gregorian.Months
	if standalone {
		names.Wide, names.Abbr, names.Narrow = names.StandWide, names.StandAbbr, names.StandNarrow
	}
	switch width {
	case 1:
		return localizedNumericField(int(month), width, numberingSystem)
	case 2:
		return localizedNumericField(int(month), width, numberingSystem)
	case 3:
		return names.Abbr[idx]
	case 4:
		return names.Wide[idx]
	default:
		return names.Narrow[idx]
	}
}

func eraName(gregorian *cldrdate.Gregorian, era string, width int) string {
	idx := 1
	if era == "BC" {
		idx = 0
	}
	if width == 5 {
		return gregorian.Eras.Narrow[idx]
	}
	if width == 4 {
		return gregorian.Eras.Wide[idx]
	}
	return gregorian.Eras.Abbr[idx]
}

func localizedNumericField(value int, width int, numberingSystem string) string {
	if width == 2 {
		return ecma402.LocalizeDigits(twoDigit(value%100), numberingSystem)
	}
	return ecma402.LocalizeDigits(strconv.Itoa(value), numberingSystem)
}

func dayPeriodPatternName(gregorian *cldrdate.Gregorian, width int, t localTime) string {
	names := gregorian.DayPeriods.AM
	fallback := "AM"
	if t.Hour >= 12 {
		names = gregorian.DayPeriods.PM
		fallback = "PM"
	}
	if width == 5 && names.Narrow != "" {
		return names.Narrow
	}
	if width == 4 && names.Wide != "" {
		return names.Wide
	}
	if names.Abbr != "" {
		return names.Abbr
	}
	return fallback
}

func flexibleDayPeriodPatternName(period string, gregorian *cldrdate.Gregorian, width int, t localTime) string {
	names := gregorian.DayPeriods.Flex[period]
	if width == 5 && names.Narrow != "" {
		return names.Narrow
	}
	if width == 4 && names.Wide != "" {
		return names.Wide
	}
	if names.Abbr != "" {
		return names.Abbr
	}
	return dayPeriodPatternName(gregorian, width, t)
}

func (f *DateTimeFormat) flexibleDayPeriodValue(t localTime) string {
	// Interval fragments can omit shared fields; precision belongs to the full format.
	if !f.pattern.hasMinute {
		t.Minute = 0
	}
	if !f.pattern.hasSecond {
		t.Second = 0
	}
	if period := cldrdate.DayPeriodFor(f.cldrLoc, t.Hour, t.Minute, t.Second); period != "" {
		return period
	}
	if t.Hour < 12 {
		return "am"
	}
	return "pm"
}
