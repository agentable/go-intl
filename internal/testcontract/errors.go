package testcontract

import (
	"errors"
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
)

// AssertIntlError verifies structured Intl error context.
func AssertIntlError(t testing.TB, err error, kind intlerr.ErrorKind, owner, name, value, loc string) {
	t.Helper()

	intlErr, ok := errors.AsType[*intlerr.Error](err)
	if !ok {
		t.Fatalf("error = %T, want intlerr.Error", err)
	}
	if intlErr.Kind != kind || intlErr.Owner != owner || intlErr.Name != name || intlErr.Value != value || intlErr.Locale != loc {
		t.Fatalf("intlerr.Error = %+v, want kind=%q owner=%q name=%q value=%q locale=%q", intlErr, kind, owner, name, value, loc)
	}
}

// AssertErrorExpected verifies structured Intl error expected guidance.
func AssertErrorExpected(t testing.TB, err error, expected string) {
	t.Helper()

	intlErr, ok := errors.AsType[*intlerr.Error](err)
	if !ok {
		t.Fatalf("error = %T, want intlerr.Error", err)
	}
	if intlErr.Expected != expected {
		t.Fatalf("intlerr.Error.Expected = %q, want %q", intlErr.Expected, expected)
	}
}

// AssertOptionError verifies structured formatter option-error context.
func AssertOptionError(t testing.TB, err error, owner string, kind intlerr.ErrorKind, name, value, loc string) {
	t.Helper()

	AssertIntlError(t, err, kind, owner, name, value, loc)
}

// AssertOptionExpected verifies structured option-error expected guidance.
func AssertOptionExpected(t testing.TB, err error, expected string) {
	t.Helper()

	AssertErrorExpected(t, err, expected)
}
