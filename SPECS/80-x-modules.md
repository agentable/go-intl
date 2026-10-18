# SPEC 80 — Experimental `x/` Satellite Module

> **Status:** Active contract (2026-10-07)
> **Audience:** Maintainers, contributors, and consumers of `github.com/agentable/go-intl/x`.
> **Authority:** This document is the charter for the `x/` satellite module. It constrains packaging and governance only; behavioral authority for each package remains with its owning SPEC (for `x/localematcher`, [SPEC 11](./11-locale-matching.md)).

---

## 1. Purpose

The main module's public surface is closed: every exported symbol maps to a
native ECMA-402 owner (SPEC 00), and a constructor enters only with a complete
implementation. Two kinds of useful code cannot meet that bar today:

1. **Go bridges for in-flight TC39 proposals** (Stage 1–3), whose API shape can
   still change before Stage 4.
2. **Non-surface exposure of finished-spec machinery** — ECMA-402 abstract
   operations that have no JavaScript owner but that external CLDR-driven
   consumers legitimately need (motivating issue: agentable/go-intl#11).

`github.com/agentable/go-intl/x` is the satellite module that hosts both. It
lets consumers adopt such code without path-squatting into `internal/`, while
the main module keeps its rules unchanged.

> **Rejected**:
>
> - **One module per proposal** — every satellite must release in lockstep with
>   the root module anyway (it consumes `internal/*`, which carries no
>   compatibility promise), so per-proposal modules multiply tag streams and
>   release toil without buying real independence. One shared v0 module with
>   per-proposal packages carries the same signal at a fraction of the cost.
> - **Relaxing the main-module native-owner rule** — the satellite exists so
>   the rule never has to bend.
> - **`go.work` for module wiring** — `go mod tidy` ignores workspaces, so the
>   committed `require` + `replace => ..` pair in `x/go.mod` remains the only
>   wiring that keeps the atomic same-commit multi-module tag push working.
>   `go.work` stays a gitignored personal option with CI pinned to
>   `GOWORK=off`.

## 2. Governance

1. **Stability.** The module stays **v0 with no compatibility promise**. Any
   release may add, change, or remove packages and APIs. The README of the
   module and every package doc must say so.
2. **Versioning.** `x/` is a member of the coordinated release set (SPEC 00 /
   CLAUDE.md Release Inventory): tagged `x/vX.Y.Z` at the same commit and with
   the same version as the root and `tools/*` modules, in one atomic tag push.
   `x/go.mod` requires the root-module version whose internal API it was
   written against.
3. **Dependency direction.** `x/` packages may import root-module packages,
   including `internal/*` (Go internal visibility is import-path based).
   Nothing in the root module may import `x/`.
4. **Thin over internal.** `x/` packages are re-exports and narrow typed
   bridges over root-module implementation. They must not reimplement
   algorithms, carry their own data copies, or invent semantics the root
   module lacks.
5. **No new runtime dependencies** beyond what the root module already
   requires, unless the owning SPEC says otherwise.

## 3. Admission criteria

A new `x/` package must have:

1. An identified external anchor: a TC39 proposal (with its current stage
   recorded in the package doc) or a finished-spec abstract operation with a
   documented consumer need that the main module's surface rules exclude.
2. An owning SPEC (new or revised) that records the package's public API, its
   stability class, and its graduation or removal path, plus a row in the
   [SPEC 72](./72-operation-ledger.md) operation ledger.
3. Table-driven tests and runnable `Example*` functions like any formatter
   package.
4. An entry in §5 and in the module README.

## 4. Graduation and removal

- A proposal that reaches Stage 4 may be promoted into the main module as a
  constructor package with a root namespace alias, subject to the usual
  completeness bar (SPEC 00 §2.1). The `x/` package is then deprecated and
  removed in a later v0 release; v0 permits removal without a compatibility
  shim.
- A package exposing finished-spec machinery (such as locale matching) may
  graduate the same way if its motivating proposal finalizes, or remain in
  `x/` long-term.
- A proposal that is withdrawn or superseded gets its `x/` package removed in
  the next release.

## 5. Package inventory

| Package | Anchor | Owning SPEC | Status |
|---------|--------|-------------|--------|
| `x/localematcher` | Stage 1 [`Intl.LocaleMatcher`](https://github.com/tc39/proposal-intl-localematcher); ECMA-402 LookupMatcher / BestFitMatcher abstract operations | [SPEC 11](./11-locale-matching.md) §11 | Active |

## 6. Acceptance criteria

- [ ] `x/go.mod` declares `module github.com/agentable/go-intl/x`, requires the
      root module, and carries `replace github.com/agentable/go-intl => ..`.
- [ ] `task modules:verify` covers `x` (tidy diff, vet, test).
- [ ] The Release Inventory in CLAUDE.md lists `github.com/agentable/go-intl/x`
      with tag prefix `x/`.
- [ ] Every `x/` package doc states the v0 no-compatibility-promise policy and
      its external anchor (proposal stage or abstract operation).
- [ ] No root-module package imports `github.com/agentable/go-intl/x/...`.

---

> This SPEC is a maintenance record for the `x/` satellite module. Admitting a
> new package, changing the versioning policy, or promoting a package into the
> main module triggers a revision of this SPEC before code changes.
