package segmenter_test

import (
	"testing"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/segmenter"
)

func BenchmarkSegmenter_WordAll_Cached(b *testing.B) {
	format := benchmarkSegmenter(b, segmenter.Options{Granularity: stringPtr(segmenter.WordGranularity)})
	segments := format.Segment("Hello, world!")

	b.ReportAllocs()
	for b.Loop() {
		wordCount := 0
		for seg := range segments.All() {
			if seg.IsWordLike {
				wordCount++
			}
		}
		if wordCount != 2 {
			b.Fatalf("word-like segment count = %d", wordCount)
		}
	}
}

func BenchmarkSegmenter_Containing_Cached(b *testing.B) {
	format := benchmarkSegmenter(b, segmenter.Options{Granularity: stringPtr(segmenter.GraphemeGranularity)})
	segments := format.Segment("a🙂b")

	b.ReportAllocs()
	for b.Loop() {
		seg, ok := segments.Containing(2)
		if !ok || seg.Segment != "🙂" {
			b.Fatalf("Containing(2) = %#v, %v", seg, ok)
		}
	}
}

func benchmarkSegmenter(b *testing.B, opts segmenter.Options) *segmenter.Segmenter {
	b.Helper()

	format, err := segmenter.New(locale.List{intltest.Locale(b, "en-US")}, opts)
	if err != nil {
		b.Fatal(err)
	}
	return format
}
