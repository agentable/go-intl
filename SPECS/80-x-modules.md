# SPEC 80 — Experimental LocaleMatcher

Owns: `github.com/agentable/go-intl/x/localematcher`, an explicit experimental
exception to the active standardized constructor surface in SPEC 00.

## Authority and packaging

The behavioral anchor is the official TC39 Stage 1
[`Intl.LocaleMatcher` draft at c35e7b759f698b16e9d73d8e98647207e35a14ce](https://github.com/tc39/proposal-intl-localematcher/blob/c35e7b759f698b16e9d73d8e98647207e35a14ce/spec.emu).
Stage 1 is experimental, not a claim of ECMA-402 standardization. This revision
uses a constructor plus `prototype.match`; older static ponyfill APIs do not
own this contract.

The package is part of the root module and uses its version. There is no
satellite module, local `replace`, extra workspace member, or extra release tag.
This keeps internal matcher/data changes atomic and lets root test/lint commands
cover the package. Its API may change with the proposal; consumers should pin a
root-module version. No LocaleMatcher alias is added to the root Intl namespace.
Existing constructors do not depend on the experimental package.

## Public contract

| Go surface | Official draft owner / Go bridge |
| --- | --- |
| `Options`, `Algorithm *string` | constructor options bag and `algorithm`; nil means `"best fit"` |
| `New(available locale.List, opts Options) (*Matcher, error)` | `new Intl.LocaleMatcher(availableLocales, options)` |
| `Matcher` | instance with frozen available locales and algorithm |
| `(*Matcher).Match(requested locale.List) (locale.Locale, bool)` | `Intl.LocaleMatcher.prototype.match`; false bridges JavaScript `undefined` |

Raw locale input enters through `locale.ParseList`, preserving the shared
identity/canonicalization boundary. Construction canonicalizes and deduplicates
available locales, rejects their Unicode `u` extensions before reading options,
and validates `algorithm` as exactly `"lookup"` or `"best fit"`. It snapshots
both the list and pointed-to option. Errors use the existing structured root
categories (`ErrInvalidValue` / `ErrInvalidOption`).

Matching returns a canonical entry from the supplied resource set or no match.
Requested Unicode extensions do not appear in the result. There is no implicit
default resource. Lookup truncates each request and then continues to the next
request; it cannot manufacture `de` from an available `de-DE`. Best fit uses the
existing CLDR written-new distance profile, maximizer and threshold; the draft
allows implementation-defined best fit. No distance, threshold, DataLocale,
internal Result or Algorithm type is exported.

The zero value is an empty matcher. Constructed instances support concurrent
calls; each owns a bounded internal cache. The available map is immutable.

## Internal boundary and proof

`internal/localematcher.NewExactMatcher` compiles the exact canonical resource
set. Existing `NewMatcher` retains service-constructor locale expansion. Both
use one private compiler and the unchanged lookup/best-fit execution logic.

- `internal/localematcher/compiled_test.go`: exact resource lookup plus existing
  service-constructor behavior.
- `x/localematcher/localematcher_test.go`: validation/order/error categories,
  canonical identities, absent parents, subsequent requests, best fit, immutable
  options, concurrent calls, zero value and a per-key translated-text consumer.
- `x/localematcher/example_test.go`: executable public usage.
- Root race tests, lint, vet and module verification cover this package; an
  external consumer must also work from a root module zip with `GOWORK=off`
  and no local replacement.

Proposal changes require revisiting this pinned contract and its tests before
changing the API. Promotion into the standardized surface requires an accepted
standard and complete implementation evidence; it is not automatic.
