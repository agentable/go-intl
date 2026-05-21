package datetimeformat

import (
	"strings"
	"time"

	ecma402dtf "github.com/agentable/go-intl/internal/ecma402/datetimeformat"
	"github.com/agentable/go-intl/internal/pattern"
)

func (f *DateTimeFormat) FormatRange(start, end time.Time) string {
	r := f.normalizeRange(start, end)
	if r.equal {
		return f.Format(start)
	}
	if parts, ok := f.formatIntervalRangeToParts(r.start, r.end); ok {
		return joinRangeParts(parts)
	}
	return joinRangeParts(f.fallbackRangeParts(f.FormatToParts(r.start), f.FormatToParts(r.end)))
}

func (f *DateTimeFormat) FormatRangeToParts(start, end time.Time) []RangePart {
	r := f.normalizeRange(start, end)
	if r.equal {
		return rangeParts(f.FormatToParts(start), SourceShared)
	}
	if parts, ok := f.formatIntervalRangeToParts(r.start, r.end); ok {
		return parts
	}
	return f.fallbackRangeParts(f.FormatToParts(r.start), f.FormatToParts(r.end))
}

type normalizedRange struct {
	start time.Time
	end   time.Time
	equal bool
}

func (f *DateTimeFormat) normalizeRange(start, end time.Time) normalizedRange {
	if start.Equal(end) {
		return normalizedRange{start: start, end: end, equal: true}
	}
	start = start.Round(0)
	end = end.Round(0)
	if f.location != nil {
		start = start.In(f.location)
		end = end.In(f.location)
	}
	return normalizedRange{start: start, end: end}
}

func (f *DateTimeFormat) formatIntervalRangeToParts(start, end time.Time) ([]RangePart, bool) {
	switch f.pattern.kind {
	case patternDate:
		pattern, ok := f.dateIntervalPattern(start, end)
		if !ok {
			return nil, false
		}
		return f.formatIntervalPattern(pattern, start, end), true
	case patternTime:
		pattern, ok := f.timeIntervalPattern(start, end)
		if !ok {
			return nil, false
		}
		return f.formatIntervalPattern(pattern, start, end), true
	case patternDateTime:
		if !sameDate(start, end) {
			return nil, false
		}
		pattern, ok := f.timeIntervalPattern(start, end)
		if !ok {
			return nil, false
		}
		dateParts := rangeParts(f.formatDatePattern(f.pattern.date, start), SourceShared)
		timeParts := f.formatIntervalPattern(pattern, start, end)
		return interpolateDateTimeRangeParts(f.pattern.dateTime, dateParts, timeParts), true
	case patternNone:
	}
	return nil, false
}

func (f *DateTimeFormat) dateIntervalPattern(start, end time.Time) (string, bool) {
	diffFields, ok := dateRangeDiffFields(start, end)
	if !ok {
		return "", false
	}
	return f.intervalPatternForSkeleton(f.pattern.dateSkeleton, diffFields, f.pattern.dateIntervalOptions)
}

func dateRangeDiffFields(start, end time.Time) ([]rune, bool) {
	switch {
	case start.Year() != end.Year():
		return []rune{'y'}, true
	case start.Month() != end.Month():
		return []rune{'M', 'L'}, true
	case start.Day() != end.Day():
		return []rune{'d'}, true
	default:
		return nil, false
	}
}

func (f *DateTimeFormat) timeIntervalPattern(start, end time.Time) (string, bool) {
	diffFields, ok := f.timeRangeDiffFields(start, end)
	if !ok {
		return "", false
	}
	return f.intervalPatternForSkeleton(f.pattern.timeSkeleton, diffFields, f.pattern.timeIntervalOptions)
}

func (f *DateTimeFormat) timeRangeDiffFields(start, end time.Time) ([]rune, bool) {
	if f.resolved.DayPeriod != "" && cldrDayPeriod(f, start) != cldrDayPeriod(f, end) {
		return []rune{'B', 'b', 'a', f.hourIntervalField(), 'h', 'H'}, true
	}
	if start.Hour() != end.Hour() {
		return []rune{f.hourIntervalField(), 'h', 'H', 'K', 'k'}, true
	}
	if start.Minute() != end.Minute() {
		return []rune{'m'}, true
	}
	if start.Second() != end.Second() {
		return []rune{'s'}, true
	}
	if f.resolved.FractionalSecondDigits != 0 && f.fractionalSecondValue(start.Nanosecond(), f.resolved.FractionalSecondDigits) != f.fractionalSecondValue(end.Nanosecond(), f.resolved.FractionalSecondDigits) {
		return []rune{'S', 's'}, true
	}
	return nil, false
}

func cldrDayPeriod(f *DateTimeFormat, t time.Time) string {
	return f.flexibleDayPeriodPatternName(4, t)
}

func (f *DateTimeFormat) hourIntervalField() rune {
	for _, r := range f.pattern.timeSkeleton {
		switch r {
		case 'h', 'H', 'K', 'k':
			return r
		}
	}
	if f.uses24HourTime() {
		return 'H'
	}
	return 'h'
}

func (f *DateTimeFormat) intervalPatternForSkeleton(skeleton string, fields []rune, opts ecma402dtf.Options) (string, bool) {
	if skeleton == "" {
		return "", false
	}
	intervals := f.gregorian.IntervalFormats[skeleton]
	if len(intervals) == 0 {
		intervals = f.gregorian.IntervalFormats[strings.TrimRight(skeleton, "S")]
	}
	for _, field := range fields {
		if pattern := intervals[string(field)]; pattern != "" {
			format := ecma402dtf.Parse(pattern, pattern, nil, "")
			format.Pattern = pattern
			return ecma402dtf.AdjustFieldTypes(format, opts).Pattern, true
		}
	}
	return "", false
}

func sameDate(start, end time.Time) bool {
	return start.Year() == end.Year() && start.Month() == end.Month() && start.Day() == end.Day()
}

func (f *DateTimeFormat) formatIntervalPattern(pattern string, start, end time.Time) []RangePart {
	tokens := tokenizeIntervalPattern(pattern)
	counts := intervalFieldCounts(tokens)
	seen := map[rune]int{}
	parts := make([]RangePart, 0, len(pattern))
	for i, token := range tokens {
		if token.literal != "" {
			parts = appendRangeLiteralPart(parts, token.literal, literalRangeSource(tokens, i, counts))
			continue
		}
		key := intervalFieldKey(token.field)
		seen[key]++
		source := SourceShared
		t := start
		if counts[key] > 1 {
			source = SourceStartRange
			if seen[key] > 1 {
				source = SourceEndRange
				t = end
			}
		}
		part := f.intervalPatternPart(token.field, token.width, t)
		parts = append(parts, RangePart{Type: part.Type, Value: part.Value, Source: source})
	}
	return parts
}

type intervalToken struct {
	literal string
	field   rune
	width   int
}

func tokenizeIntervalPattern(pattern string) []intervalToken {
	tokens := make([]intervalToken, 0, len(pattern))
	for pattern != "" {
		r := rune(pattern[0])
		if r == '\'' {
			literal, rest := consumeQuotedPatternLiteral(pattern)
			tokens = append(tokens, intervalToken{literal: literal})
			pattern = rest
			continue
		}
		if !isDatePatternField(r) && !isTimePatternField(r) {
			tokens = append(tokens, intervalToken{literal: pattern[:1]})
			pattern = pattern[1:]
			continue
		}
		width := 1
		for width < len(pattern) && rune(pattern[width]) == r {
			width++
		}
		tokens = append(tokens, intervalToken{field: r, width: width})
		pattern = pattern[width:]
	}
	return tokens
}

func intervalFieldCounts(tokens []intervalToken) map[rune]int {
	counts := map[rune]int{}
	for _, token := range tokens {
		if token.field != 0 {
			counts[intervalFieldKey(token.field)]++
		}
	}
	return counts
}

func literalRangeSource(tokens []intervalToken, idx int, counts map[rune]int) RangeSource {
	prev, prevOK := adjacentFieldSource(tokens, idx, -1, counts)
	next, nextOK := adjacentFieldSource(tokens, idx, 1, counts)
	switch {
	case prevOK && nextOK:
		if prev == next {
			return prev
		}
		if prev == SourceShared || next == SourceShared {
			return SourceShared
		}
		return SourceShared
	case prevOK:
		return prev
	case nextOK:
		return next
	default:
		return SourceShared
	}
}

func adjacentFieldSource(tokens []intervalToken, idx int, step int, counts map[rune]int) (RangeSource, bool) {
	seen := map[rune]int{}
	for i := range idx {
		if tokens[i].field != 0 {
			seen[intervalFieldKey(tokens[i].field)]++
		}
	}
	for i := idx + step; i >= 0 && i < len(tokens); i += step {
		if tokens[i].field == 0 {
			continue
		}
		key := intervalFieldKey(tokens[i].field)
		if counts[key] <= 1 {
			return SourceShared, true
		}
		if step > 0 {
			seen[key]++
		}
		if seen[key] > 1 {
			return SourceEndRange, true
		}
		return SourceStartRange, true
	}
	return "", false
}

func intervalFieldKey(field rune) rune {
	switch field {
	case 'L':
		return 'M'
	case 'e', 'c':
		return 'E'
	case 'H', 'K', 'k':
		return 'h'
	case 'b', 'B':
		return 'a'
	case 'Z', 'O', 'V', 'X', 'x':
		return 'z'
	default:
		return field
	}
}

func (f *DateTimeFormat) intervalPatternPart(field rune, width int, t time.Time) Part {
	if isDatePatternField(field) {
		return f.datePatternPart(field, width, t)
	}
	if part, ok := f.timePatternPart(field, width, t); ok {
		return part
	}
	return Part{Type: PartLiteral, Value: ""}
}

func appendRangeLiteralPart(parts []RangePart, value string, source RangeSource) []RangePart {
	if value == "" {
		return parts
	}
	if len(parts) > 0 && parts[len(parts)-1].Type == PartLiteral && parts[len(parts)-1].Source == source {
		parts[len(parts)-1].Value += value
		return parts
	}
	return append(parts, RangePart{Type: PartLiteral, Value: value, Source: source})
}

func joinRangeParts(parts []RangePart) string {
	size := 0
	for _, part := range parts {
		size += len(part.Value)
	}
	out := make([]byte, 0, size)
	for _, part := range parts {
		out = append(out, part.Value...)
	}
	return string(out)
}

func (f *DateTimeFormat) fallbackRangeParts(start, end []Part) []RangePart {
	text := f.gregorian.IntervalFallback
	if text == "" {
		text = "{0} – {1}"
	}
	patternParts, err := pattern.Partition(text)
	if err != nil {
		return []RangePart{{Type: PartLiteral, Value: text, Source: SourceShared}}
	}
	parts := make([]RangePart, 0, len(start)+len(end)+1)
	for _, part := range patternParts {
		switch part.Type {
		case "0":
			parts = append(parts, rangeParts(start, SourceStartRange)...)
		case "1":
			parts = append(parts, rangeParts(end, SourceEndRange)...)
		case pattern.Literal:
			parts = appendRangeLiteralPart(parts, part.Value, SourceShared)
		default:
			parts = appendRangeLiteralPart(parts, "{"+part.Type+"}", SourceShared)
		}
	}
	return parts
}

func rangeParts(parts []Part, source RangeSource) []RangePart {
	out := make([]RangePart, len(parts))
	for i, part := range parts {
		out[i] = RangePart{Type: part.Type, Value: part.Value, Source: source}
	}
	return out
}
