package datetimeformat

import "time"

func (f *DateTimeFormat) appendDateTime(dst []byte, t time.Time) []byte {
	switch f.pattern.kind {
	case patternDate:
		return f.appendDatePattern(dst, f.pattern.date, t)
	case patternTime:
		return f.appendTimePattern(dst, f.pattern.time, t)
	case patternDateTime:
		var dateScratch [64]byte
		var timeScratch [64]byte
		date := f.appendDatePattern(dateScratch[:0], f.pattern.date, t)
		time := f.appendTimePattern(timeScratch[:0], f.pattern.time, t)
		return appendDateTimePattern(dst, f.pattern.dateTime, date, time)
	case patternNone:
	}
	return dst
}

func (f *DateTimeFormat) appendDatePattern(dst []byte, pattern string, t time.Time) []byte {
	return f.appendPattern(dst, pattern, t)
}

func (f *DateTimeFormat) appendTimePattern(dst []byte, pattern string, t time.Time) []byte {
	return f.appendPattern(dst, pattern, t)
}

func (f *DateTimeFormat) appendPattern(dst []byte, pattern string, t time.Time) []byte {
	for pattern != "" {
		r := rune(pattern[0])
		if r == '\'' {
			var literal string
			literal, pattern = consumeQuotedPatternLiteral(pattern)
			dst = append(dst, literal...)
			continue
		}
		if !isDatePatternField(r) && !isTimePatternField(r) {
			dst = append(dst, pattern[0])
			pattern = pattern[1:]
			continue
		}
		width := 1
		for width < len(pattern) && rune(pattern[width]) == r {
			width++
		}
		switch {
		case isDatePatternField(r):
			part := f.datePatternPart(r, width, t)
			dst = append(dst, part.Value...)
		default:
			part, ok := f.timePatternPart(r, width, t)
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
			dst = append(dst, literal...)
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
		dst = append(dst, pattern[0])
		pattern = pattern[1:]
	}
	return dst
}
