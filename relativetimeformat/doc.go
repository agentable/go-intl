// Package relativetimeformat implements the ECMA-402 Intl.RelativeTimeFormat constructor.
//
//	locales, _ := locale.ParseList("en-US")
//	format, _ := relativetimeformat.New(locales, relativetimeformat.Options{})
//	out, _ := format.Format(relativetimeformat.Int(-1), relativetimeformat.Day)
//	_ = out
//
// See README.md for usage examples and SPECS/42-relativetimeformat.md for the contract.
package relativetimeformat
