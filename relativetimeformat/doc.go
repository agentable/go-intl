// Package relativetimeformat implements the ECMA-402 Intl.RelativeTimeFormat constructor.
//
//	format, _ := relativetimeformat.New(locale.MustParseList("en-US"), relativetimeformat.Options{})
//	out, _ := format.FormatInt(-1, relativetimeformat.Day)
//	_ = out
//
// See README.md for usage examples and SPECS/42-relativetimeformat.md for the contract.
package relativetimeformat
