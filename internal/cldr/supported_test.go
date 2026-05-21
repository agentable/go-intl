package cldr

import "testing"

func TestSupportedLocalesComeFromGeneratedData(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		tags []string
		has  func(Locale) bool
	}{
		{name: "number", tags: NumberSupportedLocales(), has: func(loc Locale) bool {
			_, ok := numbersByLocale[loc]
			return ok
		}},
		{name: "date", tags: DateSupportedLocales(), has: func(loc Locale) bool {
			_, ok := datesByLocale[loc]
			return ok
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if len(tc.tags) == 0 {
				t.Fatalf("%s supported locales is empty", tc.name)
			}
			for _, tag := range tc.tags {
				loc, ok := localeIndex[tag]
				if !ok {
					t.Fatalf("%s supported locale %q is missing from localeIndex", tc.name, tag)
				}
				if loc == Undefined {
					t.Fatalf("%s supported locales includes Undefined", tc.name)
				}
				if !tc.has(loc) {
					t.Fatalf("%s supported locale %q has no generated data", tc.name, tag)
				}
			}
		})
	}
}
