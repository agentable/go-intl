// Package numberformat implements ECMA-402 Intl.NumberFormat.
//
// # Errors
//
// New returns errors that callers classify with errors.Is:
//
//   - ErrInvalidOption: New rejects an invalid style, currency, currencyDisplay,
//     currencySign, unitDisplay, notation, compactDisplay, signDisplay,
//     roundingMode, roundingPriority, trailingZeroDisplay, digit range, or
//     roundingIncrement constraint. Standalone alias of
//     internal/ecma402.ErrInvalidOption. Caller pattern:
//     errors.Is(err, numberformat.ErrInvalidOption). Suggested mapping:
//     HTTP 400 Bad Request, gRPC InvalidArgument, or CLI usage error.
//   - ErrInvalidValue: decimal-string formatting methods reject malformed
//     numeric input. Caller pattern:
//     errors.Is(err, numberformat.ErrInvalidValue).
package numberformat
