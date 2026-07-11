# SPEC 50 — CLDR Data & Codegen

> **Status:** Revised (2026-06-05)
> **Priority:** High (all formatter data bottom layer; blocking SPEC 10 / 20 / 30 / 31 / 32 / 40)
> **Authority:** CLDR / ICU / tzdata release data are the upstream authorities. This SPEC records the `internal/cldr/` per-domain package structure, the on-demand blob representation, the version peg, the active locale scope, and the `tools/gen-cldr/` code generator.

---

## Overview

`internal/cldr/` is the CLDR-backed data bottom layer of go-intl. Each semantic domain is its own Go package; the CLDR payload enters the compiler only as `const` string blobs, and a hand-written decode loop expands each blob lazily on first access into the in-memory maps and slices the formatter accessors return. `tools/gen-cldr/` is the code generator for this layer, maintained as an independent Go module.

The runtime algorithm boundaries for `collator`, `segmenter`, and active best-fit locale matching are defined by their own SPECs. This SPEC defines: data source selection (direct `unicode-org/cldr-json`), embedding strategy (const blob + on-demand decode, **not** `//go:embed *.json`), active locale scope, version pinning (`cldr=48.1.0` / `icu=78` / `tzdata=2025b`), the per-domain package layout and accessor surface, the three machine gates that keep the layout honest, and the upgrade process.

> **Why**: The earlier single `internal/cldr` package compiled the CLDR tables as giant Go literals — `units.go` alone drove the `compile` subprocess to ~4.3 GB RSS and OOM-killed downstream CI (issue #3). The compile-memory cost grows superlinearly with the SSA-node count per record, not with line count, and string constants are nearly free to compile (`golang.org/x/text` is the decade-long precedent). The fix is structural and permanent: one package per domain (so Go compiles, links, and lazily loads per reachable package) and payload that enters as `const` only (so the compiler sees data, not instructions that build data).
>
> **Rejected**:
> - Single package with split files — lost on compile memory: the Go compiler is single-threaded per package, and the literal shape, not the file count, is the bomb. `units` alone still OOMs.
> - `//go:build` tags (`intl_full` / `intl_minimal`) — lost on default-correctness: the default shape must already be correct; making users opt out of a broken default is configuration cope.
> - Multi-module split — lost on cost/benefit: Go's compile and link unit is the package, not the module; splitting modules buys version-matrix, cross-module `internal/` invisibility, and release-orchestration cost for zero compile benefit.
>
> **Basis**: issue #3 reproducer measurements (Go 1.26.4, darwin/arm64): root package 5,945 → 226 MB; `unit` leaf 4,276 → 276 MB after blob-ization; all leaf domains land in the 276–317 MB band. `golang.org/x/text/internal/data` ships CLDR-scale data as const strings.

---

## CLDR Pin Rationale <a id="cldr-pin-rationale"></a>

The active data baseline is fixed at `cldr=48.1.0` / `icu=78` / `tzdata=2025b` and is exposed to tests and data contracts by `internal/cldr/locale/manifest.go` (package `cldrlocale`). The three must move together as a conformance baseline: the CLDR JSON provides the locale payload, the ICU version represents the pattern-matching and reference-behavior context, and the tzdata version determines the IANA zone/link and display-name boundaries.

Version upgrades must not be mixed with ordinary code changes as an opportunistic dependency refresh. An upgrade must update `internal/cldr/VERSION`, rerun `task data`, pass `task data:check` / `task conformance:verify` / `task verify`, and document user-observable data differences in this SPEC or release notes.

---

## 1. Data Source

### 1.1 Selection

CLDR data **MUST** directly consume the npm package set (`cldr-bcp47` / `cldr-core` / `cldr-dates-full` / `cldr-localenames-full` / `cldr-misc-full` / `cldr-numbers-full` / `cldr-segments-full` / `cldr-units-full`) of [`unicode-org/cldr-json`](https://github.com/unicode-org/cldr-json), the same source Generated reference uses.

> **Why**: Locking to the same CLDR version as Generated reference keeps conformance failures attributable to code differences, not data differences; the npm release rhythm tracks the CLDR version; and JSON connects directly to `encoding/json` (generation-time only), needing no LDML XML parser.
>
> **Rejected**:
> - `golang.org/x/text/cldr` — lost on baseline control: its data version is not the same conformance baseline as the go-intl/Generated reference pin, and its shape is Go structs, not JSON.
> - generated-reference intermediate products (`@formatjs_generated/*`) — lost on operability: TS + Bazel build artifacts, not shipped to npm; consuming them needs the Generated reference Bazel pipeline in CI.
> - ICU CGO bindings — lost on dependency policy: CGO conflicts with SPEC 00 §1.1 "no ICU C/C++ dependency".

### 1.2 Extraction scope (active scope) <a id="schema"></a>

Each semantic domain is a Go package under `internal/cldr/<domain>/` with a generated **const-only** `data.go` (one or more `const _<...>Blob` payloads plus a domain-private `const _data` string table) and hand-written `decode.go` / `accessors.go`. The locale kernel package `cldrlocale` (`internal/cldr/locale/`) owns the `Locale` handle, the locale registry, likely subtags, region/hour-cycle/week preferences, numbering data, collation candidates, IANA time-zone links, the data manifest, and `Version()`. The root `internal/cldr` directory holds only the `VERSION` text file and the domain subdirectories; it is **not** a Go package.

| `internal/cldr/<domain>` | CLDR source | Extract fields |
|---------------------------|-------------|----------------|
| `number` | `cldr-numbers-full/main/<locale>/numbers.json` | symbols (decimal/group/percent/plus/minus/NaN/Infinity/timeSeparator), decimal/percent/currency/scientific/compact formats, numbering systems |
| `date` | `cldr-dates-full/main/<locale>/ca-gregorian.json` | era / month / weekday / day-period names (stand-alone × format, wide / abbreviated / narrow), date/time/dateTime formats, availableFormats, intervalFormats, day-period rules |
| `timezone` | `cldr-core/supplemental/metaZones.json` + `cldr-dates-full/main/<locale>/timeZoneNames.json` | zone → metazone mapping, metazone display names (long / short × generic / standard / daylight), exemplarCity, metazone period boundaries |
| `currency` | `cldr-numbers-full/main/<locale>/currencies.json` + `cldr-core/supplemental/currencyData.json` | currency display names (long / short / narrow), plural forms, defaultFractionDigits, cashDigits, rounding |
| `unit` | `cldr-units-full/main/<locale>/units.json` | NumberFormat-sanctioned unit plural patterns, DurationFormat duration unit patterns (long / short / narrow), compound unit patterns, unit-supported locales |
| `list` | `cldr-misc-full/main/<locale>/listPatterns.json` | `conjunction` / `disjunction` / `unit` × `long` / `short` / `narrow` pair/start/middle/end patterns |
| `relativetime` | `cldr-dates-full/main/<locale>/dateFields.json` | long/short/narrow relative and relativeTime patterns for year/quarter/month/week/day/hour/minute/second |
| `displaynames` | `cldr-localenames-full/main/<locale>/*.json` (+ currency names imported from the `currency` domain) | language / region / script / calendar / dateTimeField display names |
| `plural` | `cldr-core/supplemental/plurals.json` + `ordinals.json` + `pluralRanges.json` | cardinal rules, ordinal rules, pluralRanges (emitted by SPEC 40 codegen; this SPEC only fixes the package location and that it passes the data-shape gate) |
| `locale` (kernel) | `cldr-core` `availableLocales.json` / `likelySubtags.json` / `timeData.json` / `weekData.json` / `calendarPreferenceData.json`, `cldr-bcp47/collation.json`, IANA `backward` | locale registry (`und` at index 0), likely subtags (maximize/minimize), hour-cycle/week/calendar preferences, numbering data, collation candidates, IANA link → canonical zone, manifest, version |

The locale-matching distance / paradigm / matchVariables / territoryContainment tables had no production consumer (the runtime matcher in `internal/localematcher` carries its own distance table), so they are no longer generated.

Runtime CLDR accessors may interpret locale subtags only to select already
generated data. When they need script or region shape checks, such as
`date.DateLocaleData` hour-cycle region lookup or display-name language-region
composition, they must delegate to `internal/localeid` instead of embedding
domain-local ASCII/length grammar.

#### Supported-value narrow indexes

`SupportedLocales`, `SupportedNumberingSystems`, `SupportedCalendars`, `SupportedCurrencies`, `SupportedTimeZones`, and `UnitSupportedLocales` are cold-start-path queries that must **not** decode the domain's main payload. Each is generated as a small dedicated `const _<...>Blob` with its own `sync.Once`, owned by the domain that proves the support:

| Narrow index | Owner | Source |
|--------------|-------|--------|
| `number.SupportedLocales()` / `number.SupportedNumberingSystems()` | `number` | generated number payload locales + ECMA-402 simple digit set |
| `date.SupportedLocales()` / `date.SupportedCalendars()` | `date` | gregorian-bearing payload locales; CLDR `"gregorian"` → ECMA-402 `"gregory"`, `+iso8601` when gregory exists |
| `currency.SupportedCurrencies()` | `currency` | `currencyData.json`-derived candidate set |
| `timezone.SupportedTimeZones()` | `timezone` | metazone + IANA link candidate set |
| `unit.SupportedLocales()` | `unit` | unit-payload locales |
| `list` / `relativetime` / `displaynames` `SupportedLocales()` | each domain | each domain's payload locales |

> **Why**: The legacy `supportedLocaleTags` walked an entire domain data map to collect keys; on the cold start of a `SupportedLocalesOf` call that triggered the full-domain decode and (before DCE) linked the whole domain blob. A narrow blob severs supported queries from main-payload decode, so root `Intl.supportedValuesOf` and constructor `SupportedLocalesOf` stay cheap.
>
> **Rejected**: Deriving supported values by scanning the decoded main map — lost on cold-start cost: it reintroduces the full-domain decode and link the narrow index exists to avoid. The `Once`-not-triggered assertion in each domain test enforces the separation.
>
> **Basis**: blob granularity follows accessor reachability; `task build:size` before/after confirms an unreferenced sub-domain blob does not enter the binary.

### 1.3 Locale profile schema

`tools/locale-profile.json` is the maintenance record of generated CLDR payload coverage. The contract is not "all constructors share one supported-locale answer" but "a constructor claims support only when the real data or algorithm suffices". For CLDR-backed surfaces, every locale in the profile must generate the corresponding payload; non-CLDR engines (`collator`, `segmenter`) define their supported set separately in their own SPECs.

MUST rules:

1. Profile JSON contains **only** the `locales` key. A new CLDR-backed surface must not introduce a new profile key; it generates payload from the same profile.
2. `tools/gen-cldr` and `tools/gen-plural-rules` **MUST** read the profile with a strict JSON decoder. Unknown keys, multiple top-level values, and empty profiles are generation errors; `task data:contract` verifies the in-repo profile is still the schema.
3. A CLDR-backed formatter `SupportedLocalesOf` **MUST** derive through the generated supported-locale accessor, not by reading the `locales` profile or `AvailableLocales()` directly.
4. Each generated supported-locale accessor must reflect the active payload and be a subset of `Manifest().LocaleProfile`, verified by `task data:contract`; each profile locale must fall to a real payload locale through ECMA-402 lookup.
5. Non-CLDR engines must not inherit the `locales` profile; they maintain a narrower, truly-supported set.
6. Changing `locales` = changing the library's conformance surface; each change goes through `task data` regeneration + `task verify`.

> **Why**: Folding the early seven surface-specific keys to one `locales` key trades a measured ~3.2 MB → 9 MB binary delta (acceptable) for removing the contributor burden of subset/default-chain rules and the user-visible surprise of "has `hi` plurals but not `hi` duration". The constructor layer still derives its supported set from real payload / engine capability, so over-claiming stays impossible.
>
> **Rejected**:
> - Multiple surface keys — lost on cognitive cost: optimizes binary size (already acceptable) at the price of contributor and user surprise.
> - `locales` + `richLocales` two-tier — lost on root cause: still encodes "semi-supported" CLDR semantics, just fewer tiers.
> - Per-locale tag map — lost on complexity: a richer JSON form for no behavioral gain.

### 1.4 Identifier source decision

| Identifier | Source | Remarks |
|------------|--------|---------|
| Currency (ISO 4217 + precision) | CLDR `currencyData.json` | Do **not** add an independent ISO 4217 table or `bojanz/currency` (separate CLDR-derived table, drifts with `internal/cldr/VERSION`) |
| Time zone (IANA zone) | `time/tzdata` transitions + CLDR `metaZones.json` display names | tzdata injected by SPEC 32; this SPEC generates the `timezone` display-name + metazone payload |
| Collation candidate identifiers | CLDR BCP47 `collation.json` | Generation filters deprecated, `ducet`, `search`, `standard`; active root support narrowed by `internal/collation` |
| Sanctioned unit identifiers | ECMA-402 hardcode in `internal/ecma402/numberformat/constants.go` | Spec list is authoritative; CLDR provides the schema, not the sanctioned list |

> **Why**: Currency precision belongs to the pinned CLDR baseline; a second CLDR-derived table would need independent verification. Sanctioned units are normative, so the spec list, not CLDR detection, is authoritative.
>
> **Rejected**:
> - Static ISO 4217 table — lost on consistency: diverges from CLDR / ICU / Generated reference.
> - `bojanz/currency` as a data source — lost on data-version control: splits the CLDR version path.
> - CLDR-detected unit list — lost on authority: CLDR's table is wider than the spec's sanctioned set.

---

## 2. Embedding Strategy

### 2.1 Const blob + on-demand decode, not `//go:embed *.json`

CLDR data **MUST** be compiled into Go `const` strings, generated by `tools/gen-cldr/`. **BANNED**:

- `//go:embed *.json` + runtime `encoding/json.Unmarshal`.
- Network pull or file I/O on startup.

The runtime contract is unchanged from the lazy-load era: no JSON, no file I/O, no network, no reflection. A blob is a const string; decode is an index loop over it, gated by `sync.Once`.

> **Why**: Const strings land in `.rodata` for O(1) access, with no cold-start JSON parse (100 locales × several files ≈ 100–500 ms, unacceptable for CLI/Lambda/short-lived services), no runtime `encoding/json` dependency, and a smaller binary (a `loadUnitData` literal compiled to hundreds of thousands of map-building instructions; a blob is data bytes plus one loop).
>
> **Rejected**:
> - `//go:embed *.json` + runtime deserialize — lost on cold start + dependency: JSON parse cost and an `encoding/json`/reflection runtime dependency.
> - Runtime network pull — lost on deployment: breaks the single-binary, offline-deployable philosophy.

### 2.2 Domain-private string tables: deduplication by ownership, not centralization

Each domain's `data.go` carries its own `const _data string` table holding **only that domain's strings**; the kernel `locale` package holds only kernel strings (locale identity, likely subtags, matching data). There is **no shared cross-domain string table**.

`StringRef` reads a `(uvarint offset, uvarint length)` pair and resolves it against the caller's own `_data`, yielding a zero-copy substring (a slice into `.rodata`).

> **Why**: The legacy global `_data` was copied verbatim three times (root + `list` + `locale`), ~38k duplicated lines. Moving the global table into the kernel would not fix it — every formatter that imports the kernel would then re-link every domain's strings, quietly defeating DCE and per-package compilation; the bomb would just change address. Dissolving the global table so each domain owns its own strings removes the duplication because ownership is correct, not because the table is centralized.
>
> **Rejected**:
> - One shared table in `locale/` — lost on DCE/compile isolation: every kernel importer relinks all domain strings.
> - Per-field `var s = "..."` — lost on binary size: duplicate strings cannot be deduplicated; slice-header and indirect-access overhead.
>
> **Basis**: `translate-agent/intl` and `golang.org/x/text` both deduplicate via a single const string + `[start:end]` indexing; the change here is that the single table is per-domain rather than per-repo.

### 2.3 Blob granularity follows accessor reachability

A large domain is split into multiple blobs, each with its own `const`, its own `sync.Once`, and its own accessor, so an unreferenced access path is dropped by the linker together with its blob. The clearest case is `timezone`: canonical links, metazone mapping, zone display names, and metazone period boundaries are independent blobs. The narrow supported-value indexes (§1.2) are the other instance — a supported query decodes only its small blob.

`task build:size` before/after is the **evidence** of this property, not its definition; DCE is a verification item, not an assumption.

> **Why**: A single giant blob per domain would force any importer to link the whole domain. Sub-domain blobs keep the link cost proportional to the access paths a consumer actually reaches.

### 2.4 Per-domain packages, not single-package files

CLDR data **MUST** be generated as one package per semantic domain (§1.2). This is the structure that lets Go compile, link, and lazily decode per reachable package — importing only `displaynames` compiles only `displaynames` + the `locale` kernel + `codec`, in parallel with other packages.

> **Why**: The Go compiler is single-threaded per package; splitting files inside one package does not parallelize compilation and does not change the per-record SSA cost. Splitting into packages does both, and aligns the compile/link/load unit with the semantic boundary.
>
> **Rejected**: Single package, split files — lost on compile time and memory: the single-thread-per-package limit and the literal SSA cost are untouched by file count.

---

## 3. Active Locale Scope

### 3.1 Curated default: 104 modern tier 1 / 2 locales

The active scope **MUST** fully embed all 104 CLDR modern tier 1 / 2 locales listed in `tools/locale-profile.json` `locales`. This is the current product portrait, not a temporary seed and not a full native-engine data commitment. `cldrlocale.Manifest()` and `task data:contract` lock the version pins and locale count. Default-profile change = conformance-surface change, evidence required. **BANNED**:

- per-locale tree-shaking (generated-reference `__addLocaleData` style).
- sub-locale lazy file I/O.
- build-tag splits (`intl_full` / `intl_minimal`); only **consumer-driven expansion** may revisit this.

The specific list is maintained in `tools/locale-profile.json`; this SPEC does not fix it.

### 3.2 Volume and cold-compile

`task build:size` reports the current profile's binary delta against an empty baseline for the root facade, each direct formatter import, the CLDR-only harness, and all formatters together. `task build:size:cold` clears the Go build cache first. Neither blocks CI, but any `tools/locale-profile.json` expansion PR **MUST** post both outputs and pass `task data:contract` + `task conformance:verify`.

Cold-compile memory — the original issue #3 failure — is now bounded structurally: the per-domain const-blob layout keeps every leaf package's `compile` RSS in the 276–317 MB band and the root facade at ~226 MB, well under the 7 GB downstream limit. Lightweight entry points (e.g. only `cldrlocale.AvailableLocales`) must not pull formatter-heavy data through shared package init; per-domain packages plus `sync.Once` decode keep large maps off the package-init path.

> **Why**: Pay-for-use — a lightweight API must not pay for the full number/date/timezone table. Go's linker trims unreferenced decoders, and the per-domain boundary keeps heavy data out of any package's init expression.
>
> **Rejected**:
> - per-locale subpackage (`numberformat/locale-data/zh.js` style) — lost on mechanism: Go has no `__addLocaleData`; inter-subpackage imports pollute the dependency graph.
> - `intl_full` / `intl_minimal` build tags in active scope — lost on need: increases generator complexity for a default that is already correct.

### 3.3 Consumer-driven expansion placeholder

Consumer-driven expansion **may** later introduce build-tag classification, but the default strategy is a curated profile, not a configuration matrix. A new build profile is allowed only when multiple hosts have a clear, repeated need the 104-locale default cannot serve, and the PR must include size, cold-compile, conformance, and supported-locale evidence. This SPEC does not schedule it.

---

## 4. Version Pinning <a id="version-pin"></a>

### 4.1 Decision

`internal/cldr/VERSION` **MUST** store three pins in one file:

```text
cldr=48.1.0
icu=78
tzdata=2025b
```

> **Why**: Generated reference's main branch locks `cldr-*: 48.1.0` (ICU 78); byte-level Generated reference alignment is a SPEC 00 §1 goal, so pinning an earlier version would institutionalize reverse divergence. CLDR 48 / ICU 78 (2025-10) is stable by public release; tzdata 2025b matches the Go 1.26.3 built-in.
>
> **Rejected**:
> - CLDR 47 / ICU 76 — lost on Generated reference alignment: one version behind, reverse divergence.
> - Follow `golang.org/x/text/cldr` — lost on baseline control: hands the conformance target to an external release rhythm.

### 4.2 Upgrade process

Any CLDR / ICU / tzdata change **requires**:

1. Update the three lines of `internal/cldr/VERSION`.
2. Update the eight `cldr-*` versions in `tools/gen-cldr/.cldr-json/package.json`, then `npm install --package-lock-only --prefix tools/gen-cldr/.cldr-json` to regenerate the lockfile.
3. Run `task data` to regenerate every domain `data.go` + kernel data. `data:preflight` catches VERSION ↔ package.json drift before `npm ci`.
4. Review the field-level diff with `git diff internal/cldr/`.
5. CI on main blocks unreviewed data increments.
6. Update conformance fixtures (`<package>/testdata/conformance/`) if necessary.

> **Why**: CLDR upgrades introduce hundreds of row-level changes; silent replacement makes a PR unreviewable. `data:preflight` + `git diff` + CI block force review and converge "drift → repair" into an executable instruction.

### 4.3 Byte-stability check

The pins in `internal/cldr/VERSION` must correspond to the data encoded in every domain `data.go` and the kernel files. CI **MUST** pass `task data:check`: regenerate and assert byte-equality with the committed files. Hand-editing a generated file is forbidden — change the generator and rerun `task data`.

> **Why**: Prevents a manual data edit without a VERSION bump (and vice versa). Byte stability is the only contract that keeps generated data auditable.

### 4.4 Generated manifest

`tools/gen-cldr` **MUST** generate `internal/cldr/locale/manifest.go` (package `cldrlocale`). The manifest is an internal audit fact, not a root public API, and must be stable and free of machine-local paths:

- `Generator`: `tools/gen-cldr`.
- `CLDR` / `ICU` / `TZData`: identical to `internal/cldr/VERSION`.
- `LocaleProfile`: the normalized `tools/locale-profile.json` locale list.
- `InputHashes`: SHA-256 of `internal/cldr/VERSION`, `tools/locale-profile.json`, and the eight `cldr-*` package metadata.

`cldrlocale.Manifest()` **MUST** return a slice clone. `task data:check` byte-equality covers `locale/manifest.go`. `cldrlocale.Version()` derives its CLDR / ICU / tzdata fields from the generated manifest — the `VERSION` text file is the codegen-time source of truth, not embedded a second time at runtime.

### 4.5 Generated-file convention <a id="cldr-data-package-convention"></a>

Every file generated by `tools/gen-cldr` begins with the Go `// Code generated by tools/gen-cldr ...` header, followed by stable metadata comments: the pinned versions, the reproducibility note (never timestamps, hostnames, usernames, or absolute paths), and the schema anchor. Hand-written files (`decode.go`, `accessors.go`, `doc.go`, and everything under `codec/`) **MUST NOT** begin with the generated header — the data-shape gate uses that header to tell generated payload from hand-written code, so a hand-written file wearing the header would be wrongly conscripted.

Each domain package follows one layout: a generated const-only `data.go`, a hand-written `decode.go` (the `sync.Once` decode loops), and a hand-written `accessors.go` (the runtime query surface). `go/format` is the final formatting step, so literal layout is deterministic across runs.

> **Why**: The generated header is the gate key in both directions — generated files must carry it, hand-written files must not. This is the contract the data-shape gate (§5.4) relies on.

---

## 5. Generator (`tools/gen-cldr/`) <a id="codegen"></a>

### 5.1 Independent Go module

`tools/gen-cldr/` **MUST** be an independent Go module (own `go.mod`) and **MUST NOT** pollute the main module's dependency graph. Because its module path shares the main module's prefix (`github.com/agentable/go-intl/tools/gen-cldr`), Go's `internal/` visibility rules let it `require` + `replace` the main module and import `internal/cldr/*` directly — so the round-trip gate (§5.4) holds the **real** production decoders, with no mirror decoder to drift.

```text
tools/gen-cldr/
├── go.mod              # standalone module; require + replace ../..
├── main.go run.go      # CLI entry + orchestration
├── extract/            # CLDR JSON → typed extract.* structs (unchanged by migration)
└── codegen/
    ├── domain.go            # the domain registry (single source of truth)
    ├── render.go            # orchestration over the registry
    ├── encoder.go           # blob encoder helpers (uvarint / delta / stringRef / zigzag)
    ├── stringtable.go       # per-domain StringTable
    ├── format.go            # go/format pass
    ├── <domain>_encode.go   # one const-only payload encoder per domain
    └── <domain>_roundtrip_test.go  # one round-trip gate per domain
```

The retired literal-rendering layer — `golang_literal.go`, `map_literal.go` (reflection → composite literal), and `wrappers.go` (the alias-facade renderer) — no longer exists. The `extract.*` extraction layer and each domain's row-shaping functions are unchanged: the data was already structured, which is what makes generator-first migration possible.

> **Why** (no mirror decoder): A separately-written decoder in the generator could silently desync from the production decoder. Sharing the prefix lets `require` + `replace` grab the production decoder, so round-trip tests the exact runtime path.
>
> **Rejected**: A generator-side mirror decoder — lost on desync risk: two decoders that can drift; the round-trip gate would test the wrong one.

### 5.2 Domain registry drives everything

`codegen/domain.go` holds the `domains` registry: one row per domain with its package directory and its const-only `emit` function. Generation, the round-trip gate, and the data-shape gate all derive their expectations from this one table — adding a domain is adding a row, not scattering functions across files. The payload path is derived (`internal/cldr/<pkg>/data.go`), and each domain gets a fresh per-domain `StringTable` so its `_data` holds only its own strings.

> **Why**: A registry makes ten domains ten rows of one table instead of ten scattered function sets, and gives every gate a single source of expected packages.

### 5.3 Generator-first encoding rules

1. **Replace the rendering layer, not the extraction layer.** Each `emit` consumes `extract.*` row streams (e.g. `unitPatternRows`) and encodes them; upstream extraction is untouched.
2. **Factory free, product pure.** The encoder may use ordinary structs, sorting, maps, and reflection; the product is only `const _data` / `const _<...>Blob`.
3. **Never reverse-parse generated `.go`.** The migration encodes from `extract.*` directly; reading old generated Go would parse the very source the migration removes.
4. **Pre-compute sort keys at generation time.** Sorted keys are written as delta streams so the decode loop carries no business logic — it is a mover.

**`mustLocaleIndex` is a generation-time guard.** Every per-locale stream resolves its locale against the kernel registry through `mustLocaleIndex`, which panics if a locale present in domain data is absent from the kernel registry. An unfiltered missing locale would silently collide onto index 0 (`und`) and overwrite its rows. Any domain locale that is intentionally unreachable at runtime **MUST** be filtered out before encoding, with a comment stating why it cannot resolve — silence is not a substitute for the guard.

> **Why**: The guard turned up a latent data bug the moment it landed (see §8). Loud failure at generation time beats a silent index-0 collision that ships wrong data.

### 5.4 The three machine gates

Migration safety rests on three independent gates, not on human review of tens of thousands of generated lines. (These are the standing CI cause-gates; the per-domain round-trip gate lives in §5.3 and behavior byte-stability is owned by SPEC 70 — together those two plus gate 3 below formed the migration's correctness locks.)

1. **Data-shape gate** (`TestGeneratedDataShape`, in `task data:contract`, `go/ast`-based). Two absolute, threshold-free invariants over every generated file, keyed by the generated header:
   - **Rule A** — a payload file (`internal/cldr/<domain>/data.go`) may contain only `const` declarations: zero func, zero var, zero non-const expression, zero import.
   - **Rule B** — any other generated file (e.g. plural rule code) may not place a `CallExpr`, `IndexExpr`, or `IndexListExpr` inside a composite literal. This kills the compile-bomb shape (`makeUnitPatternKey(localeIndex["af"], …)`) at the data-literal level, with no "under the threshold" regression path.

   The migration-era shrink-only exemption table is now **empty**: the root literal renderers it covered retired with the root package, so every generated file satisfies its invariant directly.
2. **Import-graph gate** (`TestCLDRImportGraphDirection` + `TestNoImportOfRetiredRootCLDR`, in `task data:contract`, reading production imports via `go/build.ImportDir`). Dependencies inside `internal/cldr` may only point down: `codec` imports only stdlib; the `locale` kernel imports `codec` plus stdlib-only utility leaves; each domain imports only `codec`, `locale`, the stdlib-only shared utility leaves (`localeid`, `numbering`, `pattern`, `plural`), and any sanctioned leaf→leaf edge. The single sanctioned leaf→leaf edge is `displaynames → currency` (the CLDR owner shared by NumberFormat and DisplayNames). The retired root `internal/cldr` package may never be imported again; its path is matched exactly so a domain subpackage is never mistaken for it. The leaf→leaf exception table is shrink-only: a listed edge absent from the real import graph fails the gate.
3. **Generated-byte stability** (`task data:check`): regenerate from the same input, assert byte-equality with the committed files.

Behavior byte-stability — every formatter conformance fixture output unchanged to the byte — is enforced by the formatter conformance suites (SPEC 70) and the per-domain decode snapshot tests, not by a separate gate here.

> **Why** (test the cause, not the effect): RSS numbers are observations that drift with the compiler version. The gates assert absolute shapes with no numeric threshold, so a budget cannot be approached and then exceeded — 999 elements × 10,000 records still explodes; an absolute const-only invariant does not.
>
> **Rejected**:
> - A numeric RSS threshold gate — lost on durability: a threshold is a budget, and budgets invite growth back toward the limit.
> - A schema version + decoder validation — lost on necessity: encoder, blob, and decoder always share one commit; versioning a contract that cannot desync is a statue with a seatbelt.

### 5.5 Codegen tool stack

`tools/gen-cldr/codegen/` **MUST** remain stdlib-only: deterministic string/blob construction finalized with `go/format`. **BANNED**: `dave/jennifer` or other codegen frameworks.

> **Why**: The output is const blobs plus small hand-mirrored decode loops; string construction + `go/format` is more direct than an AST/template framework, and avoids a stalled (`dave/jennifer`, no commits after 2024-09) dependency.

### 5.6 Locale list extraction

`tools/gen-cldr/extract/locales.go` merges the CLDR available locales and the active allowlist into generation records.

MUST rules:

1. The generated locale tag list **MUST** be deterministically sorted with `und` at position 0. `Locale(0)` / `Undefined` stability depends on it.
2. Sorting, key collection, and move-to-front use stdlib `maps` / `slices`; no private collection helpers or third-party collection libraries.
3. The `und`-first adjustment **MUST** happen explicitly after sorting, never relying on map iteration, allowlist order, or CLDR JSON order.

> **Why**: Any non-deterministic order produces byte-diff noise and can break the `Undefined` sentinel that every domain's locale-index stream is keyed against.

---

## 6. Data Access API

### 6.1 Per-domain accessors

Each domain package exposes its own accessors; the formatter packages consume the owning domain, not a root `cldr` package. Accessors **MUST** be O(1) after the domain's `sync.Once` decode. A returned scalar `string` is a slice into the domain `_data` and allocates nothing. Composite accessors that hand back a map or slice (for example `date.AvailableFormats` / `date.IntervalFormats` via `maps.Clone`, or `relativetime` field records) **MUST** return a defensive clone so callers cannot mutate cached domain state. The eight supported-locale index accessors (`number`/`date`/`currency`/`unit`/`list`/`relativetime`/`timezone`/`displaynames`) share one `codec.LazyStrings` owner, which decodes its `StringRefSlice` blob once and returns a `slices.Clone` on every `Get`, replacing the former per-domain `(sync.Once, []string, loader, clone-accessor)` quadruple.

Representative surface:

```go
// internal/cldr/locale (kernel)
type Locale uint16
func ResolveLocale(tag language.Tag) (Locale, bool)
func AvailableLocales() []string
func MaximizeSubtags(language, script, region string) (lang, scr, reg string, ok bool)
func MinimizeSubtags(language, script, region string) (lang, scr, reg string, ok bool)
func Version() VersionInfo
func Manifest() ManifestInfo

// internal/cldr/number
func SupportedLocales() []string
func SupportedNumberingSystems() []string
// + symbols / decimal / percent / currency / compact pattern accessors keyed by Locale

// internal/cldr/date
func SupportedLocales() []string
func SupportedCalendars() []string
// + CalendarNames / DateFormat / TimeFormat / AvailableFormats / IntervalFormats / DayPeriodFor

// internal/cldr/currency
func SupportedCurrencies() []string
// + CurrencyDigits / CurrencyDisplayName

// internal/cldr/timezone
func SupportedTimeZones() []string
// + ZoneToMetazone / MetazoneName / period-boundary accessors (per-blob)

// internal/cldr/unit
func SupportedLocales() []string
func UnitPattern(loc Locale, unit, width, plural string) string
func CompoundUnitPattern(loc Locale, width string) string

// internal/cldr/{list,relativetime,displaynames}
func SupportedLocales() []string
// + each domain's pattern / field / name accessors
```

MUST rules:

1. `cldrlocale.AvailableLocales()` is the CLDR universe; a formatter's locale matching **MUST** use the owning domain's `SupportedLocales()`, derived from generated payload, where each tag resolves to a non-`Undefined` locale through `ResolveLocale`.
2. `internal/localematcher` **MUST NOT** import any `internal/cldr` package; formatter constructors inject generated supported-locale slices and maximizers into the matcher.
3. Root `Intl.supportedValuesOf` accessors consume the owning domain's narrow index (§1.2). Root `supportedValuesOf("collation")` uses `internal/collation`, not the CLDR candidate list, so it advertises only values the active Collator can apply.
4. `SupportedCalendars()` derives from date calendar payload keys, maps CLDR `"gregorian"` → ECMA-402 `"gregory"`, and appends `"iso8601"` only when Gregorian data exists. `SupportedNumberingSystems()` includes the full ECMA-402 simple digit set even when the profile generates no matching CLDR symbol payload.
5. Number-domain decimal, percent, scientific, symbol, and currency pattern accessors must fall back from a requested numbering-system row to the locale default numbering-system row when the requested row is absent. Compact pattern accessors keep missing tuple results empty so NumberFormat can distinguish unavailable compact formats from base pattern defaults.
6. The root `internal/cldr/` directory is not a Go package. It must contain version metadata and domain subdirectories only; no production code may import a retired root CLDR package.

> **Why**: `availableLocales.json` may list locales without complete formatter payload; supported lists must come from real payload, or `ResolveLocale` could hit a payload-less locale and drift the fallback from Generated reference.

### 6.2 Visibility

The `internal/` path segment forces every domain package private. Formatter public packages do not re-export `Locale` / domain record types; they consume them through `internal/ecma402`. CLDR / ICU / tzdata pins stay in `cldrlocale.Version()` and `internal/cldr/VERSION` for internal use; the root package **MUST NOT** expose `Version()` or other diagnostic APIs, since the ECMA-402 `Intl` namespace has no such member.

---

## 7. Testing

### 7.1 Generator self-test

`tools/gen-cldr/` carries per-domain round-trip tests (`codegen/<domain>_roundtrip_test.go`). For every `extract.*` row, the test encodes through the domain's `emit` and reads back through the **production** accessor (via the `replace` to the main module), asserting per-field equality. This locks encoder, blob, decoder, and accessor together in one pass against the real runtime path — there is no mirror decoder.

### 7.2 Snapshot / contract test

`internal/cldr/locale/snapshot_test.go` (`TestCLDRGeneratedDataContract`) regenerates the CLDR files (every domain `data.go` plus the kernel `manifest.go` / `collations.go` / `timezones.go`, identified by the generator header) into a temp directory and byte-compares against the committed files; CI rejects any difference. The same package holds the data-shape and import-graph gates (`datashape_test.go`, `importgraph_test.go`), all under `task data:contract`.

### 7.3 Conformance integration

The correctness of domain data access is verified indirectly through the downstream formatter conformance suites (SPEC 70). This SPEC does not duplicate fixture design.

---

## 8. Data Semantics

### Day-period registry filtering

CLDR supplemental day-period rules cover languages beyond the kernel locale registry (kok, yue, zu, and ~15 others). The `date` encoder **encodes day-period rules only for registered locales**:

- A registry-external locale is **unreachable at runtime** — it cannot resolve to a kernel index — so its rules can never be selected; not encoding them loses nothing.
- Unfiltered, every such locale's `mustLocaleIndex` lookup would miss and (in the legacy literal renderer) collide onto index 0, overwriting `und`. The legacy renderer shipped the source-order-last ruleset (Zulu) as the effective `und` rules — a latent bug the blob migration first reproduced bug-for-bug, then fixed.
- `und` therefore takes its own genuine root am/pm rules; registry-external languages are simply not encoded.

> **Why**: This is the concrete payoff of the `mustLocaleIndex` loud-failure guard (§5.3): the guard fired on first landing and exposed the index-0 collision, and "loud failure over silent collision" is now part of the per-domain encoder template.
>
> **Rejected**: Encoding all CLDR day-period locales — lost on correctness: unreachable locales collide onto index 0 and corrupt `und`, the exact legacy bug.

---

## Forbidden

- **`//go:embed *.json` + runtime deserialize** — violates §2.1.
  - ✅ Do: generate const blobs in `tools/gen-cldr`.
  - ❌ Don't: `//go:embed numbers.json` + `json.Unmarshal(...)`.
- **Runtime network/file pull of CLDR data** — breaks single-binary, offline deployment.
  - ✅ Do: const-blob embed.
  - ❌ Don't: `http.Get(...)` / `os.Open(...)` for locale data.
- **JSON/reflection decode on startup** — breaks cold start.
  - ✅ Do: `sync.Once` index loop over a const blob.
  - ❌ Don't: `func init() { json.Unmarshal(...) }`.
- **A shared cross-domain string table** — defeats DCE and per-package compile.
  - ✅ Do: a domain-private `_data` per package; kernel holds only kernel strings.
  - ❌ Don't: move the global `_data` into `locale/` and let every importer relink it.
- **Decoding a domain's main payload to answer a supported query** — reintroduces cold-start + link cost.
  - ✅ Do: read the domain's narrow supported blob.
  - ❌ Don't: walk the decoded main map for keys.
- **A generated-header prefix on a hand-written file** — the data-shape gate would conscript it.
  - ✅ Do: keep the header on `data.go` only; `decode.go` / `accessors.go` / `codec/` are header-free.
  - ❌ Don't: start `decode.go` with `// Code generated by`.
- **Hand-editing a generated file** — breaks auditable byte stability.
  - ✅ Do: change the generator, rerun `task data`, verify with `task data:check`.
  - ❌ Don't: edit `internal/cldr/<domain>/data.go` directly.
- **A reborn root `internal/cldr` package or any import of it** — the import-graph gate fails.
  - ✅ Do: import the owning domain package.
  - ❌ Don't: `import ".../internal/cldr"`.
- **A leaf→leaf import outside the sanctioned table** — violates the dependency direction.
  - ✅ Do: depend down on `codec` / `locale` / shared utility leaves, or the one sanctioned `displaynames → currency` edge.
  - ❌ Don't: add an unlisted cross-domain import.
- **`golang.org/x/text/cldr`** in `internal/cldr/` or `tools/gen-cldr/` — baseline not go-intl-controlled.
- **`bojanz/currency` as a data source** in `numberformat/` — splits the CLDR data-version path.
- **`dave/jennifer` codegen** in `tools/gen-cldr/` — stalled framework, unneeded at this scale.
- **per-locale tree-shaking in active scope** — no Go equivalent, no main-consumer need.
- **The generator in the main module** (`internal/gen/`) — pollutes the main dependency graph.

---

## Acceptance Criteria

### Data source

- [ ] `internal/cldr/VERSION` exists with `cldr=48.1.0\nicu=78\ntzdata=2025b\n`.
- [ ] `tools/gen-cldr` verifies the local `cldr-json` `package.json` version == the `cldr=` value at startup; mismatch fails.
- [ ] `internal/cldr/` contains no `*.json` file and no `//go:embed` directive.

### Package structure

- [ ] `internal/cldr/` is one Go package per semantic domain (`number` / `date` / `timezone` / `currency` / `unit` / `list` / `relativetime` / `displaynames` / `plural`) plus the `locale` kernel and `codec`; the root directory holds only `VERSION` and the domain subdirectories and is **not** a Go package.
- [ ] Each domain has a generated const-only `data.go`, a hand-written `decode.go`, and a hand-written `accessors.go`.
- [ ] Each domain's `_data` holds only that domain's strings; there is no shared cross-domain string table.
- [ ] `cldrlocale.Version()` derives the pins from the generated manifest; the root package exposes no version diagnostic.

### Representation and gates

- [ ] `TestGeneratedDataShape` passes: Rule A (payload files const-only) and Rule B (no call/index expressions in composite literals); the exemption table is empty.
- [ ] `TestCLDRImportGraphDirection` + `TestNoImportOfRetiredRootCLDR` pass: dependencies point down only, the single sanctioned leaf→leaf edge is `displaynames → currency`, and the dead root package is never imported.
- [ ] `TestRetiredRootCLDRHasNoGoFiles` passes: `internal/cldr/` remains a non-package directory.
- [ ] `mustLocaleIndex` panics at generation time for any domain locale absent from the kernel registry; intentionally skipped locales are filtered with a stated reason.
- [ ] Per-domain round-trip tests read back through the production accessor via `require` + `replace`; no mirror decoder exists.

### Supported values

- [ ] Each domain's `SupportedLocales()` is non-empty, excludes `und`, and each tag has generated payload reachable through `ResolveLocale`.
- [ ] A supported-value query reads only its narrow blob and does not trigger the domain main-payload `sync.Once` (asserted in the domain test).
- [ ] `SupportedCalendars()` maps `"gregorian"` → `"gregory"` and contains `iso8601` when Gregorian data exists; `SupportedNumberingSystems()` contains the full ECMA-402 simple digit set.
- [ ] Root `supportedValuesOf("collation")` reflects `internal/collation`, not the CLDR candidate list.

### Generator and upgrade

- [ ] `tools/gen-cldr/go.mod` is a standalone module with `require` + `replace ../..`; `codegen/` imports only stdlib.
- [ ] The domain registry (`codegen/domain.go`) is the single source of expected domain packages; the retired literal renderers (`golang_literal.go` / `map_literal.go` / `wrappers.go`) do not exist.
- [ ] `extract/locales.go` outputs a deterministically sorted tag list with `und` at position 0.
- [ ] `task data` regenerates with no git diff (except after a deliberate VERSION change); `task data:check` fails on any byte difference.
- [ ] `tools/gen-cldr` is verified from its module root (`cd tools/gen-cldr && go test ./... && go vet ./...`), not via a root-module `./tools/gen-cldr/...` pattern.

### Volume

- [ ] `task build:size` reports root-facade, direct-formatter, CLDR-only, and all-formatter deltas.
- [ ] Per-domain `compile` RSS stays in the measured band (leaf 276–317 MB, root facade ~226 MB) under the 7 GB downstream limit (recorded in PR evidence, not a CI gate).

### Day-period semantics

- [ ] The `date` encoder encodes day-period rules only for registered locales; `und` carries its own root am/pm rules, and registry-external languages are not encoded.

---

## References

### Specification

- [Unicode CLDR](https://cldr.unicode.org/) — data source.
- [CLDR JSON Distribution](https://github.com/unicode-org/cldr-json) — `cldr-bcp47` / `cldr-core` / `cldr-dates-full` / `cldr-numbers-full` / `cldr-units-full` and others.
- [ECMA-402 §6 — Locale and Currency Identifiers](https://tc39.es/ecma402/#locale-and-currency-identifiers) — identifier spec.

### Reference implementations

- `.references/formatjs/package.json` — `cldr-*: 48.1.0` lock evidence.
- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` — CLDR JSON import list and extraction logic.
- `.references/formatjs/packages/intl-numberformat/scripts/extract-currencies.ts` — currency extraction.
- `.references/intl/internal/cldr/data.go` — const-string Go literal precedent (translate-agent/intl).

### Cross-SPEC

- [SPEC 00 §5.3 — Data strategy](./00-vision-and-scope.md#53-data-strategy)
- [SPEC 10 §Maximize / Minimize](./10-locale.md) — consumes `MaximizeSubtags` / `MinimizeSubtags`.
- [SPEC 11 §BestFitMatcher](./11-locale-matching.md) — matcher receives supported locales and maximizers from formatter constructors.
- [SPEC 20 §Currency Data](./20-numberformat.md) — consumes `currency` domain accessors.
- [SPEC 30 §DateTimeFormat Core](./30-datetimeformat.md) — consumes `date` domain accessors.
- [SPEC 31 §Skeleton Resolution](./31-datetimeformat-skeleton.md) — consumes `date` availableFormats / intervalFormats.
- [SPEC 32 §TimeZone & Calendar Data](./32-datetimeformat-tz.md) — consumes the `timezone` domain, integrated with `time/tzdata`.
- [SPEC 40 §PluralRules](./40-pluralrules.md) — the `plural` package is output here by the SPEC 40 codegen.

---

> This SPEC is the maintenance record of the CLDR data layer. Version-pin changes trigger a SPEC revision; the active locale list is maintained in `tools/locale-profile.json` (no SPEC revision); volume changes are reviewed through `task build:size`.
