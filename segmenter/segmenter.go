package segmenter

import (
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
	IsWordLike    bool   `json:"isWordLike"`
}

// New constructs a Segmenter for the requested locale and options.
func New(locales locale.List, opts Options) (*Segmenter, error) {
	validationLocale := ecma402.ValidationLocale(locales)
	cfg := defaultConfig()
	applyOptions(&cfg, opts)
	if err := cfg.validate(validationLocale); err != nil {
		return nil, err
	}

	resolution := ecma402.ResolveConstructorLocale(ecma402.ConstructorLocaleOptions{
		Locales:       locales,
		Fallback:      validationLocale,
		LocaleMatcher: cfg.localeMatcher,
		Matcher:       segmenterLocaleMatcher(),
	})

	return &Segmenter{
		resolved: ResolvedOptions{
			Locale:      resolution.Locale,
			Granularity: Granularity(cfg.granularity),
		},
	}, nil
}

// Segments holds the segmentation view of a specific input string.
type Segments struct {
	seg   *Segmenter
	input string
}

// Segment returns a Segments view over s using the resolved granularity.
func (f *Segmenter) Segment(input string) *Segments {
	return &Segments{seg: f, input: input}
}

// All iterates over every segment in order.
func (s *Segments) All() iter.Seq[Segment] {
	return func(yield func(Segment) bool) {
		switch s.seg.resolved.Granularity {
		case GraphemeGranularity:
			state := -1
			rest := s.input
			byteIndex := 0
			codeUnitIndex := 0
			for len(rest) > 0 {
				var cluster string
				cluster, rest, _, state = uniseg.FirstGraphemeClusterInString(rest, state)
				if !yield(Segment{Segment: cluster, CodeUnitIndex: codeUnitIndex, ByteIndex: byteIndex, Input: s.input}) {
					return
				}
				byteIndex += len(cluster)
				codeUnitIndex += ecma402.UTF16CodeUnitCount(cluster)
			}
		case WordGranularity:
			state := -1
			rest := s.input
			byteIndex := 0
			codeUnitIndex := 0
			for len(rest) > 0 {
				var word string
				word, rest, state = uniseg.FirstWordInString(rest, state)
				if !yield(Segment{
					Segment:       word,
					CodeUnitIndex: codeUnitIndex,
					ByteIndex:     byteIndex,
					Input:         s.input,
					IsWordLike:    isWordLike(word),
				}) {
					return
				}
				byteIndex += len(word)
				codeUnitIndex += ecma402.UTF16CodeUnitCount(word)
			}
		case SentenceGranularity:
			state := -1
			rest := s.input
			byteIndex := 0
			codeUnitIndex := 0
			for len(rest) > 0 {
				var sentence string
				sentence, rest, state = uniseg.FirstSentenceInString(rest, state)
				if !yield(Segment{Segment: sentence, CodeUnitIndex: codeUnitIndex, ByteIndex: byteIndex, Input: s.input}) {
					return
				}
				byteIndex += len(sentence)
				codeUnitIndex += ecma402.UTF16CodeUnitCount(sentence)
			}
		}
	}
}

// Containing returns the segment that contains UTF-16 code-unit index.
// The bool is false when index is out of range, matching the JS
// `Segments.prototype.containing` undefined return.
func (s *Segments) Containing(index int) (Segment, bool) {
	if index < 0 || index >= ecma402.UTF16CodeUnitCount(s.input) {
		return Segment{}, false
	}
	for seg := range s.All() {
		if index < seg.CodeUnitIndex+ecma402.UTF16CodeUnitCount(seg.Segment) {
			return seg, true
		}
	}
	return Segment{}, false
}

// ContainingByte returns the segment that contains UTF-8 byte offset index.
func (s *Segments) ContainingByte(index int) (Segment, bool) {
	if index < 0 || index >= len(s.input) {
		return Segment{}, false
	}
	for seg := range s.All() {
		if index < seg.ByteIndex+len(seg.Segment) {
			return seg, true
		}
	}
	return Segment{}, false
}

func isWordLike(word string) bool {
	for _, r := range word {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
	}
	return false
}
