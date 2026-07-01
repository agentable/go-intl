package segmenter_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/segmenter"
)

func stringPtr[T ~string](v T) *string {
	value := string(v)
	return &value
}

func collect(s *segmenter.Segments) []segmenter.Segment {
	out := []segmenter.Segment{}
	for seg := range s.All() {
		out = append(out, seg)
	}
	return out
}

func TestSegmenter_Grapheme(t *testing.T) {
	t.Parallel()
	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(segmenter.GraphemeGranularity)})
	if err != nil {
		t.Fatal(err)
	}
	input := "a🙂b"
	got := collect(s.Segment(input))
	gotText := make([]string, len(got))
	for i, seg := range got {
		gotText[i] = seg.Segment
	}
	want := []string{"a", "🙂", "b"}
	if !slices.Equal(gotText, want) {
		t.Errorf("grapheme = %v, want %v", gotText, want)
	}
	if got[2].CodeUnitIndex != 3 {
		t.Errorf("b CodeUnitIndex = %d, want 3", got[2].CodeUnitIndex)
	}
	if got[2].ByteIndex != 5 {
		t.Errorf("b ByteIndex = %d, want 5", got[2].ByteIndex)
	}
	if got[1].Input != input {
		t.Errorf("Input = %q, want %q", got[1].Input, input)
	}
}

func TestSegmenter_Word(t *testing.T) {
	t.Parallel()
	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(segmenter.WordGranularity)})
	if err != nil {
		t.Fatal(err)
	}
	got := collect(s.Segment("Hello, world!"))

	var words []string
	for _, seg := range got {
		if seg.IsWordLike {
			words = append(words, seg.Segment)
		}
	}
	want := []string{"Hello", "world"}
	if !slices.Equal(words, want) {
		t.Errorf("word-like = %v, want %v", words, want)
	}
}

func TestSegmenter_WordLikeUsesUnicodeProperties(t *testing.T) {
	t.Parallel()
	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(segmenter.WordGranularity)})
	if err != nil {
		t.Fatal(err)
	}
	got := collect(s.Segment("\u03b1\u03b2 \u4e2d\u6587 \u0663 \u2167 \u2318 \U0001f642 \u2014"))

	wordLike := map[string]bool{}
	for _, seg := range got {
		if seg.Segment != " " {
			wordLike[seg.Segment] = seg.IsWordLike
		}
	}
	for _, tc := range []struct {
		segment string
		want    bool
	}{
		{segment: "\u03b1\u03b2", want: true},
		{segment: "\u4e2d", want: true},
		{segment: "\u6587", want: true},
		{segment: "\u0663", want: true},
		{segment: "\u2167", want: true},
		{segment: "\u2318", want: false},
		{segment: "\U0001f642", want: false},
		{segment: "\u2014", want: false},
	} {
		got, ok := wordLike[tc.segment]
		if !ok {
			t.Errorf("segment %q was not returned; segments = %v", tc.segment, wordLike)
			continue
		}
		if got != tc.want {
			t.Errorf("segment %q IsWordLike = %t, want %t", tc.segment, got, tc.want)
		}
	}
}

func TestSegmenter_Sentence(t *testing.T) {
	t.Parallel()
	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(segmenter.SentenceGranularity)})
	if err != nil {
		t.Fatal(err)
	}
	got := collect(s.Segment("Hello. World!"))
	if len(got) != 2 {
		t.Errorf("sentence count = %d, want 2; got %v", len(got), got)
	}
}

func TestSegmenter_DefaultGrapheme(t *testing.T) {
	t.Parallel()
	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if got := s.ResolvedOptions().Granularity; got != segmenter.GraphemeGranularity {
		t.Errorf("default granularity = %q, want %q", got, segmenter.GraphemeGranularity)
	}
}

func TestSegmenter_Containing(t *testing.T) {
	t.Parallel()
	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(segmenter.GraphemeGranularity)})
	if err != nil {
		t.Fatal(err)
	}
	segs := s.Segment("a🙂b")
	seg, ok := segs.Containing(2)
	if !ok {
		t.Fatal("Containing(2) ok=false")
	}
	if seg.Segment != "🙂" {
		t.Errorf("Containing(2).Segment = %q, want %q", seg.Segment, "🙂")
	}

	seg, ok = segs.ContainingByte(5)
	if !ok {
		t.Fatal("ContainingByte(5) ok=false")
	}
	if seg.Segment != "b" {
		t.Errorf("ContainingByte(5).Segment = %q, want %q", seg.Segment, "b")
	}

	if _, ok := segs.Containing(-1); ok {
		t.Error("Containing(-1) ok=true, want false")
	}
	if _, ok := segs.Containing(100); ok {
		t.Error("Containing(100) ok=true, want false")
	}
}

func TestSegmenter_ContainingDistinguishesCodeUnitAndByteBoundaries(t *testing.T) {
	t.Parallel()

	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(segmenter.GraphemeGranularity)})
	if err != nil {
		t.Fatal(err)
	}
	segs := s.Segment("a🙂b")

	for _, tc := range []struct {
		name  string
		index int
		want  string
		ok    bool
	}{
		{name: "first code unit", index: 0, want: "a", ok: true},
		{name: "emoji first code unit", index: 1, want: "🙂", ok: true},
		{name: "emoji second code unit", index: 2, want: "🙂", ok: true},
		{name: "last code unit", index: 3, want: "b", ok: true},
		{name: "negative", index: -1},
		{name: "at code-unit length", index: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := segs.Containing(tc.index)
			if ok != tc.ok {
				t.Fatalf("Containing(%d) ok = %v, want %v", tc.index, ok, tc.ok)
			}
			if ok && got.Segment != tc.want {
				t.Fatalf("Containing(%d).Segment = %q, want %q", tc.index, got.Segment, tc.want)
			}
		})
	}

	for _, tc := range []struct {
		name  string
		index int
		want  string
		ok    bool
	}{
		{name: "first byte", index: 0, want: "a", ok: true},
		{name: "emoji first byte", index: 1, want: "🙂", ok: true},
		{name: "emoji interior byte", index: 3, want: "🙂", ok: true},
		{name: "last byte", index: 5, want: "b", ok: true},
		{name: "negative", index: -1},
		{name: "at byte length", index: 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, ok := segs.ContainingByte(tc.index)
			if ok != tc.ok {
				t.Fatalf("ContainingByte(%d) ok = %v, want %v", tc.index, ok, tc.ok)
			}
			if ok && got.Segment != tc.want {
				t.Fatalf("ContainingByte(%d).Segment = %q, want %q", tc.index, got.Segment, tc.want)
			}
		})
	}
}

func TestSegmenter_InvalidUTF8InputUsesReplacementRuneBoundary(t *testing.T) {
	t.Parallel()

	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(segmenter.GraphemeGranularity)})
	if err != nil {
		t.Fatal(err)
	}
	input := string([]byte{'a', 0xff, 'b'})
	got := collect(s.Segment(input))
	if len(got) != 3 {
		t.Fatalf("Segment(invalid UTF-8) yielded %d segments, want 3: %v", len(got), got)
	}

	gotText := make([]string, len(got))
	gotCodeUnitIndexes := make([]int, len(got))
	gotByteIndexes := make([]int, len(got))
	for i, seg := range got {
		gotText[i] = seg.Segment
		gotCodeUnitIndexes[i] = seg.CodeUnitIndex
		gotByteIndexes[i] = seg.ByteIndex
		if seg.Input != input {
			t.Fatalf("segment %d Input = %q, want original invalid UTF-8 string %q", i, seg.Input, input)
		}
	}
	if want := []string{"a", string([]byte{0xff}), "b"}; !slices.Equal(gotText, want) {
		t.Fatalf("segments = %q, want %q", gotText, want)
	}
	if want := []int{0, 1, 2}; !slices.Equal(gotCodeUnitIndexes, want) {
		t.Fatalf("CodeUnitIndex = %v, want %v", gotCodeUnitIndexes, want)
	}
	if want := []int{0, 1, 2}; !slices.Equal(gotByteIndexes, want) {
		t.Fatalf("ByteIndex = %v, want %v", gotByteIndexes, want)
	}

	byCodeUnit, ok := s.Segment(input).Containing(1)
	if !ok {
		t.Fatal("Containing(1) ok=false, want true")
	}
	if want := string([]byte{0xff}); byCodeUnit.Segment != want {
		t.Fatalf("Containing(1).Segment = %q, want %q", byCodeUnit.Segment, want)
	}

	byByte, ok := s.Segment(input).ContainingByte(1)
	if !ok {
		t.Fatal("ContainingByte(1) ok=false, want true")
	}
	if want := string([]byte{0xff}); byByte.Segment != want {
		t.Fatalf("ContainingByte(1).Segment = %q, want %q", byByte.Segment, want)
	}
	if _, ok := s.Segment(input).Containing(3); ok {
		t.Fatal("Containing(3) ok=true, want false at UTF-16 code-unit length")
	}
	if _, ok := s.Segment(input).ContainingByte(3); ok {
		t.Fatal("ContainingByte(3) ok=true, want false at byte length")
	}
}

func TestSegmenter_EmptyInputHasNoSegments(t *testing.T) {
	t.Parallel()

	for _, granularity := range []segmenter.Granularity{
		segmenter.GraphemeGranularity,
		segmenter.WordGranularity,
		segmenter.SentenceGranularity,
	} {
		t.Run(string(granularity), func(t *testing.T) {
			t.Parallel()

			s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(granularity)})
			if err != nil {
				t.Fatal(err)
			}
			segs := s.Segment("")
			if got := collect(segs); len(got) != 0 {
				t.Fatalf("Segment(\"\").All() = %v, want no segments", got)
			}
			if _, ok := segs.Containing(0); ok {
				t.Fatal("Containing(0) on empty input ok=true, want false")
			}
			if _, ok := segs.ContainingByte(0); ok {
				t.Fatal("ContainingByte(0) on empty input ok=true, want false")
			}
		})
	}
}

func TestSegmenter_AllStopsWhenYieldReturnsFalse(t *testing.T) {
	t.Parallel()

	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, segmenter.Options{Granularity: stringPtr(segmenter.GraphemeGranularity)})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for range s.Segment("abc").All() {
		count++
		break
	}
	if count != 1 {
		t.Fatalf("early-stopped All() yielded %d segments, want 1", count)
	}
}

func TestSegmenter_New_Errors(t *testing.T) {
	t.Parallel()
	en := intltest.Locale(t, "en")

	tests := []struct {
		name         string
		opts         segmenter.Options
		wantName     string
		wantValue    string
		wantExpected string
	}{
		{
			name:         "invalid granularity",
			opts:         segmenter.Options{Granularity: stringPtr("bogus")},
			wantName:     "granularity",
			wantValue:    "bogus",
			wantExpected: `one of "grapheme", "word", "sentence"`,
		},
		{
			name:         "invalid locale matcher",
			opts:         segmenter.Options{LocaleMatcher: stringPtr("bogus")},
			wantName:     "localeMatcher",
			wantValue:    "bogus",
			wantExpected: `one of "lookup", "best fit"`,
		},
		{
			name:         "empty locale matcher",
			opts:         segmenter.Options{LocaleMatcher: stringPtr("")},
			wantName:     "localeMatcher",
			wantExpected: `one of "lookup", "best fit"`,
		},
		{
			name:         "empty granularity",
			opts:         segmenter.Options{Granularity: stringPtr("")},
			wantName:     "granularity",
			wantExpected: `one of "grapheme", "word", "sentence"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := segmenter.New(locale.List{en}, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New() error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "segmenter", intlerr.InvalidOption, tc.wantName, tc.wantValue, en.String())
			testcontract.AssertOptionExpected(t, err, tc.wantExpected)
		})
	}
}

func TestSegmenter_NewDoesNotResolveUnsupportedTailoredLocales(t *testing.T) {
	t.Parallel()

	for _, tag := range []string{"ja", "th", "zh-Hant"} {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()

			requested := intltest.Locale(t, tag)
			format, err := segmenter.New(locale.List{requested}, segmenter.Options{Granularity: stringPtr(segmenter.WordGranularity)})
			if err != nil {
				t.Fatalf("New(%q) error = %v", tag, err)
			}
			if got := format.ResolvedOptions().Locale.String(); got == tag {
				t.Fatalf("ResolvedOptions().Locale = %q, want fallback locale until tailored segmentation is supported", got)
			}
			supported, err := segmenter.SupportedLocalesOf(locale.List{requested}, segmenter.Options{})
			if err != nil {
				t.Fatalf("SupportedLocalesOf(%q) error = %v", tag, err)
			}
			if len(supported) != 0 {
				t.Fatalf("SupportedLocalesOf(%q) = %v, want unsupported", tag, supported.Strings())
			}
		})
	}
}

func TestSupportedLocalesOf(t *testing.T) {
	t.Parallel()
	requested := locale.List{intltest.Locale(t, "en-US"), intltest.Locale(t, "xh")}
	got, err := segmenter.SupportedLocalesOf(requested, segmenter.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].String() != "en-US" {
		t.Errorf("SupportedLocalesOf = %v, want [en-US]", got)
	}
}

func TestSupportedLocalesOfErrors(t *testing.T) {
	t.Parallel()

	requested := locale.List{intltest.Locale(t, "en-US")}
	for _, tc := range []struct {
		name      string
		opts      segmenter.Options
		wantValue string
	}{
		{name: "bogus", opts: segmenter.Options{LocaleMatcher: stringPtr("bogus")}, wantValue: "bogus"},
		{name: "explicit empty", opts: segmenter.Options{LocaleMatcher: stringPtr("")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := segmenter.SupportedLocalesOf(requested, tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("SupportedLocalesOf() error = %v, want intlerr.ErrInvalidOption", err)
			}
			testcontract.AssertOptionError(t, err, "segmenter", intlerr.InvalidOption, "localeMatcher", tc.wantValue, "en-US")
			testcontract.AssertOptionExpected(t, err, `one of "lookup", "best fit"`)
		})
	}
}
