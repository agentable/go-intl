package locale

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
)

func TestParseUnicodeExtensions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in              string
		want            string
		calendar        string
		collation       string
		hourCycle       string
		caseFirst       string
		numeric         bool
		numberingSystem string
		firstDayOfWeek  string
	}{
		{in: "en-US-u-hc-h23", want: "en-US-u-hc-h23", hourCycle: "h23"},
		{in: "en-US-u-ca-buddhist-hc-h23", want: "en-US-u-ca-buddhist-hc-h23", calendar: "buddhist", hourCycle: "h23"},
		{in: "en-us-U-Ca-Gregorian-Hc-H23", want: "en-US-u-ca-gregory-hc-h23", calendar: "gregory", hourCycle: "h23"},
		{in: "en-US-u-kn", want: "en-US-u-kn", numeric: true},
		{in: "en-US-u-kn-false", want: "en-US-u-kn-false", numeric: false},
		{in: "en-US-u-kf-false", want: "en-US-u-kf-false", caseFirst: "false"},
		{in: "en-US-u-ca-islamic-civil", want: "en-US-u-ca-islamicc", calendar: "islamicc"},
		{in: "en-US-u-co-phonebk-nu-arab-fw-mon", want: "en-US-u-co-phonebk-fw-mon-nu-arab", collation: "phonebk", numberingSystem: "arab", firstDayOfWeek: "mon"},
		{in: "en-US-u-fw-0", want: "en-US-u-fw-sun", firstDayOfWeek: "sun"},
		{in: "en-US-u-foo-ca-buddhist-zz-abc", want: "en-US-u-foo-ca-buddhist-zz-abc", calendar: "buddhist"},
		{in: "en-US-u-zz-abc-ca-buddhist", want: "en-US-u-ca-buddhist-zz-abc", calendar: "buddhist"},
		{in: "en-US-u-attr2-attr1-ca-buddhist-zz-abc-aa-xyz", want: "en-US-u-attr1-attr2-aa-xyz-ca-buddhist-zz-abc", calendar: "buddhist"},
		{in: "en-u-foo-bar-nu-thai-ca-buddhist-kk-true", want: "en-u-bar-foo-ca-buddhist-kk-nu-thai", calendar: "buddhist", numberingSystem: "thai"},
		{in: "en-u-ca-buddhist-ca-gregory", want: "en-u-ca-buddhist", calendar: "buddhist"},
		{in: "en-u-kf-upper-kf-lower", want: "en-u-kf-upper", caseFirst: "upper"},
		{in: "en-u-kn-false-kn", want: "en-u-kn-false", numeric: false},
		{in: "en-u-ca-gregory-x-private", want: "en-u-ca-gregory-x-private", calendar: "gregory"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got, err := Parse(tc.in)
			if err != nil {
				t.Fatalf("Parse err = %v", err)
			}
			if got.String() != tc.want {
				t.Fatalf("String() = %q, want %q", got.String(), tc.want)
			}
			if got.Calendar() != tc.calendar || got.Collation() != tc.collation || got.HourCycle() != tc.hourCycle || got.CaseFirst() != tc.caseFirst || got.Numeric() != tc.numeric || got.NumberingSystem() != tc.numberingSystem || got.FirstDayOfWeek() != tc.firstDayOfWeek {
				t.Fatalf("extensions = %#v", got)
			}
		})
	}
}

func TestParseUnicodeExtensionBaseNameExcludesExtension(t *testing.T) {
	t.Parallel()

	loc := MustParse("en-US-u-foo-ca-buddhist-zz-abc")
	if got := loc.BaseName(); got != "en-US" {
		t.Fatalf("BaseName() = %q, want en-US", got)
	}
}

func TestNewOptionsOverrideKnownUnicodeExtensionOnly(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US-u-foo-ca-buddhist-zz-abc", Options{Calendar: "gregory"})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got, want := loc.String(), "en-US-u-foo-ca-gregory-zz-abc"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestParseInvalidUnicodeExtensions(t *testing.T) {
	t.Parallel()

	for _, in := range []string{
		"en-US-u-hc-h25",
		"en-US-u-kf-middle",
		"en-US-u-fw-funday",
		"en-US-u-ca-a",
		"en-US-u-nu-ab_cd",
	} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(in)
			if !errors.Is(err, intlerr.ErrInvalidValue) {
				t.Fatalf("Parse(%q) err = %v, want intlerr.ErrInvalidValue", in, err)
			}
		})
	}
}
