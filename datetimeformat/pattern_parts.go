package datetimeformat

import (
	"strconv"
	"strings"
	"time"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
)

func (f *DateTimeFormat) formatDatePattern(pattern string, t time.Time) []Part {
	return f.formatPattern(pattern, t)
}

func (f *DateTimeFormat) formatTimePattern(pattern string, t time.Time) []Part {
	return f.formatPattern(pattern, t)
}

func (f *DateTimeFormat) formatPattern(pattern string, t time.Time) []Part {
	parts := make([]Part, 0, len(pattern))
	for pattern != "" {
		r := rune(pattern[0])
		if r == '\'' {
			literal, rest := consumeQuotedPatternLiteral(pattern)
			parts = appendLiteralPart(parts, literal)
			pattern = rest
			continue
		}
		if !isDatePatternField(r) && !isTimePatternField(r) {
			parts = appendLiteralPart(parts, pattern[:1])
			pattern = pattern[1:]
			continue
		}
		width := 1
		for width < len(pattern) && rune(pattern[width]) == r {
			width++
		}
		if isDatePatternField(r) {
			parts = append(parts, f.datePatternPart(r, width, t))
		} else if part, ok := f.timePatternPart(r, width, t); ok {
			parts = append(parts, part)
		} else {
			parts = trimTrailingLiteralSpace(parts)
		}
		pattern = pattern[width:]
	}
	return parts
}

func (f *DateTimeFormat) datePatternPart(field rune, width int, t time.Time) Part {
	switch field {
	case 'E', 'e', 'c':
		return Part{Type: PartWeekday, Value: f.weekdayName(t.Weekday(), width)}
	case 'M', 'L':
		return Part{Type: PartMonth, Value: f.monthName(t.Month(), width)}
	case 'd':
		return Part{Type: PartDay, Value: f.numericDateValue(t.Day(), width)}
	case 'y':
		return Part{Type: PartYear, Value: f.numericDateValue(t.Year(), width)}
	case 'G':
		return Part{Type: PartEra, Value: f.eraName(t.Year(), width)}
	}
	return Part{Type: PartLiteral, Value: strings.Repeat(string(field), width)}
}

func (f *DateTimeFormat) timePatternPart(field rune, width int, t time.Time) (Part, bool) {
	switch field {
	case 'h', 'H', 'K', 'k':
		if f.uses24HourTime() && (field == 'h' || field == 'K') && width < 2 {
			width = 2
		}
		return Part{Type: PartHour, Value: f.numericDateValue(f.hourPatternValue(field, t), width)}, true
	case 'm':
		return Part{Type: PartMinute, Value: f.numericDateValue(t.Minute(), width)}, true
	case 's':
		return Part{Type: PartSecond, Value: f.numericDateValue(t.Second(), width)}, true
	case 'S':
		return Part{Type: PartFractionalSecondDigits, Value: f.fractionalSecondValue(t.Nanosecond(), width)}, true
	case 'a':
		if f.uses24HourTime() {
			return Part{}, false
		}
		return Part{Type: PartDayPeriod, Value: f.dayPeriodPatternName(width, t)}, true
	case 'B', 'b':
		return Part{Type: PartDayPeriod, Value: f.flexibleDayPeriodPatternName(width, t)}, true
	case 'z':
		return Part{Type: PartTimeZoneName, Value: f.timeZonePatternName(width, t)}, true
	case 'v':
		return Part{Type: PartTimeZoneName, Value: f.genericTimeZonePatternName(width, t)}, true
	case 'Z', 'O', 'X', 'x':
		return Part{Type: PartTimeZoneName, Value: f.offsetTimeZonePatternName(width, t)}, true
	}
	return Part{Type: PartLiteral, Value: strings.Repeat(string(field), width)}, true
}

func (f *DateTimeFormat) uses24HourTime() bool {
	if f.resolved.Hour12 != nil {
		return !*f.resolved.Hour12
	}
	switch f.resolved.HourCycle {
	case H23HourCycle, H24HourCycle:
		return true
	case H11HourCycle, H12HourCycle:
		return false
	}
	return false
}

func (f *DateTimeFormat) hourPatternValue(field rune, t time.Time) int {
	if f.uses24HourTime() {
		if field == 'k' && t.Hour() == 0 {
			return 24
		}
		return t.Hour()
	}
	switch field {
	case 'H':
		return t.Hour()
	case 'K':
		return t.Hour() % 12
	case 'k':
		if t.Hour() == 0 {
			return 24
		}
		return t.Hour()
	}
	hour := t.Hour() % 12
	if hour == 0 {
		return 12
	}
	return hour
}

func fractionalSecondValue(nanosecond, width int) string {
	if width > 9 {
		width = 9
	}
	divisor := 1_000_000_000
	for range width {
		divisor /= 10
	}
	value := nanosecond / divisor
	out := strconv.Itoa(value)
	for len(out) < width {
		out = "0" + out
	}
	return out
}

func (f *DateTimeFormat) fractionalSecondValue(nanosecond, width int) string {
	return f.localizeDigits(fractionalSecondValue(nanosecond, width))
}

func (f *DateTimeFormat) localizeDigits(s string) string {
	return ecma402.LocalizeDigits(s, f.resolved.NumberingSystem)
}

func (f *DateTimeFormat) flexibleDayPeriodPatternName(width int, t time.Time) string {
	period := cldrdate.DayPeriodFor(f.cldrLoc, t.Hour(), t.Minute())
	names := f.gregorian.DayPeriods.Flex[period]
	if width == 5 && names.Narrow != "" {
		return names.Narrow
	}
	if width == 4 && names.Wide != "" {
		return names.Wide
	}
	if names.Abbr != "" {
		return names.Abbr
	}
	return f.dayPeriodPatternName(width, t)
}
