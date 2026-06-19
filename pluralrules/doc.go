// Package pluralrules implements the ECMA-402 Intl.PluralRules constructor.
//
//	locales, _ := locale.ParseList("en-US")
//	rules, _ := pluralrules.New(locales, pluralrules.Options{})
//	category, _ := rules.Select(pluralrules.Int(1))
//	_ = category
//
// See README.md for usage examples and SPECS/40-pluralrules.md for the contract.
package pluralrules
