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
	"github.com/agentable/go-intl/internal/intltest"
	"github.com/agentable/go-intl/listformat"
	"github.com/agentable/go-intl/locale"
	"github.com/agentable/go-intl/numberformat"
	"github.com/agentable/go-intl/pluralrules"
	"github.com/agentable/go-intl/relativetimeformat"
	"github.com/agentable/go-intl/segmenter"
)

func TestECMA402RecordJSONShapes(t *testing.T) {
	t.Parallel()

	loc := intltest.Locale(t, "en-US-u-nu-latn")
	hour12 := false
	defaultNumber := mustNumberFormat(t, numberformat.Options{}).ResolvedOptions()
	significantNumber := mustNumberFormat(t, numberformat.Options{MinimumSignificantDigits: intPtr(3)}).ResolvedOptions()
	defaultPlural := mustPluralRules(t, pluralrules.Options{}).ResolvedOptions()
	significantPlural := mustPluralRules(t, pluralrules.Options{MinimumSignificantDigits: intPtr(2)}).ResolvedOptions()
	compactPlural := mustPluralRules(t, pluralrules.Options{
		Notation:       stringPtr(pluralrules.CompactNotation),
		CompactDisplay: stringPtr(pluralrules.LongCompactDisplay),
	}).ResolvedOptions()
	componentDateTime := mustDateTimeFormat(t, datetimeformat.Options{
		Year:   stringPtr(datetimeformat.NumericFieldStyle),
		Month:  stringPtr(datetimeformat.ShortMonthStyle),
		Day:    stringPtr(datetimeformat.NumericFieldStyle),
		Hour:   stringPtr(datetimeformat.NumericFieldStyle),
		Hour12: &hour12,
	}).ResolvedOptions()
	styleDateTime := mustDateTimeFormat(t, datetimeformat.Options{
		DateStyle: stringPtr(datetimeformat.MediumDateTimeStyle),
		TimeStyle: stringPtr(datetimeformat.ShortDateTimeStyle),
	}).ResolvedOptions()
	phonebookCollation := "phonebk"
	defaultCollator := mustCollator(t, collator.Options{}).ResolvedOptions()
	phonebookCollator := mustCollatorFor(t, "de", collator.Options{Collation: &phonebookCollation}).ResolvedOptions()
	defaultDuration := mustDurationFormat(t, durationformat.Options{}).ResolvedOptions()
	fractionalDuration := mustDurationFormat(t, durationformat.Options{
		Style:            stringPtr(durationformat.LongStyle),
		FractionalDigits: intPtr(2),
	}).ResolvedOptions()
	defaultList := mustListFormat(t, listformat.Options{}).ResolvedOptions()
	defaultRelativeTime := mustRelativeTimeFormat(t, relativetimeformat.Options{}).ResolvedOptions()
	regionDisplayNames := mustDisplayNames(t, displaynames.Options{Type: stringPtr(displaynames.Region)}).ResolvedOptions()
	languageDisplayNames := mustDisplayNames(t, displaynames.Options{
		Type:            stringPtr(displaynames.Language),
		LanguageDisplay: stringPtr(displaynames.StandardLanguageDisplay),
	}).ResolvedOptions()
	wordSegmenter := mustSegmenter(t, segmenter.Options{Granularity: stringPtr(segmenter.WordGranularity)}).ResolvedOptions()
	wordSegment := mustFirstSegment(t, "hello!", segmenter.WordGranularity)
	wordPunctuation := mustSegmentAt(t, "hello!", segmenter.WordGranularity, 1)
	rtlTextInfo := intltest.Locale(t, "ar").GetTextInfo()

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
			name:  "number resolved options",
			value: defaultNumber,
			want: []string{
				`"locale":"en"`,
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
			name:  "datetime resolved options",
			value: componentDateTime,
			want: []string{
				`"locale":"en"`,
				`"hourCycle":"h23"`,
				`"hour12":false`,
				`"month":"short"`,
			},
			absent: []string{`"weekday"`, `"fractionalSecondDigits"`, `"dateStyle"`, `"timeStyle"`},
		},
		{
			name:  "datetime fractional-second part",
			value: datetimeformat.Part{Type: datetimeformat.PartFractionalSecond, Value: "123"},
			want:  []string{`"type":"fractionalSecond"`, `"value":"123"`},
		},
		{
			name:  "datetime fractional-second range part",
			value: datetimeformat.RangePart{Type: datetimeformat.PartFractionalSecond, Value: "123", Source: datetimeformat.SourceStartRange},
			want:  []string{`"type":"fractionalSecond"`, `"value":"123"`, `"source":"startRange"`},
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
			name:   "plural resolved options",
			value:  defaultPlural,
			want:   []string{`"type":"cardinal"`, `"pluralCategories":["one","other"]`},
			absent: []string{`"minimumSignificantDigits"`, `"compactDisplay"`},
		},
		{
			name:  "plural significant digit resolved options",
			value: significantPlural,
			want: []string{
				`"minimumSignificantDigits":2`,
				`"maximumSignificantDigits":21`,
			},
			absent: []string{`"minimumFractionDigits"`, `"maximumFractionDigits"`, `"compactDisplay"`},
		},
		{
			name:  "plural compact resolved options",
			value: compactPlural,
			want: []string{
				`"notation":"compact"`,
				`"compactDisplay":"long"`,
				`"minimumFractionDigits":0`,
				`"maximumFractionDigits":0`,
				`"minimumSignificantDigits":1`,
				`"maximumSignificantDigits":2`,
			},
		},
		{
			name:  "list part",
			value: listformat.Part{Type: listformat.PartElement, Value: "a"},
			want:  []string{`"type":"element"`, `"value":"a"`},
		},
		{
			name:  "list resolved options",
			value: defaultList,
			want:  []string{`"locale":"en"`, `"type":"conjunction"`, `"style":"long"`},
		},
		{
			name:   "relative time part omits empty unit",
			value:  relativetimeformat.Part{Type: relativetimeformat.PartLiteral, Value: "yesterday"},
			want:   []string{`"type":"literal"`, `"value":"yesterday"`},
			absent: []string{`"unit"`},
		},
		{
			name:  "relative time resolved options",
			value: defaultRelativeTime,
			want:  []string{`"locale":"en"`, `"style":"long"`, `"numeric":"always"`, `"numberingSystem":"latn"`},
		},
		{
			name:   "duration resolved options omits fractional digits",
			value:  defaultDuration,
			want:   []string{`"style":"short"`, `"millisecondsDisplay":"auto"`},
			absent: []string{`"fractionalDigits"`},
		},
		{
			name:  "duration fractional digits resolved options",
			value: fractionalDuration,
			want:  []string{`"style":"long"`, `"fractionalDigits":2`, `"milliseconds":"long"`},
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
			name:   "display names resolved options omits language display",
			value:  regionDisplayNames,
			want:   []string{`"type":"region"`, `"fallback":"code"`},
			absent: []string{`"languageDisplay"`},
		},
		{
			name:  "display names language resolved options includes language display",
			value: languageDisplayNames,
			want:  []string{`"type":"language"`, `"languageDisplay":"standard"`},
		},
		{
			name:  "collator resolved options includes default collation",
			value: defaultCollator,
			want:  []string{`"caseFirst":"false"`, `"collation":"default"`, `"numeric":false`},
		},
		{
			name:  "collator resolved options includes backend collation",
			value: phonebookCollator,
			want:  []string{`"locale":"de"`, `"collation":"phonebk"`},
		},
		{
			name:   "segment record uses code unit index",
			value:  segmenter.Segment{Segment: "🙂", CodeUnitIndex: 2, ByteIndex: 4, Input: "a🙂", IsWordLike: false},
			want:   []string{`"segment":"🙂"`, `"index":2`, `"input":"a🙂"`},
			absent: []string{`"ByteIndex"`, `"byteIndex"`, `"isWordLike"`},
		},
		{
			name:  "segment word record includes word-like true",
			value: wordSegment,
			want:  []string{`"segment":"hello"`, `"index":0`, `"input":"hello!"`, `"isWordLike":true`},
		},
		{
			name:  "segment word record includes word-like false",
			value: wordPunctuation,
			want:  []string{`"segment":"!"`, `"index":5`, `"input":"hello!"`, `"isWordLike":false`},
		},
		{
			name:  "segmenter resolved options",
			value: wordSegmenter,
			want:  []string{`"locale":"en"`, `"granularity":"word"`},
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

func stringPtr[T ~string](v T) *string {
	value := string(v)
	return &value
}

func mustNumberFormat(t *testing.T, opts numberformat.Options) *numberformat.NumberFormat {
	t.Helper()

	format, err := numberformat.New(locale.List{intltest.Locale(t, "en")}, opts)
	if err != nil {
		t.Fatalf("numberformat.New() error = %v", err)
	}
	return format
}

func mustDateTimeFormat(t *testing.T, opts datetimeformat.Options) *datetimeformat.DateTimeFormat {
	t.Helper()

	format, err := datetimeformat.New(locale.List{intltest.Locale(t, "en-US")}, opts)
	if err != nil {
		t.Fatalf("datetimeformat.New() error = %v", err)
	}
	return format
}

func mustPluralRules(t *testing.T, opts pluralrules.Options) *pluralrules.PluralRules {
	t.Helper()

	rules, err := pluralrules.New(locale.List{intltest.Locale(t, "en")}, opts)
	if err != nil {
		t.Fatalf("pluralrules.New() error = %v", err)
	}
	return rules
}

func mustListFormat(t *testing.T, opts listformat.Options) *listformat.ListFormat {
	t.Helper()

	format, err := listformat.New(locale.List{intltest.Locale(t, "en")}, opts)
	if err != nil {
		t.Fatalf("listformat.New() error = %v", err)
	}
	return format
}

func mustRelativeTimeFormat(t *testing.T, opts relativetimeformat.Options) *relativetimeformat.RelativeTimeFormat {
	t.Helper()

	format, err := relativetimeformat.New(locale.List{intltest.Locale(t, "en")}, opts)
	if err != nil {
		t.Fatalf("relativetimeformat.New() error = %v", err)
	}
	return format
}

func mustDisplayNames(t *testing.T, opts displaynames.Options) *displaynames.DisplayNames {
	t.Helper()

	names, err := displaynames.New(locale.List{intltest.Locale(t, "en")}, opts)
	if err != nil {
		t.Fatalf("displaynames.New() error = %v", err)
	}
	return names
}

func mustCollator(t *testing.T, opts collator.Options) *collator.Collator {
	t.Helper()

	return mustCollatorFor(t, "en", opts)
}

func mustCollatorFor(t *testing.T, tag string, opts collator.Options) *collator.Collator {
	t.Helper()

	c, err := collator.New(locale.List{intltest.Locale(t, tag)}, opts)
	if err != nil {
		t.Fatalf("collator.New() error = %v", err)
	}
	return c
}

func mustDurationFormat(t *testing.T, opts durationformat.Options) *durationformat.DurationFormat {
	t.Helper()

	format, err := durationformat.New(locale.List{intltest.Locale(t, "en")}, opts)
	if err != nil {
		t.Fatalf("durationformat.New() error = %v", err)
	}
	return format
}

func mustFirstSegment(t *testing.T, input string, granularity segmenter.Granularity) segmenter.Segment {
	t.Helper()

	return mustSegmentAt(t, input, granularity, 0)
}

func mustSegmenter(t *testing.T, opts segmenter.Options) *segmenter.Segmenter {
	t.Helper()

	s, err := segmenter.New(locale.List{intltest.Locale(t, "en")}, opts)
	if err != nil {
		t.Fatalf("segmenter.New() error = %v", err)
	}
	return s
}

func mustSegmentAt(t *testing.T, input string, granularity segmenter.Granularity, index int) segmenter.Segment {
	t.Helper()

	s := mustSegmenter(t, segmenter.Options{Granularity: stringPtr(granularity)})
	i := 0
	for segment := range s.Segment(input).All() {
		if i == index {
			return segment
		}
		i++
	}
	t.Fatalf("Segment(%q).All() returned fewer than %d records", input, index+1)
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
