package gointl

import "github.com/agentable/go-intl/internal/intlerr"

type ErrorKind = intlerr.ErrorKind

const (
	InvalidOption      = intlerr.InvalidOption
	UnsupportedOption  = intlerr.UnsupportedOption
	InvalidValue       = intlerr.InvalidValue
	InvalidCode        = intlerr.InvalidCode
	InvalidKey         = intlerr.InvalidKey
	UnsupportedLocale  = intlerr.UnsupportedLocale
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
