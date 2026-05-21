// Package plural evaluates generated CLDR plural rules.
//
// It provides the shared rule engine used by PluralRules and formatter code
// that needs localized category selection.
//
// Only internal plural-aware formatters should use this package; public callers
// use pluralrules APIs.
package plural
