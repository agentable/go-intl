# go-intl

Go implementation of the active ECMA-402 Intl surface: `Locale`, `NumberFormat`, `DateTimeFormat`, `PluralRules`, `ListFormat`, `RelativeTimeFormat`, `DurationFormat`, `DisplayNames`, `Collator`, `Segmenter`, and the root `go-intl` namespace. The compatibility target is the ECMA-402 specification in `.references/ecma402/spec/`; FormatJS is the vendored readable implementation reference used to validate observable behavior.

For human usage examples, read [`README.md`](README.md). This file is the development guide for AI coding agents.

## Commands

```bash
task test                 # go test -race -p 1 ./...
task lint                 # go mod tidy diff check + pinned golangci-lint v2
task fmt                  # go fmt ./...
task vet                  # go vet ./...
task verify               # deps + fmt + vet + lint + test + conformance + data contract + vuln
task vuln                 # govulncheck ./...
task deps                 # go mod download && go mod tidy
task deps:update          # update root and nested module dependencies
task conformance:verify   # fixture schema + XFAIL + skip-list + divergence audit + coverage
task data                 # regenerate CLDR data from local npm CLDR checkout
task data:check           # verify generated CLDR data is byte-identical
task data:contract        # generated CLDR data contract tests
task build:size           # root, formatter, and CLDR binary size delta report
task bench:run            # one-shot benchmark telemetry
task bench                # benchmark report, optionally BASELINE=<file>
```

Targeted checks:

```bash
go test -race ./numberformat/...
go test -race -run TestPluralRules_Cardinal/en ./pluralrules/
go test -bench=. -benchmem ./numberformat/
(cd tools/gen-cldr && go test ./...)
(cd tools/gen-cldr && go vet ./...)
(cd tools/gen-plural-rules && go test ./...)
(cd tools/gen-plural-rules && go vet ./...)
(cd tools/gen-fixtures-from-formatjs && go test ./...)
(cd tools/gen-fixtures-from-formatjs && go vet ./...)
go test ./tools/check-conformance ./tools/conformance
```

## Architecture

```text
go-intl/
├── intl.go                # Root Intl namespace aliases and GetCanonicalLocales
├── supported.go           # Root Intl.supportedValuesOf typed accessors
├── locale/                # Intl.Locale parsing, canonicalization, maximize/minimize, info getters
├── numberformat/          # Intl.NumberFormat formatting, parts, ranges, options
├── datetimeformat/        # Intl.DateTimeFormat formatting, parts, ranges, time zones
├── pluralrules/           # Intl.PluralRules cardinal/ordinal select and selectRange
├── listformat/            # Intl.ListFormat list formatting and parts
├── relativetimeformat/    # Intl.RelativeTimeFormat relative time formatting and parts
├── durationformat/        # Intl.DurationFormat duration formatting and parts
├── displaynames/          # Intl.DisplayNames code-to-name lookup
├── collator/              # Intl.Collator locale-sensitive compare/sort
├── segmenter/             # Intl.Segmenter grapheme/word/sentence segmentation
├── internal/
│   ├── ecma402/           # Shared ECMA-402 abstract operations
│   ├── cldr/              # Generated CLDR data and runtime accessors
│   ├── collation/         # Collator backend capability metadata
│   ├── decimal/           # apd-backed decimal math for Intl mathematical values
│   ├── localematcher/     # Lookup/best-fit locale matching
│   └── tz/                # IANA time-zone resolution data
├── tools/
│   ├── conformance/       # Shared fixture, XFAIL, divergence, and coverage checks
│   ├── check-conformance/ # Unified conformance verification CLI
│   ├── gen-cldr/          # CLDR JSON -> generated Go data
│   ├── gen-plural-rules/  # CLDR plural JSON -> generated Go rules
│   └── gen-fixtures-from-formatjs/ # FormatJS/Node references -> conformance fixtures
├── SPECS/                 # Contract layer; read before design changes
├── .references/           # Reference implementations; study before coding
└── reports/               # Dependency issue reports
```

The root package mirrors the JavaScript `Intl` namespace shape as closely as Go allows. Formatter-specific constructors, options, and methods live in their formatter packages; `internal/*` stays private.

## Agent Workflow

### Design Workflow - Read ECMA-402 and SPECS First

Before designing or modifying code, read the relevant ECMA-402 source files in `.references/ecma402/spec/` and then the relevant `SPECS/` documents end to end. ECMA-402 is the normative source for constructor shape, locale/options negotiation, option names and values, resolved options, parts, range sources, and error conditions.

If a SPEC contradicts ECMA-402, update the SPEC first. Do not preserve local API convenience, FormatJS helper shape, or historical tests over the specification.

### Implementation Workflow - Find 2 References First

Before writing implementation code, find at least two relevant implementation references in `.references/` and study their behavior:

1. Start with `.references/ecma402/spec/<surface>.html` to identify the exact abstract operations and observable contract.
2. Study the matching `formatjs/packages/<polyfill>/` directory for a readable polyfill implementation.
3. Cross-reference `formatjs/packages/ecma402-abstract/` for shared algorithms.
4. Use `.references/node/` for native V8/ICU behavior and Node localization snapshots.
5. Port language-agnostic test fixtures with the implementation.

### Documentation Boundaries

- README is the usage guide: installation, examples, API overview.
- CLAUDE.md / AGENTS.md is the development guide: workflow, commands, rules, indexes.
- SPECS are the contract layer: behavior, data layout, and acceptance criteria.
- Source comments explain non-obvious implementation decisions only.

## SPECS Index

Specification documents in [`SPECS/`](SPECS/) are maintained records of design contracts.

| Spec | Topic |
|------|-------|
| [`00-vision-and-scope.md`](SPECS/00-vision-and-scope.md) | Project vision, active ECMA-402 surface, reference policy, scope boundaries |
| [`10-locale.md`](SPECS/10-locale.md) | `locale.Locale`, parsing, Unicode extension fields, maximize/minimize, info getters |
| [`11-locale-matching.md`](SPECS/11-locale-matching.md) | Lookup and best-fit matching, `ResolveLocale`, supported-locales filtering |
| [`12-abstract-operations.md`](SPECS/12-abstract-operations.md) | Production-used ECMA-402 helpers, validators, pattern parsing, digit and plural operations |
| [`20-numberformat.md`](SPECS/20-numberformat.md) | `numberformat` API, option resolution, shared digit formatting, parts, ranges, notation and unit behavior |
| [`21-number-math.md`](SPECS/21-number-math.md) | Decimal backend, mathematical value conversion, rounding modes and increments |
| [`30-datetimeformat.md`](SPECS/30-datetimeformat.md) | `datetimeformat` API, resolved options, parts, ranges |
| [`31-datetimeformat-skeleton.md`](SPECS/31-datetimeformat-skeleton.md) | Skeleton parsing, best-fit matcher, pattern scoring |
| [`32-datetimeformat-tz.md`](SPECS/32-datetimeformat-tz.md) | Time-zone resolution, canonical links, Gregorian calendar and metazone data |
| [`40-pluralrules.md`](SPECS/40-pluralrules.md) | `pluralrules` API, operands, CLDR plural rule codegen, `selectRange` |
| [`41-listformat.md`](SPECS/41-listformat.md) | `listformat` API, CLDR list patterns, formatting, parts, supported locales |
| [`42-relativetimeformat.md`](SPECS/42-relativetimeformat.md) | `relativetimeformat` API, relative field data, number/plural composition, parts |
| [`43-durationformat.md`](SPECS/43-durationformat.md) | `durationformat` API, duration records, unit options, formatting, parts |
| [`44-displaynames.md`](SPECS/44-displaynames.md) | `displaynames` API, type/style/fallback semantics, CLDR localenames data |
| [`45-collator.md`](SPECS/45-collator.md) | `collator` API, `x/text/collate` mapping, sensitivity/numeric/caseFirst behavior |
| [`46-segmenter.md`](SPECS/46-segmenter.md) | `segmenter` API, UAX #29 via `uniseg`, byte-offset bridge |
| [`50-cldr-data.md`](SPECS/50-cldr-data.md) | CLDR pins, generated data layout, formatter supported-locale accessors, generator architecture |
| [`60-facade.md`](SPECS/60-facade.md) | Root `Intl` namespace, static common functions, constructor aliases, and forbidden root one-shot APIs |
| [`61-messageformat-integration.md`](SPECS/61-messageformat-integration.md) | `messageformat-go` adapter contract and dependency direction |
| [`70-conformance.md`](SPECS/70-conformance.md) | Fixture format, FormatJS extractor, skip-list audit, divergences, conformance gates |
| [`71-benchmark.md`](SPECS/71-benchmark.md) | Benchmark layout, non-blocking performance telemetry, benchstat workflow |
| [`72-operation-ledger.md`](SPECS/72-operation-ledger.md) | Public surface to ECMA-402 owner, implementation, and verification ledger |
| [`73-json-records.md`](SPECS/73-json-records.md) | JSON field names and presence policy for resolved options, parts, segment records, and locale info |

## References Index

Reference projects in [`.references/`](.references/) are read-only implementation guides.

| Category | Path | Use |
|----------|------|-----|
| ECMA-402 specification | `.references/ecma402/spec/` | Normative source for the Intl namespace, constructors, locale/options negotiation, abstract operations, methods, parts, resolved options, and error boundaries |
| ECMA-402 TypeScript reference | `.references/formatjs/` | Readable implementation reference: Intl polyfills, ECMA-402 abstract ops, tests, fixtures |
| ECMA-402 Node/V8 reference | `.references/node/` | Native-engine tiebreaker for ICU-backed behavior, edge cases, time zones, and Node Intl output snapshots |
| ECMA-402 PHP/C scope reference | `.references/ext/` | Scope check for full Intl surface and native implementation boundaries |
| CLDR-driven Go reference | `.references/intl/` | Go pattern for embedding CLDR data and using `golang.org/x/text/language` |
| Go date/time API references | `.references/carbon/` | API ergonomics and date/time edge-case comparison; ECMA-402 remains normative |
| Go money reference | `.references/go-money/` | Currency API and data-shape comparison; do not copy semantics over ECMA-402 |

## Design Philosophy

- **KISS** - One representation per concept: one `Locale`, one option struct per formatter, one formatter type per ECMA-402 surface.
- **DRY** - Shared ECMA-402 rules live in `internal/ecma402`; generated CLDR data is consumed through one accessor layer in `internal/cldr`.
- **YAGNI** - The active surface is exactly the ten ECMA-402 constructors plus the root namespace helpers (`getCanonicalLocales` and typed supported-value accessors). New ECMA-402 additions wait for a new edition.
- **Native Intl alignment over local invention** - Public APIs must map to ECMA-402 constructors, methods, options, resolved options, part records, range sources, or error conditions. Go type bridges are allowed; new semantics are not.
- **Errors as teachers** - Constructor errors name the option, value, and locale whenever possible.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope, or documentation masquerading as code.

## API Design Principles

- **Intl namespace first** - The root package is not a constructor and should not behave like a per-locale session object. It represents the JavaScript `Intl` namespace as closely as Go allows.
- **Native API mapping is mandatory** - Before adding or changing exported API, identify the exact ECMA-402 JavaScript owner: `Intl`, `Intl.Locale`, `Intl.NumberFormat`, `Intl.DateTimeFormat`, `Intl.PluralRules`, `Intl.ListFormat`, `Intl.RelativeTimeFormat`, `Intl.DurationFormat`, `Intl.DisplayNames`, `Intl.Collator`, or `Intl.Segmenter`. If no native owner exists, do not add it unless it is a narrow Go typed bridge for an existing native operation.
- **Constructor parity** - Formatter `New` functions mirror `new Intl.<Constructor>(locales, options)`: callers pass exactly one typed `Options` value, and zero-valued `Options{}` means the ECMA-402 empty options object.
- **Typed bridges only** - Typed Go values such as `numberformat.Int` and `numberformat.Decimal` are bridges for JavaScript methods like `format(value)`; they must not introduce behavior that native Intl lacks.
- **Typed boundaries** - User locale identifiers enter through `locale.Parse`, `locale.ParseList`, or `locale.New`; formatter constructors receive `locale.List`, while `locale.FromTag` is the explicit `language.Tag` bridge. `language.Tag` is the only `golang.org/x/text` type allowed in exported signatures; secondary `x/text` types stay internal.

## Coding Rules

### Must Follow

- Go 1.26.3. Use modern stdlib features already present in the codebase: `slices`, `maps`, `for range N`, `sync.OnceValue`, `sync.Map`, `log/slog`, and `testing.B.Loop()`.
- Follow Google Go Best Practices: <https://google.github.io/go-style/best-practices>.
- Follow Google Go Style Decisions: <https://google.github.io/go-style/decisions>.
- Keep interfaces small and consumer-owned; do not expose broad internal abstractions.
- Every exported symbol must map to an ECMA-402 constructor, method, option, resolved option, part record, range source, static `Intl` function, or a Go-only typed bridge for one of those items.
- Validate formatter options in constructors. After construction, method error behavior must match the corresponding ECMA-402 operation: invalid typed inputs return errors where JavaScript would throw `TypeError` or `RangeError`; ordinary formatting does not hide constructor failures.
- Wrap user-fixable errors with package context and a sentinel error so callers can use `errors.Is` / `errors.As`.
- Keep generated CLDR/runtime data out of hot-path file I/O. Runtime data lives in generated Go source under `internal/cldr`.
- Keep formatter supported locale lists generated from actual CLDR payload maps or truthful engine capability accessors. Constructor `SupportedLocalesOf` methods must use `internal/ecma402.SupportedLocalesOf` / `SupportedLocales` with generated accessors instead of duplicating matcher, filtering, or requested-locale dedupe loops.
- Keep shared string and integer option validation in `internal/ecma402`. Formatter packages pass formatter-owned allowed values through helpers such as `RequiredStringOption`, `OptionalStringOption`, `InvalidStringOption`, and `InvalidIntegerOption`; do not hand-roll equivalent `switch` or `slices.Contains` loops.
- Keep root supported-value accessors in the root package, conventionally in `supported.go`, backed by CLDR/tz data, active collation capability, or ECMA-402 sanctioned constants. Do not create public `cldr`, `ecma402`, or `supported` packages for this data. Calendars must include `iso8601`; numbering systems must include the ECMA-402 simple digit table; do not add ad hoc runtime lists.
- Keep `DateTimeFormat` calendar support tied to `internal/cldr.SupportedCalendars()` and generated date data; do not copy calendar allow-lists into constructors.
- Keep `Segmenter` supported locales honest. Do not advertise dictionary or CJK-tailored locales such as `ja`, `th`, or `zh-Hant` until the active segmentation backend supports their word-boundary behavior.
- Keep ECMA-402 digit rounding centralized in `internal/ecma402/numberformat.FormatNumericToString`; `numberformat` and `pluralrules` both feed it resolved digit options and consume the rounded numeric value.
- Keep constructor and `SupportedLocalesOf` options aligned with the JavaScript single-options-object model. Public Go entrypoints receive exactly one typed `Options` value; use `Options{}` for omitted or empty JS options instead of variadic `Options`.
- Keep NumberFormat unit identifiers exact and case-sensitive. `UnitIdentifier("METER")` must not silently become `"meter"`; native `Intl.NumberFormat` rejects non-canonical unit casing.
- Represent optional scalar input options as pointers (`*int`, `*bool`, `*string`) and use root helpers `gointl.Int`, `gointl.Bool`, and `gointl.String` at call sites. Constructor code must copy pointed-to scalar values into internal config before storing anything on formatter instances.
- Represent JS-omitted resolved options as pointers (`*int`, `*LanguageDisplay`, etc.) rather than zero-valued primitives. ECMA-402 distinguishes "property absent" from "property = 0"; reuse a Go zero value collapses the two and makes resolved-options comparisons ambiguous.
- Each formatter package declares its own `PartType` constant set that mirrors the ECMA-402 partition record types it can emit (including types reachable only through embedded NumberFormat / list patterns). Do not share `PartType` types across packages — partition records are scoped to the constructor, and sharing creates cross-package coupling where ECMA-402 has none.
- Prefer exact modern stdlib helpers over one-off private helpers: `strings.IndexByte` for ASCII pattern-byte membership, `slices`/`maps` for deterministic collection transforms. Keep manual parser loops when the index advances by token width or quoted-literal spans.

### Forbidden

- No `panic` in production code except intentional `Must*` helpers.
- No reinventing locale parsing; use `golang.org/x/text/language` and the project canonicalization layer.
- No runtime JSON/ICU file loading for formatter data.
- No public cache controls or root-level formatter option re-exports.
- No root diagnostic APIs such as `Version()`; CLDR, ICU, and tzdata pins are internal metadata.
- No back-compat shims, alias APIs, or parallel v2 names unless a SPEC explicitly requires them.
- No root-level one-shot helpers or per-locale `Intl` session APIs unless their SPEC maps them to the ECMA-402 `Intl` namespace. JavaScript `Intl` is not a constructor.
- No documentation masquerading as code: do not encode spec prose as constants or tables no program consumes.
- No hand-written formatter supported-locale lists; update CLDR codegen, generated supported accessors, or the engine-specific support boundary instead.
- No `strconv.ParseFloat`, `math.Log10`, or `math.Pow10` in NumberFormat decimal formatting, compact scaling, scientific notation, or percent math. Use `internal/decimal`.
- No `Append*` byte-buffer methods on formatter types. ECMA-402 has no equivalent surface; if a hot-path caller needs zero-allocation output, wrap the formatter externally rather than growing the public API.
- No working around dependency bugs inline. Create `reports/<dependency-name>.md` and continue with unaffected work.

## Dependency Issue Reporting

When you encounter a bug, limitation, or unexpected behavior in a dependency:

1. Do not reimplement the dependency's functionality inside go-intl.
2. Do not silently skip the dependency or fork its logic.
3. Create `reports/<dependency-name>.md`.
4. Include dependency name and version, trigger scenario, expected behavior, actual behavior, relevant errors or traces, and any workaround suggestion without implementing it.
5. Continue with tasks that do not depend on the broken behavior.

## Testing

- Use stdlib `testing` and table-driven tests. Do not add `testify` unless a package already uses it.
- Call `t.Parallel()` in tests and subtests unless shared state makes parallel execution invalid.
- Use `b.Loop()` for benchmarks.
- Run `task test` for the race-detector gate.
- Put FormatJS-derived cases in formatter `testdata/` as JSON fixtures and assert byte-equal output.
- Generate FormatJS-derived fixtures with `tools/gen-fixtures-from-formatjs`; generated files live under `<package>/testdata/conformance/formatjs/`. The active generated gate covers statically reducible NumberFormat `format`, `formatToParts`, `formatRange`, and `formatRangeToParts`; DateTimeFormat `format`, `formatToParts`, `formatRange`, and `formatRangeToParts`; PluralRules `select` and `selectRange`; Locale `toString`, `maximize`, `minimize`, and `Intl.getCanonicalLocales`; ListFormat `format`; RelativeTimeFormat `format`; DurationFormat `format`.
- Unextracted or partially extracted FormatJS sources must appear in root `.skip-list.json` with `source`, `category`, and `reason`.
- `.skip-list.json` is extraction audit only. A generated fixture that fails must be handled through `testdata/divergences.md` or `testdata/xfail.json`, never by removing it from testdata.
- Record accepted reference mismatches in `<package>/testdata/divergences.md` only when that package has active or resolved divergence entries; empty placeholder files are not required. Keep IDs auditable by `task conformance:verify`.
- Run `task conformance:verify` after fixture, skip-list, XFAIL, or divergence changes; the unified checker validates schema, source ownership, expiration, divergence references, and coverage health.
- Public formatters should have runnable `Example*` functions when adding meaningful new surface.

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `golang.org/x/text` | BCP 47 parsing, `language.Tag`, and Unicode/CLDR building blocks |
| `github.com/cockroachdb/apd/v3` | Decimal math backend for Intl mathematical values and rounding |
| `github.com/rivo/uniseg` | Unicode text segmentation backend for `Intl.Segmenter` |

Add runtime dependencies only when an active SPEC requires them.

## Error Handling

- Root `gointl` owns the public error categories: `ErrInvalidOption`, `ErrUnsupportedOption`, `ErrInvalidValue`, `ErrInvalidCode`, `ErrInvalidKey`, `ErrUnsupportedLocale`, and `ErrUnsupportedBackend`.
- Error text should teach the caller what to fix using the three-part shape: owner/name/value/locale, `expected ...`, and `got ...`.
- Public caller-fixable errors should expose `*gointl.Error` through `errors.As`; formatter packages build those errors through `internal/intlerr` to avoid root import cycles.
- Do not expose ECMA-402 abstract operation names such as `GetOption`, `PartitionPattern`, or `ResolveLocale` in public error text; those names belong in SPECS and internal code only.
- Use `%w` wrapping for underlying dependency errors; never match errors by string.
- Root namespace helpers do not own formatter construction semantics. Constructor failures remain produced by formatter packages but classify through root sentinels and must not be hidden behind fallback output.

## Performance

- Benchmark telemetry lives in `Benchmark*` functions and `SPECS/71-benchmark.md`.
- Use representative `go test -bench` commands when changing formatter hot paths.
- Use `task bench` with `BASELINE=<file>` for non-blocking benchstat comparisons.
- Use `task build:size` when reviewing CLDR data size changes.

## Linting

`task lint` runs the pinned `bin/golangci-lint` version from `.golangci.version` and checks `go mod tidy` output. The config is `.golangci.yml`; notable enabled linters include `exhaustive`, `errorlint`, `err113`, `gocritic`, `gosec`, `misspell`, `noctx`, `prealloc`, and `revive`.

Nested tool modules are checked from their own module roots. Do not use root-module patterns such as `go test ./tools/gen-cldr/...`; run from each module root instead, for example `(cd tools/gen-cldr && go test ./...)`, `(cd tools/gen-plural-rules && go test ./...)`, and `(cd tools/gen-fixtures-from-formatjs && go test ./...)`.

## CI

GitHub Actions runs on pushes to `main` and pull requests:

- `test`: `task deps`, then `task test`
- `lint`: `task deps`, then `task lint`
- `security`: installs `govulncheck`, then runs `govulncheck ./...`
- `data-verify`: regenerates CLDR data, then runs `task data:check`

## Agent Skills

Primary local skills live in `.agents/skills/`. Use the narrowest skill that matches the requested maintenance pass.

| Skill | When to Use |
|-------|-------------|
| [agent-md-writing](.agents/skills/agent-md-writing/) | Regenerating `CLAUDE.md` and preserving the `AGENTS.md` symlink |
| [readme-writing](.agents/skills/readme-writing/) | Updating README usage documentation |
| [library-docs-maintaining](.agents/skills/library-docs-maintaining/) | Refreshing `README.md`, `CLAUDE.md`, and `AGENTS.md` together |
| [spec-writing](.agents/skills/spec-writing/) | Creating or revising individual `SPECS/*.md` contracts |
| [library-specs-maintaining](.agents/skills/library-specs-maintaining/) | Consolidating design/spec docs into `SPECS/` |
| [spec-gap-analyzing](.agents/skills/spec-gap-analyzing/) | Finding gaps between SPECS and code before implementation work |
| [spec-gap-tasking](.agents/skills/spec-gap-tasking/) | Turning a spec-gap analysis into executable tasks |
| [spec-reviewing](.agents/skills/spec-reviewing/) | Reviewing SPECS for completeness, consistency, and over-engineering before implementation |
| [go-best-practices](.agents/skills/go-best-practices/) | Applying Go API, naming, error, concurrency, and testing rules |
| [modernizing](.agents/skills/modernizing/) | Applying Go 1.20-1.26 language and stdlib idioms |
| [code-simplifying](.agents/skills/code-simplifying/) | Simplifying recently changed code without changing behavior |
| [code-deduplicating](.agents/skills/code-deduplicating/) | Extracting repeated patterns that appear three or more times |
| [code-refactoring](.agents/skills/code-refactoring/) | Planning and applying focused refactors that reduce redundancy without broad rewrites |
| [architecture-audit](.agents/skills/architecture-audit/) | Auditing package boundaries, circular dependencies, and SPECS alignment |
| [library-code-optimizing](.agents/skills/library-code-optimizing/) | Removing dead code and improving internal quality behavior-preservingly |
| [library-code-modernizing](.agents/skills/library-code-modernizing/) | Running a folder-by-folder Go modernization pass |
| [library-code-simplifying](.agents/skills/library-code-simplifying/) | Simplifying internal library code folder by folder |
| [library-comment-optimizing](.agents/skills/library-comment-optimizing/) | Trimming comments to useful godoc and non-obvious implementation notes |
| [library-config-completing](.agents/skills/library-config-completing/) | Completing baseline library config such as `.gitignore`, lint, and hooks |
| [library-ci-fixing](.agents/skills/library-ci-fixing/) | Repairing a failing GitHub Actions run from logs with the smallest root-cause fix |
| [library-error-optimizing](.agents/skills/library-error-optimizing/) | Renaming, reorganizing, or deduplicating package errors |
| [library-panic-optimizing](.agents/skills/library-panic-optimizing/) | Replacing unnecessary panics with errors and `Must*` wrappers |
| [library-symbol-naming](.agents/skills/library-symbol-naming/) | Improving public and internal symbol names |
| [library-file-naming](.agents/skills/library-file-naming/) | Aligning Go source/test filenames with package idiom |
| [library-legacy-pruning](.agents/skills/library-legacy-pruning/) | Removing deprecated APIs and old compatibility shims |
| [library-surface-lock-pruning](.agents/skills/library-surface-lock-pruning/) | Deleting low-value public-surface lock tests and handing behavior gaps to test covering |
| [library-test-covering](.agents/skills/library-test-covering/) | Raising package test coverage with production-grade tests |
| [library-upgrade-latest](.agents/skills/library-upgrade-latest/) | Refreshing tooling, dependencies, and skills to the current library baseline |
| [golangci-linting](.agents/skills/golangci-linting/) | Configuring or fixing golangci-lint v2 findings |
| [taskfile-configuring](.agents/skills/taskfile-configuring/) | Editing `Taskfile.yml` tasks and dependencies |
| [github-actions-configuring](.agents/skills/github-actions-configuring/) | Configuring or repairing GitHub Actions workflows for Go library CI |
| [dependency-selecting](.agents/skills/dependency-selecting/) | Choosing Go dependencies when a SPEC requires one |
| [research-analyzing](.agents/skills/research-analyzing/) | Structuring research over `.references/` before spec work |
| [research-writing](.agents/skills/research-writing/) | Writing research reports from reference projects |
| [committing](.agents/skills/committing/) | Creating conventional commits for scoped repository changes |
| [releasing](.agents/skills/releasing/) | Preparing and tagging a semantic version release |
| [tdd-analyzing](.agents/skills/tdd-analyzing/) | Planning TDD implementation scope from SPECS and source |
| [tdd-planning](.agents/skills/tdd-planning/) | Producing a detailed red-green-refactor plan for a single feature |
| [tdd-implementing](.agents/skills/tdd-implementing/) | Implementing features with red-green-refactor cycles |
