// Package numberformat implements the ECMA-402 Intl.NumberFormat constructor.
//
//	locales, _ := locale.ParseList("en-US")
//	format, _ := numberformat.New(locales, numberformat.Options{})
//	out := format.Format(numberformat.Int(1234))
//	_ = out
//
// See README.md for usage examples and SPECS/20-numberformat.md for the contract.
package numberformat
