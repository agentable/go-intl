package locale

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/agentable/go-intl/internal/intlerr"
)

func TestParseSimpleLocale(t *testing.T) {
	t.Parallel()

	loc, err := Parse("en-US")
	if err != nil {
		t.Fatalf("Parse err = %v", err)
	}
	if got := loc.Tag().String(); got != "en-US" {
		t.Fatalf("Tag.String() = %q, want en-US", got)
	}
	if got := loc.BaseName(); got != "en-US" {
		t.Fatalf("BaseName() = %q, want en-US", got)
	}
	if got := loc.Language(); got != "en" {
		t.Fatalf("Language() = %q, want en", got)
	}
	if got := loc.Region(); got != "US" {
		t.Fatalf("Region() = %q, want US", got)
	}
	if got := loc.String(); got != "en-US" {
		t.Fatalf("String() = %q, want en-US", got)
	}
	if loc.Calendar() != "" || loc.Collation() != "" || loc.HourCycle() != "" || loc.CaseFirst() != "" || loc.NumberingSystem() != "" || loc.FirstDayOfWeek() != "" {
		t.Fatalf("extensions = %#v, want empty", loc)
	}
	if loc.Numeric() {
		t.Fatal("Numeric = true, want false")
	}
}

func TestParseCanonicalLanguageAliases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "twi", want: "ak"},
		{in: "und-Armn-SU", want: "und-Armn-AM"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			if got := parseLocaleForTest(tc.in).String(); got != tc.want {
				t.Fatalf("Parse(%q).String() = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestParseVariantGetters(t *testing.T) {
	t.Parallel()

	loc := parseLocaleForTest("en-Latn-US-emodeng")
	if got := loc.Script(); got != "Latn" {
		t.Fatalf("Script() = %q, want Latn", got)
	}
	if got := loc.Variants(); len(got) != 1 || got[0] != "emodeng" {
		t.Fatalf("Variants() = %#v, want emodeng", got)
	}
}

func TestParsePrivateUseDoesNotPromoteUnicodeExtension(t *testing.T) {
	t.Parallel()

	loc, err := Parse("en-x-foo-u-ca-gregory")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got := loc.String(); got != "en-x-foo-u-ca-gregory" {
		t.Fatalf("String() = %q, want private-use u preserved", got)
	}
	if got := loc.Calendar(); got != "" {
		t.Fatalf("Calendar() = %q, want empty for private-use u", got)
	}
}

func TestParsePrivateUseKeepsUnicodeAliasTextOpaque(t *testing.T) {
	t.Parallel()

	loc, err := Parse("en-x-foo-u-ca-islamic-civil-fw-0")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got, want := loc.String(), "en-x-foo-u-ca-islamic-civil-fw-0"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	if loc.Calendar() != "" || loc.FirstDayOfWeek() != "" {
		t.Fatalf("private-use extensions promoted to locale fields: calendar=%q firstDayOfWeek=%q", loc.Calendar(), loc.FirstDayOfWeek())
	}
}

func TestLocaleMaximizeAndMinimize(t *testing.T) {
	t.Parallel()

	if got := parseLocaleForTest("zh").Maximize().String(); got != "zh-Hans-CN" {
		t.Fatalf("Maximize(zh) = %q, want zh-Hans-CN", got)
	}
	if got := parseLocaleForTest("zh-Hans-CN").Minimize().String(); got != "zh" {
		t.Fatalf("Minimize(zh-Hans-CN) = %q, want zh", got)
	}
	if got := parseLocaleForTest("en-Latn-US").Minimize().String(); got != "en" {
		t.Fatalf("Minimize(en-Latn-US) = %q, want en", got)
	}
}

func TestLocaleEqualUsesCanonicalForm(t *testing.T) {
	t.Parallel()

	loc := parseLocaleForTest("en-us-U-Ca-Gregorian-Hc-H23")
	same := parseLocaleForTest("en-US-u-hc-h23-ca-gregory")
	if !loc.Equal(same) {
		t.Fatalf("Equal() = false, want true for %q and %q", loc.String(), same.String())
	}
	different := parseLocaleForTest("en-US-u-ca-buddhist-hc-h23")
	if loc.Equal(different) {
		t.Fatalf("Equal() = true, want false for %q and %q", loc.String(), different.String())
	}
}

func TestLocaleNewAppliesAndValidatesOptions(t *testing.T) {
	t.Parallel()

	loc, err := New("en-US", Options{
		Language:        "FR",
		Script:          "latn",
		Region:          "ca",
		Variants:        []string{"emodeng"},
		Calendar:        "gregorian",
		Collation:       "phonebk",
		HourCycle:       "H23",
		CaseFirst:       "UPPER",
		Numeric:         boolPtr(false),
		NumberingSystem: "ARAB",
		FirstDayOfWeek:  "0",
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if got, want := loc.String(), "fr-Latn-CA-emodeng-u-ca-gregory-co-phonebk-fw-sun-hc-h23-kf-upper-kn-false-nu-arab"; got != want {
		t.Fatalf("New().String() = %q, want %q", got, want)
	}

	for _, tc := range []struct {
		name string
		opts Options
	}{
		{name: "duplicate variant", opts: Options{Variants: []string{"posix", "POSIX"}}},
		{name: "invalid script", opts: Options{Script: "1"}},
		{name: "invalid calendar", opts: Options{Calendar: "bad!"}},
		{name: "invalid hour cycle", opts: Options{HourCycle: "h99"}},
		{name: "invalid case first", opts: Options{CaseFirst: "middle"}},
		{name: "invalid first day", opts: Options{FirstDayOfWeek: "8"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := New("en", tc.opts)
			if !errors.Is(err, intlerr.ErrInvalidOption) {
				t.Fatalf("New(%s) error = %v, want intlerr.ErrInvalidOption", tc.name, err)
			}
			assertStructuredLocaleError(t, err, intlerr.InvalidOption)
		})
	}
}

func TestLocaleInfoUsesExtensionsAndRegionFallbacks(t *testing.T) {
	t.Parallel()

	loc := parseLocaleForTest("en-u-ca-buddhist-co-phonebk-hc-h23-nu-thai-fw-fri-rg-gbzzzz")
	if got := loc.GetCalendars(); len(got) != 1 || got[0] != "buddhist" {
		t.Fatalf("GetCalendars() = %v, want [buddhist]", got)
	}
	if got := loc.GetCollations(); len(got) != 1 || got[0] != "phonebk" {
		t.Fatalf("GetCollations() = %v, want [phonebk]", got)
	}
	if got := loc.GetHourCycles(); len(got) != 1 || got[0] != "h23" {
		t.Fatalf("GetHourCycles() = %v, want [h23]", got)
	}
	if got := loc.GetNumberingSystems(); len(got) != 1 || got[0] != "thai" {
		t.Fatalf("GetNumberingSystems() = %v, want [thai]", got)
	}
	if got := loc.GetWeekInfo().FirstDay; got != time.Friday {
		t.Fatalf("GetWeekInfo().FirstDay = %v, want Friday", got)
	}

	subdivision := parseLocaleForTest("en-u-sd-gbusct")
	if zones := subdivision.GetTimeZones(); len(zones) != 0 {
		t.Fatalf("GetTimeZones() for subdivision-only locale = %v, want nil without explicit region", zones)
	}
	if got := subdivision.GetWeekInfo().FirstDay; got != time.Monday {
		t.Fatalf("GetWeekInfo() first day via subdivision = %v, want Monday", got)
	}
}

func TestLocaleTextInfoAndUnmarshalText(t *testing.T) {
	t.Parallel()

	if got := parseLocaleForTest("ar").GetTextInfo().Direction; got != "rtl" {
		t.Fatalf("GetTextInfo(ar).Direction = %q, want rtl", got)
	}
	var loc Locale
	if err := loc.UnmarshalText([]byte("fr-CA")); err != nil {
		t.Fatalf("UnmarshalText(fr-CA) error = %v", err)
	}
	if got := loc.String(); got != "fr-CA" {
		t.Fatalf("UnmarshalText(fr-CA) = %q, want fr-CA", got)
	}
	if err := loc.UnmarshalText([]byte("bad_locale")); !errors.Is(err, intlerr.ErrInvalidValue) {
		t.Fatalf("UnmarshalText(bad_locale) error = %v, want intlerr.ErrInvalidValue", err)
	}
}

func TestParseInvalidLocale(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "xx-INVALID", "x-private"} {
		t.Run(in, func(t *testing.T) {
			t.Parallel()
			_, err := Parse(in)
			if !errors.Is(err, intlerr.ErrInvalidValue) {
				t.Fatalf("Parse(%q) err = %v, want errors.Is(intlerr.ErrInvalidValue)", in, err)
			}
			assertStructuredLocaleError(t, err, intlerr.InvalidValue)
		})
	}
}

func assertStructuredLocaleError(t *testing.T, err error, kind intlerr.ErrorKind) {
	t.Helper()

	detail, ok := errors.AsType[*intlerr.Error](err)
	if !ok {
		t.Fatalf("error = %v, want *intlerr.Error", err)
	}
	if detail.Kind != kind {
		t.Fatalf("error kind = %v, want %v", detail.Kind, kind)
	}
	if detail.Owner != "locale" {
		t.Fatalf("error owner = %q, want locale", detail.Owner)
	}
	text := err.Error()
	if !strings.Contains(text, "expected ") || !strings.Contains(text, "; got ") {
		t.Fatalf("error text = %q, want expected/got guidance", text)
	}
}

func TestParseLocaleForTestHelper(t *testing.T) {
	t.Parallel()

	if got := parseLocaleForTest("en").String(); got != "en" {
		t.Fatalf("parseLocaleForTest(en).String() = %q, want en", got)
	}
	defer func() {
		if recover() == nil {
			t.Fatal("parseLocaleForTest(bad) did not panic")
		}
	}()
	_ = parseLocaleForTest("")
}
