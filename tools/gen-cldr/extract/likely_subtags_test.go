package extract

import (
	"strings"
	"testing"
)

func TestExtractLikelySubtagsRejectsMalformedSource(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{name: "extra value segment", key: "en", value: "en_Latn_US_extra"},
		{name: "empty value segment", key: "en", value: "en__Latn_US"},
		{name: "misplaced value subtags", key: "en", value: "en_US_Latn"},
		{name: "invalid value language", key: "en", value: "x_Latn_US"},
		{name: "value variant", key: "en", value: "en_Latn_US_posix"},
		{name: "value extension", key: "en", value: "en_Latn_US_u_ca_gregory"},
		{name: "value private use", key: "en", value: "en_Latn_US_x_test"},
		{name: "extra key segment", key: "en_Latn_US_extra", value: "en_Latn_US"},
		{name: "empty key segment", key: "en__US", value: "en_Latn_US"},
		{name: "misplaced key subtags", key: "en_US_Latn", value: "en_Latn_US"},
		{name: "invalid key language", key: "x", value: "en_Latn_US"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := ExtractLikelySubtags(map[string]string{tt.key: tt.value})
			if err == nil {
				t.Fatal("ExtractLikelySubtags() error = nil, want malformed source error")
			}
			if !strings.Contains(err.Error(), tt.key) || !strings.Contains(err.Error(), tt.value) {
				t.Fatalf("ExtractLikelySubtags() error = %q, want source key %q and value %q", err, tt.key, tt.value)
			}
		})
	}
}

func TestExtractLikelySubtagsRejectsNormalizedDuplicateKeys(t *testing.T) {
	t.Parallel()

	raw := map[string]string{
		"en-US": "en-Latn-US",
		"en_US": "en-Latn-GB",
	}
	_, err := ExtractLikelySubtags(raw)
	if err == nil {
		t.Fatal("ExtractLikelySubtags() error = nil, want normalized duplicate error")
	}
	for _, want := range []string{"en-US", "en_US", "en-Latn-GB"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("ExtractLikelySubtags() error = %q, want %q", err, want)
		}
	}
}

func TestExtractLikelySubtagsCanonicalizesValidSourceForms(t *testing.T) {
	t.Parallel()

	raw := map[string]string{
		"EN":          "EN_latn_us",
		"pt_br":       "pt_Latn_BR",
		"sr_cyrl":     "sr_Cyrl_RS",
		"und_arab_af": "und_Arab_AF",
		"zh_hant_tw":  "zh_Hant_TW",
	}
	got, err := ExtractLikelySubtags(raw)
	if err != nil {
		t.Fatalf("ExtractLikelySubtags() error = %v", err)
	}
	want := map[string]SubtagTriple{
		"en":          {Lang: "en", Script: "Latn", Region: "US"},
		"pt-BR":       {Lang: "pt", Script: "Latn", Region: "BR"},
		"sr-Cyrl":     {Lang: "sr", Script: "Cyrl", Region: "RS"},
		"und-Arab-AF": {Lang: "und", Script: "Arab", Region: "AF"},
		"zh-Hant-TW":  {Lang: "zh", Script: "Hant", Region: "TW"},
	}
	if len(got.Maximize) != len(want) {
		t.Fatalf("Maximize length = %d, want %d", len(got.Maximize), len(want))
	}
	for key, triple := range want {
		if got.Maximize[key] != triple {
			t.Errorf("Maximize[%q] = %#v, want %#v", key, got.Maximize[key], triple)
		}
	}
}
