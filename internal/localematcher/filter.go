package localematcher

type localeIdentifier interface {
	String() string
}

// FilterLocales canonical-deduplicates requested locales, then returns the
// locales matched by supported locales while preserving requested order and
// Unicode extension state.
func FilterLocales[T localeIdentifier](supported []string, requested []T, matcher Algorithm) []T {
	return FilterLocalesWithMaximizer(supported, requested, matcher, nil)
}

func FilterLocalesWithMaximizer[T localeIdentifier](supported []string, requested []T, matcher Algorithm, maximizer Maximizer) []T {
	compiled := NewMatcher(supported, maximizer)
	seen := map[string]bool{}
	out := make([]T, 0, len(requested))
	for _, loc := range requested {
		key := loc.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		if compiled.Match([]string{key}, "", matcher).Locale != "" {
			out = append(out, loc)
		}
	}
	return out
}
