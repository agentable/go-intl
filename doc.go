// Package gointl provides the root ECMA-402 Intl namespace for Go.
//
// It exposes common namespace functions and aliases for the active
// Intl.Locale, Intl.NumberFormat, Intl.DateTimeFormat, and Intl.PluralRules
// constructor packages. Formatter construction and formatting methods live in
// those packages.
//
// # Errors
//
// SupportedValuesOf returns ErrInvalidKey for unknown keys. The root package
// also re-exports selected package sentinels for caller convenience. Callers
// should classify errors with errors.Is, not by matching strings:
//
//   - ErrInvalidKey: SupportedValuesOf received a key outside the ECMA-402
//     supported set. Caller pattern: errors.Is(err, gointl.ErrInvalidKey).
//   - ErrInvalidOption: an underlying formatter rejected an option value. It
//     matches internal ECMA-402 option validation. Caller pattern:
//     errors.Is(err, gointl.ErrInvalidOption). Suggested mapping: HTTP 400
//     Bad Request, gRPC InvalidArgument, or CLI usage error.
//   - ErrUnsupportedLocale: locale.Parse or locale.New rejected locale input.
//     Standalone alias of locale.ErrInvalidLocale. Caller pattern:
//     errors.Is(err, gointl.ErrUnsupportedLocale). Suggested mapping:
//     HTTP 400 Bad Request, gRPC InvalidArgument, or CLI usage error.
//   - ErrUnsupportedTimeZone: DateTimeFormat rejected a time-zone option.
//     Standalone. Caller pattern:
//     errors.Is(err, gointl.ErrUnsupportedTimeZone). Suggested mapping:
//     HTTP 400 Bad Request, gRPC InvalidArgument, or CLI usage error.
package gointl
