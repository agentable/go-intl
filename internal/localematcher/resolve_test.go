package localematcher

import "testing"

type testLocaleData map[string]map[string][]string

func (d testLocaleData) For(loc, key string) []string {
	if keys, ok := d[loc]; ok {
		return keys[key]
	}
	return nil
}

func TestResolveLocale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		opts      ResolveOptions
		locale    string
		data      string
		extension map[string]string
	}{
		{
			name: "best fit selects fr",
			opts: ResolveOptions{
				Algorithm:     AlgorithmBestFit,
				Supported:     []string{"fr", "en"},
				Requested:     []string{"fr-XX", "en"},
				DefaultLocale: "en",
			},
			locale: "fr",
			data:   "fr",
		},
		{
			name: "best fit selects zh hant tw",
			opts: ResolveOptions{
				Algorithm:     AlgorithmBestFit,
				Supported:     []string{"zh-Hant-TW", "en"},
				Requested:     []string{"zh-TW", "en"},
				DefaultLocale: "en",
			},
			locale: "zh-TW",
			data:   "zh-Hant-TW",
		},
		{
			name: "resolves extension keys",
			opts: ResolveOptions{
				Algorithm:             AlgorithmBestFit,
				Supported:             []string{"th", "en"},
				Requested:             []string{"th-u-ca-gregory"},
				DefaultLocale:         "en",
				RelevantExtensionKeys: []string{"ca", "nu", "hc"},
				LocaleData: testLocaleData{
					"th": {
						"nu": []string{"latn"},
						"ca": []string{"buddhist", "gregory"},
						"hc": []string{"h23", "h12"},
					},
				},
			},
			locale:    "th-u-ca-gregory",
			data:      "th",
			extension: map[string]string{"ca": "gregory", "nu": "latn", "hc": "h23"},
		},
		{
			name: "options override unicode extension without mutating other keys",
			opts: ResolveOptions{
				Algorithm:             AlgorithmLookup,
				Supported:             []string{"th", "en"},
				Requested:             []string{"th-u-ca-buddhist-nu-thai"},
				DefaultLocale:         "en",
				RelevantExtensionKeys: []string{"ca", "nu"},
				OptionValues:          []Option{{Key: "ca", Value: "gregory"}},
				LocaleData: testLocaleData{
					"th": {
						"ca": []string{"buddhist", "gregory"},
						"nu": []string{"latn", "thai"},
					},
				},
			},
			locale:    "th-u-nu-thai",
			data:      "th",
			extension: map[string]string{"ca": "gregory", "nu": "thai"},
		},
		{
			name: "unsupported option keeps supported unicode extension",
			opts: ResolveOptions{
				Algorithm:             AlgorithmLookup,
				Supported:             []string{"th", "en"},
				Requested:             []string{"th-u-ca-buddhist"},
				DefaultLocale:         "en",
				RelevantExtensionKeys: []string{"ca"},
				OptionValues:          []Option{{Key: "ca", Value: "islamic"}},
				LocaleData: testLocaleData{
					"th": {"ca": []string{"gregory", "buddhist"}},
				},
			},
			locale:    "th-u-ca-buddhist",
			data:      "th",
			extension: map[string]string{"ca": "buddhist"},
		},
		{
			name: "derived available locale keeps backed data locale",
			opts: ResolveOptions{
				Algorithm:             AlgorithmLookup,
				Supported:             []string{"zh-Hant-HK", "en"},
				Requested:             []string{"zh-HK-u-nu-hanidec"},
				DefaultLocale:         "en",
				RelevantExtensionKeys: []string{"nu"},
				LocaleData: testLocaleData{
					"zh-Hant-HK": {"nu": []string{"latn", "hanidec"}},
				},
			},
			locale:    "zh-HK-u-nu-hanidec",
			data:      "zh-Hant-HK",
			extension: map[string]string{"nu": "hanidec"},
		},
		{
			name: "unsupported extension and option fall back to locale data default",
			opts: ResolveOptions{
				Algorithm:             AlgorithmLookup,
				Supported:             []string{"en"},
				Requested:             []string{"en-u-ca-buddhist"},
				DefaultLocale:         "en",
				RelevantExtensionKeys: []string{"ca"},
				OptionValues:          []Option{{Key: "ca", Value: "islamic"}},
				LocaleData: testLocaleData{
					"en": {"ca": []string{"gregory"}},
				},
			},
			locale:    "en",
			data:      "en",
			extension: map[string]string{"ca": "gregory"},
		},
		{
			name: "empty requested",
			opts: ResolveOptions{
				Algorithm:     AlgorithmBestFit,
				Supported:     []string{"zh-Hant-TW", "en"},
				Requested:     nil,
				DefaultLocale: "en",
			},
			locale: "en",
			data:   "en",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ResolveLocale(tc.opts)
			if got.Locale != tc.locale || got.DataLocale != tc.data {
				t.Fatalf("ResolveLocale() = %#v, want locale %q data %q", got, tc.locale, tc.data)
			}
			for key, want := range tc.extension {
				if got.Extensions[key] != want {
					t.Fatalf("ResolveLocale().Extensions[%q] = %q, want %q", key, got.Extensions[key], want)
				}
			}
		})
	}
}
