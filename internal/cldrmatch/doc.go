// Package cldrmatch matches locale identifiers against generated CLDR data sets.
//
// It centralizes the small matching helpers needed when CLDR payload keys and
// requested BCP 47 locales do not have identical spelling.
//
// Only internal data accessors and locale negotiation code should use this
// package; public callers go through formatter constructors.
package cldrmatch
