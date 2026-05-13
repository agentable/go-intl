# SPEC 61 — messageformat-go Integration Contract

> **Status:** Draft (2026-05-08)
> **Type:** Interface + Decision — defines the **single direction** of dependency between go-intl and messageformat-go.
> **Authority:** This spec is SSOT for the messageformat-go ↔ go-intl boundary, the per-function migration list, and the dependency-issue reporting flow. SPEC 60 owns the root `Intl` namespace itself; SPECS 10/20/30/40 own the per-formatter constructors that messageformat-go calls.

---

## Overview

go-intl 是 **底层** ECMA-402 原语库;`messageformat-go` 是 **上层** MessageFormat 2.0 引擎。本 SPEC 锁定:

1. **依赖方向**:`messageformat-go → go-intl`,**单向**。go-intl 不感知 messageformat-go。
2. **迁移面**:`messageformat-go/pkg/functions/` 的 9 个 formatter-related 内建 function 改写为 go-intl 适配器。
3. **共享类型**:`Locale` 单一共享;`OperandsRecord` 通过 internal 通道由 NumberFormat ↔ PluralRules 共享,**禁止** messageformat-go 直接构造。
4. **依赖问题反馈流程**:遇到 go-intl bug 时,messageformat-go 提交 `reports/messageformat-go.md`;本 SPEC §5 定义归属与字段。

> **Why**: SPECS/00 §6.1 已定 messageformat-go 是 go-intl 的主消费者;R07 §4 在代码层面完成了迁移面盘点。本 SPEC 把这两份证据固化为契约,防止反向依赖与重复实现。
> **Rejected**: 共享中间包 `intl-bridge`——多一层包没有解决任何耦合问题,反而引入第三个版本释放节奏。

---

## 1. Dependency Direction

### 1.1 强制契约

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

**必须遵守**:

1. messageformat-go **必须**通过 `import "github.com/agentable/go-intl"` 与/或 `numberformat`、`datetimeformat`、`pluralrules`、`locale` 子包消费 go-intl。
2. go-intl **禁止**直接或传递地 import `kaptinlin/messageformat-go` 任何路径。
3. CI 在 go-intl 仓库通过 `go list -deps ./...` 校验,出现 `messageformat-go` 直接 block(SPEC 60 §6 已写入 acceptance)。
4. messageformat-go 仓库的 CI **应当**(非 go-intl 仓库 acceptance,但建议 messageformat-go owner 加入)校验 `import` 仅限本 SPEC §1.2 表格中的 go-intl 公开 symbol 集合。

> **Why**: 单向依赖防止 messageformat-go 在 go-intl 中"埋钩子"——例如把 RichText 元素塞进 `Locale` 字段,这种耦合一旦发生回滚成本极高。
> **Rejected**: 双向 + 共享 context 包——会形成 `(messageformat-go, go-intl, context-pkg)` 三方释放绑定,Go module 升级阻力指数级增长。

### 1.2 共享类型表面

| 类型 | 归属 SPEC | messageformat-go 使用方式 |
|------|---------|--------------------------|
| `locale.Locale` | SPEC 10 | 直接持有(由 `MessageFunctionContext.Locales() []string` → `locale.Parse` 转换) |
| `numberformat.Options` | SPEC 20 | 通过 adapter 边界构造 typed options,**不**直接持有 `*NumberFormat` 字段 |
| `datetimeformat.Options` | SPEC 30 | 通过 adapter 边界构造 typed options,**不**直接持有 `*DateTimeFormat` 字段 |
| `pluralrules.Options` | SPEC 40 | 通过 typed options 构造;messageformat-go 在 adapter 边界把 ICU 字符串映射成 Go enum |
| `OperandsRecord` | SPEC 40(internal,SPEC 20 复用) | **不可见**——messageformat-go 通过 typed `pluralrules.SelectInt` / `SelectDecimal` 调用,go-intl 内部传递 OperandsRecord |
| `numberformat.Part` / `datetimeformat.Part` | SPEC 20 / 30 | 仅在 `formatToParts` 桥接路径使用;messageformat-go `MessageValue` 内部转换 |

**必须遵守**:

1. messageformat-go **禁止** 把 `numberformat.Options` 或 `*numberformat.NumberFormat` 持久化到自身导出 API 字段中——否则 go-intl option 字段演进会成为 messageformat-go 的 breaking change。
2. messageformat-go 与 go-intl **必须**仅通过 `Locale` 类型互相传递 locale 信息;**禁止**字符串 BCP 47 跨包传递。
3. `OperandsRecord` **禁止**作为 messageformat-go 公开 API;messageformat-go 通过 `pluralrules.PluralRules` typed select 方法间接获得 plural 选择结果。

> **Why**: option type 不进入 messageformat-go 公开签名,意味着 go-intl 1.x 内部加字段时 messageformat-go 不需联动发版。
> **Rejected**: messageformat-go 暴露 `func (f *NumberFunction) Options() numberformat.Options`——会硬绑两边版本。

---

## 2. defaultRichTextElements 归属

`defaultRichTextElements` 是 formatjs `IntlShape.defaultRichTextElements?: Record<string, FormatXMLElementFn>` 的概念,服务于 React/Vue 的 rich-text MessageFormat 渲染。

**必须遵守**:

1. **禁止** `defaultRichTextElements`、`RichTextElements`、`XMLElementFn` 类符号出现在 go-intl 任何包内(`intl/`、`locale/`、`numberformat/`、`datetimeformat/`、`pluralrules/`、`internal/*`)。
2. messageformat-go 自带 `MessageValue` 已支持富类型 fallback,rich-text 渲染由 messageformat-go 自行实现,**不**通过 go-intl 暴露。
3. 若 messageformat-go 提出"go-intl 提供 rich-text 钩子" feature request,**必须**先按本 SPEC §5 提报,**不得**绕过 SPEC 流程直接 PR。

> **Why**: rich-text 是 message-formatting 范畴(模板替换 + 元素重写),不属于 ECMA-402;接进 go-intl 等于把项目目标撑爆到 SPECS/00 §1.1 明令排除的"翻译系统"。
> **Rejected**: 在 `intl.Config` 加 `WithRichTextElements(...)`——SPEC 60 §5 已显式 forbid。

---

## 3. Per-Function Migration List

`messageformat-go/pkg/functions/` 当前 11 个 function 的迁移分类。**必须**按本表完成迁移,任何偏离 **必须**先修订本 SPEC。

| messageformat-go function | 当前实现 | 迁移类型 | 目标 go-intl API |
|---------------------------|----------|----------|-------------------|
| `:integer` (`number.go`,~70 LoC) | 自实现数字解析 + 格式化 | **改写为 adapter** | `numberformat.New(loc, numberformat.Options{FractionDigits: numberformat.MaximumFractionDigits(0)})` |
| `:number` (`number.go`,~280 LoC) | 自实现 ICU 桥 | **改写为 adapter** | `numberformat.New(loc, ...)` |
| `:currency` (`currency.go`,191 LoC) | 自实现 currency 表 | **改写为 adapter** | `numberformat.New(loc, numberformat.Options{Style: numberformat.CurrencyStyle, Currency: numberformat.CurrencyCode(code)})` |
| `:percent` (`percent.go`,118 LoC) | 自实现百分比 | **改写为 adapter** | `numberformat.New(loc, numberformat.Options{Style: numberformat.PercentStyle})` |
| `:unit` (`unit.go`,120 LoC) | 自实现 unit identifier 表 | **改写为 adapter** | `numberformat.New(loc, numberformat.Options{Style: numberformat.UnitStyle, Unit: numberformat.UnitIdentifier(id)})` |
| `:offset` (`offset.go`,134 LoC) | 数值偏移 + 委托 `:number` | **partial adapter** | 偏移在 messageformat-go 自行完成,数字格式化委托 `numberformat.New(...)` |
| `:date` (`datetime.go` 子集) | 自实现 LDML 48 dateFields | **改写为 adapter** | `datetimeformat.New(loc, datetimeformat.Options{DateStyle: ...})` |
| `:datetime` (`datetime.go`,324 LoC 主体) | 自实现 dateFields/timePrecision | **改写为 adapter** | `datetimeformat.New(loc, datetimeformat.Options{DateStyle: ..., TimeStyle: ...})` |
| `:time` (`datetime.go` 子集) | 自实现 timePrecision | **改写为 adapter** | `datetimeformat.New(loc, datetimeformat.Options{TimeStyle: ...})` |
| `:string` (`string.go`,71 LoC) | 字符串透传 | **不迁移** | 与 ECMA-402 无关 |
| `:math` (`math.go`,159 LoC) | 算术运算 | **不迁移** | 与 ECMA-402 无关 |

**Adapter 形态契约**(签名,非实现):

```go
// messageformat-go 侧 adapter
package functions

import (
    "github.com/agentable/go-intl/locale"
    "github.com/agentable/go-intl/numberformat"
)

func NumberFunction(ctx MessageFunctionContext, opts map[string]any, operand any) messagevalue.MessageValue {
    loc, err := locale.Parse(firstLocale(ctx.Locales()))
    if err != nil { ctx.OnError(err); return fallback(ctx, operand) }

    nfOpts, err := mapNumberOptions(opts) // LDML 48 → ECMA-402 命名映射
    if err != nil { ctx.OnError(err); return fallback(ctx, operand) }

    nf, err := numberformat.New(loc, nfOpts)
    if err != nil { ctx.OnError(err); return fallback(ctx, operand) }

    return wrapMessageValue(nf, operand)
}
```

**必须遵守**:

1. 迁移期 **禁止** 双轨并存——一旦某 function 切到 adapter,messageformat-go 自带的旧 ICU 桥 **必须**同 PR 删除,不留 dead code。
2. **禁止** 把 LDML 48 选项命名(`dateFields`、`timePrecision`)泄漏到 go-intl;映射在 `mapXxxOptions` adapter 层完成。
3. `:string` 与 `:math` **禁止**迁移——它们与 ECMA-402 无关。
4. messageformat-go 函数签名 **必须**保持 `MessageFunction` 形态(SPEC §1.1 已固化的 `(ctx, opts, operand) → MessageValue`),迁移仅替换内部实现。

> **Why**: 迁移完成后 messageformat-go `pkg/functions/` 总 LoC 应下降 60%+;每个 adapter ≤ 30 行,真正的"薄"。
> **Rejected**: messageformat-go 在公开 adapter 类型上持有 `*numberformat.NumberFormat` 字段以"省一次构造"——缓存属于 messageformat-go 内部执行计划或调用方代码,不能泄漏到跨库公开签名。

---

## 4. OperandsRecord 共享规则

`OperandsRecord{ N, I, F, T OperandValue; V, W, C, E int }` 是 ECMA-402 §6.1.1 的 plural operands(active scope SPEC 40 owns)。

**必须遵守**:

1. `OperandsRecord` **必须**位于 `internal/ecma402/pluralrules/operands.go`,**不**作为公开 API。
2. NumberFormat compact 路径 **必须**通过 go-intl 内部 operand builder 与 generated CLDR plural rule 消费 OperandsRecord;**禁止** messageformat-go 自行构造 OperandsRecord。
3. messageformat-go 通过 `pluralrules.PluralRules` typed select 方法间接获得 plural 选择结果;**禁止**通过反射或 unsafe 拿到 OperandsRecord 内部字段。
4. SPEC 40 owns operands 字段集与计算算法;若 messageformat-go 提出新字段需求(如 `e2`),**必须**先按依赖问题反馈流程(SPEC §5)走,**不得**直接 PR `internal/ecma402/pluralrules`。

> **Why**: OperandsRecord 是 ECMA-402 内部数据结构,暴露后 messageformat-go 等于拿到 spec 内部插槽;后续 spec 演进(如 LDML 加新 operand)会同时破坏两个项目。
> **Rejected**: `pluralrules.Operands(value any) OperandsRecord` 公开函数——满足"messageformat-go 想看 operand"需求,但代价是把 internal 类型晋升到公开 surface,违反 SPEC 60 §5.

---

## 5. Dependency Issue Reporting

messageformat-go 在使用 go-intl 时遇到的 bug、限制、未预期行为 **必须**写入 `reports/messageformat-go.md`(归属仓库见下表),**禁止**通过 fork、reimplement、silent skip 绕过。

### 5.1 归属规则

| 问题类别 | 归属仓库 | 文件 |
|---------|---------|------|
| go-intl bug 触发 messageformat-go 失败 | go-intl 仓库 | `reports/messageformat-go.md`(消费方视角) |
| messageformat-go adapter 实现 bug | messageformat-go 仓库 | 本仓库 issue,不进 reports/ |
| ECMA-402 spec 解释分歧 | go-intl 仓库 | `reports/messageformat-go.md` + 触发 SPEC 修订 |
| LDML 48 → ECMA-402 选项映射不完备 | messageformat-go 仓库 | adapter 自身 PR |

### 5.2 报告格式

`reports/messageformat-go.md` 每条 issue **必须**包含:

| 字段 | 内容 |
|------|------|
| dependency | `kaptinlin/messageformat-go`,版本号(semver) |
| go-intl version | 触发问题时的 go-intl 版本 |
| problem | 1 段问题描述 |
| trigger | 最小复现 input + options |
| expected | 期望输出(引 ECMA-402 spec 条款 + formatjs 行为) |
| actual | 实际输出 + 错误消息或 stack trace |
| workaround | 调用方临时绕过方案(若有,**禁止**实现到 messageformat-go 代码中) |
| upstream issue | go-intl issue URL(若已开) |

### 5.3 处置流程

1. messageformat-go owner 在 go-intl 仓库 `reports/messageformat-go.md` 追加条目并提 PR。
2. go-intl owner 评审:
   - 确认是 go-intl bug → 创建对应 SPEC 修订或直接 fix。
   - 确认是 messageformat-go 误用 → 报告条目转 `resolved-misuse` 状态,messageformat-go 侧修复 adapter。
3. 修复发布后,reports/ 条目移入 `resolved/` 子目录或标记 `status: resolved`,**禁止**删除(保留审计轨迹)。

> **Why**: 强制 reports/ 流程防止 messageformat-go 在 adapter 层"打补丁"——任何绕过都让两个项目的 ECMA-402 行为漂移,conformance 测试失效。
> **Rejected**: messageformat-go 在 adapter 内 reimplement go-intl 函数——这会让两个项目的 ECMA-402 行为分叉。

---

## 6. Forbidden

- **禁止** go-intl 直接或传递 import `messageformat-go`(任何路径)。
- **禁止** `defaultRichTextElements` 或 rich-text 钩子出现在 go-intl 任何包内。
- **禁止** messageformat-go 公开 API 持有 `numberformat.Options`、`datetimeformat.Options` 或 `pluralrules.Options` 字段。
- **禁止** messageformat-go 直接构造 `OperandsRecord`(internal 类型)。
- **禁止** messageformat-go 在 adapter 内 reimplement go-intl 已有功能,绕过 SPEC §5 反馈流程。
- **禁止** messageformat-go 把字符串 BCP 47 跨包传递给 go-intl(必须先 `locale.New` 转 `Locale`)。
- **禁止** 迁移期同一 function 双轨并存(adapter + 旧实现共存)。
- **禁止** `:string` / `:math` function 进入迁移列表——它们与 ECMA-402 无关。
- **禁止** 将 LDML 48 命名(`dateFields`、`timePrecision`)泄漏到 go-intl 公开 API。
- **禁止** silent skip / silent fallback 替代 reports/ 提报。

---

## 7. Acceptance Criteria

### 依赖方向

- [ ] `go list -deps github.com/agentable/go-intl/...` 输出不含 `messageformat-go`。
- [ ] `messageformat-go` 仓库 CI 校验 `import` 仅限本 SPEC §1.2 表格中的公开 symbol。
- [ ] `contract_integration_test.go` 校验 `go list -deps` 不含 `messageformat-go`。

### 迁移完成度

- [ ] `messageformat-go/pkg/functions/{number,datetime,currency,unit,percent,offset}.go` 改写为 adapter(每文件 ≤ 100 LoC)。
- [ ] `:integer`、`:number`、`:currency`、`:percent`、`:unit`、`:offset`、`:date`、`:datetime`、`:time` 各自的 adapter 在 messageformat-go 测试集中保持 100% 通过率。
- [ ] `:string`、`:math` 保持原实现,**未**修改。
- [ ] messageformat-go `pkg/functions/` 总 LoC 较迁移前下降 ≥ 60%。

### 共享类型表面

- [ ] `OperandsRecord` 不在 go-intl `go doc github.com/agentable/go-intl/...` 输出中可见。
- [ ] `defaultRichTextElements` / `RichTextElements` / `XMLElementFn` 在 go-intl `git grep` 中无任何匹配。

### Dependency Issue Reporting

- [ ] `reports/messageformat-go.md` 存在,可初始无 open issue。
- [ ] `reports/messageformat-go.md` 模板提交 PR 时,SPEC §5.2 表格字段全部存在。

---

## References

- SPECS/00 §6.1(messageformat-go consumer contract)
- SPEC 60(root `Intl` namespace and forbidden messageformat-rich-text boundary)
- SPEC 40(PluralRules,OperandsRecord owner)
- `.research/R07-facade-and-caching.md` §4(迁移面盘点)
- `.references/formatjs/packages/intl/types.ts:65-78`(`defaultRichTextElements` 来源)
- messageformat-go `pkg/functions/types.go`(`MessageFunction` 签名)
- messageformat-go `pkg/functions/registry.go`(`DefaultFunctions` / `DraftFunctions` 表)
- messageformat-go `pkg/functions/number.go` / `datetime.go` / `currency.go` / `percent.go` / `unit.go` / `offset.go`(迁移源)
