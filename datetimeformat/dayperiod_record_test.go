package datetimeformat

import (
	"encoding/json/v2"
	"slices"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intltest"
)

func TestDayPeriodRecordAndParts(t *testing.T) {
	t.Parallel()

	morning := time.Date(2020, 1, 1, 1, 0, 0, 0, time.UTC)
	afternoon := morning.Add(12 * time.Hour)
	tests := []struct {
		name  string
		opts  Options
		width *string
		parts []Part
	}{
		{"automatic hour", Options{Hour: new("numeric")}, nil, []Part{{Type: PartHour, Value: "1"}, {Type: PartLiteral, Value: " "}, {Type: PartDayPeriod, Value: "AM"}}},
		{"automatic style", Options{TimeStyle: new("short")}, nil, []Part{{Type: PartHour, Value: "1"}, {Type: PartLiteral, Value: ":"}, {Type: PartMinute, Value: "00"}, {Type: PartLiteral, Value: " "}, {Type: PartDayPeriod, Value: "AM"}}},
		{"flexible long", Options{Hour: new("numeric"), DayPeriod: new("long")}, new("long"), []Part{{Type: PartHour, Value: "1"}, {Type: PartLiteral, Value: " "}, {Type: PartDayPeriod, Value: "in the morning"}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			opts := tc.opts
			opts.TimeZone = new("UTC")
			f, err := New(intltest.LocaleList(t, "en-US"), opts)
			if err != nil {
				t.Fatal(err)
			}
			data, err := json.Marshal(f.ResolvedOptions())
			if err != nil {
				t.Fatal(err)
			}
			var record map[string]any
			if err := json.Unmarshal(data, &record); err != nil {
				t.Fatal(err)
			}
			gotWidth, present := record["dayPeriod"]
			if tc.width == nil && present {
				t.Fatalf("automatic AM/PM leaked into resolved dayPeriod: %s", data)
			}
			if tc.width != nil && (!present || gotWidth != *tc.width) {
				t.Fatalf("dayPeriod = %v, present=%v, want %q", gotWidth, present, *tc.width)
			}
			parts := mustDateFormatToParts(t, f, morning)
			if !slices.Equal(parts, tc.parts) {
				t.Fatalf("FormatToParts = %#v, want %#v", parts, tc.parts)
			}
			var joined string
			for _, part := range parts {
				joined += part.Value
			}
			if got := mustDateFormat(t, f, morning); got != joined {
				t.Fatalf("Format = %q, parts join = %q", got, joined)
			}
			rangeParts := mustDateFormatRangeToParts(t, f, morning, afternoon)
			var rangeText string
			var periods []RangePart
			for _, part := range rangeParts {
				rangeText += part.Value
				if part.Type == PartDayPeriod {
					periods = append(periods, part)
				}
			}
			wantPeriods := []RangePart{{Type: PartDayPeriod, Value: "AM", Source: SourceStartRange}, {Type: PartDayPeriod, Value: "PM", Source: SourceEndRange}}
			if tc.width != nil {
				wantPeriods[0].Value = "in the morning"
				wantPeriods[1].Value = "in the afternoon"
			}
			if !slices.Equal(periods, wantPeriods) {
				t.Fatalf("range dayPeriod parts = %#v, want %#v", periods, wantPeriods)
			}
			if got := mustDateFormatRange(t, f, morning, afternoon); got != rangeText {
				t.Fatalf("FormatRange = %q, parts join = %q", got, rangeText)
			}
		})
	}
}

func TestDayPeriodDisplayPrecision(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                       string
		opts                       Options
		hour, minute, second, nano int
		want                       string
	}{
		{"hour ignores minute", Options{Hour: new("numeric")}, 12, 1, 0, 0, "12 noon"},
		{"minute ignores second", Options{Hour: new("numeric"), Minute: new("2-digit")}, 12, 0, 1, 0, "12:00 noon"},
		{"second before noon", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit")}, 11, 59, 59, 0, "11:59:59 in the morning"},
		{"exact noon", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit")}, 12, 0, 0, 0, "12:00:00 noon"},
		{"second after noon", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit")}, 12, 0, 1, 0, "12:00:01 in the afternoon"},
		{"last second", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit")}, 12, 0, 59, 0, "12:00:59 in the afternoon"},
		{"minute after noon", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit")}, 12, 1, 0, 0, "12:01:00 in the afternoon"},
		{"only period", Options{}, 12, 1, 1, 0, "noon"},
		{"fraction one", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit"), FractionalSecondDigits: new(1)}, 12, 0, 0, 100000000, "12:00:00.1 noon"},
		{"fraction two", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit"), FractionalSecondDigits: new(2)}, 12, 0, 0, 10000000, "12:00:00.01 noon"},
		{"fraction three", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit"), FractionalSecondDigits: new(3)}, 12, 0, 0, 1000000, "12:00:00.001 noon"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := tc.opts
			opts.TimeZone = new("UTC")
			opts.DayPeriod = new("long")
			f, err := New(intltest.LocaleList(t, "en"), opts)
			if err != nil {
				t.Fatal(err)
			}
			instant := time.Date(2026, 1, 1, tc.hour, tc.minute, tc.second, tc.nano, time.UTC)
			got := mustDateFormat(t, f, instant)
			if got != tc.want {
				t.Errorf("Format = %q, want %q", got, tc.want)
			}
			var joined string
			for _, part := range mustDateFormatToParts(t, f, instant) {
				joined += part.Value
			}
			if joined != tc.want {
				t.Errorf("parts join = %q, want %q", joined, tc.want)
			}
		})
	}
}

func TestDayPeriodPrecisionRanges(t *testing.T) {
	t.Parallel()
	noon := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name       string
		opts       Options
		start, end time.Time
		want       []RangePart
	}{
		{"forward", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit")}, noon, noon.Add(time.Second), []RangePart{{Type: PartDayPeriod, Value: "noon", Source: SourceStartRange}, {Type: PartDayPeriod, Value: "in the afternoon", Source: SourceEndRange}}},
		{"reverse", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit")}, noon.Add(time.Second), noon, []RangePart{{Type: PartDayPeriod, Value: "in the afternoon", Source: SourceStartRange}, {Type: PartDayPeriod, Value: "noon", Source: SourceEndRange}}},
		{"equal", Options{Hour: new("numeric"), Minute: new("2-digit"), Second: new("2-digit")}, noon, noon, []RangePart{{Type: PartDayPeriod, Value: "noon", Source: SourceShared}}},
		{"hidden minute", Options{Hour: new("numeric")}, noon, noon.Add(time.Minute), []RangePart{{Type: PartDayPeriod, Value: "noon", Source: SourceShared}}},
		{"shared hour", Options{Hour: new("numeric"), Minute: new("2-digit")}, noon.Add(time.Minute), noon.Add(2 * time.Minute), []RangePart{{Type: PartDayPeriod, Value: "in the afternoon", Source: SourceShared}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			opts := tc.opts
			opts.TimeZone = new("UTC")
			opts.DayPeriod = new("long")
			f, err := New(intltest.LocaleList(t, "en"), opts)
			if err != nil {
				t.Fatal(err)
			}
			var text string
			var periods []RangePart
			for _, p := range mustDateFormatRangeToParts(t, f, tc.start, tc.end) {
				text += p.Value
				if p.Type == PartDayPeriod {
					periods = append(periods, p)
				}
			}
			if !slices.Equal(periods, tc.want) {
				t.Errorf("periods = %#v, want %#v", periods, tc.want)
			}
			if got := mustDateFormatRange(t, f, tc.start, tc.end); got != text {
				t.Errorf("FormatRange = %q, parts join = %q", got, text)
			}
		})
	}
}

func TestDayPeriodSelectedFieldsAndLocales(t *testing.T) {
	t.Parallel()
	for _, tag := range []string{"en", "fr"} {
		for _, opts := range []Options{
			{}, {Minute: new("numeric")}, {Second: new("numeric")}, {Hour: new("numeric"), Second: new("numeric")},
		} {
			opts.TimeZone = new("UTC")
			opts.DayPeriod = new("long")
			if tag == "fr" {
				opts.Hour = new("numeric")
				opts.Hour12 = new(true)
			}
			f, err := New(intltest.LocaleList(t, tag), opts)
			if err != nil {
				t.Fatal(err)
			}
			for _, clock := range [][3]int{{12, 0, 0}, {12, 1, 1}, {0, 0, 0}, {0, 0, 1}, {23, 59, 59}} {
				instant := time.Date(2026, 1, 1, clock[0], clock[1], clock[2], 0, time.UTC)
				minute, second := clock[1], clock[2]
				// Presence in the selected full program, not in the requested options.
				var hasMinute, hasSecond bool
				for _, token := range f.pattern.timeProgram {
					hasMinute = hasMinute || token.field == 'm'
					hasSecond = hasSecond || token.field == 's'
				}
				if !hasMinute {
					minute = 0
				}
				if !hasSecond {
					second = 0
				}
				want := "in the afternoon"
				if tag == "fr" {
					want = "de l’après-midi"
				}
				if clock[0] == 12 && minute == 0 && second == 0 {
					want = "noon"
					if tag == "fr" {
						want = "midi"
					}
				}
				if clock[0] == 0 {
					want = "in the morning"
					if tag == "fr" {
						want = "du matin"
					}
					if minute == 0 && second == 0 {
						want = "midnight"
						if tag == "fr" {
							want = "minuit"
						}
					}
				}
				if clock[0] == 23 {
					want = "at night"
					if tag == "fr" {
						want = "du soir"
					}
				}
				var got string
				for _, part := range mustDateFormatToParts(t, f, instant) {
					if part.Type == PartDayPeriod {
						got = part.Value
					}
				}
				if got != want {
					t.Errorf("%s %s %v: period = %q, want %q", tag, f.pattern.time, clock, got, want)
				}
			}
		}
	}
}
