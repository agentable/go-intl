package gointl

import "github.com/agentable/go-intl/internal/intlerr"

// ErrorKind is the category of an [Error], readable from Error.Kind after
// extracting the detail with errors.AsType[*gointl.Error]. The constants below
// are its legal values. Prefer errors.Is with a category sentinel when the
// structured fields are not needed.
type ErrorKind = intlerr.ErrorKind

const (
	// InvalidOption is the Error.Kind for ECMA-402 RangeError-equivalent option failures.
	InvalidOption = intlerr.InvalidOption
	// UnsupportedOption is the Error.Kind for valid options the active backend cannot apply.
	UnsupportedOption = intlerr.UnsupportedOption
	// InvalidValue is the Error.Kind for malformed or non-finite runtime values.
	InvalidValue = intlerr.InvalidValue
	// InvalidCode is the Error.Kind for invalid DisplayNames code inputs.
	InvalidCode = intlerr.InvalidCode
)

// Error records structured context for caller-fixable Intl failures. Kind is
// the root category; Owner is the Intl package or root namespace; Name is the
// rejected option, argument, key, code, or field; Value is the rejected value
// and can be empty; Locale is empty unless the failure is locale-dependent;
// Expected is human guidance; and Err wraps the category sentinel plus any
// underlying cause.
//
// Caller pattern:
//
//	detail, ok := errors.AsType[*gointl.Error](err)
//
// Suggested mapping: choose the external invalid-input or unsupported-
// capability response from detail.Kind; do not expose Error() text directly.
type Error = intlerr.Error

var (
	// ErrInvalidOption classifies constructor and SupportedLocalesOf option
	// validation failures. It is a standalone category sentinel.
	//
	// Caller pattern:
	//
	//	errors.Is(err, gointl.ErrInvalidOption)
	//
	// Suggested mapping: invalid caller input, such as HTTP 400 or CLI usage exit.
	ErrInvalidOption = intlerr.ErrInvalidOption
	// ErrUnsupportedOption classifies valid ECMA-402 options not backed by the
	// active implementation. It is a standalone category sentinel and also
	// matches errors.ErrUnsupported.
	//
	// Caller pattern:
	//
	//	errors.Is(err, gointl.ErrUnsupportedOption)
	//
	// Suggested mapping: an unsupported-capability response.
	ErrUnsupportedOption = intlerr.ErrUnsupportedOption
	// ErrInvalidValue classifies malformed, non-finite, or otherwise invalid
	// runtime formatting values. It is a standalone category sentinel.
	//
	// Caller pattern:
	//
	//	errors.Is(err, gointl.ErrInvalidValue)
	//
	// Suggested mapping: invalid caller input, such as HTTP 400 or CLI usage exit.
	ErrInvalidValue = intlerr.ErrInvalidValue
	// ErrInvalidCode classifies DisplayNames.Of codes outside the formatter's
	// resolved type. It is a standalone category sentinel.
	//
	// Caller pattern:
	//
	//	errors.Is(err, gointl.ErrInvalidCode)
	//
	// Suggested mapping: invalid caller input, such as HTTP 400 or CLI usage exit.
	ErrInvalidCode = intlerr.ErrInvalidCode
)
