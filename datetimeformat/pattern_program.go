package datetimeformat

import (
	"strconv"
	"strings"
	"unicode/utf8"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
)

type patternToken struct {
	literal string
	field   rune
	width   int
}

type patternProgram []patternToken

func compilePatternProgram(pattern string) patternProgram {
	program := make(patternProgram, 0, len(pattern))
	for pattern != "" {
		r := rune(pattern[0])
		if r == '\'' {
			literal, rest := consumeQuotedPatternLiteral(pattern)
			if literal = normalizePatternLiteral(literal); literal != "" {
				program = append(program, patternToken{literal: literal})
			}
			pattern = rest
			continue
		}
		if !isDatePatternField(r) && !isTimePatternField(r) {
			r, size := utf8.DecodeRuneInString(pattern)
			program = append(program, patternToken{literal: normalizePatternLiteral(string(r))})
			pattern = pattern[size:]
			continue
		}
		width := 1
		for width < len(pattern) && rune(pattern[width]) == r {
			width++
		}
		program = append(program, patternToken{field: r, width: width})
		pattern = pattern[width:]
	}
	return program
}

func (p *selectedPattern) compilePrograms() {
	p.dateProgram = compilePatternProgram(p.date)
	p.timeProgram = compilePatternProgram(p.time)
	p.dateTimeProgram = compileDateTimeProgram(p.dateTime)
}

func (p selectedPattern) parts(f *DateTimeFormat, t localTime) []Part {
	switch p.kind {
	case patternDate:
		return p.dateProgram.parts(f, t)
	case patternTime:
		return p.timeProgram.parts(f, t)
	case patternDateTime:
		dateParts := p.dateProgram.parts(f, t)
		timeParts := p.timeProgram.parts(f, t)
		return interpolateDateTimeParts(p.dateTimeProgram, dateParts, timeParts)
	case patternNone:
	}
	return nil
}

func (p selectedPattern) appendTo(f *DateTimeFormat, dst []byte, t localTime) []byte {
	switch p.kind {
	case patternDate:
		return p.dateProgram.appendTo(f, dst, t)
	case patternTime:
		return p.timeProgram.appendTo(f, dst, t)
	case patternDateTime:
		var dateScratch [64]byte
		var timeScratch [64]byte
		date := p.dateProgram.appendTo(f, dateScratch[:0], t)
		time := p.timeProgram.appendTo(f, timeScratch[:0], t)
		return appendDateTimeProgram(dst, p.dateTimeProgram, date, time)
	case patternNone:
	}
	return dst
}

func (p patternProgram) parts(f *DateTimeFormat, t localTime) []Part {
	parts := make([]Part, 0, len(p))
	for _, token := range p {
		if token.literal != "" {
			parts = appendLiteralPart(parts, token.literal)
			continue
		}
		part, ok := f.patternTokenPart(token, t)
		if ok {
			parts = append(parts, part)
		} else {
			parts = trimTrailingLiteralSpace(parts)
		}
	}
	return parts
}

func (p patternProgram) appendTo(f *DateTimeFormat, dst []byte, t localTime) []byte {
	for _, token := range p {
		if token.literal != "" {
			dst = append(dst, token.literal...)
			continue
		}
		part, ok := f.patternTokenPart(token, t)
		if ok {
			dst = append(dst, part.Value...)
		} else {
			dst = trimTrailingSpaceBytes(dst)
		}
	}
	return dst
}

func appendDateTimeProgram(dst []byte, program ecma402.Pattern, date, time []byte) []byte {
	for _, part := range program {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			dst = append(dst, time...)
		case ecma402.PatternPartPlaceholder1:
			dst = append(dst, date...)
		case ecma402.PatternPartLiteral:
			dst = append(dst, part.Value...)
		default:
			dst = append(dst, "{"+part.Type+"}"...)
		}
	}
	return dst
}

func (f *DateTimeFormat) patternTokenPart(token patternToken, t localTime) (Part, bool) {
	numberingSystem := f.resolved.NumberingSystem
	if isDatePatternField(token.field) {
		return datePatternPart(token.field, token.width, t, &f.gregorian, numberingSystem), true
	}
	return f.timePatternPart(token.field, token.width, t, numberingSystem, f.uses24Hour, f.cldrLoc, &f.gregorian)
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
