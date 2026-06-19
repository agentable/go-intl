# SPEC 61 — messageformat-go Integration Contract

> **Status:** Draft (2026-05-08)
> **Type:** Interface + Decision — defines the **single direction** of dependency between go-intl and messageformat-go.
> **Authority:** This spec records the messageformat-go ↔ go-intl boundary, the per-function migration list, and the dependency-issue reporting flow. SPEC 60 records the root `Intl` namespace itself; SPECS 10/20/30/40 record the per-formatter constructors that messageformat-go calls.

---

## Overview

go-intl is the **lower** ECMA-402 primitive library; `messageformat-go` is the **upper** MessageFormat 2.0 engine. This SPEC locks:

1. **Dependence direction**: `messageformat-go → go-intl`, **One-way**. go-intl is not aware of messageformat-go.
2. **Migration surface**: 9 formatter-related built-in functions of `messageformat-go/pkg/functions/` are rewritten as go-intl adapters.
3. **Shared type**: `Locale` is single shared; `OperandsRecord` is shared by NumberFormat ↔ PluralRules through the internal channel, **FORBIDDEN** messageformat-go direct construction.
4. **Dependency problem feedback process**: When encountering go-intl bug, messageformat-go submits `reports/messageformat-go.md`; this SPEC §5 defines ownership and fields.

> **Why**: SPECS/00 §6.1 identifies messageformat-go as the main consumer of go-intl, and the current codebase already exposes the migration surface. This SPEC turns that relationship into a contract to prevent reverse dependencies and duplicate implementations.
> **Rejected**: Shared intermediate package `intl-bridge` - One more layer of packages does not solve any coupling problems, but instead introduces a third version release rhythm.

---

## 1. Dependency Direction

### 1.1 Mandatory contract

```text
              ┌──────────────────────┐
              │  messageformat-go    │
              │  pkg/functions/*.go  │
              └─────────┬────────────┘
                        │ import
                        ▼
              ┌──────────────────────┐
              │  github.com/agentable │
              │      /go-intl         │
              └──────────────────────┘
```

**Rules**:

1. messageformat-go **MUST** consume go-intl through `import "github.com/agentable/go-intl"` and/or `numberformat`, `datetimeformat`, `pluralrules`, `locale` sub-packages.
2. go-intl **forbids** to import `kaptinlin/messageformat-go` directly or transitively into any path.
3. CI passed the `go list -deps ./...` verification in the go-intl warehouse, and `messageformat-go` appeared as a direct block (SPEC 60 §6 has been written as acceptance).
4. The CI of the messageformat-go warehouse should** (not accepted by the go-intl warehouse, but it is recommended that the messageformat-go owner joins it) check `import` is limited to the go-intl public symbol collection in this SPEC §1.2 table.

> **Why**: One-way dependency prevents messageformat-go from "burying hooks" in go-intl - such as inserting RichText elements into the `Locale` field. Once this coupling occurs, the cost of rollback is extremely high.
> **Rejected**: Two-way + shared context package - will form `(messageformat-go, go-intl, context-pkg)` three-party release binding, and the resistance to Go module upgrade will increase exponentially.

### 1.2 Shared type surface

| Type | Attribution to SPEC | messageformat-go Usage |
|------|---------|--------------------------|
| `locale.Locale` | SPEC 10 | Directly held (converted from `MessageFunctionContext.Locales() []string` → `locale.Parse`) |
| `numberformat.Options` | SPEC 20 | Constructed through adapter boundary typed options, **not** directly hold the `*NumberFormat` field |
| `datetimeformat.Options` | SPEC 30 | Constructed through adapter boundary typed options, **not** directly hold the `*DateTimeFormat` field |
| `pluralrules.Options` | SPEC 40 | Constructed through typed options; messageformat-go maps ICU strings to Go enum at adapter boundaries |
| `OperandsRecord` | SPEC 40 (internal, SPEC 20 reuse) | **Invisible**——messageformat-go is called through `pluralrules.Value` + `Select`, go-intl internally passes OperandsRecord |
| `numberformat.Part` / `datetimeformat.Part` | SPEC 20 / 30 | Only used in `formatToParts` bridge path; messageformat-go `MessageValue` internal conversion |

**Rules**:

1. messageformat-go **forbids** to persist `numberformat.Options` or `*numberformat.NumberFormat` into its own exported API fields - otherwise go-intl option field evolution will become a breaking change of messageformat-go.
2. messageformat-go and go-intl **MUST** only pass locale information to each other through the `Locale` type; **it is prohibited** to pass string BCP 47 across packages.
3. `OperandsRecord` is **FORBIDDEN** as messageformat-go exposes the API; messageformat-go indirectly obtains the plural selection result through the `Select` method of `pluralrules.PluralRules`.

> **Why**: option type does not enter the public signature of messageformat-go, which means that messageformat-go does not need to be released in conjunction with go-intl 1.x when adding fields internally.
> **Rejected**: messageformat-go exposes `func (f *NumberFunction) Options() numberformat.Options` - will hard-bind both versions.

---

## 2. defaultRichTextElements ownership

`defaultRichTextElements` is the concept of generated-reference `IntlShape.defaultRichTextElements?: Record<string, FormatXMLElementFn>`, serving the rich-text MessageFormat rendering of React/Vue.

**Rules**:

1. **BANNED** `defaultRichTextElements`, `RichTextElements`, `XMLElementFn` class symbols appear in any go-intl package (`intl/`, `locale/`, `numberformat/`, `datetimeformat/`, `pluralrules/`, `internal/*`).
2. messageformat-go comes with `MessageValue`, which already supports rich type fallback. Rich-text rendering is implemented by messageformat-go itself and is not exposed through go-intl.
3. If messageformat-go proposes a "go-intl provides rich-text hook" feature request, it must be reported in accordance with SPEC §5 first, and it must not bypass the SPEC process and PR directly.

> **Why**: rich-text is in the message-formatting category (template replacement + element rewriting) and does not belong to ECMA-402; connecting to go-intl is equivalent to pushing the project goal to a "translation system" that is explicitly excluded by SPECS/00 §1.1.
> **Rejected**: Add `WithRichTextElements(...)` to `intl.Config` - SPEC 60 §5 explicitly forbid.

---

## 3. Per-Function Migration List

`messageformat-go/pkg/functions/` The migration classification of the current 11 functions. Migration **MUST** be completed according to this table, any deviation **MUST** be revised to this SPEC first.

| messageformat-go function | current implementation | migration type | target go-intl API |
|---------------------------|----------|----------|-------------------|
| `:integer` (`number.go`,~70 LoC) | Self-implemented number parsing + formatting | **Rewritten as adapter** | `numberformat.New(locale.List{loc}, numberformat.Options{MaximumFractionDigits: gointl.Int(0)})` |
| `:number` (`number.go`,~280 LoC) | Self-implemented ICU bridge | **Rewritten as adapter** | `numberformat.New(locale.List{loc}, ...)` |
| `:currency` (`currency.go`,191 LoC) | Self-implemented currency table | **Rewritten as adapter** | `numberformat.New(locale.List{loc}, numberformat.Options{Style: numberformat.CurrencyStyle, Currency: numberformat.Currency(code)})` |
| `:percent` (`percent.go`,118 LoC) | Self-implementation percentage | **Rewritten as adapter** | `numberformat.New(locale.List{loc}, numberformat.Options{Style: numberformat.PercentStyle})` |
| `:unit` (`unit.go`,120 LoC) | Self-implemented unit identifier table | **Rewritten as adapter** | `numberformat.New(locale.List{loc}, numberformat.Options{Style: numberformat.UnitStyle, Unit: numberformat.Unit(id)})` |
| `:offset` (`offset.go`,134 LoC) | Numeric offset + delegate `:number` | **partial adapter** | Offset is done by itself in messageformat-go, number format delegate `numberformat.New(...)` |
| `:date` (`datetime.go` subset) | Self-implemented LDML 48 dateFields | **Rewritten as adapter** | `datetimeformat.New(locale.List{loc}, datetimeformat.Options{DateStyle: ...})` |
| `:datetime` (`datetime.go`,324 LoC body) | Self-implemented dateFields/timePrecision | **Rewritten as adapter** | `datetimeformat.New(locale.List{loc}, datetimeformat.Options{DateStyle: ..., TimeStyle: ...})` |
| `:time` (`datetime.go` subset) | Self-implemented timePrecision | **Rewritten as adapter** | `datetimeformat.New(locale.List{loc}, datetimeformat.Options{TimeStyle: ...})` |
| `:string` (`string.go`,71 LoC) | String transparent transmission | **Not migrated** | Not related to ECMA-402 |
| `:math` (`math.go`,159 LoC) | Arithmetic operations | **Not migrated** | Not related to ECMA-402 |

**Adapter form contract** (signature, non-implementation):

```go
// messageformat-go side adapter
package functions

import (
    "github.com/agentable/go-intl/locale"
    "github.com/agentable/go-intl/numberformat"
)

func NumberFunction(ctx MessageFunctionContext, opts map[string]any, operand any) messagevalue.MessageValue {
    loc, err := locale.Parse(firstLocale(ctx.Locales()))
    if err != nil { ctx.OnError(err); return fallback(ctx, operand) }

nfOpts, err := mapNumberOptions(opts) // LDML 48 → ECMA-402 named mapping
    if err != nil { ctx.OnError(err); return fallback(ctx, operand) }

    nf, err := numberformat.New(locale.List{loc}, nfOpts)
    if err != nil { ctx.OnError(err); return fallback(ctx, operand) }

    return wrapMessageValue(nf, operand)
}
```

**Rules**:

1. Dual-track coexistence is **prohibited** during the migration period - once a function is switched to the adapter, the old ICU bridge that comes with messageformat-go **MUST** be deleted with the PR, without leaving dead code.
2. **FORBIDDEN** Leak LDML 48 option names (`dateFields`, `timePrecision`) into go-intl; mapping is done at the `mapXxxOptions` adapter layer.
3. `:string` and `:math` are **disabled** from migration - they have nothing to do with ECMA-402.
4. The messageformat-go function signature **MUST** maintain the `MessageFunction` form (SPEC §1.1 solidified `(ctx, opts, operand) → MessageValue`), and migration only replaces the internal implementation.

> **Why**: After the migration is complete, the total LoC of messageformat-go `pkg/functions/` should drop by 60%+; each adapter is ≤ 30 lines, truly "thin".
> **Rejected**: messageformat-go holds the `*numberformat.NumberFormat` field on the public adapter type to "save once construction" - the cache belongs to messageformat-go's internal execution plan or caller code and cannot be leaked to cross-library public signatures.

---

## 4. OperandsRecord sharing rules

`OperandsRecord{ N, I, F, T OperandValue; V, W, C, E int }` is the plural operands (active scope SPEC 40 owns) of ECMA-402 §6.1.1.

**Rules**:

1. `OperandsRecord` **MUST** be located in `internal/ecma402/pluralrules/operands.go`, **not** as a public API.
2. NumberFormat compact path **MUST** consume OperandsRecord through go-intl internal operand builder and generated CLDR plural rule; **FORBIDDEN** messageformat-go to construct OperandsRecord by itself.
3. messageformat-go indirectly obtains the plural selection result through `pluralrules.PluralRules.Select`; it is **FORBIDDEN** to obtain the internal fields of OperandsRecord through reflection or unsafe.
4. SPEC 40 owns operands field set and calculation algorithm; if messageformat-go proposes new field requirements (such as `e2`), you must first follow the dependency issue feedback process (SPEC §5), and you must not PR `internal/ecma402/pluralrules` directly.

> **Why**: OperandsRecord is an ECMA-402 internal data structure. After exposure, messageformat-go is equivalent to getting the spec internal slot; subsequent spec evolution (such as adding a new operand to LDML) will destroy both projects at the same time.
> **Rejected**: `pluralrules.Operands(value any) OperandsRecord` public function - meets the requirement of "messageformat-go wants to see operand", but at the cost of promoting the internal type to a public surface, which violates SPEC 60 §5.

---

## 5. Dependency Issue Reporting

The bugs, limitations, and unexpected behaviors encountered by messageformat-go when using go-intl **MUST** be written to `reports/messageformat-go.md` (see the table below for the ownership warehouse), and are **FORBIDDEN** to be bypassed through fork, reimplement, and silent skip.

### 5.1 Attribution rules

| Problem category | Attribution warehouse | File |
|---------|---------|------|
| go-intl bug triggers messageformat-go failure | go-intl warehouse | `reports/messageformat-go.md` (consumer perspective) |
| messageformat-go adapter implementation bug | messageformat-go warehouse | This warehouse issue, not included reports/ |
| ECMA-402 spec interpretation divergence | go-intl repository | `reports/messageformat-go.md` + trigger SPEC revision |
| LDML 48 → ECMA-402 Incomplete option mapping | messageformat-go repository | adapter self PR |

### 5.2 Report format

`reports/messageformat-go.md` Each issue **MUST** contain:

| Field | Content |
|------|------|
| dependency | `kaptinlin/messageformat-go`, version number (semver) |
| go-intl version | The go-intl version when the problem was triggered |
| problem | 1 paragraph problem description |
| trigger | minimum recurrence input + options |
| expected | expected output (quoting ECMA-402 spec clause + reference behavior) |
| actual | actual output + error message or stack trace |
| workaround | Temporary bypass solution for the caller (if any, it is **prohibited** to be implemented into the messageformat-go code) |
| upstream issue | go-intl issue URL (if already opened) |

### 5.3 Handling Process

1. messageformat-go owner adds an entry to the go-intl warehouse `reports/messageformat-go.md` and submits a PR.
2. go-intl owner review:
- Confirmed to be a go-intl bug → Create corresponding SPEC revision or fix directly.
- Confirmed that messageformat-go is misused → report the entry to `resolved-misuse` status, messageformat-go side repair adapter.
3. After the fix is released, the reports/ entries are moved to the `resolved/` subdirectory or marked `status: resolved`, and are **FORBIDDEN** to be deleted (to retain audit trails).

> **Why**: Forcing the reports/ process prevents messageformat-go from "patching" the adapter layer - any bypass will cause the ECMA-402 behavior of both projects to drift and invalidate conformance tests.
> **Rejected**: messageformat-go reimplement go-intl function inside adapter - this will bifurcate the ECMA-402 behavior of both projects.

---

## 6. Forbidden

- **BANNED** go-intl directly or pass import `messageformat-go` (any path).
- **FORBIDDEN** `defaultRichTextElements` or rich-text hooks appear inside any go-intl package.
- **BANNED** The messageformat-go public API holds a `numberformat.Options`, `datetimeformat.Options` or `pluralrules.Options` field.
- **FORBIDDEN** messageformat-go constructs `OperandsRecord` (internal type) directly.
- **BANNED** messageformat-go has the function of reimplementing go-intl in the adapter, bypassing the SPEC §5 feedback process.
- **BANNED** messageformat-go passes the string BCP 47 across packages to go-intl (must convert `locale.New` to `Locale` first).
- **FORBIDDEN** Dual-track coexistence of the same function during the migration period (adapter + old implementation coexist).
- **BANNED** `:string` / `:math` functions enter the migration list - they are not related to ECMA-402.
- **BANNED** Leaking LDML 48 names (`dateFields`, `timePrecision`) into the go-intl public API.
- **BANNED** silent skip / silent fallback replace reports/ report.

---

## 7. Acceptance Criteria

### Dependence direction

- [ ] `go list -deps github.com/agentable/go-intl/...` output does not contain `messageformat-go`.
- [ ] `messageformat-go` Warehouse CI check `import` Only public symbols in this SPEC §1.2 table.
- [ ] `contract_integration_test.go` checks `go list -deps` without `messageformat-go`.

### Migration completion

- [ ] `messageformat-go/pkg/functions/{number,datetime,currency,unit,percent,offset}.go` rewritten as adapter (per file ≤ 100 LoC).
- [ ] The respective adapters of `:integer`, `:number`, `:currency`, `:percent`, `:unit`, `:offset`, `:date`, `:datetime`, `:time` maintain 100% pass rate in the messageformat-go test set.
- [ ] `:string`, `:math` maintain the original implementation and have not been modified.
- [ ] messageformat-go `pkg/functions/` The total LoC dropped by ≥ 60% compared to before migration.

### Shared type surface

- [ ] `OperandsRecord` is not visible in go-intl `go doc github.com/agentable/go-intl/...` output.
- [ ] `defaultRichTextElements` / `RichTextElements` / `XMLElementFn` has no match in go-intl `git grep`.

### Dependency Issue Reporting

- [ ] `reports/messageformat-go.md` exists, but there is no open issue initially.
- [ ] When submitting PR for `reports/messageformat-go.md` template, all SPEC §5.2 table fields exist.

---

## References

- SPECS/00 §6.1(messageformat-go consumer contract)
- SPEC 60(root `Intl` namespace and forbidden messageformat-rich-text boundary)
- SPEC 40(PluralRules,OperandsRecord owner)
- `.references/formatjs/packages/intl/types.ts:65-78`(`defaultRichTextElements` source)
- messageformat-go `pkg/functions/types.go`(`MessageFunction` signature)
- messageformat-go `pkg/functions/registry.go`(`DefaultFunctions` / `DraftFunctions` table)
- messageformat-go `pkg/functions/number.go` / `datetime.go` / `currency.go` / `percent.go` / `unit.go` / `offset.go` (migration source)
