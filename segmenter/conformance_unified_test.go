package segmenter

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/tools/conformance"
)

func TestUnifiedConformanceFixtures(t *testing.T) {
	t.Parallel()

	conformance.RunFixtures(t, ".", func(t *testing.T, fixture conformance.Fixture) {
		if fixture.IsSupportedLocalesOf() {
			runSupportedLocalesFixture(t, fixture)
			return
		}

		format, err := New(locale.List{intltest.Locale(t, fixture.Locale)}, conformanceSegmenterOptions(t, fixture))
		if testcontract.AssertErrorCode(t, "New()", err, fixture.ErrorCode, func(code string) error {
			return conformanceSegmenterError(t, code)
		}) {
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

func TestConformanceSegmenterOptionsPreserveExplicitEmptyString(t *testing.T) {
	t.Parallel()

	_, err := New(intltest.LocaleList(t, "en"), conformanceSegmenterOptions(t, conformance.Fixture{
		Options: json.RawMessage(`{"granularity":""}`),
	}))
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New() error = %v, want %v", err, intlerr.ErrInvalidOption)
	}
	testcontract.AssertOptionError(t, err, "segmenter", intlerr.InvalidOption, "granularity", "", "en")
	testcontract.AssertOptionExpected(t, err, `one of "grapheme", "word", "sentence"`)
}

func runSupportedLocalesFixture(t *testing.T, fixture conformance.Fixture) {
	t.Helper()

	testcontract.AssertSupportedLocalesOfFixture(t, fixture, intltest.LocaleListJSON, func(locales locale.List) (locale.List, error) {
		return SupportedLocalesOf(locales, conformanceSegmenterOptions(t, fixture))
	}, func(code string) error {
		return conformanceSegmenterError(t, code)
	})
}

func conformanceSegmenterOptions(t *testing.T, fixture conformance.Fixture) Options {
	t.Helper()

	var options struct {
		LocaleMatcher *string `json:"localeMatcher"`
		Granularity   *string `json:"granularity"`
	}
	if err := json.Unmarshal(fixture.Options, &options); err != nil {
		t.Fatal(err)
	}
	return Options{
		LocaleMatcher: options.LocaleMatcher,
		Granularity:   options.Granularity,
	}
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

	return testcontract.IntlErrorCode(t, "segmenter", code, "invalid_option")
}
