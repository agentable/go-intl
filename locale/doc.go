// Package locale implements the ECMA-402 Intl.Locale model on top of
// golang.org/x/text/language.Tag.
//
// # Errors
//
// Parse and New return errors that callers classify with errors.Is:
//
//   - ErrInvalidLocale: Parse rejects an empty tag, private-use-only tag, tag
//     containing underscores, BCP 47 parse failure, or invalid Unicode extension
//     field; New rejects invalid option-derived fields. Standalone. Caller
//     pattern: errors.Is(err, locale.ErrInvalidLocale). Suggested mapping:
//     HTTP 400 Bad Request, gRPC InvalidArgument, or CLI usage error.
//
// The package also re-exports the shared ECMA-402 option sentinel for locale
// adjacent option parsing:
//
//   - ErrInvalidOption: ECMA-402 RangeError-equivalent option failure.
//     Standalone alias of internal/ecma402.ErrInvalidOption. Caller pattern:
//     errors.Is(err, locale.ErrInvalidOption). Suggested mapping: HTTP 400
//     Bad Request, gRPC InvalidArgument, or CLI usage error.
package locale
