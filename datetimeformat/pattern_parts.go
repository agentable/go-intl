package datetimeformat

import (
	"strconv"
	"strings"
	"unicode/utf8"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
)

func (f *DateTimeFormat) formatPattern(pattern string, t localTime) []Part {
	numberingSystem := f.resolved.NumberingSystem
	uses24Hour := f.uses24Hour
	cldrLoc := f.cldrLoc
	gregorian := &f.gregorian
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
			r, size := utf8.DecodeRuneInString(pattern)
			parts = appendLiteralPart(parts, string(r))
			pattern = pattern[size:]
			continue
		}
		width := 1
		for width < len(pattern) && rune(pattern[width]) == r {
			width++
		}
		if isDatePatternField(r) {
			parts = append(parts, datePatternPart(r, width, t, gregorian, numberingSystem))
		} else if part, ok := f.timePatternPart(r, width, t, numberingSystem, uses24Hour, cldrLoc, gregorian); ok {
			parts = append(parts, part)
		} else {
			parts = trimTrailingLiteralSpace(parts)
		}
		pattern = pattern[width:]
	}
	return parts
}

func datePatternPart(field rune, width int, t localTime, gregorian *cldrdate.Gregorian, numberingSystem string) Part {
	switch field {
	case 'E', 'e', 'c':
		return Part{Type: PartWeekday, Value: weekdayName(gregorian, t.Weekday, width)}
	case 'M', 'L':
		return Part{Type: PartMonth, Value: monthName(gregorian, t.Month, width, numberingSystem)}
	case 'd':
		return Part{Type: PartDay, Value: localizedNumericField(t.Day, width, numberingSystem)}
	case 'y':
		return Part{Type: PartYear, Value: localizedNumericField(t.displayYear(), width, numberingSystem)}
	case 'G':
		return Part{Type: PartEra, Value: eraName(gregorian, t.Era, width)}
	}
	return Part{Type: PartLiteral, Value: strings.Repeat(string(field), width)}
}

func (f *DateTimeFormat) timePatternPart(field rune, width int, t localTime, numberingSystem string, uses24Hour bool, cldrLoc cldrdate.Locale, gregorian *cldrdate.Gregorian) (Part, bool) {
	switch field {
	case 'h', 'H', 'K', 'k':
		if uses24Hour && (field == 'h' || field == 'K') && width < 2 {
			width = 2
		}
		return Part{Type: PartHour, Value: localizedNumericField(hourPatternValue(field, t, uses24Hour), width, numberingSystem)}, true
	case 'm':
		return Part{Type: PartMinute, Value: localizedNumericField(t.Minute, width, numberingSystem)}, true
	case 's':
		return Part{Type: PartSecond, Value: localizedNumericField(t.Second, width, numberingSystem)}, true
	case 'S':
		value := fractionalSecondValue(t.Nanosecond, width)
		return Part{Type: PartFractionalSecond, Value: ecma402.LocalizeDigits(value, numberingSystem)}, true
	case 'a':
		if uses24Hour {
			return Part{}, false
		}
		return Part{Type: PartDayPeriod, Value: dayPeriodPatternName(gregorian, width, t)}, true
	case 'B', 'b':
		return Part{Type: PartDayPeriod, Value: flexibleDayPeriodPatternName(cldrLoc, gregorian, width, t)}, true
	case 'z':
		return Part{Type: PartTimeZoneName, Value: f.timeZonePatternName(width, t.Time)}, true
	case 'v':
		return Part{Type: PartTimeZoneName, Value: f.genericTimeZonePatternName(width, t.Time)}, true
	case 'Z', 'O', 'X', 'x':
		return Part{Type: PartTimeZoneName, Value: f.offsetTimeZonePatternName(t.Time)}, true
	}
	return Part{Type: PartLiteral, Value: strings.Repeat(string(field), width)}, true
}

func resolvedUses24HourTime(resolved ResolvedOptions) bool {
	if resolved.Hour12 != nil {
		return !*resolved.Hour12
	}
	if implied := hourCycleImpliesHour12(HourCycle(ecma402.ResolvedScalarValue(resolved.HourCycle))); implied != nil {
		return !*implied
	}
	return false
}

func hourPatternValue(field rune, t localTime, uses24Hour bool) int {
	if uses24Hour {
		if field == 'k' && t.Hour == 0 {
			return 24
		}
		return t.Hour
	}
	switch field {
	case 'H':
		return t.Hour
	case 'K':
		return t.Hour % 12
	case 'k':
		if t.Hour == 0 {
			return 24
		}
		return t.Hour
	}
	hour := t.Hour % 12
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
