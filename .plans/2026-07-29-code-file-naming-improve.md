# Code File Naming Improvement

## Scope And Conventions

- Baseline: `bd0724a65eb8b236c12be6dbfae72f09f7e351e3`, equal to
  `origin/main` at audit start.
- Scope: all 520 tracked Go paths across the root module and three nested tool
  modules. The current worktree contains unrelated user edits and deletions;
  none is part of this pass.
- Form: lowercase ASCII snake_case, with package and directory context omitted.
  A second subject segment is used only for real in-package disambiguation.
- Reserved inventory: 8 `main.go`, 33 `doc.go`, 11 `example_test.go`, and 14
  benchmark files retain their Go/tool semantics.
- Generated inventory: 18 files marked `Code generated ... DO NOT EDIT` retain
  generator-owned paths. The documented CLDR `data.go`, `decode.go`, and
  `accessors.go` roles remain stable.
- Baseline `task verify` passed in an isolated detached worktree with
  `GOTOOLCHAIN=go1.26.5` and `GOWORK=off`.

## Accepted Renames

| From | To | Primary subject | Context removed | Rule/convention | Related paths | Required reference updates |
|---|---|---|---|---|---|---|
| `internal/ecma402/constants.go` | `internal/ecma402/unit_identifiers.go` | Sanctioned unit-identifier accessors | Generic and now-false `constants` bucket | Name the current declaration family; history moved the literal table to `internal/unitid` | `internal/ecma402/identifier.go`, `internal/ecma402/identifier_test.go`, root `supported.go` | Update the package layout in `SPECS/12-abstract-operations.md` |
| `internal/ecma402/types.go` | `internal/ecma402/pattern.go` | `Part` and `Pattern` contracts | Generic `types` bucket and retired `MathematicalValue` ownership | Name the single remaining contract subject | `internal/ecma402/partition.go`, `internal/ecma402/partition_test.go` | Update the package layout, rule, and acceptance path in `SPECS/12-abstract-operations.md` |
| `numberformat/test_helpers_test.go` | `numberformat/decimal_fixture_test.go` | `mustDecimalValue` test fixture | Generic `test_helpers` bucket | Name the single fixture subject | `numberformat/format_test.go`, `numberformat/range_test.go` | None |
| `listformat/format_invariant_test.go` | `listformat/template_test.go` | List-template compilation invariants | Broad `format_invariant` label | Name the tested `compileListTemplate` subject | `listformat/format.go` | None |
| `listformat/format_invariant2_test.go` | `listformat/format_parts_invariant_test.go` | `Format`/`FormatToParts` concatenation invariant | Arbitrary collision suffix `2` | Use the established `format_parts_invariant_test.go` suite name | `durationformat/format_parts_invariant_test.go`, `listformat/format.go` | None |
| `locale/must_test.go` | `locale/parse_fixture_test.go` | `parseLocaleForTest` fixture | Stale `must` label after helper contraction | Name the current single fixture subject | Locale package tests and benchmark | None |

## Explicit Keeps

- `intl.go` and each constructor-named file such as `numberformat.go` name a
  documented public ECMA-402 owner; their package repetition is intentional.
- `conformance_unified_test.go` is the documented cross-source fixture harness.
- `bcp47.go`, `log10.go`, and `ucanonicalize.go` use stable standards or
  ECMA-402 operation vocabulary rather than ordinary word-number suffixes.
- `internal/decimal/from.go` and `ops.go` form exact source/test pairs and name
  cohesive construction and arithmetic concerns.
- CLDR generator files such as `domain.go`, `payload.go`, and `render.go` name
  cohesive generator stages and are not workflow prefixes.
- Focused contract and regression suites may keep distinct owners instead of
  being forced into source basename pairing.

## Routed Structural Work

- `datetimeformat/internal_edges_test.go` contains pattern-literal, style,
  interval/time-zone, small-formatting, and range-fallback subjects. A future
  test-organization pass should split those tests among their owning source
  concerns; no umbrella rename is truthful today.
- `locale/construct_test.go` spans constructor options, Unicode aliases,
  language-tag bridges, equality, and text marshaling. A future split should
  separate those owners before changing its path.
- Broad shared files such as `internal/ecma402/options.go` and
  `tools/gen-cldr/cldr/numbers.go` remain cohesive domain owners. If their
  declarations diverge later, route a split rather than inventing a new bucket.

## Validation

- Focused race tests passed for `internal/ecma402`, `numberformat`,
  `listformat`, and `locale` against the exact staged tree.
- `task verify` passed from an independent staged-tree checkout with isolated
  Go and golangci-lint caches: lint reported zero issues, race tests passed,
  all 370 conformance fixtures passed, data contracts passed, and
  `govulncheck` found no vulnerabilities.
- Repository-wide tests with coverage passed at 79.5% total statement
  coverage. `task bench:run` executed every benchmark successfully with race
  detection and allocation reporting.
- Tests with race detection and `go vet` passed in `tools/gen-cldr`,
  `tools/gen-plural-rules`, and `tools/gen-fixtures-from-formatjs`, using the
  repository's pinned CLDR and tzdata inputs.
- All root-module packages built for Linux amd64 and Windows amd64 with CGO
  disabled.
- Real local consumers passed against the staged module through a temporary Go
  workspace with module writes disabled: `go-time`, `messageformat-go`,
  `messageformat-go/mf1`, and `go-typescript`.
- Stale-path searches returned no matches, `git diff --cached --check` passed,
  all six moves remained `R100`, generated paths were untouched, and unrelated
  user changes remained unstaged.
