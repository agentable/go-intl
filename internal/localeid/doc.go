// Package localeid canonicalizes and expands BCP 47 locale identifiers.
//
// It isolates language.Tag bridging, likely-subtag maximize behavior, and
// locale identifier normalization used across Locale and formatter setup.
//
// Only internal locale construction and matching code should use this package;
// public callers use the locale package.
package localeid
