// Package displaynames implements the ECMA-402 Intl.DisplayNames constructor.
//
//	locales, _ := locale.ParseList("en-US")
//	names, _ := displaynames.New(locales, displaynames.Options{Type: displaynames.Language})
//	name, ok, _ := names.Of("fr")
//	_, _ = name, ok
//
// See README.md for usage examples and SPECS/44-displaynames.md for the contract.
package displaynames
