package tz

import (
	"errors"
	"fmt"
)

// ErrUnsupportedTimeZone classifies unresolved IANA names and invalid fixed offsets.
//
// datetimeformat maps this internal sentinel to its public
// ErrUnsupportedTimeZone at the package boundary.
var ErrUnsupportedTimeZone error = unsupportedTimeZoneError{}

type unsupportedTimeZoneError struct {
	reason string
	name   string
}

func (e unsupportedTimeZoneError) Error() string {
	reason := e.reason
	if reason == "" {
		reason = "unsupported time zone"
	}
	if e.name != "" {
		return fmt.Sprintf("tz: %s %q", reason, e.name)
	}
	return "tz: " + reason
}

func (unsupportedTimeZoneError) Is(target error) bool {
	if target == errors.ErrUnsupported {
		return true
	}
	_, ok := target.(unsupportedTimeZoneError)
	return ok
}

func unsupportedTimeZone(name string) error {
	return unsupportedTimeZoneError{name: name}
}

func invalidOffset(name string) error {
	return unsupportedTimeZoneError{reason: "invalid offset", name: name}
}
