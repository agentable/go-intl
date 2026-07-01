package datetimeformat

import (
	"slices"
	"testing"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
)

func TestAvailableFormatCandidatesSortsSkeletons(t *testing.T) {
	t.Parallel()

	candidates := availableFormatCandidates(cldrdate.Gregorian{
		AvailableFormats: map[string]string{
			"yMd":   "M/d/y",
			"Hms":   "HH:mm:ss",
			"yMMMd": "MMM d, y",
		},
	})

	gotSkeletons := make([]string, len(candidates))
	gotPatterns := make([]string, len(candidates))
	for i, candidate := range candidates {
		gotSkeletons[i] = candidate.Skeleton
		gotPatterns[i] = candidate.Pattern
	}
	if want := []string{"Hms", "yMMMd", "yMd"}; !slices.Equal(gotSkeletons, want) {
		t.Fatalf("available format skeletons = %#v, want %#v", gotSkeletons, want)
	}
	if want := []string{"HH:mm:ss", "MMM d, y", "M/d/y"}; !slices.Equal(gotPatterns, want) {
		t.Fatalf("available format patterns = %#v, want %#v", gotPatterns, want)
	}
}

func TestDateTimeStyleIndexFollowsStyleOrder(t *testing.T) {
	t.Parallel()

	for i, style := range dateTimeStyles {
		if got := dateTimeStyleIndex(style); got != i {
			t.Fatalf("dateTimeStyleIndex(%q) = %d, want %d", style, got, i)
		}
	}

	if got := dateTimeStyleIndex(Style("unknown")); got != defaultDateTimeStyleIndex {
		t.Fatalf("dateTimeStyleIndex(%q) = %d, want %d", Style("unknown"), got, defaultDateTimeStyleIndex)
	}
}
