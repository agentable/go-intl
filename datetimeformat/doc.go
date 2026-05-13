// Package datetimeformat implements ECMA-402 Intl.DateTimeFormat.
//
// # Errors
//
// New returns errors that callers classify with errors.Is:
//
//   - ErrInvalidOption: New rejects an invalid date/time field option,
//     dateStyle/timeStyle combination, fractionalSecondDigits value, or hour
//     cycle option. Standalone alias of internal/ecma402.ErrInvalidOption.
//     Caller pattern: errors.Is(err, datetimeformat.ErrInvalidOption).
//     Suggested mapping: HTTP 400 Bad Request, gRPC InvalidArgument, or CLI
//     usage error.
//   - ErrUnsupportedTimeZone: New rejects an invalid IANA time zone, unknown
//     zone, or invalid fixed-offset time zone from Options.TimeZone. Standalone.
//     Caller pattern: errors.Is(err, datetimeformat.ErrUnsupportedTimeZone).
//     Suggested mapping: HTTP 400 Bad Request, gRPC InvalidArgument, or CLI
//     usage error.
package datetimeformat
