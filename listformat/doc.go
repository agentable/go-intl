// Package listformat implements the ECMA-402 Intl.ListFormat constructor.
//
//	locales, _ := locale.ParseList("en-US")
//	format, _ := listformat.New(locales, listformat.Options{})
//	out := format.Format([]string{"apples", "bananas"})
//	_ = out
//
// See README.md for usage examples and SPECS/41-listformat.md for the contract.
package listformat
