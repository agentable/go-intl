package datetimeformat

import "unicode/utf8"

func (p selectedPattern) appendTo(f *DateTimeFormat, dst []byte, t localTime) []byte {
	switch p.kind {
	case patternDate:
		return f.appendPattern(dst, p.date, t)
	case patternTime:
		return f.appendPattern(dst, p.time, t)
	case patternDateTime:
		var dateScratch [64]byte
		var timeScratch [64]byte
		date := f.appendPattern(dateScratch[:0], p.date, t)
		time := f.appendPattern(timeScratch[:0], p.time, t)
		return appendDateTimePattern(dst, p.dateTime, date, time)
	case patternNone:
	}
	return dst
}

func (f *DateTimeFormat) appendPattern(dst []byte, pattern string, t localTime) []byte {
	numberingSystem := f.resolved.NumberingSystem
	uses24Hour := f.uses24Hour
	cldrLoc := f.cldrLoc
	gregorian := &f.gregorian
	for pattern != "" {
		r := rune(pattern[0])
		if r == '\'' {
			var literal string
			literal, pattern = consumeQuotedPatternLiteral(pattern)
			dst = append(dst, normalizePatternLiteral(literal)...)
			continue
		}
		if !isDatePatternField(r) && !isTimePatternField(r) {
			r, size := utf8.DecodeRuneInString(pattern)
			dst = append(dst, normalizePatternLiteral(string(r))...)
			pattern = pattern[size:]
			continue
		}
		width := 1
		for width < len(pattern) && rune(pattern[width]) == r {
			width++
		}
		switch {
		case isDatePatternField(r):
			part := datePatternPart(r, width, t, gregorian, numberingSystem)
			dst = append(dst, part.Value...)
		default:
			part, ok := f.timePatternPart(r, width, t, numberingSystem, uses24Hour, cldrLoc, gregorian)
			if ok {
				dst = append(dst, part.Value...)
			} else {
				dst = trimTrailingSpaceBytes(dst)
			}
		}
		pattern = pattern[width:]
	}
	return dst
}

func appendDateTimePattern(dst []byte, pattern string, date, time []byte) []byte {
	for pattern != "" {
		if pattern[0] == '\'' {
			var literal string
			literal, pattern = consumeQuotedPatternLiteral(pattern)
			dst = append(dst, normalizePatternLiteral(literal)...)
			continue
		}
		if len(pattern) >= 3 && pattern[0] == '{' && pattern[2] == '}' {
			switch pattern[1] {
			case '0':
				dst = append(dst, time...)
			case '1':
				dst = append(dst, date...)
			default:
				dst = append(dst, pattern[:3]...)
			}
			pattern = pattern[3:]
			continue
		}
		r, size := utf8.DecodeRuneInString(pattern)
		dst = append(dst, normalizePatternLiteral(string(r))...)
		pattern = pattern[size:]
	}
	return dst
}
