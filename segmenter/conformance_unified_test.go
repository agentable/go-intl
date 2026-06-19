package segmenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		if fixture.Feature == "supportedLocalesOf" {
			runSupportedLocalesFixture(t, fixture)
			return
		}

		format, err := New(locale.List{intltest.Locale(t, fixture.Locale)}, conformanceSegmenterOptions(t, fixture))
		if fixture.ErrorCode != "" {
			if !errors.Is(err, conformanceSegmenterError(t, fixture.ErrorCode)) {
				t.Fatalf("New() error = %v, want %q", err, fixture.ErrorCode)
			}
			return
		}
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

func runSupportedLocalesFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	var tags []string
	if err := json.Unmarshal(fixture.Input, &tags); err != nil {
		t.Fatal(err)
	}
	got, err := SupportedLocalesOf(intltest.LocaleList(t, tags...), conformanceSegmenterOptions(t, fixture))
	if fixture.ErrorCode != "" {
		if !errors.Is(err, conformanceSegmenterError(t, fixture.ErrorCode)) {
			t.Fatalf("SupportedLocalesOf() error = %v, want %q", err, fixture.ErrorCode)
		}
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(fixture.ExpectedLocales) {
		t.Fatalf("SupportedLocalesOf(%v) = %v, want %v", tags, got.Strings(), fixture.ExpectedLocales)
	}
	for i, want := range fixture.ExpectedLocales {
		if got[i].String() != want {
			t.Fatalf("SupportedLocalesOf(%v)[%d] = %q, want %q", tags, i, got[i].String(), want)
		}
	}
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

func conformanceSegmenterError(t *testing.T, code string) error {
	t.Helper()

	switch code {
	case "invalid_option":
		return intlerr.ErrInvalidOption
	default:
		t.Fatalf("unsupported segmenter errorCode %q", code)
		return nil
	}
}
