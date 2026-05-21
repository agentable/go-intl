// Package durationformat implements the ECMA-402 Intl.DurationFormat constructor.
//
//	format, _ := durationformat.New(locale.MustParseList("en-US"), durationformat.Options{})
//	out, _ := format.Format(durationformat.Duration{Hours: 1, Minutes: 2})
//	_ = out
//
// See README.md for usage examples and SPECS/43-durationformat.md for the contract.
package durationformat
