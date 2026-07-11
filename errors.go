package gointl

import "github.com/agentable/go-intl/internal/intlerr"

// ErrorKind is the category of an [Error], readable via the exported Error.Kind
// field after errors.As. The constants below are its legal values; classify by
// category either with errors.Is against a sentinel (e.g. ErrInvalidOption) or by
// switching on Error.Kind.
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
	// InvalidKey is the Error.Kind for invalid root namespace keys.
	InvalidKey = intlerr.InvalidKey
	// UnsupportedLocale is the Error.Kind for locale requests outside the active data set.
	UnsupportedLocale = intlerr.UnsupportedLocale
	// UnsupportedBackend is the Error.Kind for unavailable required implementation support.
	UnsupportedBackend = intlerr.UnsupportedBackend
)

// Error records structured Intl error context and human guidance.
type Error = intlerr.Error

var (
	// ErrInvalidOption classifies ECMA-402 RangeError-equivalent option failures.
	ErrInvalidOption = intlerr.ErrInvalidOption
	// ErrUnsupportedOption classifies valid options not backed by the active implementation.
	ErrUnsupportedOption = intlerr.ErrUnsupportedOption
	// ErrInvalidValue classifies malformed or non-finite runtime values.
	ErrInvalidValue = intlerr.ErrInvalidValue
	// ErrInvalidCode classifies invalid DisplayNames code inputs.
	ErrInvalidCode = intlerr.ErrInvalidCode
	// ErrInvalidKey classifies invalid root namespace keys.
	ErrInvalidKey = intlerr.ErrInvalidKey
	// ErrUnsupportedLocale classifies locale requests outside the active data set.
	ErrUnsupportedLocale = intlerr.ErrUnsupportedLocale
	// ErrUnsupportedBackend classifies unavailable required implementation support.
	ErrUnsupportedBackend = intlerr.ErrUnsupportedBackend
)
