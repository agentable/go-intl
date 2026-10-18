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
		{in: "en-us-U-Ca-Gregory-Hc-H23", want: "en-US-u-ca-gregory-hc-h23", calendar: "gregory", hourCycle: "h23"},
		{in: "en-US-u-kn", want: "en-US-u-kn", numeric: true},
		{in: "en-US-u-kn-false", want: "en-US-u-kn-false", numeric: false},
		{in: "en-US-u-kf-false", want: "en-US-u-kf-false", caseFirst: "false"},
		{in: "en-US-u-ca-islamicc", want: "en-US-u-ca-islamic-civil", calendar: "islamic-civil"},
		{in: "en-US-u-co-phonebk-nu-arab-fw-mon", want: "en-US-u-co-phonebk-fw-mon-nu-arab", collation: "phonebk", numberingSystem: "arab", firstDayOfWeek: "mon"},
		{in: "en-US-u-foo-ca-buddhist-zz-abc", want: "en-US-u-foo-ca-buddhist-zz-abc", calendar: "buddhist"},
		{in: "en-US-u-zz-abc-ca-buddhist", want: "en-US-u-ca-buddhist-zz-abc", calendar: "buddhist"},
		{in: "en-US-u-attr2-attr1-ca-buddhist-zz-abc-aa-xyz", want: "en-US-u-attr1-attr2-aa-xyz-ca-buddhist-zz-abc", calendar: "buddhist"},
		{in: "en-u-foo-bar-nu-thai-ca-buddhist-kk-true", want: "en-u-bar-foo-ca-buddhist-kk-nu-thai", calendar: "buddhist", numberingSystem: "thai"},
		{in: "en-u-ca-buddhist-ca-gregory", want: "en-u-ca-buddhist", calendar: "buddhist"},
		{in: "en-u-kf-upper-kf-lower", want: "en-u-kf-upper", caseFirst: "upper"},
		{in: "en-u-kn-false-kn", want: "en-u-kn-false", numeric: false},
		{in: "en-u-ca-gregory-x-private", want: "en-u-ca-gregory-x-private", calendar: "gregory"},
		{in: "en-1901-t-de-1901-u-ca-gregory-x-a-a", want: "en-1901-t-de-1901-u-ca-gregory-x-a-a", calendar: "gregory"},
		{in: "en-x-a-a", want: "en-x-a-a"},
		{in: "en-a-foo-x-a-foo", want: "en-a-foo-x-a-foo"},
		{in: "en-x-foo-bar", want: "en-x-foo-bar"},
		{in: "en-a-foo-bar", want: "en-a-foo-bar"},
		{in: "en-t-en-US", want: "en-t-en-us"},
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

func TestParseAllowsDuplicateTransformedKeys(t *testing.T) {
	t.Parallel()

	// IsWellFormedLanguageTag permits repeated tkeys, unlike repeated singletons.
	loc, err := Parse("en-t-m0-ungegn-m0-alaloc")
	if err != nil {
		t.Fatal(err)
	}
	again, err := Parse(loc.String())
	if err != nil || !again.Equal(loc) {
		t.Fatalf("round trip = %v, %v", again, err)
	}
}

func TestParseRejectsNumericFirstDayOfWeekType(t *testing.T) {
	t.Parallel()

	_, err := Parse("en-US-u-fw-0")
	if !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("Parse(en-US-u-fw-0) error = %v, want intlerr.ErrInvalidValue", err)
	}
}

func TestParseUnicodeExtensionBaseNameExcludesExtension(t *testing.T) {
	t.Parallel()

	loc := parseLocaleForTest("en-US-u-foo-ca-buddhist-zz-abc")
	if got := loc.BaseName(); got != "en-US" {
		t.Fatalf("BaseName() = %q, want en-US", got)
	}
}

func TestNewOptionsOverrideKnownUnicodeExtensionOnly(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US-u-foo-ca-buddhist-zz-abc", Options{Calendar: stringPtr("gregory")})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got, want := loc.String(), "en-US-u-foo-ca-gregory-zz-abc"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestNewOptionsPreserveUnknownUnicodeKeywordsWhileReplacingKnownKeywords(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US-u-foo-ca-buddhist-kk-true-zz-abc", Options{
		Calendar: stringPtr("gregory"),
		Numeric:  boolPtr(true),
	})
	if err != nil {
		t.Fatalf("New err = %v", err)
	}
	if got, want := loc.String(), "en-US-u-foo-ca-gregory-kk-kn-zz-abc"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestParseInvalidUnicodeExtensions(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		in       string
		name     string
		value    string
		expected string
	}{
		{in: "en-US-u-ca-a", name: "languageTag", value: "en-US-u-ca-a", expected: localeLanguageTagExpected},
		{in: "en-US-u-nu-ab_cd", name: "languageTag", value: "en-US-u-nu-ab_cd", expected: localeLanguageTagExpected},
	} {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(tc.in)
			if !errors.Is(err, intlerr.ErrInvalidValue) {
				t.Fatalf("Parse(%q) err = %v, want intlerr.ErrInvalidValue", tc.in, err)
			}
			detail := assertStructuredLocaleError(t, err, intlerr.InvalidValue)
			if detail.Name != tc.name || detail.Value != tc.value || detail.Expected != tc.expected {
				t.Fatalf("Parse(%q) error detail = %+v, want name=%q value=%q expected=%q", tc.in, detail, tc.name, tc.value, tc.expected)
			}
		})
	}
}

func TestParsePreservesUnicodeTypePresence(t *testing.T) {
	t.Parallel()
	for _, tag := range []string{"en-u-hc-h25", "en-u-kf-middle", "en-US-u-fw-funday", "en-u-ca", "en-u-co", "en-u-hc", "en-u-kf", "en-u-fw", "en-u-nu", "en-u-ca-hc-kf-nu", "en-t-fr-u-ca-x-private"} {
		t.Run(tag, func(t *testing.T) {
			t.Parallel()
			loc, err := Parse(tag)
			if err != nil {
				t.Fatal(err)
			}
			if got := loc.String(); got != tag {
				t.Fatalf("String() = %q, want %q", got, tag)
			}
			again, err := Parse(loc.String())
			if err != nil || !again.Equal(loc) {
				t.Fatalf("round trip = %v, %v", again, err)
			}
		})
	}
	for _, tag := range []string{"en-u-ca-true", "en-u-ca-ca-gregory"} {
		loc, err := Parse(tag)
		if err != nil {
			t.Fatal(err)
		}
		if got := loc.String(); got != "en-u-ca" {
			t.Errorf("Parse(%q) = %q, want en-u-ca", tag, got)
		}
	}
}

func TestUnknownWeekdayUsesRegionPreference(t *testing.T) {
	t.Parallel()
	loc, err := Parse("en-US-u-fw-funday")
	if err != nil {
		t.Fatal(err)
	}
	base := parseLocaleForTest("en-US")
	if got, want := loc.GetWeekInfo().FirstDay, base.GetWeekInfo().FirstDay; got != want {
		t.Fatalf("firstDay = %v, want %v", got, want)
	}
}

func TestEmptyUnicodeTypesHavePresentCandidateLists(t *testing.T) {
	t.Parallel()
	loc, err := Parse("en-u-ca-co-hc-nu")
	if err != nil {
		t.Fatal(err)
	}
	for name, got := range map[string][]string{"calendars": loc.GetCalendars(), "collations": loc.GetCollations(), "hourCycles": loc.GetHourCycles(), "numberingSystems": loc.GetNumberingSystems()} {
		if len(got) != 1 || got[0] != "" {
			t.Errorf("%s = %q, want one empty type", name, got)
		}
	}
}
