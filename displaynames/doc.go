// Package displaynames implements the ECMA-402 Intl.DisplayNames constructor.
//
//	names, _ := displaynames.New(locale.MustParseList("en-US"), displaynames.Options{Type: displaynames.Language})
//	name, ok, _ := names.Of("fr")
//	_, _ = name, ok
//
// See README.md for usage examples and SPECS/44-displaynames.md for the contract.
package displaynames
