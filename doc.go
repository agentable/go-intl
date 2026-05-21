// Package gointl provides the root ECMA-402 Intl namespace for Go.
//
// It exposes common namespace functions and aliases for the active
// Intl.Locale, Intl.NumberFormat, Intl.DateTimeFormat, Intl.PluralRules,
// Intl.ListFormat, Intl.RelativeTimeFormat, Intl.DurationFormat,
// Intl.DisplayNames, Intl.Collator, and Intl.Segmenter constructor packages.
// Formatter construction and formatting methods live in those packages.
//
// Import formatter subpackages directly when an application needs one
// constructor. Importing this root package is an aggregate facade and pulls the
// active constructor packages so namespace aliases remain available. Measure
// root dependency and binary-size reports as aggregate facade cost; measure
// formatter subpackages separately for single-constructor services.
//
// # Errors
//
// Formatter construction and formatting errors wrap the root error categories.
// Callers classify errors with errors.Is, not by matching strings. Suggested
// mappings are for applications that expose this library through HTTP, gRPC,
// CLI, or another user boundary.
//
// ErrInvalidOption is returned when constructor or SupportedLocalesOf option
// validation rejects an option name/value combination. It is a standalone
// category sentinel. Caller pattern:
//
//	errors.Is(err, gointl.ErrInvalidOption)
//
// Suggested mapping: invalid caller input, such as HTTP 400 or CLI usage exit.
//
// ErrUnsupportedOption is returned when a valid ECMA-402 option is not backed by
// the active implementation. It is a standalone category sentinel. Caller
// pattern:
//
//	errors.Is(err, gointl.ErrUnsupportedOption)
//
// It also matches errors.ErrUnsupported. Caller pattern:
//
//	errors.Is(err, errors.ErrUnsupported)
//
// Suggested mapping: unsupported capability, such as HTTP 501 or a clear CLI
// unsupported-feature message.
//
// ErrInvalidValue is returned when a runtime formatting value is malformed,
// non-finite, or otherwise outside the target Intl operation. It is a
// standalone category sentinel. Caller pattern:
//
//	errors.Is(err, gointl.ErrInvalidValue)
//
// Suggested mapping: invalid caller input, such as HTTP 400 or CLI usage exit.
//
// ErrInvalidCode is returned when DisplayNames.Of receives a code outside its
// resolved type. It is a standalone category sentinel. Caller pattern:
//
//	errors.Is(err, gointl.ErrInvalidCode)
//
// Suggested mapping: invalid caller input, such as HTTP 400 or CLI usage exit.
//
// ErrInvalidKey is returned when a root namespace key is outside the ECMA-402
// supported set. It is a standalone category sentinel. Caller pattern:
//
//	errors.Is(err, gointl.ErrInvalidKey)
//
// Suggested mapping: invalid caller input, such as HTTP 400 or CLI usage exit.
//
// ErrUnsupportedLocale is returned when locale negotiation cannot satisfy a
// requested locale with the active data set. It is a standalone category
// sentinel. Caller pattern:
//
//	errors.Is(err, gointl.ErrUnsupportedLocale)
//
// It also matches errors.ErrUnsupported. Caller pattern:
//
//	errors.Is(err, errors.ErrUnsupported)
//
// Suggested mapping: unsupported locale input, such as HTTP 400 or 406 depending
// on the host API contract.
//
// ErrUnsupportedBackend is returned when required implementation support is
// unavailable. It is a standalone category sentinel. Caller pattern:
//
//	errors.Is(err, gointl.ErrUnsupportedBackend)
//
// It also matches errors.ErrUnsupported. Caller pattern:
//
//	errors.Is(err, errors.ErrUnsupported)
//
// Suggested mapping: service/configuration capability failure, such as HTTP 500
// or 503.
//
// Public caller-fixable errors also expose *gointl.Error. Caller pattern:
//
//	detail, ok := errors.AsType[*gointl.Error](err)
//
// Error fields are stable machine-readable context: Kind is the root category;
// Owner is the Intl package or root namespace; Name is the option, argument,
// key, code, or field; Value is the rejected value and can be empty when the
// invalid input was omitted; Locale is set only for locale-dependent failures;
// Expected is human guidance; Err is the wrapped sentinel or underlying cause.
package gointl
