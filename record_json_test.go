package gointl_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/agentable/go-intl/collator"
	"github.com/agentable/go-intl/datetimeformat"
	"github.com/agentable/go-intl/displaynames"
	"github.com/agentable/go-intl/durationformat"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
	"github.com/agentable/go-intl/segmenter"
)

func TestECMA402RecordJSONShapes(t *testing.T) {
	t.Parallel()

	loc := locale.MustParse("en-US-u-nu-latn")
	minFrac, maxFrac := 0, 3
	hour12 := false
	fractionalDigits := 2
	significantNumber := mustNumberFormat(t, numberformat.Options{MinimumSignificantDigits: intPtr(3)}).ResolvedOptions()
	significantPlural := mustPluralRules(t, pluralrules.Options{MinimumSignificantDigits: intPtr(2)}).ResolvedOptions()
	styleDateTime := mustDateTimeFormat(t, datetimeformat.Options{
		DateStyle: datetimeformat.MediumDateTimeStyle,
		TimeStyle: datetimeformat.ShortDateTimeStyle,
	}).ResolvedOptions()
	defaultCollator := mustCollator(t, collator.Options{}).ResolvedOptions()
	wordSegment := mustFirstSegment(t, "hello!", segmenter.WordGranularity)
	rtlTextInfo := locale.MustParse("ar").GetTextInfo()

	tests := []struct {
		name   string
		value  any
		want   []string
		absent []string
	}{
		{
			name:  "locale",
			value: loc,
			want:  []string{`"en-US-u-nu-latn"`},
		},
		{
			name: "number resolved options",
			value: numberformat.ResolvedOptions{
				Locale:                loc,
				NumberingSystem:       "latn",
				Style:                 numberformat.DecimalStyle,
				MinimumIntegerDigits:  1,
				MinimumFractionDigits: &minFrac,
				MaximumFractionDigits: &maxFrac,
				UseGrouping:           numberformat.UseGroupingAuto,
				Notation:              numberformat.StandardNotation,
				SignDisplay:           numberformat.AutoSignDisplay,
				RoundingIncrement:     1,
				RoundingMode:          numberformat.HalfExpandRoundingMode,
				RoundingPriority:      numberformat.AutoRoundingPriority,
				TrailingZeroDisplay:   numberformat.AutoTrailingZeroDisplay,
			},
			want: []string{
				`"locale":"en-US-u-nu-latn"`,
				`"minimumFractionDigits":0`,
				`"maximumFractionDigits":3`,
				`"trailingZeroDisplay":"auto"`,
			},
			absent: []string{`"currency"`, `"unit"`, `"minimumSignificantDigits"`},
		},
		{
			name:  "number significant digit resolved options",
			value: significantNumber,
			want: []string{
				`"minimumSignificantDigits":3`,
				`"maximumSignificantDigits":21`,
			},
			absent: []string{`"minimumFractionDigits"`, `"maximumFractionDigits"`},
		},
		{
			name:  "number range part",
			value: numberformat.RangePart{Type: numberformat.PartInteger, Value: "10", Source: numberformat.SourceStartRange},
			want:  []string{`"type":"integer"`, `"value":"10"`, `"source":"startRange"`},
		},
		{
			name: "datetime resolved options",
			value: datetimeformat.ResolvedOptions{
				Locale:          loc,
				Calendar:        "gregory",
				NumberingSystem: "latn",
				TimeZone:        "UTC",
				HourCycle:       datetimeformat.H23HourCycle,
				Hour12:          &hour12,
				Year:            datetimeformat.NumericFieldStyle,
				Month:           datetimeformat.ShortMonthStyle,
				Day:             datetimeformat.NumericFieldStyle,
			},
			want: []string{
				`"locale":"en-US-u-nu-latn"`,
				`"hour12":false`,
				`"month":"short"`,
			},
			absent: []string{`"weekday"`, `"fractionalSecondDigits"`},
		},
		{
			name:  "datetime style shortcut resolved options",
			value: styleDateTime,
			want:  []string{`"dateStyle":"medium"`, `"timeStyle":"short"`},
			absent: []string{
				`"year"`,
				`"month"`,
				`"day"`,
				`"hour"`,
				`"minute"`,
				`"second"`,
			},
		},
		{
			name: "plural resolved options",
			value: pluralrules.ResolvedOptions{
				Locale:                loc,
				Type:                  pluralrules.Cardinal,
				MinimumIntegerDigits:  1,
				MinimumFractionDigits: &minFrac,
				MaximumFractionDigits: &maxFrac,
				PluralCategories:      []pluralrules.Category{pluralrules.One, pluralrules.Other},
				Notation:              pluralrules.StandardNotation,
				CompactDisplay:        pluralrules.ShortCompactDisplay,
				RoundingIncrement:     1,
				RoundingMode:          pluralrules.HalfExpandRoundingMode,
				RoundingPriority:      pluralrules.AutoRoundingPriority,
				TrailingZeroDisplay:   pluralrules.AutoTrailingZeroDisplay,
			},
			want:   []string{`"type":"cardinal"`, `"pluralCategories":["one","other"]`},
			absent: []string{`"minimumSignificantDigits"`},
		},
		{
			name:  "plural significant digit resolved options",
			value: significantPlural,
			want: []string{
				`"minimumSignificantDigits":2`,
				`"maximumSignificantDigits":21`,
			},
			absent: []string{`"minimumFractionDigits"`, `"maximumFractionDigits"`},
		},
		{
			name:  "list part",
			value: listformat.Part{Type: listformat.PartElement, Value: "a"},
			want:  []string{`"type":"element"`, `"value":"a"`},
		},
		{
			name: "list resolved options",
			value: listformat.ResolvedOptions{
				Locale: loc,
				Type:   listformat.Conjunction,
				Style:  listformat.LongStyle,
			},
			want: []string{`"locale":"en-US-u-nu-latn"`, `"type":"conjunction"`, `"style":"long"`},
		},
		{
			name:   "relative time part omits empty unit",
			value:  relativetimeformat.Part{Type: relativetimeformat.PartLiteral, Value: "yesterday"},
			want:   []string{`"type":"literal"`, `"value":"yesterday"`},
			absent: []string{`"unit"`},
		},
		{
			name: "relative time resolved options",
			value: relativetimeformat.ResolvedOptions{
				Locale:          loc,
				Style:           relativetimeformat.LongStyle,
				Numeric:         relativetimeformat.NumericAlways,
				NumberingSystem: "latn",
			},
			want: []string{`"locale":"en-US-u-nu-latn"`, `"style":"long"`, `"numeric":"always"`, `"numberingSystem":"latn"`},
		},
		{
			name: "duration resolved options",
			value: durationformat.ResolvedOptions{
				Locale:              loc,
				NumberingSystem:     "latn",
				Style:               durationformat.ShortStyle,
				Years:               durationformat.ShortUnitStyle,
				YearsDisplay:        durationformat.AutoDisplay,
				Months:              durationformat.ShortUnitStyle,
				MonthsDisplay:       durationformat.AutoDisplay,
				Weeks:               durationformat.ShortUnitStyle,
				WeeksDisplay:        durationformat.AutoDisplay,
				Days:                durationformat.ShortUnitStyle,
				DaysDisplay:         durationformat.AutoDisplay,
				Hours:               durationformat.ShortUnitStyle,
				HoursDisplay:        durationformat.AutoDisplay,
				Minutes:             durationformat.ShortUnitStyle,
				MinutesDisplay:      durationformat.AutoDisplay,
				Seconds:             durationformat.ShortUnitStyle,
				SecondsDisplay:      durationformat.AutoDisplay,
				Milliseconds:        durationformat.ShortUnitStyle,
				MillisecondsDisplay: durationformat.AutoDisplay,
				Microseconds:        durationformat.ShortUnitStyle,
				MicrosecondsDisplay: durationformat.AutoDisplay,
				Nanoseconds:         durationformat.ShortUnitStyle,
				NanosecondsDisplay:  durationformat.AutoDisplay,
				FractionalDigits:    &fractionalDigits,
			},
			want: []string{`"fractionalDigits":2`, `"millisecondsDisplay":"auto"`},
		},
		{
			name: "duration record includes subsecond fields",
			value: durationformat.Duration{
				Seconds:      1,
				Milliseconds: 2,
				Microseconds: 3,
				Nanoseconds:  4,
			},
			want: []string{`"seconds":1`, `"milliseconds":2`, `"microseconds":3`, `"nanoseconds":4`},
		},
		{
			name: "display names resolved options omits language display",
			value: displaynames.ResolvedOptions{
				Locale:   loc,
				Style:    displaynames.LongStyle,
				Type:     displaynames.Region,
				Fallback: displaynames.CodeFallback,
			},
			want:   []string{`"type":"region"`, `"fallback":"code"`},
			absent: []string{`"languageDisplay"`},
		},
		{
			name: "collator resolved options omits empty collation",
			value: collator.ResolvedOptions{
				Locale:            loc,
				Usage:             collator.SortUsage,
				Sensitivity:       collator.VariantSensitivity,
				CaseFirst:         collator.FalseCaseFirst,
				Numeric:           true,
				IgnorePunctuation: false,
			},
			want:   []string{`"caseFirst":"false"`, `"numeric":true`},
			absent: []string{`"collation"`},
		},
		{
			name:  "collator resolved options includes default collation",
			value: defaultCollator,
			want:  []string{`"collation":"default"`},
		},
		{
			name:   "segment record uses code unit index",
			value:  segmenter.Segment{Segment: "🙂", CodeUnitIndex: 2, ByteIndex: 4, Input: "a🙂", IsWordLike: false},
			want:   []string{`"segment":"🙂"`, `"index":2`, `"input":"a🙂"`, `"isWordLike":false`},
			absent: []string{`"ByteIndex"`, `"byteIndex"`},
		},
		{
			name:  "segment word record includes word-like true",
			value: wordSegment,
			want:  []string{`"segment":"hello"`, `"index":0`, `"input":"hello!"`, `"isWordLike":true`},
		},
		{
			name: "segmenter resolved options",
			value: segmenter.ResolvedOptions{
				Locale:      loc,
				Granularity: segmenter.WordGranularity,
			},
			want: []string{`"locale":"en-US-u-nu-latn"`, `"granularity":"word"`},
		},
		{
			name:  "week info uses ECMA weekday numbers",
			value: locale.WeekInfo{FirstDay: time.Sunday, Weekend: []time.Weekday{time.Saturday, time.Sunday}},
			want:  []string{`"firstDay":7`, `"weekend":[6,7]`},
		},
		{
			name:  "text info",
			value: locale.TextInfo{Direction: "ltr"},
			want:  []string{`"direction":"ltr"`},
		},
		{
			name:  "text info from rtl locale",
			value: rtlTextInfo,
			want:  []string{`"direction":"rtl"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := mustMarshalJSON(t, tc.value)
			for _, want := range tc.want {
				if !strings.Contains(got, want) {
					t.Fatalf("json.Marshal(%s) = %s, want substring %s", tc.name, got, want)
				}
			}
			for _, absent := range tc.absent {
				if strings.Contains(got, absent) {
					t.Fatalf("json.Marshal(%s) = %s, did not want substring %s", tc.name, got, absent)
				}
			}
		})
	}
}

func TestDurationJSONAcceptsCamelCaseRecord(t *testing.T) {
	t.Parallel()

	var duration durationformat.Duration
	if err := json.Unmarshal([]byte(`{"hours":1,"milliseconds":2}`), &duration); err != nil {
		t.Fatal(err)
	}
	if duration.Hours != 1 || duration.Milliseconds != 2 {
		t.Fatalf("json.Unmarshal(Duration) = %+v, want hours=1 milliseconds=2", duration)
	}
}

func intPtr(v int) *int {
	return &v
}

func mustNumberFormat(t *testing.T, opts numberformat.Options) *numberformat.NumberFormat {
	t.Helper()

	format, err := numberformat.New(locale.List{locale.MustParse("en")}, opts)
	if err != nil {
		t.Fatalf("numberformat.New() error = %v", err)
	}
	return format
}

func mustDateTimeFormat(t *testing.T, opts datetimeformat.Options) *datetimeformat.DateTimeFormat {
	t.Helper()

	format, err := datetimeformat.New(locale.List{locale.MustParse("en-US")}, opts)
	if err != nil {
		t.Fatalf("datetimeformat.New() error = %v", err)
	}
	return format
}

func mustPluralRules(t *testing.T, opts pluralrules.Options) *pluralrules.PluralRules {
	t.Helper()

	rules, err := pluralrules.New(locale.List{locale.MustParse("en")}, opts)
	if err != nil {
		t.Fatalf("pluralrules.New() error = %v", err)
	}
	return rules
}

func mustCollator(t *testing.T, opts collator.Options) *collator.Collator {
	t.Helper()

	c, err := collator.New(locale.List{locale.MustParse("en")}, opts)
	if err != nil {
		t.Fatalf("collator.New() error = %v", err)
	}
	return c
}

func mustFirstSegment(t *testing.T, input string, granularity segmenter.Granularity) segmenter.Segment {
	t.Helper()

	s, err := segmenter.New(locale.List{locale.MustParse("en")}, segmenter.Options{Granularity: granularity})
	if err != nil {
		t.Fatalf("segmenter.New() error = %v", err)
	}
	for segment := range s.Segment(input).All() {
		return segment
	}
	t.Fatalf("Segment(%q).All() returned no records", input)
	return segmenter.Segment{}
}

func mustMarshalJSON(t *testing.T, value any) string {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal(%T) error = %v", value, err)
	}
	return string(data)
}
