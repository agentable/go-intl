// Hand-written accessor layer for the relativetime domain. It exposes
// relative-field data and the narrow supported-locale index over lazily decoded
// const blobs.

package relativetime

import (
	"maps"

	cldrlocale "github.com/agentable/go-intl/internal/cldr/locale"
)

// ResolveLocale resolves a tag to its kernel locale handle, forwarding to the
// shared locale kernel so relativetime handles index identically to every other
// domain.
func ResolveLocale(tag string) (Locale, bool) {
	return cldrlocale.ResolveLocale(tag)
}

// FieldsFor returns a caller-owned copy of the relative-time field data for the
// locale, or nil when the locale carries none.
func FieldsFor(loc Locale) RelativeTimeFields {
	fieldOnce.Do(loadFields)
	return cloneFields(relativeTimeFieldsForLocale(loc))
}

func relativeTimeFieldsForLocale(loc Locale) RelativeTimeFields {
	fields, ok := fieldsByLocale[loc]
	if !ok {
		return nil
	}
	return fields
}

func cloneFields(fields RelativeTimeFields) RelativeTimeFields {
	if fields == nil {
		return nil
	}
	out := make(RelativeTimeFields, len(fields))
	for unit, styles := range fields {
		out[unit] = cloneStyleFields(styles)
	}
	return out
}

func cloneStyleFields(styles map[string]RelativeTimeField) map[string]RelativeTimeField {
	if styles == nil {
		return nil
	}
	out := make(map[string]RelativeTimeField, len(styles))
	for style, field := range styles {
		out[style] = cloneField(field)
	}
	return out
}

func cloneField(field RelativeTimeField) RelativeTimeField {
	return RelativeTimeField{
		Future:   maps.Clone(field.Future),
		Past:     maps.Clone(field.Past),
		Relative: maps.Clone(field.Relative),
	}
}

// SupportedLocales returns the relative-time-supported locale tags in
// sorted-locale order. It reads only the narrow supported blob and never
// triggers the field blob decode.
func SupportedLocales() []string {
	return supported.Get()
}
