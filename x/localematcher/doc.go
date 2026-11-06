// Package localematcher matches requested locales against an application's
// available resource locales using lookup or CLDR-backed best fit.
//
// The API follows the official Stage 1 Intl.LocaleMatcher constructor and
// prototype.match draft, pinned in SPECS/80-x-modules.md. New snapshots the
// available locales and algorithm; Match returns an available locale or no
// match. Applications own their default fallback.
//
// Experimental: this package ships in the root go-intl module, without a root
// Intl namespace alias. Its API may change with the official proposal. Pin a
// root-module version when depending on this draft API.
package localematcher
