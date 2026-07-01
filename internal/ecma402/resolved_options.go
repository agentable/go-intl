package ecma402

// ResolvedScalar returns a fresh pointer for a scalar resolved option property
// whose ECMA-402 omission is represented by nil in Go.
func ResolvedScalar[T any](value T) *T {
	return &value
}

// CloneResolvedScalar copies a scalar resolved option pointer so ResolvedOptions
// never exposes mutable formatter state to callers.
func CloneResolvedScalar[T any](value *T) *T {
	if value == nil {
		return nil
	}
	return ResolvedScalar(*value)
}

// ResolvedScalarValue returns the scalar value for a resolved option pointer,
// using the zero value when ECMA-402 omits that property.
func ResolvedScalarValue[T any](value *T) T {
	if value == nil {
		var zero T
		return zero
	}
	return *value
}
