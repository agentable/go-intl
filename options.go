package gointl

// Int returns a pointer to v for optional integer Options fields.
func Int(v int) *int {
	return &v
}

// Bool returns a pointer to v for optional boolean Options fields.
func Bool(v bool) *bool {
	return &v
}

// String returns a pointer to v for optional string Options fields.
func String(v string) *string {
	return &v
}
