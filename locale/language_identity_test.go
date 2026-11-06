package locale

import (
	"encoding/json/v2"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
)

func TestExplicitUnknownLanguageSubtags(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		tag, script, region string
	}{
		{"en", "", ""},
		{"en-Zzzz", "Zzzz", ""},
		{"en-ZZ", "", "ZZ"},
		{"en-Zzzz-ZZ", "Zzzz", "ZZ"},
	} {
		for _, suffix := range []struct{ tag, base string }{
			{},
			{"-emodeng-t-en-u-ca-gregory-x-private", "-emodeng"},
		} {
			t.Run(tc.tag+suffix.tag, func(t *testing.T) {
				t.Parallel()
				input := tc.tag + suffix.tag
				loc, err := Parse(input)
				if err != nil {
					t.Fatal(err)
				}
				if got := loc.String(); got != input {
					t.Errorf("String() = %q, want %q", got, input)
				}
				if got, want := loc.BaseName(), tc.tag+suffix.base; got != want {
					t.Errorf("BaseName() = %q, want %q", got, want)
				}
				if loc.Language() != "en" || loc.Script() != tc.script || loc.Region() != tc.region {
					t.Errorf("getters = %q/%q/%q, want en/%q/%q", loc.Language(), loc.Script(), loc.Region(), tc.script, tc.region)
				}
				data, err := json.Marshal(loc)
				if err != nil || string(data) != `"`+input+`"` {
					t.Errorf("JSON = %s, %v; want %q", data, err, input)
				}
				var roundTrip Locale
				if err := json.Unmarshal(data, &roundTrip); err != nil || !roundTrip.Equal(loc) {
					t.Errorf("JSON round trip = %v, %v", roundTrip, err)
				}
				if got, want := loc.Maximize().String(), "en-Latn-US"+suffix.tag; got != want {
					t.Errorf("Maximize() = %q, want %q", got, want)
				}
				if got, want := loc.Minimize().String(), "en"+suffix.tag; got != want {
					t.Errorf("Minimize() = %q, want %q", got, want)
				}
			})
		}
	}
}

func TestCanonicalAliasPreservesSuffix(t *testing.T) {
	t.Parallel()

	const suffix = "-Latn-GH-emodeng-t-en-us-u-ca-gregory-x-private"
	loc, err := Parse("twi" + suffix)
	if err != nil {
		t.Fatal(err)
	}
	if got := loc.String(); got != "ak"+suffix {
		t.Fatalf("Parse identity = %q, want %q", got, "ak"+suffix)
	}
	loc, err = New("en"+suffix, Options{Language: new("TWI")})
	if err != nil {
		t.Fatal(err)
	}
	if got := loc.String(); got != "ak"+suffix {
		t.Fatalf("New language identity = %q, want %q", got, "ak"+suffix)
	}
	if got := loc.Maximize().String(); got != "ak"+suffix {
		t.Fatalf("Maximize = %q, want retained suffix", got)
	}
	if got := loc.Minimize().String(); got != "ak-emodeng-t-en-us-u-ca-gregory-x-private" {
		t.Fatalf("Minimize = %q, want retained suffix", got)
	}
}

func TestNorwegianLanguageIdentity(t *testing.T) {
	t.Parallel()
	const suffix = "-Latn-NO-emodeng-t-en-us-u-ca-gregory-x-private"
	for _, base := range []string{"no", "nb"} {
		t.Run(base, func(t *testing.T) {
			t.Parallel()
			for _, input := range []string{base, base + "-NO", base + suffix} {
				loc, err := Parse(input)
				if err != nil {
					t.Fatal(err)
				}
				if got := loc.String(); got != input {
					t.Errorf("Parse(%q) = %q", input, got)
				}
			}
			loc, err := New("en"+suffix, Options{Language: new(base)})
			if err != nil {
				t.Fatal(err)
			}
			if got := loc.String(); got != base+suffix {
				t.Errorf("New language identity = %q, want %q", got, base+suffix)
			}
			if got := loc.Maximize().String(); got != base+suffix {
				t.Errorf("Maximize = %q, want %q", got, base+suffix)
			}
		})
	}
}

func TestSyntaxIdentityIndependentOfBackend(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"und-Gara", "sjc", "xx-invalid", "abcdefgh", "en-foobar", "qaa", "no", "nb", "en-Zzzz-ZZ"} {
		t.Run(input, func(t *testing.T) {
			loc, err := Parse(input)
			if err != nil {
				t.Fatal(err)
			}
			if loc.String() != input {
				t.Fatalf("identity %q, want %q", loc.String(), input)
			}
			const suffix = "-t-en-us-u-ca-gregory-x-private"
			suffixed, err := Parse(input + suffix)
			if err != nil {
				t.Fatal(err)
			}
			for _, value := range []Locale{suffixed, suffixed.Maximize(), suffixed.Minimize()} {
				if !strings.HasSuffix(value.String(), suffix) {
					t.Fatalf("lost suffix: %s", value)
				}
				data, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				var roundtrip Locale
				if err := json.Unmarshal(data, &roundtrip); err != nil || !roundtrip.Equal(value) {
					t.Fatalf("roundtrip: %v %v", roundtrip, err)
				}
			}
		})
	}
	loc, err := New("und-Armn", Options{Region: new("SU")})
	if err != nil {
		t.Fatal(err)
	}
	if loc.String() != "und-Armn-AM" {
		t.Fatalf("region alias override: %s", loc)
	}
}

func TestSubdivisionGettersAcceptUnknownRegion(t *testing.T) {
	t.Parallel()
	for input, control := range map[string]string{
		"en-u-rg-000zzzz": "en", "en-u-rg-aazzzz": "en",
		"en-u-sd-000abc": "en-001", "en-u-sd-aaabc": "en-001",
		"en-u-rg-buzzzz": "en-u-rg-mmzzzz", "en-u-rg-ukzzzz": "en-u-rg-gbzzzz",
		"en-u-sd-buabc": "en-MM", "en-u-sd-ukabc": "en-GB",
	} {
		t.Run(input, func(t *testing.T) {
			loc, err := Parse(input)
			if err != nil {
				t.Fatal(err)
			}
			want, err := Parse(control)
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(loc.GetCalendars(), want.GetCalendars()) {
				t.Errorf("calendar fallback %v, want %v", loc.GetCalendars(), want.GetCalendars())
			}
			wantHours := want.GetHourCycles()
			if input == "en-u-sd-000abc" || input == "en-u-sd-aaabc" {
				wantHours = []string{"h23"}
			}
			if !slices.Equal(loc.GetHourCycles(), wantHours) {
				t.Errorf("hour cycle fallback %v, want %v", loc.GetHourCycles(), wantHours)
			}
			gotWeek, wantWeek := loc.GetWeekInfo(), want.GetWeekInfo()
			if gotWeek.FirstDay != wantWeek.FirstDay || !slices.Equal(gotWeek.Weekend, wantWeek.Weekend) {
				t.Errorf("week fallback %v, want %v", gotWeek, wantWeek)
			}
		})
	}
}

func TestCompleteTagBridge(t *testing.T) {
	t.Parallel()
	for _, input := range []string{"en-US-u-ca-gregory", "no-Latn-NO-t-en-u-nu-latn-x-private", "en-Zzzz-ZZ"} {
		loc, err := Parse(input)
		if err != nil {
			t.Fatal(err)
		}
		tag, err := loc.Tag()
		if err != nil {
			t.Fatal(err)
		}
		roundtrip, err := FromTag(tag, Options{})
		if err != nil || !roundtrip.Equal(loc) {
			t.Fatalf("lossy bridge: %s -> %s -> %s, %v", loc, tag, roundtrip, err)
		}
	}
	for _, input := range []string{"und-Gara", "sjc", "abcdefgh", "en-foobar"} {
		loc, err := Parse(input)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := loc.Tag(); !errors.Is(err, intlerr.ErrInvalidValue) {
			t.Fatalf("unrepresentable bridge %s: %v", loc, err)
		}
	}
}

func TestCanonicalIdentityGrammarAndAliases(t *testing.T) {
	t.Parallel()
	for input, want := range map[string]string{
		"en-t-en-m0-names":                      "en-t-en-m0-prprname",
		"en-u-sd-cn11":                          "en-u-sd-cnbj",
		"und-u-rg-fi01":                         "und-u-rg-axzzzz",
		"EN-b-FOO-a-BAR-u-ca-gregory-v-BAZ-x-A": "en-a-bar-b-foo-u-ca-gregory-v-baz-x-a",
		"en-t-IW-b1-FOO-a0-BAR-a0-BAZ":          "en-t-he-a0-bar-b1-foo",
		"hy-SU":                                 "hy-AM", "und-Armn-SU": "und-Armn-AM", "en-BU": "en-MM", "en-UK": "en-GB",
		"ja-Latn-hepburn-heploc": "ja-Latn-alalc97", "sh": "sr-Latn", "twi": "ak", "sgn-US": "ase",
		"en-foobar-1996": "en-1996-foobar",
	} {
		loc, err := Parse(input)
		if err != nil || loc.String() != want {
			t.Errorf("Parse(%q)=%s,%v; want %s", input, loc, err, want)
			continue
		}
		again, err := Parse(loc.String())
		if err != nil || !again.Equal(loc) {
			t.Errorf("not idempotent: %s, %v", again, err)
		}
	}
	for _, input := range []string{"en-x", "en-x-123456789", "en-x-!", "en-a", "en-a-1", "en-t-a0", "en-t-en-abcde-abcde", "en-u-a1", "en-u-ca-gregorian", "en-a-foo-A-bar", "en-POSIX-posix"} {
		if _, err := Parse(input); !errors.Is(err, intlerr.ErrInvalidValue) {
			t.Errorf("invalid %q: %v", input, err)
		}
	}
}
