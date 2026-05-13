package localematcher

import "github.com/agentable/go-intl/locale"

// FilterLocales canonical-deduplicates requested locales, then returns the
// locales matched by supported locales while preserving requested order and
// Unicode extension state.
func FilterLocales(supported []string, requested []locale.Locale, matcher Algorithm) []locale.Locale {
	seen := map[string]bool{}
	out := make([]locale.Locale, 0, len(requested))
	for _, loc := range requested {
		key := loc.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		if Match([]string{key}, supported, "", matcher).Locale != "" {
			out = append(out, loc)
		}
	}
	return out
}
