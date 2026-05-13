package localematcher

import (
	"fmt"

	"github.com/agentable/go-intl/locale"
)

func CanonicalizeLocaleList(locales any) ([]string, error) {
	seen := map[string]struct{}{}
	out := []string{}
	add := func(loc locale.Locale) {
		canonical := loc.String()
		if _, ok := seen[canonical]; ok {
			return
		}
		seen[canonical] = struct{}{}
		out = append(out, canonical)
	}
	switch v := locales.(type) {
	case nil:
		return out, nil
	case string:
		loc, err := locale.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("localematcher: invalid locale %q: %w", v, ErrInvalidLocale)
		}
		add(loc)
	case locale.Locale:
		add(v)
	case []string:
		for _, s := range v {
			loc, err := locale.Parse(s)
			if err != nil {
				return nil, fmt.Errorf("localematcher: invalid locale %q: %w", s, ErrInvalidLocale)
			}
			add(loc)
		}
	case []locale.Locale:
		for _, loc := range v {
			add(loc)
		}
	default:
		return nil, fmt.Errorf("localematcher: invalid locale list %T: %w", locales, ErrInvalidLocale)
	}
	return out, nil
}
