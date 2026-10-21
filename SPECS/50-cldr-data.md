# SPEC 50 — CLDR Data & Codegen

> **Status:** Revised (2026-08-09)
> **Priority:** High (all formatter data bottom layer; blocking SPEC 10 / 20 / 30 / 31 / 32 / 40)
> **Authority:** CLDR / ICU / tzdata release data are the upstream authorities. This SPEC records the `internal/cldr/` per-domain package structure, the on-demand blob representation, the version peg, the active locale scope, and the `tools/gen-cldr/` code generator.

---

## Overview

`internal/cldr/` is the CLDR-backed data bottom layer of go-intl. Each semantic domain is its own Go package; the CLDR payload enters the compiler only as `const` string blobs, and a hand-written decode loop expands each blob lazily on first access into the in-memory maps and slices the formatter accessors return. `tools/gen-cldr/` is the code generator for this layer, maintained as an independent Go module.

The active best-fit locale-matching algorithm is defined by SPEC 11. This SPEC defines: data source selection (direct `unicode-org/cldr-json`), embedding strategy (const blob + on-demand decode, **not** `//go:embed *.json`), active locale scope, version pinning (`cldr=48.1.0` / `icu=78` / `tzdata=2025b`), the per-domain package layout and accessor surface, the generated-data machine gates, and the upgrade process.

> **Why**: The earlier single `internal/cldr` package compiled the CLDR tables as giant Go literals — `units.go` alone drove the `compile` subprocess to ~4.3 GB RSS and OOM-killed downstream CI (issue #3). The compile-memory cost grows superlinearly with the SSA-node count per record, not with line count, and string constants are nearly free to compile (`golang.org/x/text` is the decade-long precedent). The fix is structural and permanent: one package per domain (so Go compiles, links, and lazily loads per reachable package) and payload that enters as `const` only (so the compiler sees data, not instructions that build data).
>
> **Rejected**:
> - Single package with split files — lost on compile memory: the Go compiler is single-threaded per package, and the literal shape, not the file count, is the bomb. `units` alone still OOMs.
> - `//go:build` tags (`intl_full` / `intl_minimal`) — lost on default-correctness: the default shape must already be correct; making users opt out of a broken default is configuration cope.
> - Multi-module split — lost on cost/benefit: Go's compile and link unit is the package, not the module; splitting modules buys version-matrix, cross-module `internal/` invisibility, and release-orchestration cost for zero compile benefit.
>
> **Basis**: issue #3 reproducer measurements (darwin/arm64): root package 5,945 → 226 MB; `unit` leaf 4,276 → 276 MB after blob-ization; all leaf domains land in the 276–317 MB band. These measurements remain structural evidence rather than a toolchain-patch benchmark. `golang.org/x/text/internal/data` ships CLDR-scale data as const strings.

---

## CLDR Pin Rationale <a id="cldr-pin-rationale"></a>

The active data baseline is fixed at `cldr=48.1.0` / `icu=78` / `tzdata=2025b` and is exposed to tests and data contracts by `internal/cldr/locale/manifest.go` (package `cldrlocale`). The manifest records all three in one auditable baseline, but their owners remain distinct: CLDR JSON provides locale/display payloads, ICU identifies the reference-behavior context, and the pinned IANA release determines Zone/Link identity and region records. Go's embedded transition data may be newer than the identity pin but must never be older.

Version upgrades must not be mixed with ordinary code changes as an opportunistic dependency refresh. An upgrade must update `internal/cldr/VERSION`, rerun `task data`, pass `task data:check` / `task conformance:verify` / `task verify`, and document user-observable data differences in this SPEC or release notes.

---

## 1. Data Source

### 1.1 Selection

CLDR locale data directly consumes the pinned npm package set (`cldr-bcp47` / `cldr-core` / `cldr-dates-full` / `cldr-localenames-full` / `cldr-misc-full` / `cldr-numbers-full` / `cldr-units-full`) of [`unicode-org/cldr-json`](https://github.com/unicode-org/cldr-json). Number symbols additionally consume the same release's original `common/main/root.xml`, because npm JSON omits its root rows and aliases.

> **Why**: The pinned npm release supplies resolved locale payloads through generation-time `encoding/json/v2`. Root symbols need a separate same-release input: replacing digits alone does not produce the correct arab/arabext separators and signs. The generator parses only `numbers/symbols` with stdlib XML; runtime packages retain const blobs and no XML/JSON I/O. Reference algorithms and reference output versions are evaluated separately.
>
> **Rejected**:
> - `golang.org/x/text/cldr` — lost on baseline control: its data version is not the same conformance baseline as the go-intl/Generated reference pin, and its shape is Go structs, not JSON.
> - generated-reference intermediate products (`@formatjs_generated/*`) — lost on operability: TS + Bazel build artifacts, not shipped to npm; consuming them needs the Generated reference Bazel pipeline in CI.
> - ICU CGO bindings — lost on dependency policy: CGO conflicts with SPEC 00 §1.1 "no ICU C/C++ dependency".

### 1.2 Extraction scope (active scope) <a id="schema"></a>

Each semantic domain is a Go package under `internal/cldr/<domain>/` with a generated **const-only** `data.go` (one or more `const _<...>Blob` payloads plus a domain-private `const _data` string table) and hand-written `decode.go` / `accessors.go`. The locale kernel package `cldrlocale` (`internal/cldr/locale/`) owns the `Locale` handle, the locale registry, likely subtags, script directions, region/hour-cycle/week preferences, numbering data, the data manifest, and `Version()`. IANA identifier records live in `internal/tz`, outside the CLDR kernel and localized time-zone display domain. The root `internal/cldr` directory holds only the `VERSION` text file and the domain subdirectories; it is **not** a Go package.

| `internal/cldr/<domain>` | CLDR source | Extract fields |
|---------------------------|-------------|----------------|
| `number` | `cldr-numbers-full/main/<locale>/numbers.json` + same-release `common/main/root.xml` symbols | minimumGroupingDigits, symbols (decimal/group/percent/plus/minus/NaN/Infinity/timeSeparator and optional currencyDecimal/currencyGroup), decimal/percent/currency/scientific/compact formats, currency-spacing insertion text, plural-sensitive currency-name placement patterns, numbering systems |
| `date` | `cldr-dates-full/main/<locale>/ca-gregorian.json` | era / month / weekday / day-period names (stand-alone × format, wide / abbreviated / narrow), date/time/dateTime formats, availableFormats, intervalFormats, day-period rules |
| `timezone` | `cldr-core/supplemental/metaZones.json` + `cldr-dates-full/main/<locale>/timeZoneNames.json` | zone → metazone mapping, metazone display names (long / short × generic / standard / daylight), exemplarCity, GMT/hour/region formatting patterns, metazone period boundaries |
| `currency` | `cldr-numbers-full/main/<locale>/currencies.json` + `cldr-core/supplemental/currencyData.json` | currency display names (long / short / narrow), plural forms, default fraction digits; all selected currency name/symbol rows |
| `unit` | `cldr-units-full/main/<locale>/units.json` | NumberFormat-sanctioned simple and direct compound plural patterns, optional denominator-specific `perUnitPattern`, generic compound unit patterns, DurationFormat duration unit patterns (long / short / narrow), unit-supported locales |
| `list` | `cldr-misc-full/main/<locale>/listPatterns.json` | `conjunction` / `disjunction` / `unit` × `long` / `short` / `narrow` pair/start/middle/end patterns |
| `relativetime` | `cldr-dates-full/main/<locale>/dateFields.json` | long/short/narrow relative and relativeTime patterns for year/quarter/month/week/day/hour/minute/second |
| `displaynames` | `cldr-localenames-full/main/<locale>/*.json` (+ currency names imported from the `currency` domain) | language / region / script / calendar / dateTimeField display names |
| `plural` | `cldr-core/supplemental/plurals.json` + `ordinals.json` + `pluralRanges.json` | cardinal rules, ordinal rules, pluralRanges (emitted by SPEC 40 codegen; this SPEC only fixes the package location and that it passes the data-shape gate) |
| `locale` (kernel) | `cldr-core` `availableLocales.json` / `likelySubtags.json` / `scriptMetadata.json` / `timeData.json` / `weekData.json` / `calendarPreferenceData.json` | locale registry (`und` at index 0), likely-subtag maximize data, known script directions, hour-cycle/week/calendar preferences, numbering data, manifest, version |

Currency fractions provide only default precision: each wire row contains a
currency code reference and one default-digits uvarint. A missing precision
entry uses two digits. SupportedCurrencies enumerates the selected profile's
currency name keys independently of these exceptions; name and symbol rows
are retained in full. Evidence: `tools/gen-cldr/codegen/currency_encode.go`,
`tools/gen-cldr/codegen/currency_roundtrip_test.go`, and `internal/cldr/currency/currency_test.go`.

Three private identity products are emitted directly to their runtime owners:

Decimal and currency compact rows retain separate raw LDML pattern families
in the number-domain data. Currency rows come from
`currencyFormats-numberSystem-*/{short|long}/standard`; they are not derived
from decimal compact rows. The loader
validates quote balance with `internal/pattern.ParseCompact`,
the same parser used by NumberFormat and PluralRules in their constructors.
Quoted digits and semicolons are literals, doubled apostrophes are escaped,
and a plural-specific compact pattern may omit its number field. A malformed
quote is a generation error carrying the locale/file, display and row key;
validation does not rewrite the stored source pattern.
Decimal count keys accept plural categories or numeric counts; only currency
input recognizes the `-alt-alphaNextToNumber` suffix and validates its base
count. Unknown count syntax is a generation error. Production compact
accessors apply category-to-`other` fallback within the selected tuple and
leave a missing tuple empty. NumberFormat owns complete-family fallback and
construction-time currency-placeholder compilation, as defined by SPEC 20;
neither the decoder nor the accessor substitutes decimal data for currency.
Evidence: `tools/gen-cldr/cldr/numbers_test.go`,
`tools/gen-cldr/codegen/number_roundtrip_test.go`, and
`numberformat/testdata/conformance/manual/compact-patterns.json`.


Number symbols preserve optional `currencyDecimal` / `currencyGroup` source
text separately from ordinary separators; omitted and explicitly empty values
remain empty in the wire. NumberFormat chooses nonempty monetary overrides
only for currency style at construction (SPEC 20 §4.2). Pinned source witnesses
are `tools/gen-cldr/.cldr-json/node_modules/cldr-numbers-full/main/de-AT/numbers.json`
(group NBSP, currencyGroup `.`) and `tools/gen-cldr/.cldr-json/node_modules/cldr-numbers-full/main/fr-CH/numbers.json`
(decimal `,`, currencyDecimal `.`). The latter is source evidence, not an
expansion of the default profile. Parser and production round-trip tests own
this projection in `tools/gen-cldr/cldr/numbers_test.go` and
`tools/gen-cldr/codegen/number_roundtrip_test.go`.

`miscPatterns-numberSystem-*.range`, when supplied, must be
`{0}<nonempty separator>{1}` with each argument exactly once and no other
braces. `parseRangeSign` preserves all middle text, including whitespace;
pinned pt-PT supplies `{0} - {1}`. An omitted range field or misc-pattern row
uses the loader's en-dash default. Invalid supplied patterns fail generation.
`tools/gen-cldr/cldr/numbers_test.go`, the production number round trip, and
SPEC 20 §5 verify extraction and shared-literal consumption independently.

Currency spacing is projected from each numbering-system currency row into
two insertion strings: `BeforeCurrency` and `AfterCurrency`. Each supplied side
must have `currencyMatch=[[:^S:]&[:^Z:]]`, `surroundingMatch=[:digit:]`, and
a string `insertBetween`; unsupported conditions or invalid insertion values
fail generation with the source locale, numbering system and field.
The validated classes mean a currency edge outside Unicode Symbol and
Separator categories, adjacent to a decimal digit. The wire carries only the
insertion text, not a runtime UnicodeSet expression or configurable rule object.
A missing side or empty insertion text disables that side. The production
accessor falls back to the same locale's default numbering-system row only
when the requested spacing row is absent.
Source-value, omission and invalid-rule tests in
`tools/gen-cldr/cldr/numbers_test.go` and the production query/round-trip tests
in `tools/gen-cldr/codegen/number_roundtrip_test.go` verify this projection.

| Runtime owner | CLDR source | Generated contract |
|---------------|-------------|--------------------|
| `internal/localeid/unicode_alias_data.go` | `cldr-bcp47/bcp47/*.json` `keyword.u` records | key-aware Unicode type aliases; malformed aliases never become valid through lookup |
| `internal/localematcher/profile_data.go` | `languageMatching.json` + `territoryContainment.json` | all ordered `written-new` rules, `oneway`, source and expanded match variables, and paradigm locales |
| `internal/tz/registry_data.go` | SHA-256-pinned official IANA archive + CLDR BCP47 `timezone.json` | complete Zone/Link records, ECMA/ICU primary selection, and `zone.tab` region projection |

These files do not create a public or root CLDR package. Their hand-written
owners provide syntax/matching behavior, while the generator owns pinned facts
and validation.

Runtime CLDR accessors may interpret locale subtags only to select already
generated data. When they need script or region shape checks, such as
`date.DateLocaleData` hour-cycle region lookup or display-name language-region
composition, they must delegate to `internal/localeid` instead of embedding
domain-local ASCII/length grammar.

#### Supported-value narrow indexes

`SupportedLocales`, `SupportedNumberingSystems`, `SupportedCalendars`, `SupportedCurrencies`, and `UnitSupportedLocales` are cold-start-path queries that must **not** decode the domain's main payload. Each is generated as a small dedicated `const _<...>Blob` with its own `sync.Once`, owned by the domain that proves the support:

| Narrow index | Owner | Source |
|--------------|-------|--------|
| `number.SupportedLocales()` / `number.SupportedNumberingSystems()` | `number` | generated number payload locales + ECMA-402 simple digit set |
| `date.SupportedLocales()` / `date.SupportedCalendars()` | `date` | gregorian-bearing payload locales; CLDR `"gregorian"` → ECMA-402 `"gregory"`, `+iso8601` when gregory exists |
| `currency.SupportedCurrencies()` | `currency` | union of currency name keys in the selected profile |
| `unit.SupportedLocales()` | `unit` | unit-payload locales |
| `list` / `relativetime` / `displaynames` `SupportedLocales()` | each domain | each domain's payload locales |

`internal/tz.SupportedTimeZones()` is different by design: it is the primary projection of the generated IANA/CLDR identifier records, not a CLDR payload index. It must not decode localized metazone/display-name data.

> **Why**: The legacy `supportedLocaleTags` walked an entire domain data map to collect keys; on the cold start of a `SupportedLocalesOf` call that triggered the full-domain decode and (before DCE) linked the whole domain blob. A narrow blob severs supported queries from main-payload decode, so root `Intl.supportedValuesOf` and constructor `SupportedLocalesOf` stay cheap.
>
> **Rejected**: Deriving supported values by scanning the decoded main map — lost on cold-start cost: it reintroduces the full-domain decode and link the narrow index exists to avoid. The `Once`-not-triggered assertion in each domain test enforces the separation.
>
> **Basis**: blob granularity follows accessor reachability; `task build:size` before/after confirms an unreferenced sub-domain blob does not enter the binary.

### 1.3 Locale profile schema

`tools/locale-profile.json` is the maintenance record of generated CLDR payload coverage. The contract is not "all constructors share one supported-locale answer" but "a constructor claims support only when its real payload suffices". Every locale in the profile must generate the corresponding active payload.

MUST rules:

1. Profile JSON contains **only** the `locales` key. A new CLDR-backed surface must not introduce a new profile key; it generates payload from the same profile.
2. `tools/gen-cldr` and `tools/gen-plural-rules` **MUST** read the profile with a strict JSON decoder. Unknown keys, multiple top-level values, and empty profiles are generation errors; `task data:contract` verifies the in-repo profile is still the schema.
3. A CLDR-backed formatter `SupportedLocalesOf` **MUST** derive through the generated supported-locale accessor, not by reading the `locales` profile or `AvailableLocales()` directly.
4. `Manifest().LocaleProfile` records requested tags; each domain's generated supported-locale accessor records its actual payload keys. Payload keys may be lookup parents of the requests: an `en-US` profile can carry `en` number/list data while its manifest remains `["en-US"]`. Supported keys must be sorted, unique, equal to a request or its subtag prefix, and resolvable to a kernel handle; unrelated payloads are rejected. Plural rule keys such as `de-DE` can resolve to the kernel's `de` handle without a duplicate registry row. Every requested tag must reach a real domain payload through the shared lookup matcher. `task data:contract` checks this coverage, while generator round trips compare the exact domain payload keys.
5. Changing `locales` = changing the library's conformance surface; each change goes through `task data` regeneration + `task verify`.

> **Why**: Folding the early seven surface-specific keys to one `locales` key trades a measured ~3.2 MB → 9 MB binary delta (acceptable) for removing the contributor burden of subset/default-chain rules and the user-visible surprise of "has `hi` plurals but not `hi` duration". The constructor layer still derives its supported set from real payload / engine capability, so over-claiming stays impossible.
>
> **Rejected**:
> - Multiple surface keys — lost on cognitive cost: optimizes binary size (already acceptable) at the price of contributor and user surprise.
> - `locales` + `richLocales` two-tier — lost on root cause: still encodes "semi-supported" CLDR semantics, just fewer tiers.
> - Per-locale tag map — lost on complexity: a richer JSON form for no behavioral gain.

### 1.4 Identifier source decision

| Identifier | Source | Remarks |
|------------|--------|---------|
| Currency identifiers and precision | selected-profile `currencies.json` name keys for membership; `currencyData.json` fractions for default digits | Precision exceptions do not define membership. Unknown precision uses two digits; retain all selected name/symbol rows. No independent ISO 4217 table. |
| Time zone (IANA zone) | pinned official IANA Zone/Link archive + CLDR BCP47 `timezone.json` primary metadata + Go `time.LoadLocation` transitions (embedded fallback) + CLDR display data | `internal/tz` owns legal identifiers/primary/regions and delegates transition lookup to Go; `internal/cldr/timezone` owns localized names/metazones only |
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

- `//go:embed *.json` + runtime `encoding/json/v2.Unmarshal`.
- Network pull or file I/O on startup.

The runtime contract is unchanged from the lazy-load era: no JSON, no file I/O, no network, no reflection. A blob is a const string; decode is an index loop over it, gated by `sync.Once`.

> **Why**: Const strings land in `.rodata` for O(1) access, with no cold-start JSON parse (100 locales × several files ≈ 100–500 ms, unacceptable for CLI/Lambda/short-lived services), no runtime `encoding/json/v2` dependency, and a smaller binary (a `loadUnitData` literal compiled to hundreds of thousands of map-building instructions; a blob is data bytes plus one loop).
>
> **Rejected**:
> - `//go:embed *.json` + runtime deserialize — lost on cold start + dependency: JSON parse cost and an `encoding/json/v2`/reflection runtime dependency.
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

A large domain is split into multiple blobs, each with its own `const`, its own `sync.Once`, and its own accessor, so an unreferenced access path is dropped by the linker together with its blob. The clearest CLDR case is `timezone`: metazone mapping, zone display names, formatting patterns, and metazone period boundaries are independently reachable. Identifier records are a separate typed generated product in `internal/tz`. The narrow supported-value indexes (§1.2) are the other instance — a supported query decodes only its small blob.

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

---

DisplayNames source loading validates supplied localePattern and localeSeparator with the indexed pattern parser: both {0} and {1} must occur, other placeholders and malformed openings fail with source path, locale and field. Repeated arguments remain legal. Missing separator keeps the existing parent/default behavior; ICU LocaleDisplayNamesImpl::initialize supplies {0}, {1} and {0} ({1}) defaults. This check does not change valid generated bytes.

Relative-time source loading validates each supplied future/past plural pattern
with `internal/pattern.Partition`. Only `{0}` is an argument; it may be absent or
repeated. Unknown arguments (including `{literal}`) and unmatched openings fail
with source file, unit/style field, tense and plural-key context. Relative literal
values remain uninterpreted text. `tools/gen-cldr/cldr/relative_time_test.go`
verifies rejection and unchanged valid patterns; `task data:check` proves that
the pinned payload remains byte-identical. Runtime formatting does not repeat
this source validation.

## 4. Version Pinning <a id="version-pin"></a>

### 4.1 Decision

`internal/cldr/VERSION` **MUST** store three pins in one file:

```text
cldr=48.1.0
icu=78
tzdata=2025b
```

> **Why**: These pins define the library's reproducible data baseline. FormatJS and native hosts may use different releases; their algorithms and versioned observations do not replace the pinned inputs. Number root symbols must use the same CLDR release as the npm data; tzdata 2025b owns the identity registry, while Go owns transition lookup as described in [SPEC 32](32-datetimeformat-tz.md).
>
> **Rejected**:
> - Mixing root symbols from another CLDR release — loses coherent source identity and reproducibility.
> - Follow `golang.org/x/text/cldr` — lost on baseline control: hands the conformance target to an external release rhythm.

Pin syntax is owned by `tools/internal/datapin/version.go` and `tools/internal/datapin/tzdata.go`,
with valid/invalid inputs in `tools/internal/datapin/pins_test.go`. `tools/data-preflight` owns package
pin agreement and the bundled Go transition floor; the generator's
`tools/gen-cldr/cldr/version.go` and `tools/gen-cldr/tzdb/tzdb.go` validate generation inputs and archive
identity/SHA-256. Preflight does not identify the host's active transition data.

`tools/fetch-tzdata/main.go` owns verified cache reuse, download, SHA-256
checking, and atomic cache replacement; `tools/fetch-tzdata/main_test.go` verifies those paths.
`task data:fetch:tzdb` uses that same command. An archive can be maintained directly:

```sh
go run ./tools/fetch-tzdata -lock tools/gen-cldr/tzdata.json -cache tools/gen-cldr/.tzdata
```

### 4.2 Upgrade process

Any CLDR / ICU / tzdata change **requires**:

1. Update the three lines of `internal/cldr/VERSION`.
2. Update the required `cldr-*` package versions (the input set is owned by `tools/gen-cldr/cldr/fetch.go:requiredPackages`) in `tools/gen-cldr/.cldr-json/package.json`, then `npm install --package-lock-only --prefix tools/gen-cldr/.cldr-json` to regenerate the lockfile. Update the same-release `cldr/number-root/root.xml` and its `source.json` provenance/hash alongside the npm pin; an IANA upgrade also updates `tools/gen-cldr/tzdata.json`.
3. Run `task data` to regenerate every domain `data.go` + kernel data. `tools/data-preflight` structurally validates VERSION, all `cldr-*` package pins, the tzdata lock/hash, and the Go transition-data floor before `npm ci`.
4. Review the field-level diff with `git diff internal/cldr/`.
5. The `Generated data` CI job runs the same `task data:check` gate on main pushes and pull requests, then requires a clean worktree.
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
- `InputHashes`: SHA-256 of `internal/cldr/VERSION`, `tools/locale-profile.json`, the required `cldr-*` package metadata, the verified IANA pin/archive, and `tools/gen-cldr/cldr/number-root/{source.json,root.xml}`. Embedded input hashes name repository paths, never machine paths.

`cldrlocale.Manifest()` **MUST** return a slice clone. `task data:check` byte-equality covers `locale/manifest.go`. `cldrlocale.Version()` derives its CLDR / ICU / tzdata fields from the generated manifest — the `VERSION` text file is the codegen-time source of truth, not embedded a second time at runtime.

### 4.5 Generated-file convention <a id="cldr-data-package-convention"></a>

Every file generated by `tools/gen-cldr` begins with the Go `// Code generated by tools/gen-cldr ...` header, followed by stable metadata comments: the pinned versions, the reproducibility note (never timestamps, hostnames, usernames, or absolute paths), and the schema anchor. Hand-written files (`decode.go`, `accessors.go`, `doc.go`, and everything under `codec/`) **MUST NOT** begin with the generated header — the data-shape gate uses that header to tell generated payload from hand-written code, so a hand-written file wearing the header would be wrongly conscripted.

Each domain package follows one layout: a generated const-only `data.go`, a hand-written `decode.go` (the `sync.Once` decode loops), and a hand-written `accessors.go` (the runtime query surface). The three auxiliary identity products above are typed generated tables in their owning utility packages rather than CLDR domains. `go/format` is the final formatting step for every output, so literal layout is deterministic across runs.

> **Why**: The generated header is the gate key in both directions — generated files must carry it, hand-written files must not. This is the contract the data-shape gate (§5.4) relies on.

---

## 5. Generator (`tools/gen-cldr/`) <a id="codegen"></a>

### 5.1 Independent Go module

`tools/gen-cldr/` **MUST** be an independent Go module (own `go.mod`) and **MUST NOT** pollute the main module's dependency graph. Because its module path shares the main module's prefix (`github.com/agentable/go-intl/tools/gen-cldr`), Go's `internal/` visibility rules let it import `internal/cldr/*` directly from the required root module — so the round-trip gate (§5.4) holds the **real** production decoders, with no mirror decoder to drift.

```text
tools/gen-cldr/
├── go.mod              # standalone module; pinned root requirement
├── main.go run.go      # CLI entry + orchestration
├── extract/            # CLDR JSON → typed extract.* structs (unchanged by migration)
└── codegen/
    ├── domain.go            # the domain registry (single source of truth)
    ├── render.go            # orchestration over the registry
    ├── encoder.go           # blob encoder helpers (uvarint / delta / stringRef / zigzag)
    ├── stringtable.go       # per-domain StringTable
    ├── format.go            # go/format pass
    ├── unicode_alias.go     # auxiliary localeid alias table
    ├── language_matching.go # auxiliary localematcher profile
    ├── timezone_registry.go # auxiliary internal/tz identifier registry
    ├── <domain>_encode.go   # one const-only payload encoder per domain
    └── <domain>_roundtrip_test.go  # one round-trip gate per domain
```

The retired literal-rendering layer — `golang_literal.go`, `map_literal.go` (reflection → composite literal), and `wrappers.go` (the alias-facade renderer) — no longer exists. The `extract.*` extraction layer and each domain's row-shaping functions are unchanged: the data was already structured, which is what makes generator-first migration possible.

> **Why** (no mirror decoder): A separately-written decoder in the generator could silently desync from the production decoder. The shared module-path prefix permits access to the production decoder. The repository workspace selects the current checkout for local tests; `GOWORK=off` selects the pinned release for consumer checks.
>
> **Rejected**: A generator-side mirror decoder — lost on desync risk: two decoders that can drift; the round-trip gate would test the wrong one.

### 5.2 Domain registry drives everything

`codegen/domain.go` holds the `domains` registry: one row per CLDR payload domain with its package directory and its const-only `emit` function. Generation, the round-trip gate, and the data-shape gate all derive domain expectations from this one table. The payload path is derived (`internal/cldr/<pkg>/data.go`), and each domain gets a fresh per-domain `StringTable` so its `_data` holds only its own strings. Auxiliary identity products have explicit CLI output paths because their runtime owners sit outside `internal/cldr`; `Taskfile.yml` supplies their output paths and invokes the check. `tools/check-generated-data/compare.go` discovers files from recognized generator headers in both trees, requires the same complete relative-path set, and compares corresponding bytes. `tools/check-generated-data/compare_test.go` covers missing, extra, and changed files; no hand-maintained diff list owns completeness.

Time-zone display generation always validates the IANA pin/archive and CLDR
BCP47 aliases, even when the auxiliary `-timezone-out` path is omitted. The CLI
defaults `-tzdata-lock` to `tzdata.json`; direct `Run` callers supply that required
input. Missing or invalid join inputs fail before rendering. Omitting registry
output must leave the runtime timezone payload byte-identical.

> **Why**: A registry makes all domains rows of one table instead of scattered function sets, and gives every gate a single source of expected packages.

### 5.3 Generator-first encoding rules

1. **Replace the rendering layer, not the extraction layer.** Each `emit` consumes `extract.*` row streams (e.g. `unitPatternRows`) and encodes them; upstream extraction is untouched.
2. **Factory free, product pure.** The encoder may use ordinary structs, sorting, maps, and reflection; the product is only `const _data` / `const _<...>Blob`.
3. **Never reverse-parse generated `.go`.** The migration encodes from `extract.*` directly; reading old generated Go would parse the very source the migration removes.
4. **Pre-compute sort keys at generation time.** Sorted keys are written as delta streams so the decode loop carries no business logic — it is a mover.

**Locale-index validation is a generation-time error boundary.** Per-locale encoders resolve their locale against the kernel registry through `tools/gen-cldr/codegen/encoder.go:localeIndexValue`. A missing locale returns an error naming that locale, and encoding stops before writing the record; `encoder_test.go` verifies that the record callback is not invoked. An unfiltered missing locale would silently collide onto index 0 (`und`) and overwrite its rows. Any domain locale that is intentionally unreachable at runtime **MUST** be filtered out before encoding, with a comment stating why it cannot resolve — silence is not a substitute for the guard.

> **Why**: The guard turned up a latent data bug the moment it landed (see §8). Loud failure at generation time beats a silent index-0 collision that ships wrong data.

### 5.4 The two machine gates

Migration safety rests on two independent generated-data gates, not on human review of tens of thousands of generated lines. (These are the standing CI cause-gates; the per-domain round-trip gate lives in §5.3 and behavior byte-stability is owned by SPEC 70 — together those two plus gate 2 below formed the migration's correctness locks.)

1. **Data-shape gate** (`TestGeneratedDataShape`, in `task data:contract`, `go/ast`-based). Two absolute, threshold-free invariants over every generated file, keyed by the generated header:
   - **Rule A** — a payload file (`internal/cldr/<domain>/data.go`) may contain only `const` declarations: zero func, zero var, zero non-const expression, zero import.
   - **Rule B** — any other generated file (e.g. plural rule code) may not place a `CallExpr`, `IndexExpr`, or `IndexListExpr` inside a composite literal. This kills the compile-bomb shape (`makeUnitPatternKey(localeIndex["af"], …)`) at the data-literal level, with no "under the threshold" regression path.

   The migration-era shrink-only exemption table is now **empty**: the root literal renderers it covered retired with the root package, so every generated file satisfies its invariant directly.
2. **Generated path, byte stability, and generator proof** (`task data:check`): regenerate every CLDR, plural, locale-alias, matcher-profile, and time-zone registry output under one repository-shaped temporary `internal/` root. `tools/check-generated-data` uses the two generated-header prefixes as ownership metadata, compares relative paths in both directions, then requires byte equality for every matched path. After comparison, the gate runs `go -C tools/gen-cldr test ./...` and `go -C tools/gen-cldr vet ./...`, so encoder/production-decoder round trips and generator static checks cannot be skipped by root-module test patterns. Hand-written files remain outside the generated path set because they do not carry a generated header.

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

### 5.7 Likely-subtag source validation

`tools/gen-cldr/extract.ExtractLikelySubtags` validates the pinned CLDR source
before producing maximize records. Keys must be fully consumed
language, language-script, language-region, or language-script-region forms;
values must be full language-script-region triples. Both sides reuse
`internal/localeid` subtag validators and canonicalizers. Empty, misplaced,
extra, variant, extension, private-use, and normalized-duplicate source forms
fail generation with the original key and value in the error; none may be
silently truncated or overwritten.

---

## 6. Data Access API

### 6.1 Per-domain accessors

Each domain package exposes its own accessors; the formatter packages consume the owning domain, not a root `cldr` package. Access cost follows the representation: direct map lookups, binary searches over compact tables, and rule scans have different complexity. Defensive collection copies cost proportionally to the returned collection. Scalar string references can borrow the domain `_data` without allocating; this is not a blanket O(1) guarantee for every accessor. Composite accessors that hand back a map or slice (for example `date.AvailableFormats` / `date.IntervalFormats` via `maps.Clone`, or `relativetime` field records) **MUST** return a defensive clone so callers cannot mutate cached domain state. The eight supported-locale index accessors (`number`/`date`/`currency`/`unit`/`list`/`relativetime`/`timezone`/`displaynames`) share one `codec.LazyStrings` owner, which decodes its `StringRefSlice` blob once and returns a `slices.Clone` on every `Get`, replacing the former per-domain `(sync.Once, []string, loader, clone-accessor)` quadruple.

Representative surface:

```go
// internal/cldr/locale (kernel)
type Locale uint16
func ResolveLocale(tag string) (Locale, bool)
func AvailableLocales() []string
func MaximizeSubtags(language, script, region string) (lang, scr, reg string, ok bool)
func Version() VersionInfo
func Manifest() ManifestInfo

// internal/cldr/number
func SupportedLocales() []string
func SupportedNumberingSystems() []string
func (Locale) CurrencyNamePattern(numberingSystem, plural string) string
// + symbols / decimal / percent / currency / compact pattern accessors keyed by Locale

// internal/cldr/date
func SupportedLocales() []string
func SupportedCalendars() []string
// + CalendarNames / DateFormat / TimeFormat / AvailableFormats / IntervalFormats / DayPeriodFor

// internal/cldr/currency
func SupportedCurrencies() []string
// + CurrencyDigits / CurrencyDisplayName

// internal/cldr/timezone
// ZoneToMetazone / MetazoneName / display-pattern / period-boundary accessors

// internal/tz (generated identifier facts, hand-written behavior)
func LookupIdentifier(name string) (IdentifierRecord, bool)
func SupportedTimeZones() []string
func TimeZonesForRegion(region string) []string

// internal/cldr/unit
func SupportedLocales() []string
func UnitPattern(loc Locale, unit, width, plural string) string
func PerUnitPattern(loc Locale, unit, width string) string
func CompoundUnitPattern(loc Locale, width string) string

// internal/cldr/{list,relativetime,displaynames}
func SupportedLocales() []string
// + each domain's pattern / field / name accessors
```

MUST rules:

1. `cldrlocale.AvailableLocales()` is the CLDR universe; a formatter's locale matching **MUST** use the owning domain's `SupportedLocales()`, derived from generated payload, where each tag resolves to a non-`Undefined` locale through `ResolveLocale`.
2. `internal/localematcher` **MUST NOT** import any `internal/cldr` package. Its own generated profile supplies distance facts; formatter constructors inject generated supported-locale slices and the locale-kernel maximizer.
3. Root `Intl.supportedValuesOf` accessors consume the owning domain's narrow index (§1.2), except time zones, which consume the exact primary projection of `internal/tz` identifier records.
4. `SupportedCalendars()` derives from date calendar payload keys, maps CLDR `"gregorian"` → ECMA-402 `"gregory"`, and appends `"iso8601"` only when Gregorian data exists. `SupportedNumberingSystems()` includes the full ECMA-402 simple digit set even when the profile generates no matching CLDR symbol payload.
5. Number-domain decimal, percent, scientific, and currency-symbol accessors must fall back from a requested numbering-system row to the locale default numbering-system row when the requested row is absent. Currency-name placement additionally falls back from a missing plural category to `other` before applying the same-locale numbering-system fallback. Compact pattern accessors keep missing tuple results empty so NumberFormat can distinguish unavailable compact formats from base pattern defaults. Number symbols instead use an explicit locale row, then a concrete root row; a validated root alias resolves through the same locale’s latn symbols. An empty numbering-system argument still selects the locale default.
6. Number generation must preserve every recognized `currencyFormats-numberSystem-*/unitPattern-count-{zero,one,two,few,many,other}` row, validate exactly one `{0}` and `{1}` in each present row, reject unknown plural suffixes, and require `other` for the locale's default numbering-system row. Missing categories and non-default rows remain absent so `CurrencyNamePattern` can apply its bounded same-locale fallback.
7. Unit generation must preserve every CLDR row whose de-namespaced identifier passes `IsWellFormedUnitIdentifier`, including direct compound rows. It validates that every generic compound pattern contains exactly one `{0}` and `{1}`, and that every present `perUnitPattern` contains exactly one `{0}`. A missing direct or `perUnitPattern` row is preserved as absence so NumberFormat can choose the next composition path; plural unit patterns may legitimately omit `{0}` and must not be repaired.
8. The root `internal/cldr/` directory is not a Go package. It must contain version metadata and domain subdirectories only; no production code may import a retired root CLDR package.

> **Why**: `availableLocales.json` may list locales without complete formatter payload; supported lists must come from real payload, or `ResolveLocale` could hit a payload-less locale and drift the fallback from Generated reference.

### 6.2 Visibility

The `internal/` path segment forces every domain package private. Formatter public packages do not re-export `Locale` / domain record types; they consume them through `internal/ecma402`. CLDR / ICU / tzdata pins stay in `cldrlocale.Version()` and `internal/cldr/VERSION` for internal use; the root package **MUST NOT** expose `Version()` or other diagnostic APIs, since the ECMA-402 `Intl` namespace has no such member.

---

## 7. Testing

### 7.1 Generator self-test

`tools/gen-cldr/` carries per-domain round-trip tests (`codegen/<domain>_roundtrip_test.go`). The tests derive expected values from the pinned source through the existing loader and compare them with **production** accessors over the committed generated blobs (via the repository workspace during local development, or the pinned root version with `GOWORK=off`). They do not run a mirror decoder or independently prove all loader semantics. Date comparisons include AM/PM and flexible day-period names at wide, abbreviated and narrow widths, preserving category associations and absence. Day-period rules are independently painted into a full-day minute table from source intervals, with fixed points overlaid; every registered locale is compared, including intentional absence. Unregistered supplemental keys are excluded by the encoding contract rather than resolved through parent fallback. Encoder unit tests and byte regeneration separately verify generation against these committed products. `task data:check` runs these tests and `go vet` from the nested module root as part of the standing generated-data gate.

### 7.2 Snapshot / contract test

`task data:check` is the sole byte-identity orchestration gate. It regenerates
all products into one temporary `internal/` tree, compares the complete
generated-header-owned path set and bytes in both directions (including the
private identity products in `internal/localeid`, `internal/localematcher`, and
`internal/tz`), runs the nested generator test/vet proof, then runs
`task data:contract`. The locale package retains manifest, header, runtime
contract, and data-shape tests; per-domain generator tests retain production
decoder/accessor round trips.

### 7.3 Conformance integration

The correctness of domain data access is verified indirectly through the downstream formatter conformance suites (SPEC 70). This SPEC does not duplicate fixture design.

---

## 8. Data Semantics

### Day-period registry filtering

CLDR supplemental day-period rules cover languages beyond the kernel locale registry (kok, yue, zu, and ~15 others). The `date` encoder **encodes day-period rules only for registered locales**:

- A registry-external locale is **unreachable at runtime** — it cannot resolve to a kernel index — so its rules can never be selected; not encoding them loses nothing.
- In the legacy literal renderer, an unchecked lookup for such a locale missed and collided onto index 0, overwriting `und`. The legacy renderer shipped the source-order-last ruleset (Zulu) as the effective `und` rules — a latent bug the blob migration first reproduced bug-for-bug, then fixed.
- `und` therefore takes its own genuine root am/pm rules; registry-external languages are simply not encoded.

> **Why**: The historical `mustLocaleIndex` guard exposed the index-0 collision on first landing. The current error-returning boundary (§5.3) preserves the same data-integrity guarantee without a generation-time panic.
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
- **A reborn root `internal/cldr` package or any import of it** — violates the domain ownership boundary. This is an architecture review rule, not a separate import-graph test gate.
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
- [ ] Key-aware Unicode type aliases, the ordered language-matching profile, and the pinned IANA/CLDR time-zone identifier registry are generated directly into their private runtime owners and byte-compared by `task data:check`.
- [ ] Each domain's `_data` holds only that domain's strings; there is no shared cross-domain string table.
- [ ] `cldrlocale.Version()` derives the pins from the generated manifest; the root package exposes no version diagnostic.

### Representation and gates

- [ ] `TestGeneratedDataShape` passes: Rule A (payload files const-only) and Rule B (no call/index expressions in composite literals); the exemption table is empty.
- [ ] Locale-index resolution returns a contextual error and prevents encoding a record whose locale is absent from the kernel registry; intentionally skipped locales are filtered with a stated reason.
- [ ] Per-domain round-trip tests read back through the production accessor via a pinned `require` plus the repository workspace; no mirror decoder exists.
- [ ] Language-matching loader tests reject malformed fields, unknown variables, duplicate paradigms, invalid distances, containment cycles, and a missing final catch-all while preserving all 378 source rows in order.
- [ ] Time-zone registry generation rejects pin/hash/license/version drift, malformed Zone/Link rows, link cycles, unknown CLDR primary targets, and ASCII-fold collisions; it preserves every legal IANA identifier and `zone.tab` region record.

### Supported values

- [ ] Each domain's `SupportedLocales()` is non-empty, excludes `und`, and each tag has generated payload reachable through `ResolveLocale`.
- [ ] A supported-value query reads only its narrow blob and does not trigger the domain main-payload `sync.Once` (asserted in the domain test).
- [ ] `SupportedCalendars()` maps `"gregorian"` → `"gregory"` and contains `iso8601` when Gregorian data exists; `SupportedNumberingSystems()` contains the full ECMA-402 simple digit set.
- [ ] Root `supportedValuesOf("timeZone")` is exactly the sorted primary projection from `internal/tz`, and no time-zone supported-value query decodes `internal/cldr/timezone` display data.

### Generator and upgrade

- [ ] `tools/gen-cldr/go.mod` is a standalone module with a pinned root requirement and repository workspace membership; round-trip tests import the production accessors.
- [ ] The domain registry (`codegen/domain.go`) is the single source of expected domain packages; the retired literal renderers (`golang_literal.go` / `map_literal.go` / `wrappers.go`) do not exist.
- [ ] `extract/locales.go` outputs a deterministically sorted tag list with `und` at position 0.
- [ ] `task data` regenerates with no git diff (except after a deliberate VERSION change); `task data:check` fails on any byte difference.
- [ ] `task data:check` verifies `tools/gen-cldr` from its module root with `go -C tools/gen-cldr test ./...` and `go -C tools/gen-cldr vet ./...`, not via a root-module `./tools/gen-cldr/...` pattern.

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
- [SPEC 10 §Maximize / Minimize](./10-locale.md) — uses `MaximizeSubtags` for maximization and candidate-equivalent minimization.
- [SPEC 11 §BestFitMatcher](./11-locale-matching.md) — matcher receives supported locales and maximizers from formatter constructors.
- [SPEC 20 §Currency Data](./20-numberformat.md) — consumes `currency` domain accessors.
- [SPEC 30 §DateTimeFormat Core](./30-datetimeformat.md) — consumes `date` domain accessors.
- [SPEC 31 §Skeleton Resolution](./31-datetimeformat-skeleton.md) — consumes `date` availableFormats / intervalFormats.
- [SPEC 32 §TimeZone & Calendar Data](./32-datetimeformat-tz.md) — consumes the `timezone` domain, integrated with `time/tzdata`.
- [SPEC 40 §PluralRules](./40-pluralrules.md) — the `plural` package is output here by the SPEC 40 codegen.

---

> This SPEC is the maintenance record of the CLDR data layer. Version-pin changes trigger a SPEC revision; the active locale list is maintained in `tools/locale-profile.json` (no SPEC revision); volume changes are reviewed through `task build:size`.

### Calendar preference identifiers

The generator adapts legacy CLDR calendar identifiers (`gregorian` → `gregory`,
`ethiopic-amete-alem` → `ethioaa`) and validates/canonicalizes Unicode types
before encoding calendar preferences. Invalid identifiers report the source
path, region and value. Locale consumes these canonical values directly and
filters against actual supported calendars; an empty supported intersection
still returns the Gregorian fallback. Explicit locale calendar keywords retain
their own presence semantics.

Generation input integrity: `tools/gen-cldr/cldr.CrossCheck` validates the name and pinned version in `package.json` for every required CLDR package before writing output. `tools/gen-plural-rules` validates `cldr-core` identity and equal nonempty versions beside cardinal, ordinal, and range inputs, including separate source directories. Package metadata proves release identity, not the integrity of individual JSON bytes.

The selected profile retains every currency name/symbol row from its pinned
`cldr-numbers-full/main/<locale>/currencies.json`. The narrow supported-currency
index is derived from those name keys alone; `currencyData.json` precision
exceptions cannot add membership. DisplayNames language data carries CLDR
variant names and the locale separator alongside the locale pattern.

Number generation carries the positive `minimumGroupingDigits` from resolved
CLDR full JSON into the number blob and production accessor. An absent field
uses ICU’s minimum of one; malformed supplied values fail with source context.
The source/accessor round trip verifies this locale-level value independently
of numbering-system rows.

Number source loading preserves every explicit `symbols-numberSystem-*` row
and each supplied decimal, percent, scientific or currency format family.
Default/latn rows require all four families; non-default families are parsed
independently, and a symbols-only row does not require formats. Supplied
families retain their existing compact, currency spacing and currency-name
fields; absent families remain absent. Malformed supplied families fail with
locale, system and field context. Malformed supplied symbols fail with the
file and numbering-system key; their range separator
comes from the corresponding `miscPatterns` row. The ar/arab and ur/arabext
overrides remain locale-owned. `TestLoadNumbersPreservesNonDefaultSymbols`,
`TestLoadNumbersPreservesPinnedSymbolOverrides`, and
`TestLoadNumbersRejectsMalformedNonDefaultSymbols` in
`tools/gen-cldr/cldr/numbers_test.go` verify this source boundary.

`TestLoadNumbersPreservesPinnedNonDefaultFormats` independently checks raw
CLDR 48.1.0 gu/gujr standard and compact rows; individual-family and malformed
family tests cover absence and invalid supplied input.
`tools/gen-cldr/source_test.go:TestLoadAllPreservesPinnedNumberFormats`
verifies source composition. `TestNumberRoundTrip` verifies those loaded rows
through the production accessor, and the public NumberFormat
`manual/number-format-families.json` fixture verifies style grouping. Roundtrip
alone cannot detect input omitted by both source loading and generation.

### Number root symbols source

`tools/gen-cldr/cldr/number-root/source.json` owns the CLDR release, upstream
tag/commit/URL, SHA-256 and license of the complete original `root.xml` beside
it. CLDR JSON 48.1.0 names DATA=release-48-1-final1; that annotated tag points
to commit `225136a3bd3eff573f3d64c8fa25d6f9afa974ad`. The immutable input is
`https://raw.githubusercontent.com/unicode-org/cldr/225136a3bd3eff573f3d64c8fa25d6f9afa974ad/common/main/root.xml`,
SHA-256 `c16e0df1d410c8f481fcb9c4dd195709e08ec9f8a6519d2d9905078e32990f17`.
Both snapshot and source record are generator-only embedded inputs; VERSION
agreement and actual-byte SHA validation precede loading, with no fetch at
generation or runtime. A CLDR upgrade must update this input with its npm pin.

The loader preserves explicit locale symbols before filling absent concrete
root rows. It accepts only the pinned locale-to-latn alias form and proves
that every constructor-supported simple system has a concrete row or that
alias. It does not materialize alias rows for every locale. Root arab uses
NaN and exponent `NaN` / `اس`; ar/arab retains its local `ليس رقمًا` / `أس`.
Inherited root symbols retain the locale default's range pattern separator;
root symbols do not own miscPatterns, decimal formats or currency placement.
Other LDML/XPath semantics are outside this loader. Missing required root
symbols, unsupported aliases, duplicate systems, missing system coverage,
version mismatch or damaged snapshot bytes fail with source context.

`tools/gen-cldr/cldr/number_root_test.go`, the LoadAll source tests and
`TestRunRecordsNumberRootSources` verify damaged inputs, locale overrides,
range-pattern ownership and reproducible manifest hashes. Production
roundtrip and formatter conformance verify the generated projection.
`internal/cldr/number/number_test.go:TestNumberSymbolsLocaleAlias` compares the
complete pinned ar-EG records: deva follows locale latn, while an empty argument
selects the default arab row. The positive and negative public fixtures in
`numberformat/testdata/conformance/manual/symbols-alias.json` verify separators,
adjacent LRM literals and complete resolved options without changing the
requested numbering system.
Algorithm references: `.references/formatjs/packages/intl-numberformat/scripts/number-data/number-data.ts`
(root aliases restart at the requested locale),
`.references/node/deps/icu-small/source/i18n/dcfmtsym.cpp` (symbol inheritance
and latn fallback). FormatJS's `.references/formatjs/MODULE.bazel` pins XML
48.2; it is an algorithm reference, not this library's data source.

## Canonical identity data and generation publication

`tools/gen-cldr/cldr/identity_alias.go` loads language, script, territory, variant
and subdivision replacements from pinned `cldr-core/supplemental/aliases.json`.
`bcp47.go` loads Unicode and transformed extension type aliases, retaining their
separate key namespaces. `codegen/unicode_alias.go` emits these with existing
Unicode aliases into `internal/localeid/unicode_alias_data.go`. Language rules
are grouped by source language; multi-target territory choice consumes the
existing CLDR maximizer rather than duplicating likely-subtag data. Locale and
DisplayNames share this identity owner; no registry knowledge is required for
syntax acceptance. `task data:check` includes this generated file.

`task data` invokes `tools/generate-data`. Both CLDR and plural generation write
to one fresh `.tmp/generated-*` candidate before installation. Generator or
input-validation failure leaves the official generated path set and bytes
unchanged. `tools/internal/generateddata` supplies the same header-derived
ownership inventory used by the byte checker. Installation preflights all
handwritten-file collisions, writes sorted generated paths, then removes only
obsolete generated files. Handwritten decoders/accessors and VERSION survive.

Installation is not a filesystem transaction: I/O failure may leave a partial
install, returns a contextual nonzero error, and retains the candidate. The
post-install `data:contract` verifies the checkout, not the candidate before
publication. `task data:check` remains the separate fresh-byte comparison gate.
`tools/generate-data/main_test.go` runs the real generators against auxiliary
output failure and missing/malformed plural inputs, comparing original target
paths and bytes; installation tests cover exact set and handwritten preservation.
