package datetimeformat

import (
	"strconv"
	"time"
)

func (f *DateTimeFormat) weekdayName(weekday time.Weekday, width int) string {
	if width == 5 {
		return f.gregorian.Weekdays.Narrow[int(weekday)]
	}
	if width == 4 {
		return f.gregorian.Weekdays.Wide[int(weekday)]
	}
	return f.gregorian.Weekdays.Abbr[int(weekday)]
}

func (f *DateTimeFormat) monthName(month time.Month, width int) string {
	idx := int(month) - 1
	switch width {
	case 1:
		return f.localizeDigits(strconv.Itoa(int(month)))
	case 2:
		return f.localizeDigits(twoDigit(int(month)))
	case 3:
		return f.gregorian.Months.Abbr[idx]
	case 4:
		return f.gregorian.Months.Wide[idx]
	default:
		return f.gregorian.Months.Narrow[idx]
	}
}

func (f *DateTimeFormat) eraName(year int, width int) string {
	idx := 1
	if year <= 0 {
		idx = 0
	}
	if width == 5 {
		return f.gregorian.Eras.Narrow[idx]
	}
	if width == 4 {
		return f.gregorian.Eras.Wide[idx]
	}
	return f.gregorian.Eras.Abbr[idx]
}

func (f *DateTimeFormat) numericDateValue(value int, width int) string {
	if width == 2 {
		return f.localizeDigits(twoDigit(value % 100))
	}
	return f.localizeDigits(strconv.Itoa(value))
}

func (f *DateTimeFormat) dayPeriodPatternName(width int, t time.Time) string {
	names := f.gregorian.DayPeriods.AM
	fallback := "AM"
	if t.Hour() >= 12 {
		names = f.gregorian.DayPeriods.PM
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
