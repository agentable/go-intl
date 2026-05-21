// Package decimal provides decimal arithmetic for ECMA-402 mathematical values.
//
// It wraps the apd backend with the finite, NaN, infinity, rounding, and digit
// operations required by NumberFormat and PluralRules.
//
// Only internal ECMA-402 algorithms and formatter implementations should use
// this package; public callers pass ordinary Go values to formatter methods.
package decimal
