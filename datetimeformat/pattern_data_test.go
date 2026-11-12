package datetimeformat

import (
	"testing"

	ecma402dtf "github.com/agentable/go-intl/internal/ecma402/datetimeformat"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
)

func TestAvailableFormatCandidatesPreserveLocalizedTies(t *testing.T) {
	t.Parallel()
	g := cldrdate.Gregorian{AvailableFormats: map[string]string{"y": "y 'localized'", "Hms": "HH:mm:ss"}}
	candidates := availableFormatCandidates(g)
	selected := ecma402dtf.Match(ecma402dtf.Options{Year: ecma402dtf.NumericNumeric}, candidates)
	if selected.Pattern != "y 'localized'" {
		t.Fatalf("selected pattern = %q", selected.Pattern)
	}
	for _, opts := range []ecma402dtf.Options{
		{Minute: ecma402dtf.NumericNumeric}, {Second: ecma402dtf.NumericNumeric}, {DayPeriod: ecma402dtf.FieldLong},
	} {
		format := ecma402dtf.Match(opts, candidates)
		if format.Hour != "" || format.Year != "" {
			t.Errorf("sparse request added fields: %#v", format)
		}
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
