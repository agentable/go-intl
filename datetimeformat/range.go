package datetimeformat

import (
	"strings"
	"time"
	"unicode"

	"github.com/agentable/go-intl/internal/ecma402"
)

func (f *DateTimeFormat) FormatRange(start, end time.Time) string {
	r := f.normalizeRange(start, end)
	if r.relation.equal {
		return f.Format(r.start)
	}
	return string(f.appendRange(nil, f.pattern, r))
}

func (f *DateTimeFormat) FormatRangeToParts(start, end time.Time) []RangePart {
	r := f.normalizeRange(start, end)
	if r.relation.equal {
		return rangeParts(f.FormatToParts(r.start), SourceShared)
	}
	return f.formatRangeParts(f.pattern, r)
}

func (f *DateTimeFormat) formatRangeParts(pattern selectedPattern, r normalizedRange) []RangePart {
	if parts, ok := f.formatIntervalRangeToParts(pattern, r.relation, r.startLocal, r.endLocal); ok {
		return parts
	}
	startParts := pattern.parts(f, r.startLocal)
	endParts := pattern.parts(f, r.endLocal)
	if r.relation.fallbackProgram != nil {
		startParts = r.relation.fallbackProgram.parts(f, r.startLocal)
		endParts = r.relation.fallbackProgram.parts(f, r.endLocal)
	}
	return f.fallbackRangeParts(startParts, endParts)
}

func (f *DateTimeFormat) appendRange(dst []byte, pattern selectedPattern, r normalizedRange) []byte {
	if out, ok := f.appendIntervalRange(dst, pattern, r.relation, r.startLocal, r.endLocal); ok {
		return out
	}
	var startScratch [64]byte
	var endScratch [64]byte
	var start, end []byte
	if r.relation.fallbackProgram != nil {
		start = r.relation.fallbackProgram.appendTo(f, startScratch[:0], r.startLocal)
		end = r.relation.fallbackProgram.appendTo(f, endScratch[:0], r.endLocal)
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

func (f *DateTimeFormat) normalizeRange(start, end time.Time) normalizedRange {
	location := f.location
	start = start.Round(0)
	end = end.Round(0)
	if start.Equal(end) {
		return normalizedRange{start: start, relation: rangeRelation{equal: true}}
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
	}
}

func (f *DateTimeFormat) formatIntervalRangeToParts(pattern selectedPattern, relation rangeRelation, start, end localTime) ([]RangePart, bool) {
	program := relation.program
	if program == nil {
		return nil, false
	}
	switch pattern.kind {
	case patternDate, patternTime:
		return program.parts(f, start, end), true
	case patternDateTime:
		dateParts := rangeParts(pattern.dateProgram.parts(f, start), SourceShared)
		timeParts := program.parts(f, start, end)
		return interpolateDateTimeRangeParts(pattern.dateTimeProgram, dateParts, timeParts), true
	case patternNone:
	}
	return nil, false
}

func (f *DateTimeFormat) appendIntervalRange(dst []byte, pattern selectedPattern, relation rangeRelation, start, end localTime) ([]byte, bool) {
	program := relation.program
	if program == nil {
		return nil, false
	}
	switch pattern.kind {
	case patternDate, patternTime:
		return program.appendTo(f, dst, start, end), true
	case patternDateTime:
		var dateScratch [64]byte
		var timeScratch [64]byte
		date := pattern.dateProgram.appendTo(f, dateScratch[:0], start)
		time := program.appendTo(f, timeScratch[:0], start, end)
		return appendDateTimeProgram(dst, pattern.dateTimeProgram, date, time), true
	case patternNone:
	}
	return nil, false
}

func (p intervalProgram) parts(f *DateTimeFormat, start, end localTime) []RangePart {
	parts := make([]RangePart, 0, len(p))
	for _, step := range p {
		token := step.token
		if token.literal != "" {
			parts = appendRangeLiteralPart(parts, token.literal, step.source)
			continue
		}
		part, ok := f.patternTokenPart(token, step.time(start, end))
		if ok {
			parts = append(parts, RangePart{Type: part.Type, Value: part.Value, Source: step.source})
		} else {
			parts = trimTrailingRangeLiteralSpace(parts)
		}
	}
	return parts
}

func (p intervalProgram) appendTo(f *DateTimeFormat, dst []byte, start, end localTime) []byte {
	for _, step := range p {
		token := step.token
		if token.literal != "" {
			dst = append(dst, token.literal...)
			continue
		}
		part, ok := f.patternTokenPart(token, step.time(start, end))
		if ok {
			dst = append(dst, part.Value...)
		} else {
			dst = trimTrailingSpaceBytes(dst)
		}
	}
	return dst
}

type intervalStep struct {
	token    patternToken
	source   RangeSource
	endpoint intervalEndpoint
}

type intervalProgram []intervalStep

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

func compileIntervalPattern(pattern string) intervalProgram {
	tokens := compilePatternProgram(pattern)
	steps := make(intervalProgram, len(tokens))
	counts := map[rune]int{}
	for i, token := range tokens {
		steps[i].token = token
		if token.field != 0 {
			counts[intervalFieldKey(token.field)]++
		}
	}
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
	case previousOK && nextOK && previous == next:
		return previous
	case previousOK && nextOK:
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

func trimTrailingRangeLiteralSpace(parts []RangePart) []RangePart {
	if len(parts) == 0 || parts[len(parts)-1].Type != PartLiteral {
		return parts
	}
	parts[len(parts)-1].Value = strings.TrimRightFunc(parts[len(parts)-1].Value, unicode.IsSpace)
	if parts[len(parts)-1].Value == "" {
		return parts[:len(parts)-1]
	}
	return parts
}

const defaultIntervalFallback = "{0} – {1}"

type rangeFallbackStepKind uint8

const (
	rangeFallbackLiteral rangeFallbackStepKind = iota
	rangeFallbackStart
	rangeFallbackEnd
)

type rangeFallbackStep struct {
	kind    rangeFallbackStepKind
	literal string
}

type rangeFallbackProgram []rangeFallbackStep

func compileRangeFallbackProgram(text string) rangeFallbackProgram {
	if text == "" {
		text = defaultIntervalFallback
	}
	patternParts, err := ecma402.PartitionPattern(text)
	if err != nil {
		return rangeFallbackProgram{{kind: rangeFallbackLiteral, literal: text}}
	}
	program := make(rangeFallbackProgram, 0, len(patternParts))
	for _, part := range patternParts {
		switch part.Type {
		case ecma402.PatternPartPlaceholder0:
			program = append(program, rangeFallbackStep{kind: rangeFallbackStart})
		case ecma402.PatternPartPlaceholder1:
			program = append(program, rangeFallbackStep{kind: rangeFallbackEnd})
		case ecma402.PatternPartLiteral:
			program = append(program, rangeFallbackStep{kind: rangeFallbackLiteral, literal: part.Value})
		default:
			program = append(program, rangeFallbackStep{kind: rangeFallbackLiteral, literal: "{" + part.Type + "}"})
		}
	}
	return program
}

func (f *DateTimeFormat) fallbackRangeParts(start, end []Part) []RangePart {
	return f.pattern.rangeRecord.fallback.parts(start, end)
}

func (p rangeFallbackProgram) parts(start, end []Part) []RangePart {
	parts := make([]RangePart, 0, len(start)+len(end)+len(p))
	for _, step := range p {
		switch step.kind {
		case rangeFallbackStart:
			parts = append(parts, rangeParts(start, SourceStartRange)...)
		case rangeFallbackEnd:
			parts = append(parts, rangeParts(end, SourceEndRange)...)
		case rangeFallbackLiteral:
			parts = appendRangeLiteralPart(parts, step.literal, SourceShared)
		}
	}
	return parts
}

func (f *DateTimeFormat) appendFallbackRange(dst []byte, start, end []byte) []byte {
	return f.pattern.rangeRecord.fallback.appendTo(dst, start, end)
}

func (p rangeFallbackProgram) appendTo(dst, start, end []byte) []byte {
	for _, step := range p {
		switch step.kind {
		case rangeFallbackStart:
			dst = append(dst, start...)
		case rangeFallbackEnd:
			dst = append(dst, end...)
		case rangeFallbackLiteral:
			dst = append(dst, step.literal...)
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
