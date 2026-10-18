package locale

import (
	"encoding/json/v2"
	"testing"
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
