// Package collator implements the ECMA-402 Intl.Collator constructor.
//
//	compare, _ := collator.New(locale.MustParseList("en-US"), collator.Options{})
//	ok := compare.Compare("a", "b") < 0
//	_ = ok
//
// See README.md for usage examples and SPECS/45-collator.md for the contract.
package collator
