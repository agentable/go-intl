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
		Calendar:        stringPtr("GREGORIAN"),
		HourCycle:       stringPtr("H23"),
		CaseFirst:       stringPtr("UPPER"),
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

func TestNewCanonicalizesNumericFirstDayOfWeekZero(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US", Options{FirstDayOfWeek: stringPtr("0")})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got, want := loc.String(), "en-US-u-fw-sun"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if got := loc.FirstDayOfWeek(); got != "sun" {
		t.Fatalf("FirstDayOfWeek() = %q, want sun", got)
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

	a := parseLocaleForTest("en-us-u-hc-h23-ca-gregorian")
	b := parseLocaleForTest("en-US-u-ca-gregory-hc-h23")
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
