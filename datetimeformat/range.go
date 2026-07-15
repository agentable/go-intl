package datetimeformat

import (
	"strings"
	"time"
	"unicode/utf8"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
	ecma402dtf "github.com/agentable/go-intl/internal/ecma402/datetimeformat"
)

func (f *DateTimeFormat) FormatRange(start, end time.Time) (string, error) {
	r, err := f.normalizeRange(start, end)
	if err != nil {
		return "", err
	}
	if r.equal {
		return f.Format(r.start), nil
	}
	pattern := f.pattern
	return string(f.appendRange(nil, pattern, r)), nil
}

func (f *DateTimeFormat) FormatRangeToParts(start, end time.Time) ([]RangePart, error) {
	r, err := f.normalizeRange(start, end)
	if err != nil {
		return nil, err
	}
	if r.equal {
		return rangeParts(f.FormatToParts(r.start), SourceShared), nil
	}
	pattern := f.pattern
	return f.formatRangeParts(pattern, r), nil
}

func (f *DateTimeFormat) formatRangeParts(pattern selectedPattern, r normalizedRange) []RangePart {
	if parts, ok := f.formatIntervalRangeToParts(pattern, r.startLocal, r.endLocal); ok {
		return parts
	}
	return f.fallbackRangeParts(
		pattern.parts(f, r.startLocal),
		pattern.parts(f, r.endLocal),
	)
}

func (f *DateTimeFormat) appendRange(dst []byte, pattern selectedPattern, r normalizedRange) []byte {
	if out, ok := f.appendIntervalRange(dst, pattern, r.startLocal, r.endLocal); ok {
		return out
	}
	var startScratch [64]byte
	var endScratch [64]byte
	start := pattern.appendTo(f, startScratch[:0], r.startLocal)
	end := pattern.appendTo(f, endScratch[:0], r.endLocal)
	return f.appendFallbackRange(dst, start, end)
}

type normalizedRange struct {
	start      time.Time
	startLocal localTime
	endLocal   localTime
	equal      bool
}

func (f *DateTimeFormat) normalizeRange(start, end time.Time) (normalizedRange, error) {
	location := f.location
	start = start.Round(0)
	end = end.Round(0)
	if start.Equal(end) {
		return normalizedRange{start: start, equal: true}, nil
	}
	start, startLocal := gregoryTimeInLocation(start, location)
	_, endLocal := gregoryTimeInLocation(end, location)
	return normalizedRange{
		start:      start,
		startLocal: startLocal,
		endLocal:   endLocal,
	}, nil
}

func (f *DateTimeFormat) formatIntervalRangeToParts(pattern selectedPattern, start, end localTime) ([]RangePart, bool) {
	switch pattern.kind {
	case patternDate:
		rangePattern, ok := f.dateIntervalPattern(pattern, start, end)
		if !ok {
			return nil, false
		}
		return f.formatIntervalPattern(rangePattern, start, end), true
	case patternTime:
		rangePattern, ok := f.timeIntervalPattern(pattern, start, end)
		if !ok {
			return nil, false
		}
		return f.formatIntervalPattern(rangePattern, start, end), true
	case patternDateTime:
		if !sameDate(start, end) {
			return nil, false
		}
		rangePattern, ok := f.timeIntervalPattern(pattern, start, end)
		if !ok {
			return nil, false
		}
		dateParts := rangeParts(f.formatPattern(pattern.date, start), SourceShared)
		timeParts := f.formatIntervalPattern(rangePattern, start, end)
		return interpolateDateTimeRangeParts(pattern.dateTime, dateParts, timeParts), true
	case patternNone:
	}
	return nil, false
}

func (f *DateTimeFormat) appendIntervalRange(dst []byte, pattern selectedPattern, start, end localTime) ([]byte, bool) {
	switch pattern.kind {
	case patternDate:
		rangePattern, ok := f.dateIntervalPattern(pattern, start, end)
		if !ok {
			return nil, false
		}
		return f.appendIntervalPattern(dst, rangePattern, start, end), true
	case patternTime:
		rangePattern, ok := f.timeIntervalPattern(pattern, start, end)
		if !ok {
			return nil, false
		}
		return f.appendIntervalPattern(dst, rangePattern, start, end), true
	case patternDateTime:
		if !sameDate(start, end) {
			return nil, false
		}
		rangePattern, ok := f.timeIntervalPattern(pattern, start, end)
		if !ok {
			return nil, false
		}
		var dateScratch [64]byte
		var timeScratch [64]byte
		date := f.appendPattern(dateScratch[:0], pattern.date, start)
		time := f.appendIntervalPattern(timeScratch[:0], rangePattern, start, end)
		return appendDateTimePattern(dst, pattern.dateTime, date, time), true
	case patternNone:
	}
	return nil, false
}

func (f *DateTimeFormat) dateIntervalPattern(pattern selectedPattern, start, end localTime) (string, bool) {
	diffFields, ok := dateRangeDiffFields(start, end)
	if !ok {
		return "", false
	}
	intervalFormats := f.gregorian.IntervalFormats
	return intervalPatternForSkeleton(intervalFormats, pattern.dateSkeleton, diffFields, pattern.dateIntervalOptions)
}

func dateRangeDiffFields(start, end localTime) ([]rune, bool) {
	switch {
	case start.Year != end.Year:
		return []rune{'y'}, true
	case start.Month != end.Month:
		return []rune{'M', 'L'}, true
	case start.Day != end.Day:
		return []rune{'d'}, true
	default:
		return nil, false
	}
}

func (f *DateTimeFormat) timeIntervalPattern(pattern selectedPattern, start, end localTime) (string, bool) {
	resolved := f.resolved
	uses24Hour := f.uses24Hour
	cldrLoc := f.cldrLoc
	gregorian := &f.gregorian
	intervalFormats := gregorian.IntervalFormats
	diffFields, ok := timeRangeDiffFields(pattern, start, end, resolved, uses24Hour, cldrLoc, gregorian)
	if !ok {
		return "", false
	}
	return intervalPatternForSkeleton(intervalFormats, pattern.timeSkeleton, diffFields, pattern.timeIntervalOptions)
}

func timeRangeDiffFields(pattern selectedPattern, start, end localTime, resolved ResolvedOptions, uses24Hour bool, cldrLoc cldrdate.Locale, gregorian *cldrdate.Gregorian) ([]rune, bool) {
	dayPeriodStyle := ecma402.ResolvedScalarValue(resolved.DayPeriod)
	hourField := hourIntervalField(pattern.timeSkeleton, uses24Hour)
	if dayPeriodStyle != "" && flexibleDayPeriodPatternName(cldrLoc, gregorian, 4, start) != flexibleDayPeriodPatternName(cldrLoc, gregorian, 4, end) {
		return []rune{'B', 'b', 'a', hourField, 'h', 'H'}, true
	}
	if start.Hour != end.Hour {
		return []rune{hourField, 'h', 'H', 'K', 'k'}, true
	}
	if start.Minute != end.Minute {
		return []rune{'m'}, true
	}
	if start.Second != end.Second {
		return []rune{'s'}, true
	}
	fractionalSecondDigits := ecma402.ResolvedScalarValue(resolved.FractionalSecondDigits)
	if fractionalSecondDigits != 0 && fractionalSecondValue(start.Nanosecond, fractionalSecondDigits) != fractionalSecondValue(end.Nanosecond, fractionalSecondDigits) {
		return []rune{'S', 's'}, true
	}
	return nil, false
}

func hourIntervalField(timeSkeleton string, uses24Hour bool) rune {
	for _, r := range timeSkeleton {
		switch r {
		case 'h', 'H', 'K', 'k':
			return r
		}
	}
	if uses24Hour {
		return 'H'
	}
	return 'h'
}

func intervalPatternForSkeleton(intervalFormats map[string]map[string]string, skeleton string, fields []rune, opts ecma402dtf.Options) (string, bool) {
	if skeleton == "" {
		return "", false
	}
	intervals := intervalFormats[skeleton]
	if len(intervals) == 0 {
		intervals = intervalFormats[strings.TrimRight(skeleton, "S")]
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

func sameDate(start, end localTime) bool {
	return start.Year == end.Year && start.Month == end.Month && start.Day == end.Day
}

func (f *DateTimeFormat) formatIntervalPattern(pattern string, start, end localTime) []RangePart {
	numberingSystem := f.resolved.NumberingSystem
	uses24Hour := f.uses24Hour
	cldrLoc := f.cldrLoc
	gregorian := &f.gregorian
	steps := compileIntervalPattern(pattern, start, end)
	parts := make([]RangePart, 0, len(pattern))
	for _, step := range steps {
		token := step.token
		if token.literal != "" {
			parts = appendRangeLiteralPart(parts, token.literal, step.source)
			continue
		}
		part := f.intervalPatternPart(token.field, token.width, step.time(start, end), numberingSystem, uses24Hour, cldrLoc, gregorian)
		parts = append(parts, RangePart{Type: part.Type, Value: part.Value, Source: step.source})
	}
	return parts
}

func (f *DateTimeFormat) appendIntervalPattern(dst []byte, pattern string, start, end localTime) []byte {
	numberingSystem := f.resolved.NumberingSystem
	uses24Hour := f.uses24Hour
	cldrLoc := f.cldrLoc
	gregorian := &f.gregorian
	for _, step := range compileIntervalPattern(pattern, start, end) {
		token := step.token
		if token.literal != "" {
			dst = append(dst, token.literal...)
			continue
		}
		part := f.intervalPatternPart(token.field, token.width, step.time(start, end), numberingSystem, uses24Hour, cldrLoc, gregorian)
		dst = append(dst, part.Value...)
	}
	return dst
}

type intervalToken struct {
	literal string
	field   rune
	width   int
}

type intervalStep struct {
	token    intervalToken
	source   RangeSource
	endpoint intervalEndpoint
}

type intervalEndpoint uint8

const (
	intervalStart intervalEndpoint = iota
	intervalEnd
)

func (s intervalStep) time(start, end localTime) localTime {
	if s.endpoint == intervalEnd {
		return end
	}
	return start
}

func compileIntervalPattern(pattern string, start, end localTime) []intervalStep {
	steps, counts := tokenizeIntervalPattern(pattern)
	seen := map[rune]int{}
	previousSource := SourceShared
	previousOK := false
	for i := range steps {
		token := steps[i].token
		if token.field != 0 {
			key := intervalFieldKey(token.field)
			seen[key]++
			source := intervalOccurrenceSource(counts[key], seen[key])
			steps[i].source = source
			steps[i].endpoint = intervalStart
			if source == SourceEndRange {
				steps[i].endpoint = intervalEnd
			}
			previousSource, previousOK = source, true
			continue
		}

		nextSource, nextOK := nextIntervalFieldSource(steps[i+1:], counts, seen)
		steps[i].source = surroundingIntervalSource(previousSource, previousOK, nextSource, nextOK)
	}
	return steps
}

func tokenizeIntervalPattern(pattern string) ([]intervalStep, map[rune]int) {
	steps := make([]intervalStep, 0, len(pattern))
	counts := map[rune]int{}
	for pattern != "" {
		r := rune(pattern[0])
		if r == '\'' {
			literal, rest := consumeQuotedPatternLiteral(pattern)
			steps = append(steps, intervalStep{token: intervalToken{literal: normalizePatternLiteral(literal)}})
			pattern = rest
			continue
		}
		if !isDatePatternField(r) && !isTimePatternField(r) {
			r, size := utf8.DecodeRuneInString(pattern)
			steps = append(steps, intervalStep{token: intervalToken{literal: normalizePatternLiteral(string(r))}})
			pattern = pattern[size:]
			continue
		}
		width := 1
		for width < len(pattern) && rune(pattern[width]) == r {
			width++
		}
		steps = append(steps, intervalStep{token: intervalToken{field: r, width: width}})
		counts[intervalFieldKey(r)]++
		pattern = pattern[width:]
	}
	return steps, counts
}

func nextIntervalFieldSource(steps []intervalStep, counts, seen map[rune]int) (RangeSource, bool) {
	for _, step := range steps {
		token := step.token
		if token.field == 0 {
			continue
		}
		key := intervalFieldKey(token.field)
		return intervalOccurrenceSource(counts[key], seen[key]+1), true
	}
	return "", false
}

func intervalOccurrenceSource(count, occurrence int) RangeSource {
	if count <= 1 {
		return SourceShared
	}
	if occurrence == 1 {
		return SourceStartRange
	}
	return SourceEndRange
}

func surroundingIntervalSource(previous RangeSource, previousOK bool, next RangeSource, nextOK bool) RangeSource {
	switch {
	case previousOK && nextOK:
		if previous == next {
			return previous
		}
		return SourceShared
	case previousOK:
		return previous
	case nextOK:
		return next
	default:
		return SourceShared
	}
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

func (f *DateTimeFormat) intervalPatternPart(field rune, width int, t localTime, numberingSystem string, uses24Hour bool, cldrLoc cldrdate.Locale, gregorian *cldrdate.Gregorian) Part {
	if isDatePatternField(field) {
		return datePatternPart(field, width, t, gregorian, numberingSystem)
	}
	if part, ok := f.timePatternPart(field, width, t, numberingSystem, uses24Hour, cldrLoc, gregorian); ok {
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

const defaultIntervalFallback = "{0} – {1}"

func partitionRangeFallbackPattern(text string) ecma402.Pattern {
	if text == "" {
		text = defaultIntervalFallback
	}
	patternParts, err := ecma402.PartitionPattern(text)
	if err != nil {
		return ecma402.Pattern{{Type: ecma402.PatternPartLiteral, Value: text}}
	}
	return patternParts
}

func (f *DateTimeFormat) fallbackRangeParts(start, end []Part) []RangePart {
	patternParts := f.fallbackRangePattern
	if len(patternParts) == 0 {
		patternParts = partitionRangeFallbackPattern("")
	}
	parts := make([]RangePart, 0, len(start)+len(end)+1)
	for _, part := range patternParts {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			parts = append(parts, rangeParts(start, SourceStartRange)...)
		case ecma402.PatternPartPlaceholder1:
			parts = append(parts, rangeParts(end, SourceEndRange)...)
		case ecma402.PatternPartLiteral:
			parts = appendRangeLiteralPart(parts, part.Value, SourceShared)
		default:
			parts = appendRangeLiteralPart(parts, "{"+part.Type+"}", SourceShared)
		}
	}
	return parts
}

func (f *DateTimeFormat) appendFallbackRange(dst []byte, start, end []byte) []byte {
	patternParts := f.fallbackRangePattern
	if len(patternParts) == 0 {
		patternParts = partitionRangeFallbackPattern("")
	}
	for _, part := range patternParts {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			dst = append(dst, start...)
		case ecma402.PatternPartPlaceholder1:
			dst = append(dst, end...)
		case ecma402.PatternPartLiteral:
			dst = append(dst, part.Value...)
		default:
			dst = append(dst, "{"+part.Type+"}"...)
		}
	}
	return dst
}

func rangeParts(parts []Part, source RangeSource) []RangePart {
	out := make([]RangePart, len(parts))
	for i, part := range parts {
		out[i] = RangePart{Type: part.Type, Value: part.Value, Source: source}
	}
	return out
}
