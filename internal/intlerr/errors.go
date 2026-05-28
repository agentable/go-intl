// Package intlerr provides the cycle-free implementation behind the root Intl
// error surface.
package intlerr

import (
	"errors"
	"fmt"
	"strings"
)

// ErrorKind classifies Intl errors for stable errors.Is category matching.
type ErrorKind string

const (
	InvalidOption      ErrorKind = "invalidOption"
	UnsupportedOption  ErrorKind = "unsupportedOption"
	InvalidValue       ErrorKind = "invalidValue"
	InvalidCode        ErrorKind = "invalidCode"
	InvalidKey         ErrorKind = "invalidKey"
	UnsupportedLocale  ErrorKind = "unsupportedLocale"
	UnsupportedBackend ErrorKind = "unsupportedBackend"
)

var (
	// ErrInvalidOption classifies ECMA-402 RangeError-equivalent option failures.
	ErrInvalidOption = sentinelOf(InvalidOption)
	// ErrUnsupportedOption classifies valid options not backed by the active implementation.
	ErrUnsupportedOption = sentinelOf(UnsupportedOption)
	// ErrInvalidValue classifies malformed or non-finite runtime values.
	ErrInvalidValue = sentinelOf(InvalidValue)
	// ErrInvalidCode classifies invalid DisplayNames code inputs.
	ErrInvalidCode = sentinelOf(InvalidCode)
	// ErrInvalidKey classifies invalid root namespace keys.
	ErrInvalidKey = sentinelOf(InvalidKey)
	// ErrUnsupportedLocale classifies locale requests outside the active data set.
	ErrUnsupportedLocale = sentinelOf(UnsupportedLocale)
	// ErrUnsupportedBackend classifies unavailable required implementation support.
	ErrUnsupportedBackend = sentinelOf(UnsupportedBackend)
)

type kindError struct {
	kind ErrorKind
}

func sentinelOf(kind ErrorKind) error {
	return kindError{kind: kind}
}

func (e kindError) Error() string {
	return "intl: " + e.kind.label()
}

func (e kindError) Is(target error) bool {
	if target == errors.ErrUnsupported {
		return e.kind.isUnsupported()
	}
	switch target := target.(type) {
	case kindError:
		return e.kind == target.kind
	case *Error:
		return e.kind == target.Kind
	default:
		return false
	}
}

// Error records structured Intl error context while preserving the wrapped
// error for errors.Is and errors.AsType.
type Error struct {
	Kind     ErrorKind
	Owner    string
	Name     string
	Value    string
	Locale   string
	Expected string
	Err      error
}

// New returns an Error carrying stable Intl context and a wrapped error.
func New(kind ErrorKind, owner, name, value, locale string, cause error) error {
	return NewExpected(kind, owner, name, value, locale, "", cause)
}

// NewExpected returns an Error with caller-provided "expected" guidance.
func NewExpected(kind ErrorKind, owner, name, value, locale, expected string, cause error) error {
	if cause == nil {
		cause = sentinelOf(kind)
	}
	err := &Error{
		Kind:     kind,
		Owner:    owner,
		Name:     name,
		Value:    value,
		Locale:   locale,
		Expected: expected,
		Err:      cause,
	}
	if err.Expected == "" {
		err.Expected = err.expectedText()
	}
	return err
}

func (e *Error) Error() string {
	var b strings.Builder
	b.WriteString(e.Owner)
	b.WriteString(": ")
	b.WriteString(e.Kind.label())
	if e.Name != "" {
		b.WriteByte(' ')
		b.WriteString(e.Name)
	}
	if e.Value != "" {
		fmt.Fprintf(&b, " %q", e.Value)
	}
	if e.Locale != "" {
		fmt.Fprintf(&b, " for locale %q", e.Locale)
	}
	if expected := e.expectedText(); expected != "" {
		b.WriteString(": expected ")
		b.WriteString(expected)
	}
	b.WriteString("; got ")
	if e.Value == "" {
		b.WriteString("empty value")
	} else {
		fmt.Fprintf(&b, "%q", e.Value)
	}
	return b.String()
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Is(target error) bool {
	if target == errors.ErrUnsupported {
		return e.Kind.isUnsupported()
	}
	switch target := target.(type) {
	case kindError:
		return e.Kind == target.kind || errors.Is(e.Err, target)
	case *Error:
		return e.Kind == target.Kind
	default:
		return false
	}
}

func (e *Error) expectedText() string {
	if e.Expected != "" {
		return e.Expected
	}
	switch e.Kind {
	case InvalidOption:
		if e.Name != "" {
			return fmt.Sprintf("a supported value for option %q", e.Name)
		}
		return "a supported option value"
	case UnsupportedOption:
		if e.Name != "" {
			return fmt.Sprintf("an implementation-supported value for option %q", e.Name)
		}
		return "an implementation-supported option value"
	case InvalidValue:
		if e.Name != "" {
			return fmt.Sprintf("a well-formed Intl value for %q", e.Name)
		}
		return "a well-formed Intl value"
	case InvalidCode:
		if e.Name != "" {
			return fmt.Sprintf("a well-formed code for %q", e.Name)
		}
		return "a well-formed code"
	case InvalidKey:
		return "a supported Intl key"
	case UnsupportedLocale:
		return "a locale supported by the active data set"
	case UnsupportedBackend:
		return "an available implementation backend"
	default:
		return ""
	}
}

func (k ErrorKind) label() string {
	switch k {
	case InvalidOption:
		return "invalid option"
	case UnsupportedOption:
		return "unsupported option"
	case InvalidValue:
		return "invalid value"
	case InvalidCode:
		return "invalid code"
	case InvalidKey:
		return "invalid key"
	case UnsupportedLocale:
		return "unsupported locale"
	case UnsupportedBackend:
		return "unsupported backend"
	default:
		return string(k)
	}
}

func (k ErrorKind) isUnsupported() bool {
	switch k {
	case UnsupportedOption, UnsupportedLocale, UnsupportedBackend:
		return true
	case InvalidOption, InvalidValue, InvalidCode, InvalidKey:
		return false
	default:
		return false
	}
}
