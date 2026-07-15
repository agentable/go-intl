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
	InvalidOption     ErrorKind = "invalidOption"
	UnsupportedOption ErrorKind = "unsupportedOption"
	InvalidValue      ErrorKind = "invalidValue"
	InvalidCode       ErrorKind = "invalidCode"
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

// NewInvalidOptionExpected returns a structured invalid-option error with
// caller-provided expected guidance.
func NewInvalidOptionExpected(owner, name, value, locale, expected string, cause error) error {
	return NewExpected(InvalidOption, owner, name, value, locale, expected, cause)
}

// NewUnsupportedOptionExpected returns a structured unsupported-option error
// with caller-provided expected guidance.
func NewUnsupportedOptionExpected(owner, name, value, locale, expected string, cause error) error {
	return NewExpected(UnsupportedOption, owner, name, value, locale, expected, cause)
}

// NewInvalidValueExpected returns a structured invalid-value error with
// caller-provided expected guidance.
func NewInvalidValueExpected(owner, name, value, locale, expected string, cause error) error {
	return NewExpected(InvalidValue, owner, name, value, locale, expected, cause)
}

// NewInvalidCodeExpected returns a structured invalid-code error with
// caller-provided expected guidance.
func NewInvalidCodeExpected(owner, name, value, locale, expected string, cause error) error {
	return NewExpected(InvalidCode, owner, name, value, locale, expected, cause)
}

// NewExpected returns an Error with caller-provided "expected" guidance.
func NewExpected(kind ErrorKind, owner, name, value, locale, expected string, cause error) error {
	err := &Error{
		Kind:     kind,
		Owner:    owner,
		Name:     name,
		Value:    value,
		Locale:   locale,
		Expected: expected,
		Err:      causeWithKind(kind, cause),
	}
	if err.Expected == "" {
		err.Expected = err.expectedText()
	}
	return err
}

func causeWithKind(kind ErrorKind, cause error) error {
	sentinel := sentinelOf(kind)
	if cause == nil {
		return sentinel
	}
	if errors.Is(cause, sentinel) {
		return cause
	}
	return errors.Join(sentinel, cause)
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
	default:
		return string(k)
	}
}

func (k ErrorKind) isUnsupported() bool {
	switch k {
	case UnsupportedOption:
		return true
	case InvalidOption, InvalidValue, InvalidCode:
		return false
	default:
		return false
	}
}
