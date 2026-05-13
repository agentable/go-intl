// Package ecma402 implements the ECMA-402 abstract-operation layer used by
// every formatter package (numberformat, datetimeformat, pluralrules).
// It keeps production-used algorithms and validators from the ECMA-402 spec
// and the formatjs reference implementation; see SPECS/12-abstract-operations.md.
package ecma402

import "errors"

// ErrInvalidOption is the sentinel for ECMA-402 RangeError equivalents — a
// value outside its allowed enumeration, an out-of-range numeric option, or a
// malformed identifier (currency code, unit identifier, time zone name).
var ErrInvalidOption = errors.New("ecma402: invalid option")
