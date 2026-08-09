// Package option provides the pointer constructors for optional scalar Intl
// options (*int, *bool, *string). It is a zero-dependency leaf so a service that
// uses a single formatter package (e.g. numberformat) can set options without
// importing the aggregate root, which pulls in all active constructor packages.
//
// The root gointl package re-exports these as gointl.Int/Bool/String for
// namespace fidelity; prefer this package directly in single-formatter services.
package option

// Int returns a pointer to v for optional integer Options fields.
func Int(v int) *int {
	return &v
}

// Bool returns a pointer to v for optional boolean Options fields.
func Bool(v bool) *bool {
	return &v
}

// String returns a pointer to v for optional string Options fields, accepting
// any string-kind type (such as a formatter's typed option enum).
func String[T ~string](v T) *string {
	value := string(v)
	return &value
}
