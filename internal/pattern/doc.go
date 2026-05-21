// Package pattern parses brace-delimited formatter patterns.
//
// It provides the small tokenizer used by ECMA-402 partition algorithms before
// formatter packages substitute localized field values.
//
// Only internal ECMA-402 and formatter code should use this package; public
// callers receive already formatted strings or parts.
package pattern
