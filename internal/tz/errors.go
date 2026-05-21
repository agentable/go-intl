package tz

import "errors"

// ErrUnsupportedTimeZone classifies unresolved IANA names and invalid fixed offsets.
//
// datetimeformat maps this internal sentinel to its public
// ErrUnsupportedTimeZone at the package boundary.
var ErrUnsupportedTimeZone error = unsupportedTimeZoneError{}

type unsupportedTimeZoneError struct{}

func (unsupportedTimeZoneError) Error() string {
	return "tz: unsupported time zone"
}

func (unsupportedTimeZoneError) Is(target error) bool {
	return target == errors.ErrUnsupported
}
