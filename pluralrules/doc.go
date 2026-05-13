// Package pluralrules implements ECMA-402 Intl.PluralRules.
//
// # Errors
//
// New returns errors that callers classify with errors.Is:
//
//   - ErrInvalidOption: New rejects an invalid type or digit range option.
//     Standalone alias of internal/ecma402.ErrInvalidOption. Caller pattern:
//     errors.Is(err, pluralrules.ErrInvalidOption). Suggested mapping:
//     HTTP 400 Bad Request, gRPC InvalidArgument, or CLI usage error.
//   - ErrInvalidValue: typed select methods reject malformed or non-finite
//     numeric input. Caller pattern:
//     errors.Is(err, pluralrules.ErrInvalidValue).
package pluralrules
