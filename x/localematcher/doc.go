// Package localematcher matches requested locales against an application's
// available locales with the same ECMA-402 lookup and best-fit algorithms
// that back the go-intl Intl constructors.
//
// The API shape follows the Stage 1 TC39 Intl.LocaleMatcher proposal
// (https://github.com/tc39/proposal-intl-localematcher): Match is the Go
// bridge for Intl.LocaleMatcher.match, and New compiles a reusable matcher
// for repeated matching against a fixed available-locale set, the same
// constructor-side compilation the Intl formatters perform.
//
// Both entry points pre-wire the CLDR likely-subtags maximizer used by the
// Intl constructors, so best-fit results match constructor locale
// negotiation. Result carries the matched locale plus the data locale,
// Unicode extension, and CLDR distance produced by the match.
//
// Experimental: this package lives in the github.com/agentable/go-intl/x
// module. It is not part of the ECMA-402 surface, stays v0 with no
// compatibility promise, and may change or be removed as the proposal
// evolves. SPECS/80-x-modules.md owns the module charter.
package localematcher
