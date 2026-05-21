package locale

// List is an ordered Intl locale request list.
//
// A nil or empty List represents omitted locales at constructor boundaries.
type List []Locale

// ParseList parses locale identifiers into an ordered request list.
func ParseList(tags ...string) (List, error) {
	out := make(List, 0, len(tags))
	for _, tag := range tags {
		loc, err := Parse(tag)
		if err != nil {
			return nil, err
		}
		out = append(out, loc)
	}
	return out, nil
}

// MustParseList is like ParseList but panics on invalid input.
func MustParseList(tags ...string) List {
	locales, err := ParseList(tags...)
	if err != nil {
		panic(err)
	}
	return locales
}

// CanonicalizeList returns the first occurrence of each canonical locale while
// preserving request order.
func CanonicalizeList(locales List) List {
	seen := map[string]bool{}
	out := make(List, 0, len(locales))
	for _, loc := range locales {
		key := loc.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, loc)
	}
	return out
}

// Strings returns the canonical locale identifiers in request order.
func (l List) Strings() []string {
	canonical := CanonicalizeList(l)
	out := make([]string, len(canonical))
	for i, loc := range canonical {
		out[i] = loc.String()
	}
	return out
}
