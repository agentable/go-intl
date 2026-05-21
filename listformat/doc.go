// Package listformat implements the ECMA-402 Intl.ListFormat constructor.
//
//	format, _ := listformat.New(locale.MustParseList("en-US"), listformat.Options{})
//	out := format.Format([]string{"apples", "bananas"})
//	_ = out
//
// See README.md for usage examples and SPECS/41-listformat.md for the contract.
package listformat
