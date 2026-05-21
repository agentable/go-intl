// Package conformance loads and filters ECMA-402 conformance fixtures.
//
// It gives formatter tests one shared representation for skip lists, expected
// failures, divergences, and product coverage checks.
//
// Only tests and conformance tools should use this package; runtime formatter
// code must not depend on fixture metadata.
package conformance
