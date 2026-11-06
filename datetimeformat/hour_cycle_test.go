package datetimeformat

import (
	"encoding/json"
	"encoding/json/jsontext"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/internal/testcontract"
)

func TestStyleHourCycles(t *testing.T) {
	t.Parallel()
	for _, loc := range []string{"en-US", "en-GB", "ja-JP", "fr-CA"} {
		for _, cycle := range []string{"h11", "h12", "h23", "h24"} {
			for _, style := range []string{"short", "medium"} {
				for _, dateStyle := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/%s/%t", loc, cycle, style, dateStyle), func(t *testing.T) {
						t.Parallel()
						opts := Options{TimeZone: new("UTC"), TimeStyle: new(style), HourCycle: new(cycle)}
						if dateStyle {
							opts.DateStyle = new("short")
						}
						f, err := New(intltest.LocaleList(t, loc), opts)
						if err != nil {
							t.Fatal(err)
						}
						record, err := json.Marshal(f.ResolvedOptions())
						if err != nil {
							t.Fatal(err)
						}
						var values map[string]any
						if err := json.Unmarshal(record, &values); err != nil {
							t.Fatal(err)
						}
						for _, key := range []string{"hour", "minute", "second", "dayPeriod"} {
							if _, ok := values[key]; ok {
								t.Errorf("style reports component %s", key)
							}
						}
						if values["hourCycle"] != cycle {
							t.Errorf("record = %s", record)
						}
						midnight := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
						assertHourCycle(t, f, cycle, midnight)
						assertHourCycle(t, f, cycle, midnight.Add(15*time.Hour))
						assertCycleRange(t, f, cycle, midnight, midnight.Add(3*time.Hour))
						assertCycleRange(t, f, cycle, midnight, midnight.Add(27*time.Hour))
					})
				}
			}
		}
	}
}

func TestHourCycleForSelectedPattern(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		tag    string
		opts   Options
		cycle  string
		noHour bool
	}{
		{"minute h11", "en", Options{Minute: new("numeric"), HourCycle: new("h11")}, "h11", false},
		{"minute h12", "en", Options{Minute: new("numeric"), HourCycle: new("h12")}, "h12", false},
		{"minute h23", "en", Options{Minute: new("numeric"), HourCycle: new("h23")}, "h23", false},
		{"minute h24", "en", Options{Minute: new("numeric"), HourCycle: new("h24")}, "h24", false},
		{"hour12 overrides hc", "en-u-hc-h24", Options{Minute: new("numeric"), Hour12: new(true)}, "h12", false},
		{"date components", "en", Options{Year: new("numeric"), HourCycle: new("h24")}, "", true},
		{"date style", "en", Options{DateStyle: new("long"), HourCycle: new("h24")}, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts := tc.opts
			opts.TimeZone = new("UTC")
			f, err := New(intltest.LocaleList(t, tc.tag), opts)
			if err != nil {
				t.Fatal(err)
			}
			r := f.ResolvedOptions()
			midnight := time.Date(2026, 5, 8, 0, 7, 0, 0, time.UTC)
			parts, err := f.FormatToParts(midnight)
			if err != nil {
				t.Fatal(err)
			}
			hasHour := false
			for _, p := range parts {
				hasHour = hasHour || p.Type == PartHour
			}
			if hasHour != (r.Hour != nil) || (tc.noHour && hasHour) {
				t.Fatalf("hour presence disagrees: resolved=%#v parts=%v", r, parts)
			}
			if tc.opts.Hour12 != nil && strings.Contains(r.Locale.String(), "hc-") {
				t.Errorf("hour12 retained overridden hc: %s", r.Locale)
			}
			if !hasHour {
				if r.HourCycle != nil || r.Hour12 != nil {
					t.Errorf("hourCycle/hour12 present without hour: %#v", r)
				}
				return
			}
			want12 := tc.cycle == "h11" || tc.cycle == "h12"
			if r.HourCycle == nil || string(*r.HourCycle) != tc.cycle || r.Hour12 == nil || *r.Hour12 != want12 {
				t.Errorf("resolved = %#v, want %s/%t", r, tc.cycle, want12)
			}
			assertHourCycle(t, f, tc.cycle, midnight)
			assertHourCycle(t, f, tc.cycle, midnight.Add(21*time.Hour))
			assertCycleRange(t, f, tc.cycle, midnight, midnight.Add(3*time.Hour))
			assertCycleRange(t, f, tc.cycle, midnight, midnight.Add(27*time.Hour))
		})
	}
}

func assertCycleRange(t *testing.T, f *DateTimeFormat, cycle string, start, end time.Time) {
	t.Helper()
	parts, err := f.FormatRangeToParts(start, end)
	if err != nil {
		t.Fatal(err)
	}
	joined := ""
	var hours, periods int
	for _, p := range parts {
		joined += p.Value
		if p.Source != SourceStartRange && p.Source != SourceEndRange && p.Source != SourceShared {
			t.Errorf("invalid source %q", p.Source)
		}
		if p.Type == PartDayPeriod {
			periods++
		}
		if p.Type != PartHour {
			continue
		}
		hours++
		instant := start
		if p.Source == SourceEndRange {
			instant = end
		}
		want := instant.Hour()
		if cycle == "h11" || cycle == "h12" {
			want %= 12
		}
		if cycle == "h12" && want == 0 {
			want = 12
		}
		if cycle == "h24" && want == 0 {
			want = 24
		}
		if got, err := strconv.Atoi(p.Value); err != nil || got != want {
			t.Errorf("range hour = %q (%s), want %d: %v", p.Value, p.Source, want, parts)
		}
		if p.Source == SourceShared {
			t.Errorf("different hours marked shared: %v", parts)
		}
	}
	if hours != 2 {
		t.Errorf("range hour count %d: %v", hours, parts)
	}
	if want12 := cycle == "h11" || cycle == "h12"; (periods > 0) != want12 {
		t.Errorf("range day periods = %d for %s: %v", periods, cycle, parts)
	}
	if text, err := f.FormatRange(start, end); err != nil || text != joined {
		t.Errorf("range %q, %v differs from parts %q", text, err, joined)
	}
}

func TestExplicitHourCycles(t *testing.T) {
	t.Parallel()
	for _, loc := range []string{"en-US", "en-GB", "fr-CA"} {
		for _, cycle := range []string{"h11", "h12", "h23", "h24"} {
			for _, matcher := range []string{"basic", "best fit"} {
				for _, extension := range []bool{false, true} {
					t.Run(fmt.Sprintf("%s/%s/%s/%t", loc, cycle, matcher, extension), func(t *testing.T) {
						t.Parallel()
						opts := Options{TimeZone: new("UTC"), Hour: new("numeric"), FormatMatcher: new(matcher)}
						tag := loc
						if extension {
							tag += "-u-hc-" + cycle
						} else {
							opts.HourCycle = new(cycle)
						}
						f, err := New(intltest.LocaleList(t, tag), opts)
						if err != nil {
							t.Fatal(err)
						}
						resolved := f.ResolvedOptions()
						if resolved.HourCycle == nil || string(*resolved.HourCycle) != cycle {
							t.Errorf("hourCycle = %v, want %s", resolved.HourCycle, cycle)
						}
						if extension && !strings.Contains(resolved.Locale.String(), "hc-"+cycle) {
							t.Errorf("locale = %s, want hc retained", resolved.Locale)
						}
						assertHourCycle(t, f, cycle, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
					})
				}
			}
		}
	}
	for _, optionCycle := range []string{"h11", "h24"} {
		f, err := New(intltest.LocaleList(t, "en-US-u-hc-h23"), Options{TimeZone: new("UTC"), Hour: new("numeric"), HourCycle: new(optionCycle)})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(f.ResolvedOptions().Locale.String(), "hc-") {
			t.Errorf("overridden hc retained: %s", f.ResolvedOptions().Locale)
		}
		assertHourCycle(t, f, optionCycle, time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	}
}

func TestBasicHourCyclePatterns(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ locale, cycle, hour, minute string }{
		{"en-US", "h11", "0 AM", "0:07 AM"},
		{"en-GB", "h12", "12 am", "12:07 am"},
		{"fr-CA", "h12", "12 h a.m.", "12 h 07 a.m."},
		{"ja-JP", "h12", "午前12時", "午前12:07"},
		{"en-US", "h23", "00", "00:07"},
	} {
		t.Run(tc.locale+"/"+tc.cycle, func(t *testing.T) {
			t.Parallel()
			for _, withMinute := range []bool{false, true} {
				opts := Options{Hour: new("numeric"), HourCycle: new(tc.cycle), FormatMatcher: new("basic"), TimeZone: new("UTC")}
				want := tc.hour
				if withMinute {
					opts.Minute = new("numeric")
					want = tc.minute
				}
				f, err := New(intltest.LocaleList(t, tc.locale), opts)
				if err != nil {
					t.Fatal(err)
				}
				start := time.Date(2026, 1, 1, 0, 7, 0, 0, time.UTC)
				if got := mustDateFormat(t, f, start); got != want {
					t.Errorf("minute=%t: Format = %q, want %q", withMinute, got, want)
				}
				if got := joinPartValues(mustDateFormatToParts(t, f, start)); got != want {
					t.Errorf("minute=%t: parts = %q, want %q", withMinute, got, want)
				}
				assertCycleRange(t, f, tc.cycle, start, start.Add(3*time.Hour))
				assertCycleRange(t, f, tc.cycle, start, start.Add(27*time.Hour))
			}
		})
	}
}

func TestBasicHourCyclePreservesFlexibleDayPeriod(t *testing.T) {
	t.Parallel()
	for _, cycle := range []string{"h23", "h24"} {
		t.Run(cycle, func(t *testing.T) {
			f, err := New(intltest.LocaleList(t, "en-US"), Options{
				FormatMatcher: new("basic"), TimeZone: new("UTC"),
				HourCycle: new(cycle), Hour: new("numeric"),
				Minute: new("2-digit"), Second: new("2-digit"), DayPeriod: new("short"),
			})
			if err != nil {
				t.Fatal(err)
			}
			if got := f.ResolvedOptions().DayPeriod; got == nil || *got != ShortFieldStyle {
				t.Errorf("resolved dayPeriod = %v, want short", got)
			}
			start := time.Date(2026, 1, 1, 13, 7, 8, 0, time.UTC)
			want := "13:07:08 in the afternoon"
			if got := mustDateFormat(t, f, start); got != want {
				t.Errorf("Format = %q, want %q", got, want)
			}
			if got := joinPartValues(mustDateFormatToParts(t, f, start)); got != want {
				t.Errorf("parts = %q, want %q", got, want)
			}
			for _, delta := range []time.Duration{time.Hour, 25 * time.Hour} {
				end := start.Add(delta)
				parts := mustDateFormatRangeToParts(t, f, start, end)
				var joined strings.Builder
				periods := 0
				for _, part := range parts {
					joined.WriteString(part.Value)
					if part.Type == PartDayPeriod {
						periods++
						if part.Value != "in the afternoon" {
							t.Errorf("range dayPeriod = %q", part.Value)
						}
					}
				}
				if periods == 0 {
					t.Errorf("range %v omitted flexible dayPeriod: %v", delta, parts)
				}
				if got := mustDateFormatRange(t, f, start, end); got != joined.String() {
					t.Errorf("range text = %q, parts = %q", got, joined.String())
				}
			}
		})
	}
}

func TestLocaleHour12Preference(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		locale string
		hour12 bool
		cycle  string
	}{
		{"fr-CA", true, "h12"}, {"fr-CA", false, "h23"}, {"ja", true, "h11"}, {"ja-JP", true, "h11"}, {"en-US", true, "h12"},
		{"ja", false, "h23"}, {"en-US", false, "h23"}, {"fr-FR", false, "h23"},
	} {
		for _, style := range []bool{false, true} {
			for _, extension := range []string{"", "-u-hc-h11", "-u-hc-h12", "-u-hc-h23", "-u-hc-h24"} {
				t.Run(fmt.Sprintf("%s/%t/%t/%s", tc.locale, tc.hour12, style, extension), func(t *testing.T) {
					t.Parallel()
					opts := Options{TimeZone: new("UTC"), Hour12: new(tc.hour12), HourCycle: new("h24")}
					if style {
						opts.TimeStyle = new("short")
					} else {
						opts.Hour = new("numeric")
						opts.Minute = new("2-digit")
					}
					f, err := New(intltest.LocaleList(t, tc.locale+extension), opts)
					if err != nil {
						t.Fatal(err)
					}
					r := f.ResolvedOptions()
					if r.HourCycle == nil || string(*r.HourCycle) != tc.cycle || r.Hour12 == nil || *r.Hour12 != tc.hour12 {
						t.Errorf("resolved = %#v, want %s/%t", r, tc.cycle, tc.hour12)
					}
					if strings.Contains(r.Locale.String(), "hc-") {
						t.Errorf("hour12 retained overridden hc: %s", r.Locale)
					}
					if tc.locale == "ja" && r.Locale.String() != "ja" {
						t.Errorf("maximization leaked into locale: %s", r.Locale)
					}
					midnight := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
					assertHourCycle(t, f, tc.cycle, midnight)
					assertCycleRange(t, f, tc.cycle, midnight, midnight.Add(3*time.Hour))
				})
			}
		}
	}
}

func assertHourCycle(t *testing.T, f *DateTimeFormat, cycle string, instant time.Time) {
	t.Helper()
	want := instant.Hour()
	if cycle == "h11" || cycle == "h12" {
		want %= 12
	}
	if cycle == "h12" && want == 0 {
		want = 12
	}
	if cycle == "h24" && want == 0 {
		want = 24
	}
	parts, err := f.FormatToParts(instant)
	if err != nil {
		t.Fatal(err)
	}
	var joined string
	var hours, periods int
	for _, p := range parts {
		joined += p.Value
		if p.Type == PartDayPeriod {
			periods++
		}
		if p.Type == PartHour {
			hours++
			if got, err := strconv.Atoi(p.Value); err != nil || got != want {
				t.Errorf("hour = %q, want %d (%s)", p.Value, want, cycle)
			}
			if width := f.ResolvedOptions().Hour; width != nil && ((*width == TwoDigitFieldStyle) != (len(p.Value) == 2)) && want < 10 {
				t.Errorf("hour width %q disagrees with %s", p.Value, *width)
			}
		}
	}
	if hours != 1 {
		t.Errorf("hours = %d, parts %v", hours, parts)
	}
	if want12 := cycle == "h11" || cycle == "h12"; (periods > 0) != want12 {
		t.Errorf("day periods = %d for %s: %v", periods, cycle, parts)
	}
	if text, err := f.Format(instant); err != nil || text != joined {
		t.Errorf("text %q, %v differs from parts %q", text, err, joined)
	}
}

func TestLanguageRegionDefaultHourCycle(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, tag, cycle, resolved string
		options                    Options
	}{
		{"French components", "fr-CA", "h23", `{"locale":"fr-CA","calendar":"gregory","numberingSystem":"latn","timeZone":"UTC","hourCycle":"h23","hour12":false,"hour":"2-digit"}`, Options{Hour: new("numeric")}},
		{"English components", "en-CA", "h12", `{"locale":"en-CA","calendar":"gregory","numberingSystem":"latn","timeZone":"UTC","hourCycle":"h12","hour12":true,"hour":"numeric"}`, Options{Hour: new("numeric")}},
		{"French style", "fr-CA", "h23", `{"locale":"fr-CA","calendar":"gregory","numberingSystem":"latn","timeZone":"UTC","hourCycle":"h23","hour12":false,"timeStyle":"short"}`, Options{TimeStyle: new("short")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			opts := tc.options
			opts.TimeZone = new("UTC")
			f, err := New(intltest.LocaleList(t, tc.tag), opts)
			if err != nil {
				t.Fatal(err)
			}
			testcontract.AssertResolvedOptionsJSON(t, f.ResolvedOptions(), jsontext.Value(tc.resolved))
			midnight := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			assertHourCycle(t, f, tc.cycle, midnight)
			assertCycleRange(t, f, tc.cycle, midnight, midnight.Add(3*time.Hour))
			assertCycleRange(t, f, tc.cycle, midnight, midnight.Add(27*time.Hour))
		})
	}
}
