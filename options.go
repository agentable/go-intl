package gointl

import "github.com/agentable/go-intl/option"

// Int returns a pointer to v for optional integer Options fields.
//
// It re-exports [option.Int] for namespace fidelity. Single-formatter services
// should import the zero-dependency github.com/agentable/go-intl/option package
// directly instead of the aggregate root.
func Int(v int) *int {
	return option.Int(v)
}

// Bool returns a pointer to v for optional boolean Options fields. Re-exports
// [option.Bool]; see [Int] for the single-formatter import guidance.
func Bool(v bool) *bool {
	return option.Bool(v)
}

// String returns a pointer to v for optional string Options fields. Re-exports
// [option.String]; see [Int] for the single-formatter import guidance.
func String[T ~string](v T) *string {
	return option.String(v)
}
