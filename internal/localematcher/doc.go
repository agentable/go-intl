// Package localematcher implements ECMA-402 locale matching.
//
// It provides lookup, best-fit, supported-locale filtering, and ResolveLocale
// behavior over generated locale availability data.
//
// Only constructors and root supported-locale helpers should use this package;
// public callers use constructor SupportedLocalesOf methods.
package localematcher
