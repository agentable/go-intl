package testcontract

import (
	"testing"

	"github.com/agentable/go-intl/internal/intlerr"
)

func TestAssertOptionError(t *testing.T) {
	t.Parallel()

	err := intlerr.NewInvalidOptionExpected("datetimeformat", "dateStyle/timeStyle", "year", "en-US", "no explicit component options", nil)
	AssertIntlError(t, err, intlerr.InvalidOption, "datetimeformat", "dateStyle/timeStyle", "year", "en-US")
	AssertErrorExpected(t, err, "no explicit component options")
	AssertOptionError(t, err, "datetimeformat", intlerr.InvalidOption, "dateStyle/timeStyle", "year", "en-US")
	AssertOptionExpected(t, err, "no explicit component options")
}
