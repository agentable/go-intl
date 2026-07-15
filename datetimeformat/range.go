package datetimeformat

import (
	"time"
	"unicode/utf8"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	"github.com/agentable/go-intl/internal/ecma402"
)

func (f *DateTimeFormat) FormatRange(start, end time.Time) (string, error) {
	r, err := f.normalizeRange(start, end)
	if err != nil {
		return "", err
	}
	if r.relation.equal {
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
	if r.relation.equal {
		return rangeParts(f.FormatToParts(r.start), SourceShared), nil
	}
	pattern := f.pattern
	return f.formatRangeParts(pattern, r), nil
}

func (f *DateTimeFormat) formatRangeParts(pattern selectedPattern, r normalizedRange) []RangePart {
	if parts, ok := f.formatIntervalRangeToParts(pattern, r.relation, r.startLocal, r.endLocal); ok {
		return parts
	}
	startParts := pattern.parts(f, r.startLocal)
	endParts := pattern.parts(f, r.endLocal)
	if r.relation.fallbackPattern != "" {
		startParts = f.formatPattern(r.relation.fallbackPattern, r.startLocal)
		endParts = f.formatPattern(r.relation.fallbackPattern, r.endLocal)
	}
	return f.fallbackRangeParts(
		startParts,
		endParts,
	)
}

func (f *DateTimeFormat) appendRange(dst []byte, pattern selectedPattern, r normalizedRange) []byte {
	if out, ok := f.appendIntervalRange(dst, pattern, r.relation, r.startLocal, r.endLocal); ok {
		return out
	}
	var startScratch [64]byte
	var endScratch [64]byte
	var start, end []byte
	if r.relation.fallbackPattern != "" {
		start = f.appendPattern(startScratch[:0], r.relation.fallbackPattern, r.startLocal)
		end = f.appendPattern(endScratch[:0], r.relation.fallbackPattern, r.endLocal)
	} else {
		start = pattern.appendTo(f, startScratch[:0], r.startLocal)
		end = pattern.appendTo(f, endScratch[:0], r.endLocal)
	}
	return f.appendFallbackRange(dst, start, end)
}

type normalizedRange struct {
	start      time.Time
	startLocal localTime
	endLocal   localTime
	relation   rangeRelation
}

func (f *DateTimeFormat) normalizeRange(start, end time.Time) (normalizedRange, error) {
	location := f.location
	start = start.Round(0)
	end = end.Round(0)
	if start.Equal(end) {
		return normalizedRange{start: start, relation: rangeRelation{equal: true}}, nil
	}
	start, startLocal := gregoryTimeInLocation(start, location)
	_, endLocal := gregoryTimeInLocation(end, location)
	var relation rangeRelation
	switch f.pattern.kind {
	case patternDate:
		relation = f.pattern.rangeRecord.dateRelation(startLocal, endLocal)
	case patternTime, patternDateTime:
		relation = f.pattern.rangeRecord.timeRelation(f, startLocal, endLocal)
	case patternNone:
	}
	return normalizedRange{
		start:      start,
		startLocal: startLocal,
		endLocal:   endLocal,
		relation:   relation,
	}, nil
}

func (f *DateTimeFormat) formatIntervalRangeToParts(pattern selectedPattern, relation rangeRelation, start, end localTime) ([]RangePart, bool) {
	rangePattern := relation.pattern
	if rangePattern == "" {
		return nil, false
	}
	switch pattern.kind {
	case patternDate:
		return f.formatIntervalPattern(rangePattern, start, end), true
	case patternTime:
		return f.formatIntervalPattern(rangePattern, start, end), true
	case patternDateTime:
		dateParts := rangeParts(f.formatPattern(pattern.date, start), SourceShared)
		timeParts := f.formatIntervalPattern(rangePattern, start, end)
		return interpolateDateTimeRangeParts(pattern.dateTime, dateParts, timeParts), true
	case patternNone:
	}
	return nil, false
}

func (f *DateTimeFormat) appendIntervalRange(dst []byte, pattern selectedPattern, relation rangeRelation, start, end localTime) ([]byte, bool) {
	rangePattern := relation.pattern
	if rangePattern == "" {
		return nil, false
	}
	switch pattern.kind {
	case patternDate:
		return f.appendIntervalPattern(dst, rangePattern, start, end), true
	case patternTime:
		return f.appendIntervalPattern(dst, rangePattern, start, end), true
	case patternDateTime:
		var dateScratch [64]byte
		var timeScratch [64]byte
		date := f.appendPattern(dateScratch[:0], pattern.date, start)
		time := f.appendIntervalPattern(timeScratch[:0], rangePattern, start, end)
		return appendDateTimePattern(dst, pattern.dateTime, date, time), true
	case patternNone:
	}
	return nil, false
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
	patternParts := f.pattern.rangeRecord.fallback
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
	patternParts := f.pattern.rangeRecord.fallback
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
