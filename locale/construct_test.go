package locale

import (
	"errors"
	"testing"

	"golang.org/x/text/language"
)

func TestNewWithOptions(t *testing.T) {
	t.Parallel()

	loc, err := New(language.Japanese, Options{Calendar: "japanese", HourCycle: "h23"})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got := loc.String(); got != "ja-u-ca-japanese-hc-h23" {
		t.Fatalf("String() = %q, want ja-u-ca-japanese-hc-h23", got)
	}
}

func TestNewOptionsOverrideTagExtensions(t *testing.T) {
	t.Parallel()

	tag := language.MustParse("en-US-u-ca-buddhist-hc-h12")
	loc, err := New(tag, Options{Calendar: "gregory", HourCycle: "h23"})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got := loc.String(); got != "en-US-u-ca-gregory-hc-h23" {
		t.Fatalf("String() = %q, want en-US-u-ca-gregory-hc-h23", got)
	}
}

func TestNewLanguageIdentifierOptions(t *testing.T) {
	t.Parallel()

	loc, err := New(language.MustParse("en-US-u-nu-arab"), Options{
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

	_, err := New(language.English, Options{Variants: []string{"emodeng", "emodeng"}})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New duplicate variants err = %v, want ErrInvalidOption", err)
	}
}

func TestNewInvalidOption(t *testing.T) {
	t.Parallel()

	_, err := New(language.English, Options{HourCycle: "h25"})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New err = %v, want ErrInvalidOption", err)
	}
}

func TestNewCanonicalizesNumericFirstDayOfWeekZero(t *testing.T) {
	t.Parallel()

	loc, err := New(language.MustParse("en-US"), Options{FirstDayOfWeek: "0"})
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

func TestNewRejectsMultipleOptions(t *testing.T) {
	t.Parallel()

	_, err := New(language.English, Options{Calendar: "gregory"}, Options{HourCycle: "h23"})
	if !errors.Is(err, ErrInvalidOption) {
		t.Fatalf("New err = %v, want ErrInvalidOption", err)
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	a := MustParse("en-us-u-hc-h23-ca-gregorian")
	b := MustParse("en-US-u-ca-gregory-hc-h23")
	if !a.Equal(b) {
		t.Fatalf("%q Equal(%q) = false, want true", a.String(), b.String())
	}
	if a.Equal(MustParse("en-US-u-ca-buddhist-hc-h23")) {
		t.Fatal("Equal returned true for different calendar")
	}
}

func TestTextMarshaling(t *testing.T) {
	t.Parallel()

	loc := MustParse("en-US-u-hc-h23")
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
