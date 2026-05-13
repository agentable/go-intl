# go-intl

Go implementation of the active ECMA-402 Intl surface: `Locale`, `NumberFormat`, `DateTimeFormat`, `PluralRules`, and the root `go-intl` namespace. The compatibility target is the ECMA-402 specification in `.references/ecma402/spec/`; FormatJS is the vendored readable implementation reference used to validate observable behavior.

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
task conformance:verify   # fixture schema + divergence audit validation
task data:update          # regenerate CLDR data from local npm CLDR checkout
task data:verify          # verify generated CLDR data is byte-identical
task data:contract        # generated CLDR data contract tests
task build:size           # CLDR binary size budget check
task bench:gate           # perf-tag benchmark budget tests
task bench                # benchmark report, optionally BASELINE=<file>
```

Targeted checks:

```bash
go test -race ./numberformat/...
go test -race -run TestPluralRules_Cardinal/en ./pluralrules/
go test -bench=. -benchmem ./numberformat/
(cd tools/gen-cldr && go test ./...)
(cd tools/gen-cldr && go vet ./...)
(cd tools/gen-fixtures-from-formatjs && go test ./...)
```

## Architecture

```text
go-intl/
├── intl.go                # Root Intl namespace helpers and active constructor aliases
├── locale/                # Intl.Locale parsing, canonicalization, maximize/minimize, info getters
├── numberformat/          # Intl.NumberFormat formatting, parts, ranges, options
├── datetimeformat/        # Intl.DateTimeFormat formatting, parts, ranges, time zones
├── pluralrules/           # Intl.PluralRules cardinal/ordinal select and selectRange
├── internal/
│   ├── ecma402/           # Shared ECMA-402 abstract operations
│   ├── cldr/              # Generated CLDR data and runtime accessors
│   ├── decimal/           # apd-backed decimal math for Intl mathematical values
│   ├── localematcher/     # Lookup/best-fit locale matching
│   └── tz/                # IANA time-zone resolution data
├── tools/
│   ├── gen-cldr/          # CLDR JSON -> generated Go data
│   ├── gen-plural-rules/  # CLDR plural JSON -> generated Go rules
│   ├── gen-fixtures-from-formatjs/ # FormatJS/Node references -> conformance fixtures
│   ├── check-fixtures/    # Conformance fixture schema validator
│   └── check-divergences/ # Divergence ID audit tool
├── SPECS/                 # Contract layer; read before design changes
├── .references/           # Reference implementations; study before coding
└── reports/               # Dependency issue reports
```

The root package mirrors the JavaScript `Intl` namespace shape as closely as Go allows. Formatter-specific constructors, options, and methods live in their formatter packages; `internal/*` stays private.

## Agent Workflow

### Design Phase - Read ECMA-402 and SPECS First

Before designing or modifying code, read the relevant ECMA-402 source files in `.references/ecma402/spec/` and then the relevant `SPECS/` documents end to end. ECMA-402 is the normative source for constructor shape, locale/options negotiation, option names and values, resolved options, parts, range sources, and error conditions.

If a SPEC contradicts ECMA-402, update the SPEC first. Do not preserve local API convenience, FormatJS helper shape, or historical tests over the specification.

### Implementation Phase - Find 2 References First

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

Specification documents in [`SPECS/`](SPECS/) are the single source of truth for design contracts.

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
| [`50-cldr-data.md`](SPECS/50-cldr-data.md) | CLDR pins, generated data layout, formatter supported-locale accessors, generator architecture |
| [`60-facade.md`](SPECS/60-facade.md) | Root `Intl` namespace, static common functions, constructor aliases, and forbidden root one-shot APIs |
| [`61-messageformat-integration.md`](SPECS/61-messageformat-integration.md) | `messageformat-go` adapter contract and dependency direction |
| [`70-conformance.md`](SPECS/70-conformance.md) | Fixture format, FormatJS extractor, skip-list audit, divergences, conformance gates |
| [`71-benchmark.md`](SPECS/71-benchmark.md) | Benchmark layout, performance thresholds, benchstat workflow |

## References Index

Reference projects in [`.references/`](.references/) are read-only implementation guides.

| Category | Path | Use |
|----------|------|-----|
| ECMA-402 specification | `.references/ecma402/spec/` | Normative source for the Intl namespace, constructors, locale/options negotiation, abstract operations, methods, parts, resolved options, and error boundaries |
| ECMA-402 TypeScript reference | `.references/formatjs/` | Readable implementation reference: Intl polyfills, ECMA-402 abstract ops, tests, fixtures |
| ECMA-402 Node/V8 reference | `.references/node/` | Native-engine tiebreaker for ICU-backed behavior, edge cases, time zones, and Node Intl output snapshots |
| ECMA-402 PHP/C scope reference | `.references/ext/` | Scope check for full Intl surface and native implementation boundaries |
| CLDR-driven Go reference | `.references/intl/` | Go pattern for embedding CLDR data and using `golang.org/x/text/language` |

## Design Philosophy

- **KISS** - One representation per concept: one `Locale`, one option struct per formatter, one formatter type per ECMA-402 surface.
- **DRY** - Shared ECMA-402 rules live in `internal/ecma402`; generated CLDR data is consumed through one accessor layer in `internal/cldr`.
- **YAGNI** - Only `Locale`, `NumberFormat`, `DateTimeFormat`, `PluralRules`, and common `Intl` namespace functions are active; other Intl constructors wait for a consumer.
- **Native Intl alignment over local invention** - Public APIs must map to ECMA-402 constructors, methods, options, resolved options, part records, range sources, or error conditions. Go type bridges are allowed; new semantics are not.
- **Errors as teachers** - Constructor errors name the option, value, and locale whenever possible.
- **Never:** accidental complexity, feature gravity, abstraction theater, configurability cope, or documentation masquerading as code.

## API Design Principles

- **Intl namespace first** - The root package is not a constructor and should not behave like a per-locale session object. It represents the JavaScript `Intl` namespace as closely as Go allows.
- **Native API mapping is mandatory** - Before adding or changing exported API, identify the exact ECMA-402 JavaScript owner: `Intl`, `Intl.Locale`, `Intl.NumberFormat`, `Intl.DateTimeFormat`, or `Intl.PluralRules`. If no native owner exists, do not add it unless it is a narrow Go typed bridge for an existing native operation.
- **Constructor parity** - Formatter `New` functions mirror `new Intl.<Constructor>(locales, options)`: omitted options use ECMA-402 defaults, and zero-valued `Options{}` means an empty options object.
- **Typed bridges only** - Typed Go methods such as `FormatInt64` are bridges for JavaScript methods like `format(value)`; they must not introduce behavior that native Intl lacks.
- **Typed boundaries** - Locale inputs are `locale.Locale` or `language.Tag`; raw strings are parsed once at the boundary.

## Coding Rules

### Must Follow

- Go 1.26.2. Use modern stdlib features already present in the codebase: `slices`, `maps`, `for range N`, `sync.Map`, `log/slog`, and `testing.B.Loop()`.
- Follow Google Go Best Practices: <https://google.github.io/go-style/best-practices>.
- Follow Google Go Style Decisions: <https://google.github.io/go-style/decisions>.
- Keep interfaces small and consumer-owned; do not expose broad internal abstractions.
- Every exported symbol must map to an ECMA-402 constructor, method, option, resolved option, part record, range source, static `Intl` function, or a Go-only typed bridge for one of those items.
- Validate formatter options in constructors. After construction, method error behavior must match the corresponding ECMA-402 operation: invalid typed inputs return errors where JavaScript would throw `TypeError` or `RangeError`; ordinary formatting does not hide constructor failures.
- Wrap user-fixable errors with package context and a sentinel error so callers can use `errors.Is` / `errors.As`.
- Keep generated CLDR/runtime data out of hot-path file I/O. Runtime data lives in generated Go source under `internal/cldr`.
- Keep formatter supported locale lists generated from actual CLDR payload maps. Constructor `SupportedLocalesOf` methods must call `internal/localematcher.FilterLocales` with `cldr.NumberSupportedLocales()`, `cldr.DateSupportedLocales()`, or `plural.SupportedLocales()` instead of duplicating filtering or requested-locale dedupe loops.
- Keep root `SupportedValuesOf` values generated from CLDR/tz data or ECMA-402 sanctioned constants. Calendars must include `iso8601`; numbering systems must include the ECMA-402 simple digit table; do not add ad hoc runtime lists.
- Keep ECMA-402 digit rounding centralized in `internal/ecma402/numberformat.FormatNumericToString`; `numberformat` and `pluralrules` both feed it resolved digit options and consume the rounded numeric value.
- Keep constructor options aligned with the JavaScript single-options-object model. Variadic `Options` is only the Go bridge for omission versus one object; passing more than one options object must remain invalid.
- Keep NumberFormat unit identifiers exact and case-sensitive. `UnitIdentifier("METER")` must not silently become `"meter"`; native `Intl.NumberFormat` rejects non-canonical unit casing.
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
- No hand-written NumberFormat/DateTimeFormat supported-locale lists; update CLDR codegen and `supported.go` instead.
- No `strconv.ParseFloat`, `math.Log10`, or `math.Pow10` in NumberFormat decimal formatting, compact scaling, scientific notation, or percent math. Use `internal/decimal`.
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
- Generate FormatJS-derived fixtures with `tools/gen-fixtures-from-formatjs`; generated files live under `<package>/testdata/conformance/formatjs/`. The active generated gate covers statically reducible NumberFormat `format`, `formatToParts`, `formatRange`, and `formatRangeToParts`; DateTimeFormat `format`, `formatToParts`, `formatRange`, and `formatRangeToParts`; PluralRules `select` and `selectRange`; Locale `toString`, `maximize`, `minimize`, and `Intl.getCanonicalLocales`.
- Unextracted or partially extracted FormatJS sources must appear in root `.skip-list.json` with `source`, `category`, and `reason`.
- `.skip-list.json` is extraction audit only. A generated fixture that fails must be handled through `testdata/divergences.md` or `testdata/xfail.json`, never by removing it from testdata.
- Record accepted reference mismatches in the relevant divergence file and keep IDs auditable by `task conformance:verify`.
- Public formatters should have runnable `Example*` functions when adding meaningful new surface.

## Dependencies

| Dependency | Purpose |
|------------|---------|
| `golang.org/x/text` | BCP 47 parsing, `language.Tag`, and Unicode/CLDR building blocks |
| `github.com/cockroachdb/apd/v3` | Decimal math backend for Intl mathematical values and rounding |

Add runtime dependencies only when an active SPEC requires them.

## Error Handling

- Package sentinels live in package-level `errors.go` files, for example `ErrInvalidOption` and `ErrUnsupportedLocale`.
- Error text should teach the caller what to fix: option name, value, and locale when available.
- Use `fmt.Errorf("package: context: %w", ErrX)` style wrapping; never match errors by string.
- Root namespace helpers do not own formatter construction semantics. Constructor failures belong to the formatter packages and must not be hidden behind fallback output.

## Performance

- Benchmark budgets live in `*_perf_test.go` and `SPECS/71-benchmark.md`.
- Use `task bench:gate` before merging changes that touch formatter hot paths.
- Use `task bench` with `BASELINE=<file>` for benchstat comparisons.
- Use `task build:size` after CLDR data changes.

## Linting

`task lint` runs the pinned `bin/golangci-lint` version from `.golangci.version` and checks `go mod tidy` output. The config is `.golangci.yml`; notable enabled linters include `exhaustive`, `errorlint`, `err113`, `gocritic`, `gosec`, `misspell`, `noctx`, `prealloc`, and `revive`.

Nested tool modules are checked from their own module roots. Do not use root-module patterns such as `go test ./tools/gen-cldr/...`; run `(cd tools/gen-cldr && go test ./...)` and `(cd tools/gen-cldr && go vet ./...)` instead.

## CI

GitHub Actions runs on pushes to `main` and pull requests:

- `test`: `task deps`, then `task test`
- `lint`: `task deps`, then `task lint`
- `security`: installs `govulncheck`, then runs `govulncheck ./...`
- `data-verify`: regenerates CLDR data, runs `task data:verify`, then `task build:size`

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
| [go-best-practices](.agents/skills/go-best-practices/) | Applying Go API, naming, error, concurrency, and testing rules |
| [modernizing](.agents/skills/modernizing/) | Applying Go 1.20-1.26 language and stdlib idioms |
| [code-simplifying](.agents/skills/code-simplifying/) | Simplifying recently changed code without changing behavior |
| [code-deduplicating](.agents/skills/code-deduplicating/) | Extracting repeated patterns that appear three or more times |
| [library-code-optimizing](.agents/skills/library-code-optimizing/) | Removing dead code and improving internal quality behavior-preservingly |
| [library-code-modernizing](.agents/skills/library-code-modernizing/) | Running a folder-by-folder Go modernization pass |
| [library-code-simplifying](.agents/skills/library-code-simplifying/) | Simplifying internal library code folder by folder |
| [library-error-optimizing](.agents/skills/library-error-optimizing/) | Renaming, reorganizing, or deduplicating package errors |
| [library-panic-optimizing](.agents/skills/library-panic-optimizing/) | Replacing unnecessary panics with errors and `Must*` wrappers |
| [library-symbol-naming](.agents/skills/library-symbol-naming/) | Improving public and internal symbol names |
| [library-file-naming](.agents/skills/library-file-naming/) | Aligning Go source/test filenames with package idiom |
| [library-legacy-pruning](.agents/skills/library-legacy-pruning/) | Removing deprecated APIs and old compatibility shims |
| [library-test-covering](.agents/skills/library-test-covering/) | Raising package test coverage with production-grade tests |
| [golangci-linting](.agents/skills/golangci-linting/) | Configuring or fixing golangci-lint v2 findings |
| [taskfile-configuring](.agents/skills/taskfile-configuring/) | Editing `Taskfile.yml` tasks and dependencies |
| [dependency-selecting](.agents/skills/dependency-selecting/) | Choosing Go dependencies when a SPEC requires one |
| [research-analyzing](.agents/skills/research-analyzing/) | Structuring research over `.references/` before spec work |
| [research-writing](.agents/skills/research-writing/) | Writing research reports from reference projects |
| [tdd-analyzing](.agents/skills/tdd-analyzing/) | Planning TDD implementation scope from SPECS and source |
| [tdd-implementing](.agents/skills/tdd-implementing/) | Implementing features with red-green-refactor cycles |
