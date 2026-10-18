package locale

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"golang.org/x/text/language"
)

func TestNewWithOptions(t *testing.T) {
	t.Parallel()

	loc, err := New("ja", Options{Calendar: stringPtr("japanese"), HourCycle: stringPtr("h23")})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got := loc.String(); got != "ja-u-ca-japanese-hc-h23" {
		t.Fatalf("String() = %q, want ja-u-ca-japanese-hc-h23", got)
	}
}

func TestUnicodeTypeAliasesAreCanonicalizedByKey(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   string
		want string
	}{
		{name: "deprecated calendar", in: "en-u-ca-islamicc", want: "en-u-ca-islamic-civil"},
		{name: "canonical calendar", in: "en-u-ca-islamic-civil", want: "en-u-ca-islamic-civil"},
		{name: "measurement system", in: "en-u-ms-imperial", want: "en-u-ms-uksystem"},
		{name: "calendar value under collation key", in: "en-u-co-islamic-civil", want: "en-u-co-islamic-civil"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			loc, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse(%q) error = %v", tc.in, err)
			}
			if got := loc.String(); got != tc.want {
				t.Fatalf("Parse(%q).String() = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestUnicodeTypeAliasDoesNotMakeMalformedInputValid(t *testing.T) {
	t.Parallel()

	if _, err := Parse("en-u-ca-gregorian"); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("Parse(en-u-ca-gregorian) error = %v, want intlerr.ErrInvalidValue", err)
	}
	if _, err := New("en", Options{Calendar: stringPtr("gregorian")}); !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New(calendar=gregorian) error = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNewOptionsOverrideTagExtensions(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US-u-ca-buddhist-hc-h12", Options{Calendar: stringPtr("gregory"), HourCycle: stringPtr("h23")})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got := loc.String(); got != "en-US-u-ca-gregory-hc-h23" {
		t.Fatalf("String() = %q, want en-US-u-ca-gregory-hc-h23", got)
	}
}

func TestNewLanguageIdentifierOptions(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US-pinyin-u-nu-arab", Options{
		Language: stringPtr("ZH"),
		Script:   stringPtr("hANS"),
		Region:   stringPtr("cn"),
	})
	if err != nil {
		t.Fatalf("New language options err = %v", err)
	}
	if got, want := loc.String(), "zh-Hans-CN-pinyin-u-nu-arab"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if got := loc.Variants(); len(got) != 1 || got[0] != "pinyin" {
		t.Fatalf("Variants() = %#v, want pinyin", got)
	}
}

func TestNewLanguageIdentifierOptionsPreserveExtensions(t *testing.T) {
	t.Parallel()

	loc, err := New("de-1901-t-en-u-ca-gregory-x-private", Options{
		Language: stringPtr("fr"),
		Script:   stringPtr("Latn"),
		Region:   stringPtr("CA"),
	})
	if err != nil {
		t.Fatalf("New language options error = %v", err)
	}
	const want = "fr-Latn-CA-1901-t-en-u-ca-gregory-x-private"
	if got := loc.String(); got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if got, want := loc.Tag().String(), "fr-Latn-CA-1901-t-en-x-private"; got != want {
		t.Fatalf("Tag().String() = %q, want %q", got, want)
	}
	if got, want := loc.BaseName(), "fr-Latn-CA-1901"; got != want {
		t.Fatalf("BaseName() = %q, want %q", got, want)
	}
	if got := loc.Variants(); len(got) != 1 || got[0] != "1901" {
		t.Fatalf("Variants() = %#v, want [1901]", got)
	}
	if got, want := loc.Calendar(), "gregory"; got != want {
		t.Fatalf("Calendar() = %q, want %q", got, want)
	}
}

func TestNewLanguageOptionsPreserveExplicitUnknownSubtags(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		tag  string
		opts Options
		want string
	}{
		{"en-Zzzz-ZZ", Options{Language: stringPtr("fr")}, "fr-Zzzz-ZZ"},
		{"en-Zzzz-ZZ", Options{Script: stringPtr("Latn")}, "en-Latn-ZZ"},
		{"en-Zzzz-ZZ", Options{Region: stringPtr("GB")}, "en-Zzzz-GB"},
		{"en-Zzzz", Options{Region: stringPtr("GB")}, "en-Zzzz-GB"},
		{"en-ZZ", Options{Script: stringPtr("Latn")}, "en-Latn-ZZ"},
		{"en-Latn-US", Options{Script: stringPtr("Zzzz")}, "en-Zzzz-US"},
		{"en-Latn-US", Options{Region: stringPtr("ZZ")}, "en-Latn-ZZ"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			for _, suffix := range []string{"", "-emodeng-t-en-u-ca-gregory-x-private"} {
				loc, err := New(tc.tag+suffix, tc.opts)
				if err != nil {
					t.Fatal(err)
				}
				if got, want := loc.String(), tc.want+suffix; got != want {
					t.Errorf("New(%q) = %q, want %q", tc.tag+suffix, got, want)
				}
			}
		})
	}
}

func TestNewRejectsExplicitEmptyStringOptions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		opts     Options
		expected string
	}{
		{name: "language", opts: Options{Language: stringPtr("")}, expected: localeLanguageExpected},
		{name: "script", opts: Options{Script: stringPtr("")}, expected: localeScriptExpected},
		{name: "region", opts: Options{Region: stringPtr("")}, expected: localeRegionExpected},
		{name: "calendar", opts: Options{Calendar: stringPtr("")}, expected: localeUnicodeTypeExpected},
		{name: "collation", opts: Options{Collation: stringPtr("")}, expected: localeUnicodeTypeExpected},
		{name: "hourCycle", opts: Options{HourCycle: stringPtr("")}, expected: localeHourCycleExpected},
		{name: "caseFirst", opts: Options{CaseFirst: stringPtr("")}, expected: localeCaseFirstExpected},
		{name: "numberingSystem", opts: Options{NumberingSystem: stringPtr("")}, expected: localeUnicodeTypeExpected},
		{name: "firstDayOfWeek", opts: Options{FirstDayOfWeek: stringPtr("")}, expected: localeFirstDayExpected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New("en", tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(empty %s) error = %v, want intlerr.ErrInvalidOption", tc.name, err)
			}
			detail := assertStructuredLocaleError(t, err, intlerr.InvalidOption)
			if detail.Name != tc.name || detail.Value != "" || detail.Expected != tc.expected {
				t.Fatalf("New(empty %s) error detail = %+v, want name=%q value empty expected %q", tc.name, detail, tc.name, tc.expected)
			}
		})
	}
}

func TestNewNormalizesRawOptionValuesDuringValidation(t *testing.T) {
	t.Parallel()

	loc, err := New("en", Options{
		Calendar:        stringPtr("GREGORY"),
		HourCycle:       stringPtr("h23"),
		CaseFirst:       stringPtr("upper"),
		NumberingSystem: stringPtr("ARAB"),
		FirstDayOfWeek:  stringPtr("MON"),
	})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got, want := loc.String(), "en-u-ca-gregory-fw-mon-hc-h23-kf-upper-nu-arab"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if got := loc.Calendar(); got != "gregory" {
		t.Fatalf("Calendar() = %q, want gregory", got)
	}
	if got := loc.HourCycle(); got != "h23" {
		t.Fatalf("HourCycle() = %q, want h23", got)
	}
	if got := loc.CaseFirst(); got != "upper" {
		t.Fatalf("CaseFirst() = %q, want upper", got)
	}
	if got := loc.NumberingSystem(); got != "arab" {
		t.Fatalf("NumberingSystem() = %q, want arab", got)
	}
	if got := loc.FirstDayOfWeek(); got != "mon" {
		t.Fatalf("FirstDayOfWeek() = %q, want mon", got)
	}
}

func TestNewRejectsUnicodeFoldedTypeOptions(t *testing.T) {
	t.Parallel()

	const foldedASCII = "\u212Aarab"
	for _, tc := range []struct {
		name string
		opts Options
	}{
		{name: "calendar", opts: Options{Calendar: stringPtr(foldedASCII)}},
		{name: "collation", opts: Options{Collation: stringPtr(foldedASCII)}},
		{name: "numberingSystem", opts: Options{NumberingSystem: stringPtr(foldedASCII)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New("en", tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s) error = %v, want intlerr.ErrInvalidOption", tc.name, err)
			}
			detail := assertStructuredLocaleError(t, err, intlerr.InvalidOption)
			if detail.Name != tc.name || detail.Value != foldedASCII || detail.Expected != localeUnicodeTypeExpected {
				t.Fatalf("New(%s) error detail = %+v, want name=%q value=%q expected %q",
					tc.name, detail, tc.name, foldedASCII, localeUnicodeTypeExpected)
			}
		})
	}
}

func TestNewRejectsNonASCIIEnumOptions(t *testing.T) {
	t.Parallel()

	const foldedASCII = "\u212A"
	for _, tc := range []struct {
		name     string
		opts     Options
		value    string
		expected string
	}{
		{name: "hourCycle", opts: Options{HourCycle: stringPtr(foldedASCII + "23")}, value: foldedASCII + "23", expected: localeHourCycleExpected},
		{name: "caseFirst", opts: Options{CaseFirst: stringPtr(foldedASCII + "upper")}, value: foldedASCII + "upper", expected: localeCaseFirstExpected},
		{name: "firstDayOfWeek", opts: Options{FirstDayOfWeek: stringPtr(foldedASCII + "mon")}, value: foldedASCII + "mon", expected: localeFirstDayExpected},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New("en", tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s) error = %v, want intlerr.ErrInvalidOption", tc.name, err)
			}
			detail := assertStructuredLocaleError(t, err, intlerr.InvalidOption)
			if detail.Name != tc.name || detail.Value != tc.value || detail.Expected != tc.expected {
				t.Fatalf("New(%s) error detail = %+v, want name=%q value=%q expected %q",
					tc.name, detail, tc.name, tc.value, tc.expected)
			}
		})
	}
}

func TestLocaleNewRejectsInvalidOption(t *testing.T) {
	t.Parallel()

	_, err := New("en", Options{HourCycle: stringPtr("h25")})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New err = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNewCanonicalizesNumericFirstDayOfWeek(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		want  string
	}{
		{value: "0", want: "sun"},
		{value: "1", want: "mon"},
		{value: "2", want: "tue"},
		{value: "3", want: "wed"},
		{value: "4", want: "thu"},
		{value: "5", want: "fri"},
		{value: "6", want: "sat"},
		{value: "7", want: "sun"},
	}
	for _, tc := range tests {
		t.Run(tc.value, func(t *testing.T) {
			t.Parallel()

			loc, err := New("en-US", Options{FirstDayOfWeek: stringPtr(tc.value)})
			if err != nil {
				t.Fatalf("New err = %v", err)
			}
			if got, want := loc.String(), "en-US-u-fw-"+tc.want; got != want {
				t.Fatalf("String() = %q, want %q", got, want)
			}
			if got := loc.FirstDayOfWeek(); got != tc.want {
				t.Fatalf("FirstDayOfWeek() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFromTag(t *testing.T) {
	t.Parallel()

	loc, err := FromTag(language.Japanese, Options{Calendar: stringPtr("japanese")})
	if err != nil {
		t.Fatalf("FromTag err = %v", err)
	}
	if got, want := loc.String(), "ja-u-ca-japanese"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestFromTagTagRoundTrip(t *testing.T) {
	t.Parallel()

	locales := []Locale{
		parseLocaleForTest("en"),
		parseLocaleForTest("en-US"),
		parseLocaleForTest("zh-Hans-CN"),
		parseLocaleForTest("de-AT-1901"),
		parseLocaleForTest("ar-u-ca-islamic-hc-h12-nu-arab"),
	}
	for _, loc := range locales {
		t.Run(loc.String(), func(t *testing.T) {
			t.Parallel()

			got, err := FromTag(loc.Tag(), Options{})
			if err != nil {
				t.Fatalf("FromTag(%v) error = %v", loc.Tag(), err)
			}
			if got.Tag() != loc.Tag() {
				t.Fatalf("FromTag(%v).Tag() = %v, want %v", loc.Tag(), got.Tag(), loc.Tag())
			}
		})
	}
}

func TestNewNumericOptionPresence(t *testing.T) {
	t.Parallel()

	enabled, err := New("en-US", Options{Numeric: boolPtr(true)})
	if err != nil {
		t.Fatalf("New numeric=true err = %v", err)
	}
	if got, want := enabled.String(), "en-US-u-kn"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}

	disabled, err := New("en-US-u-kn", Options{Numeric: boolPtr(false)})
	if err != nil {
		t.Fatalf("New numeric=false err = %v", err)
	}
	if got, want := disabled.String(), "en-US-u-kn-false"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if disabled.Numeric() {
		t.Fatal("Numeric() = true, want false")
	}

	enabledOverride, err := New("en-US-u-kn-false", Options{Numeric: boolPtr(true)})
	if err != nil {
		t.Fatalf("New numeric=true overriding false err = %v", err)
	}
	if got, want := enabledOverride.String(), "en-US-u-kn"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if !enabledOverride.Numeric() {
		t.Fatal("Numeric() = false, want true")
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	a := parseLocaleForTest("en-us-u-hc-h23-ca-islamicc")
	b := parseLocaleForTest("en-US-u-ca-islamic-civil-hc-h23")
	if !a.Equal(b) {
		t.Fatalf("%q Equal(%q) = false, want true", a.String(), b.String())
	}
	if a.Equal(parseLocaleForTest("en-US-u-ca-buddhist-hc-h23")) {
		t.Fatal("Equal returned true for different calendar")
	}
}

func TestTextMarshaling(t *testing.T) {
	t.Parallel()

	loc := parseLocaleForTest("en-US-u-hc-h23")
	text, err := loc.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText err = %v", err)
	}
	if string(text) != "en-US-u-hc-h23" {
		t.Fatalf("MarshalText = %q", text)
	}
	var got Locale
	if err := got.UnmarshalText(text); err != nil {
		t.Fatalf("UnmarshalText err = %v", err)
	}
	if !got.Equal(loc) {
		t.Fatalf("UnmarshalText = %q, want %q", got.String(), loc.String())
	}
}

func TestNewPreservesExistingVariantsWhenLanguageOptionsChange(t *testing.T) {
	t.Parallel()

	loc, err := New("de-1901", Options{Region: stringPtr("AT")})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got, want := loc.String(), "de-AT-1901"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	variants := loc.Variants()
	if len(variants) != 1 || variants[0] != "1901" {
		t.Fatalf("Variants() = %#v, want [1901]", variants)
	}
}

func TestUnmarshalTextRejectsInvalidLocale(t *testing.T) {
	t.Parallel()

	var loc Locale
	if err := loc.UnmarshalText([]byte("")); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("UnmarshalText(empty) error = %v, want intlerr.ErrInvalidValue", err)
	}
}

func TestNewValidatesOnlyExplicitUnicodeOptions(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, tag string
		opts      Options
		want      string
		bad       bool
	}{
		{name: "unknown tag", tag: "en-u-hc-h25-kf-middle", want: "en-u-hc-h25-kf-middle"},
		{name: "override unknown", tag: "en-u-hc-h25", opts: Options{HourCycle: stringPtr("h23")}, want: "en-u-hc-h23"},
		{name: "hour casing", tag: "en", opts: Options{HourCycle: stringPtr("H23")}, bad: true},
		{name: "case first casing", tag: "en", opts: Options{CaseFirst: stringPtr("UPPER")}, bad: true},
		{name: "unknown weekday", tag: "en-US", opts: Options{FirstDayOfWeek: stringPtr("funday")}, want: "en-US-u-fw-funday"},
		{name: "weekday compound type", tag: "en", opts: Options{FirstDayOfWeek: stringPtr("foo-bar")}, want: "en-u-fw-foo-bar"},
		{name: "weekday malformed", tag: "en", opts: Options{FirstDayOfWeek: stringPtr("foo_b ar")}, bad: true},
		{name: "empty presence", tag: "en-u-ca", want: "en-u-ca"},
		{name: "true option", tag: "en", opts: Options{Calendar: stringPtr("true")}, want: "en-u-ca"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			loc, err := New(tc.tag, tc.opts)
			if tc.bad {
				if !errors.Is(err, intlerr.ErrInvalidOption) {
					t.Fatalf("error = %v, want invalid option", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := loc.String(); got != tc.want {
				t.Fatalf("String() = %q, want %q", got, tc.want)
			}
			if tc.name == "true option" && loc.Calendar() != "" {
				t.Fatalf("Calendar() = %q, want empty type", loc.Calendar())
			}
		})
	}
}

func TestUnicodeKeyGrammar(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		tag   string
		valid bool
	}{
		{"en-u-00", false}, {"en-u-a0", false}, {"en-u-A0-foobar", false},
		{"en-u-0a", true}, {"en-u-aa", true}, {"EN-u-0A", true},
		{"en-u-attr-0a-foobar-aa-x-private", true},
	} {
		t.Run(tc.tag, func(t *testing.T) {
			t.Parallel()
			for name, parse := range map[string]func(string) (Locale, error){"Parse": Parse, "New": func(tag string) (Locale, error) { return New(tag, Options{}) }} {
				_, err := parse(tc.tag)
				if tc.valid {
					if err != nil {
						t.Fatalf("%s(%q): %v", name, tc.tag, err)
					}
				} else {
					if !errors.Is(err, intlerr.ErrInvalidValue) {
						t.Fatalf("%s(%q) error = %v, want invalid value", name, tc.tag, err)
					}
					detail := assertStructuredLocaleError(t, err, intlerr.InvalidValue)
					if detail.Value != tc.tag {
						t.Fatalf("error value = %q, want %q", detail.Value, tc.tag)
					}
				}
			}
		})
	}
}
