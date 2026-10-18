package localematcher

type localeIdentifier interface {
	String() string
}

// FilterLocales returns requested locales matched by supported locales while
// preserving requested order and Unicode extension state.
func FilterLocales[T localeIdentifier](compiled *Matcher, requested []T, matcher Algorithm) []T {
	var singleRequested [1]string
	out := make([]T, 0, len(requested))
	for _, loc := range requested {
		singleRequested[0] = loc.String()
		if compiled.Match(singleRequested[:], "", matcher).Locale != "" {
			out = append(out, loc)
		}
	}
	return out
}
