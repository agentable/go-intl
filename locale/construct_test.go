package locale

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"

	"golang.org/x/text/language"
)

func TestNewWithOptions(t *testing.T) {
	t.Parallel()

	loc, err := New("ja", Options{Calendar: "japanese", HourCycle: "h23"})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got := loc.String(); got != "ja-u-ca-japanese-hc-h23" {
		t.Fatalf("String() = %q, want ja-u-ca-japanese-hc-h23", got)
	}
}

func TestNewOptionsOverrideTagExtensions(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US-u-ca-buddhist-hc-h12", Options{Calendar: "gregory", HourCycle: "h23"})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got := loc.String(); got != "en-US-u-ca-gregory-hc-h23" {
		t.Fatalf("String() = %q, want en-US-u-ca-gregory-hc-h23", got)
	}
}

func TestNewLanguageIdentifierOptions(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US-u-nu-arab", Options{
		Language: "zh",
		Script:   "Hans",
		Region:   "CN",
		Variants: []string{"pinyin"},
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

func TestNewRejectsDuplicateVariantOptions(t *testing.T) {
	t.Parallel()

	_, err := New("en", Options{Variants: []string{"emodeng", "emodeng"}})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New duplicate variants err = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNewInvalidOption(t *testing.T) {
	t.Parallel()

	_, err := New("en", Options{HourCycle: "h25"})
	if !errors.Is(err, intlerr.ErrInvalidOption) {
		t.Fatalf("New err = %v, want intlerr.ErrInvalidOption", err)
	}
}

func TestNewCanonicalizesNumericFirstDayOfWeekZero(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US", Options{FirstDayOfWeek: "0"})
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

	loc, err := FromTag(language.Japanese, Options{Calendar: "japanese"})
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

func TestNewPreservesExistingVariantsWhenOptionsVariantsUnset(t *testing.T) {
	t.Parallel()

	loc, err := New("de-1901", Options{Region: "AT"})
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
