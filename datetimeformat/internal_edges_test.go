package datetimeformat

import (
	"reflect"
	"testing"
	"time"

	cldrdate "github.com/agentable/go-intl/internal/cldr/date"
	ecma402dtf "github.com/agentable/go-intl/internal/ecma402/datetimeformat"
	"github.com/agentable/go-intl/locale"
)

func TestDateTimeFormatPatternLiteralEdges(t *testing.T) {
	t.Parallel()

	parts := appendLiteralPart(nil, "")
	if len(parts) != 0 {
		t.Fatalf("appendLiteralPart(empty) = %#v, want no part", parts)
	}
	parts = appendLiteralPart([]Part{{Type: PartLiteral, Value: "a"}}, "b")
	if len(parts) != 1 || parts[0].Value != "ab" {
		t.Fatalf("appendLiteralPart(merge) = %#v, want one merged literal", parts)
	}
	literal, rest := consumeQuotedPatternLiteral("'it''s' y")
	if literal != "it's" || rest != " y" {
		t.Fatalf("consumeQuotedPatternLiteral() = %q, %q; want it's, space-y", literal, rest)
	}
	literal, rest = consumeQuotedPatternLiteral("'unterminated")
	if literal != "unterminated" || rest != "" {
		t.Fatalf("consumeQuotedPatternLiteral(unterminated) = %q, %q", literal, rest)
	}

	hour12 := false
	format := DateTimeFormat{resolved: ResolvedOptions{Hour12: &hour12, NumberingSystem: "latn"}}
	got := format.formatPattern("H a", gregoryLocalTime(time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC)))
	want := []Part{{Type: PartHour, Value: "9"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("formatPattern(24h day period) = %#v, want trailing literal trimmed", got)
	}

	empty := DateTimeFormat{}
	if got := empty.FormatToParts(time.Date(2026, time.May, 8, 0, 0, 0, 0, time.UTC)); got != nil {
		t.Fatalf("FormatToParts() without selected pattern = %#v, want nil", got)
	}
}

func TestDateTimeFormatUsesStandardDateTimePatternWhenAtPatternMissing(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{DateStyle: MediumDateTimeStyle, TimeStyle: ShortDateTimeStyle, Hour12: boolPtr(false)})
	if err != nil {
		t.Fatalf("New(dateStyle+timeStyle) error = %v", err)
	}
	format.gregorian.DateTimeAtFormats = [4]string{}
	format.pattern = format.selectPattern()
	got := format.Format(time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC))
	if got != "May 8, 2026, 09:07" {
		t.Fatalf("Format() with standard datetime fallback = %q, want May 8, 2026, 09:07", got)
	}
}

func TestDateTimeFormatShortComponentDateTimePattern(t *testing.T) {
	t.Parallel()

	format, err := New(locale.List{locale.MustParse("en-US")}, Options{
		Year:   TwoDigitFieldStyle,
		Month:  NumericMonthStyle,
		Day:    NumericFieldStyle,
		Hour:   NumericFieldStyle,
		Minute: TwoDigitFieldStyle,
		Hour12: boolPtr(true),
	})
	if err != nil {
		t.Fatalf("New(short date+time fields) error = %v", err)
	}
	got := format.Format(time.Date(2026, time.May, 8, 9, 7, 0, 0, time.UTC))
	if got != "5/8/26, 9:07\u202fAM" {
		t.Fatalf("Format(short date+time fields) = %q, want short connector pattern", got)
	}
}

func TestDateTimeFormatIntervalAndTimeZoneHelpers(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		field rune
		want  rune
	}{
		{field: 'L', want: 'M'},
		{field: 'e', want: 'E'},
		{field: 'c', want: 'E'},
		{field: 'H', want: 'h'},
		{field: 'K', want: 'h'},
		{field: 'k', want: 'h'},
		{field: 'b', want: 'a'},
		{field: 'B', want: 'a'},
		{field: 'Z', want: 'z'},
		{field: 'O', want: 'z'},
		{field: 'V', want: 'z'},
		{field: 'X', want: 'z'},
		{field: 'x', want: 'z'},
		{field: 'y', want: 'y'},
	} {
		if got := intervalFieldKey(tc.field); got != tc.want {
			t.Fatalf("intervalFieldKey(%q) = %q, want %q", tc.field, got, tc.want)
		}
	}

	for _, tc := range []struct {
		style ecma402dtf.TimeZoneName
		want  string
	}{
		{style: ecma402dtf.TimeZoneNameShort, want: "z"},
		{style: ecma402dtf.TimeZoneNameLong, want: "zzzz"},
		{style: ecma402dtf.TimeZoneNameShortOffset, want: "O"},
		{style: ecma402dtf.TimeZoneNameLongOffset, want: "OOOO"},
		{style: ecma402dtf.TimeZoneNameShortGeneric, want: "v"},
		{style: ecma402dtf.TimeZoneNameLongGeneric, want: "vvvv"},
		{style: ecma402dtf.TimeZoneName("unknown"), want: "z"},
	} {
		if got := timeZonePatternField(tc.style); got != tc.want {
			t.Fatalf("timeZonePatternField(%q) = %q, want %q", tc.style, got, tc.want)
		}
	}

	if got := skipQuotedPatternLiteral("'z''z' x", 0); got != len("'z''z'") {
		t.Fatalf("skipQuotedPatternLiteral(escaped) = %d, want closing index", got)
	}
	if got := skipQuotedPatternLiteral("'unterminated", 0); got != len("'unterminated") {
		t.Fatalf("skipQuotedPatternLiteral(unterminated) = %d, want len", got)
	}
}

func TestDateTimeFormatSmallFormattingHelpers(t *testing.T) {
	t.Parallel()

	gregorian := cldrdate.Gregorian{
		DateFormats:     [4]string{"date-full", "date-long", "date-medium", "date-short"},
		TimeFormats:     [4]string{"time-full", "time-long", "time-medium", "time-short"},
		DateTimeFormats: [4]string{"both-full", "both-long", "both-medium", "both-short"},
	}
	if got := dateStylePattern(gregorian, LongDateTimeStyle); got != "date-long" {
		t.Fatalf("dateStylePattern(long) = %q, want date-long", got)
	}
	if got := timeStylePattern(gregorian, FullDateTimeStyle); got != "time-full" {
		t.Fatalf("timeStylePattern(full) = %q, want time-full", got)
	}
	if got := dateTimeStandardStylePattern(gregorian, ShortDateTimeStyle); got != "both-short" {
		t.Fatalf("dateTimeStandardStylePattern(short) = %q, want both-short", got)
	}
	if got := dateTimeStandardStylePattern(gregorian, Style("unknown")); got != "" {
		t.Fatalf("dateTimeStandardStylePattern(unknown) = %q, want empty", got)
	}
	if got := fractionalSecondValue(123456789, 12); got != "123456789" {
		t.Fatalf("fractionalSecondValue(width>9) = %q, want 123456789", got)
	}

	format := DateTimeFormat{resolved: ResolvedOptions{NumberingSystem: "latn"}}
	format.gregorian.DayPeriods.AM.Abbr = "AM"
	morning := gregoryLocalTime(time.Date(2026, time.May, 8, 9, 0, 0, 0, time.UTC))
	afternoon := gregoryLocalTime(time.Date(2026, time.May, 8, 13, 0, 0, 0, time.UTC))
	if got := format.dayPeriodPatternName(5, morning); got != "AM" {
		t.Fatalf("dayPeriodPatternName(narrow fallback) = %q, want AM", got)
	}
	if got := format.dayPeriodPatternName(4, afternoon); got != "PM" {
		t.Fatalf("dayPeriodPatternName(pm default fallback) = %q, want PM", got)
	}
	if got := format.flexibleDayPeriodPatternName(4, morning); got != "AM" {
		t.Fatalf("flexibleDayPeriodPatternName(fallback) = %q, want AM", got)
	}

	datePart := format.datePatternPart('Q', 2, morning)
	if datePart.Type != PartLiteral || datePart.Value != "QQ" {
		t.Fatalf("datePatternPart(unknown) = %#v, want literal QQ", datePart)
	}
	timePart, ok := format.timePatternPart('?', 2, morning)
	if !ok || timePart.Type != PartLiteral || timePart.Value != "??" {
		t.Fatalf("timePatternPart(unknown) = %#v, %v; want literal ??", timePart, ok)
	}
}

func TestDateTimeFormatFallbackRangeParts(t *testing.T) {
	t.Parallel()

	start := []Part{{Type: PartHour, Value: "9"}}
	end := []Part{{Type: PartHour, Value: "10"}}
	format := DateTimeFormat{gregorian: cldrdate.Gregorian{IntervalFallback: "{1} <- {0}"}}
	got := format.fallbackRangeParts(start, end)
	want := []RangePart{
		{Type: PartHour, Value: "10", Source: SourceEndRange},
		{Type: PartLiteral, Value: " <- ", Source: SourceShared},
		{Type: PartHour, Value: "9", Source: SourceStartRange},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("fallbackRangeParts(custom) = %#v, want %#v", got, want)
	}

	format.gregorian.IntervalFallback = ""
	got = format.fallbackRangeParts(start, end)
	if joined := joinRangeParts(got); joined != "9 – 10" {
		t.Fatalf("fallbackRangeParts(default) = %q, want 9 – 10", joined)
	}
}
