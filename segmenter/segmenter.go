package segmenter

import (
	"encoding/json"
	"iter"
	"sync"
	"unicode"

	"github.com/rivo/uniseg"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
	"github.com/agentable/go-intl/internal/ecma402"
	"github.com/agentable/go-intl/internal/localematcher"
	cldrseg "github.com/agentable/go-intl/internal/segmentation"
	"github.com/agentable/go-intl/locale"
)

// Segmenter splits strings into graphemes, words, or sentences.
type Segmenter struct {
	resolved ResolvedOptions
	boundary segmentBoundary
}

var segmenterLocaleMatcher = sync.OnceValue(func() *localematcher.Matcher {
	return localematcher.NewMatcher(cldrseg.SupportedLocales(), cldrlocale.Maximize)
})

// Segment represents one boundary-delimited chunk of a string.
type Segment struct {
	Segment       string `json:"segment"`
	CodeUnitIndex int    `json:"index"`
	ByteIndex     int    `json:"-"`
	Input         string `json:"input"`
	IsWordLike    bool   `json:"-"`

	hasWordLike bool
}

type nextSegment func(rest string, state int) (segment, remaining string, nextState int)

type segmentBoundary struct {
	next     nextSegment
	wordLike func(string) bool
}

func (s Segment) MarshalJSON() ([]byte, error) {
	type segmentJSON struct {
		Segment       string `json:"segment"`
		CodeUnitIndex int    `json:"index"`
		Input         string `json:"input"`
		IsWordLike    *bool  `json:"isWordLike,omitempty"`
	}
	out := segmentJSON{
		Segment:       s.Segment,
		CodeUnitIndex: s.CodeUnitIndex,
		Input:         s.Input,
	}
	if s.hasWordLike || s.IsWordLike {
		out.IsWordLike = &s.IsWordLike
	}
	return json.Marshal(out)
}

// New constructs a Segmenter for the requested locale and options.
func New(locales locale.List, opts Options) (*Segmenter, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	validationLocaleName := validationLocale.String()
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocaleName); err != nil {
		return nil, err
	}

	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:       locales,
		Fallback:      validationLocale,
		LocaleMatcher: cfg.localeMatcher,
		Matcher:       segmenterLocaleMatcher(),
	})

	granularity := Granularity(cfg.granularity)
	return &Segmenter{
		resolved: ResolvedOptions{
			Locale:      resolution.Locale,
			Granularity: granularity,
		},
		boundary: segmentBoundaryFor(granularity),
	}, nil
}

// Segments holds the segmentation view of a specific input string.
type Segments struct {
	boundary segmentBoundary
	input    string
}

// Segment returns a Segments view over s using the resolved granularity.
func (f *Segmenter) Segment(input string) *Segments {
	return &Segments{
		boundary: f.boundary,
		input:    input,
	}
}

// All iterates over every segment in order.
func (s *Segments) All() iter.Seq[Segment] {
	boundary := s.boundary
	input := s.input
	return func(yield func(Segment) bool) {
		yieldSegments(boundary, input, yield)
	}
}

func segmentBoundaryFor(granularity Granularity) segmentBoundary {
	switch granularity {
	case GraphemeGranularity:
		return segmentBoundary{next: nextGraphemeSegment}
	case WordGranularity:
		return segmentBoundary{next: uniseg.FirstWordInString, wordLike: isWordLike}
	case SentenceGranularity:
		return segmentBoundary{next: uniseg.FirstSentenceInString}
	}
	return segmentBoundary{}
}

func nextGraphemeSegment(rest string, state int) (segment, remaining string, nextState int) {
	segment, remaining, _, nextState = uniseg.FirstGraphemeClusterInString(rest, state)
	return segment, remaining, nextState
}

func yieldSegments(boundary segmentBoundary, input string, yield func(Segment) bool) {
	if boundary.next == nil {
		return
	}
	iteratorState := -1
	rest := input
	byteIndex := 0
	codeUnitIndex := 0
	for len(rest) > 0 {
		part, remaining, nextState := boundary.next(rest, iteratorState)
		segment := Segment{
			Segment:       part,
			CodeUnitIndex: codeUnitIndex,
			ByteIndex:     byteIndex,
			Input:         input,
		}
		if boundary.wordLike != nil {
			segment.hasWordLike = true
			segment.IsWordLike = boundary.wordLike(part)
		}
		if !yield(segment) {
			return
		}
		rest = remaining
		iteratorState = nextState
		byteIndex += len(part)
		codeUnitIndex += ecma402.UTF16CodeUnitCount(part)
	}
}

// Containing returns the segment that contains UTF-16 code-unit index.
// The bool is false when index is out of range, matching the JS
// `Segments.prototype.containing` undefined return.
func (s *Segments) Containing(index int) (Segment, bool) {
	return findContainingSegment(s.boundary, s.input, index, ecma402.UTF16CodeUnitCount(s.input), segmentCodeUnitEnd)
}

// ContainingByte returns the segment that contains UTF-8 byte offset index.
func (s *Segments) ContainingByte(index int) (Segment, bool) {
	return findContainingSegment(s.boundary, s.input, index, len(s.input), segmentByteEnd)
}

func findContainingSegment(boundary segmentBoundary, input string, index, limit int, segmentEnd func(Segment) int) (Segment, bool) {
	if index < 0 || index >= limit {
		return Segment{}, false
	}
	var found Segment
	ok := false
	yieldSegments(boundary, input, func(seg Segment) bool {
		if index < segmentEnd(seg) {
			found = seg
			ok = true
			return false
		}
		return true
	})
	return found, ok
}

func segmentCodeUnitEnd(seg Segment) int {
	return seg.CodeUnitIndex + ecma402.UTF16CodeUnitCount(seg.Segment)
}

func segmentByteEnd(seg Segment) int {
	return seg.ByteIndex + len(seg.Segment)
}

func isWordLike(word string) bool {
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
	}
	return false
}
