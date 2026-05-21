package segmenter

import (
	"encoding/json"
	"testing"

	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		format, err := New(locale.List{locale.MustParse(fixture.Locale)}, conformanceSegmenterOptions(t, fixture))
		if err != nil {
			t.Fatal(err)
		}
		var input string
		if err := json.Unmarshal(fixture.Input, &input); err != nil {
			t.Fatal(err)
		}
		var got []Segment
		for segment := range format.Segment(input).All() {
			got = append(got, segment)
		}
		if len(got) != len(fixture.ExpectedSegments) {
			t.Fatalf("Segment(%q) returned %d segments, want %d: %v", input, len(got), len(fixture.ExpectedSegments), got)
		}
		for i, want := range fixture.ExpectedSegments {
			assertSegmentRecord(t, input, i, got[i], want)
		}
	})
}

func conformanceSegmenterOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher string `json:"localeMatcher"`
		Granularity   string `json:"granularity"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	var opts Options
	if options.LocaleMatcher != "" {
		opts.LocaleMatcher = LocaleMatcher(options.LocaleMatcher)
	}
	if options.Granularity != "" {
		opts.Granularity = Granularity(options.Granularity)
	}
	return opts
}

func assertSegmentRecord(t *testing.T, input string, i int, got Segment, want conformance.SegmentRecord) {
	t.Helper()

	if got.Segment != want.Segment {
		t.Fatalf("Segment(%q)[%d].Segment = %q, want %q", input, i, got.Segment, want.Segment)
	}
	if got.CodeUnitIndex != want.CodeUnitIndex {
		t.Fatalf("Segment(%q)[%d].CodeUnitIndex = %d, want %d", input, i, got.CodeUnitIndex, want.CodeUnitIndex)
	}
	if want.ByteIndex != nil && got.ByteIndex != *want.ByteIndex {
		t.Fatalf("Segment(%q)[%d].ByteIndex = %d, want %d", input, i, got.ByteIndex, *want.ByteIndex)
	}
	if want.IsWordLike != nil && got.IsWordLike != *want.IsWordLike {
		t.Fatalf("Segment(%q)[%d].IsWordLike = %v, want %v", input, i, got.IsWordLike, *want.IsWordLike)
	}
}
