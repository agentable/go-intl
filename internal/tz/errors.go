package tz

import "errors"

// ErrUnsupportedTimeZone classifies unresolved IANA names and invalid fixed offsets.
//
// datetimeformat maps this internal sentinel to its public
// ErrUnsupportedTimeZone at the package boundary.
var ErrUnsupportedTimeZone = errors.New("tz: unsupported time zone")
