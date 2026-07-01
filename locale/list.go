package locale

// List is an ordered Intl locale request list.
//
// A nil or empty List represents omitted locales at constructor boundaries.
type List []Locale

// ParseList parses locale identifiers into an ordered request list.
func ParseList(tags ...string) (List, error) {
	out := make(List, len(tags))
	for i, tag := range tags {
		loc, err := Parse(tag)
		if err != nil {
			return nil, err
		}
		out[i] = loc
	}
	return out, nil
}

// Strings returns the canonical locale identifiers in list order.
func (l List) Strings() []string {
	out := make([]string, len(l))
	for i, loc := range l {
		out[i] = loc.String()
	}
	return out
}
