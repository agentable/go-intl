// Package tz resolves ECMA-402 time-zone identifiers.
//
// It combines generated IANA links, offset parsing, and default-zone handling
// for DateTimeFormat construction.
//
// Only internal date-time formatting code should use this package; public
// callers pass time-zone option strings to DateTimeFormat.
package tz
