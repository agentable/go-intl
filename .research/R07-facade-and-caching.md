---
id: R07
title: 顶层 façade、缓存、错误传播与 messageformat-go 集成边界
task: r07
date: 2026-05-08
status: draft
scope:
  - createIntl/IntlShape 在 Go 中的等价物
  - formatter 缓存 key 设计与淘汰策略
  - onError 回调 vs Go error 返回的双轨调和
  - messageformat-go 的迁移边界与共享类型
  - node_i18n.cc host binding 的 Phase 1/4 取舍
  - defaultRichTextElements 在 Go 端的去留
tags: [facade, cache, error-model, messageformat-go, host-binding]
---

# R07 — 顶层 façade、缓存、错误传播与 messageformat-go 集成边界

## 执行摘要

| 决策点 | 推荐 | 置信度 | 主要依据 |
|-------|------|--------|---------|
| Façade 形态 | 提供 `intl.New(IntlConfig) *Intl` 持久对象 + 顶层 `intl.FormatNumber(loc, v, opts...)` 一行 helper 双轨 | 高 | `formatjs/packages/intl/create-intl.ts:57-162` 的 `createIntl` 即此模式；SPEC §3 已要求双层 |
| 缓存 key 设计 | `string` 拼接（locale + sorted JSON 类似的 canonical options），用 `sync.Map` 持有 | 高 | `formatjs` 借 `@formatjs/fast-memoize` 的 `strategies.variadic` 走相同字符串 key 路径 (`utils.ts:127-134`)；Go 端 `language.Tag.String()` 已经规范化 |
| 缓存淘汰 | 默认无淘汰 + 显式 `IntlCache` 接口允许调用方注入 LRU；进程级别允许 `Reset()` | 中 | `formatjs` 同样默认无淘汰但暴露 `IntlCache` 由调用方控制 (`utils.ts:81-91`)，服务端用例需要可注入 |
| 错误模型 | 构造期返回 `error`；运行期 `Format`/`FormatToParts` 不返回错误，遵循 ECMA-402 fallback；façade 可接 `WithOnError(func(error))` 钩子 | 高 | `formatjs/packages/intl/error.ts:11-110` 定义的 `IntlError` 家族 + `onError` 回调；ECMA-402 规定运行期不抛 |
| messageformat-go 边界 | go-intl 为下游被依赖，messageformat-go 改造为薄 adapter；通过 `locale.Locale` + `numberformat.Options` 等显式类型对接 | 高 | `messageformat-go/pkg/functions/{number,datetime,currency,unit,percent,offset}.go` 当前以 `map[string]any` + 直接 ICU 调用实现，迁移面与 SPEC §6.1 一致 |
| host-binding | Phase 1 不交付；保留 `intl.New` 出口结构使 Phase 4 可在 `go-typescript` 中桥接 `globalThis.Intl` | 高 | `node/src/node_i18n.{cc,h}` 是 ICU 数据加载/Buffer 转码胶水（733+132 行），与 ECMA-402 对象暴露解耦 |
| defaultRichTextElements | 不在 go-intl 中保留，归 messageformat-go 自行支持（其 `MessageValue` 已有富类型） | 高 | formatjs `intl/types.ts:75` 的 `defaultRichTextElements` 仅服务于 `formatMessage`；go-intl 不承担消息格式化 |

---

## 1. createIntl/IntlShape 模型与 Go 等价物

### 1.1 formatjs 的 façade 形态

`createIntl(config, cache?)` 返回一个 `IntlShape`：把已解析的 `ResolvedIntlConfig` 与一组以 `bind(null, config, formatters.getXxx)` 部分应用过的 `formatXxx` 方法拼到同一个对象上 <!-- ref: formatjs/packages/intl/create-intl.ts:57-162 -->。`IntlShape extends ResolvedIntlConfig & IntlFormatters` <!-- ref: formatjs/packages/intl/types.ts:257-260 -->，约束是「config 与 formatter 不可分离」——函数体看到的 `locale` / `timeZone` / `formats` / `onError` 来自构造时冻结的 config。

每个 `formatXxx` 自身只是一层薄包装：把允许选项 `filterProps` 出来、再交给缓存 getter（`getNumberFormat`、`getDateTimeFormat` 等）取/造一个原生 `Intl.NumberFormat` 实例做实际格式化 <!-- ref: formatjs/packages/intl/number.ts:41-66, formatjs/packages/intl/dateTime.ts:35-76 -->。

### 1.2 PHP 的对应

PHP `ext/intl` 没有跨 formatter 的 façade，每个 ICU 类各自独立。这意味着 façade 是 FormatJS 在 polyfill 上层为 React/Vue 等使用方加的便利层，而不是 ECMA-402 规范成员。

### 1.3 Go 端的双轨建议

SPEC §3 已经要求两个入口（根包 `intl.FormatNumber/...` + 子包 `numberformat.New`）。基于 formatjs 的实证经验，根包应该再补一个「持久 façade」类型 `intl.Intl`，承担两个职责：

- 锁定一组配置（locale、timeZone、formats、onError）；
- 持有 formatter 缓存以共享一次解析的 ECMA-402 选项。

```go
// 根包入口签名
package intl

type Intl struct { /* 持久 IntlShape 等价物 */ }

func New(opts ...Option) (*Intl, error)            // createIntl 等价
func (i *Intl) FormatNumber(v any, opts ...numberformat.Option) string
func (i *Intl) FormatDate(t time.Time, opts ...datetimeformat.Option) string
func (i *Intl) ResolvedLocale() locale.Locale

// 一行 helper：内部走全局 LRU 或 sync.Map 单例
func FormatNumber(loc locale.Locale, v any, opts ...numberformat.Option) string
func FormatDate(loc locale.Locale, t time.Time, opts ...datetimeformat.Option) string
```

调用示例（3-5 行，文档导出用）：

```go
fr, _ := intl.New(intl.WithLocale(locale.MustParse("fr-FR")))
out := fr.FormatNumber(1234.5, numberformat.WithStyle(numberformat.StyleCurrency), numberformat.WithCurrency("EUR"))
// "1 234,50 €"
```

---

## 2. Formatter 缓存：key、容器与淘汰

### 2.1 formatjs 的缓存机制

缓存通过 `@formatjs/fast-memoize` 的 `strategies.variadic` 实现：把所有参数序列化为字符串作 key，命中则复用，否则走 `new Intl.NumberFormat(...)` 等构造 <!-- ref: formatjs/packages/intl/utils.ts:114-170 -->。`IntlCache` 只是 `Record<string, FormatterInstance>` 的命名空间集合，按 formatter 类型分桶（`dateTime`、`number`、`message`、`relativeTime`、`pluralRules`、`list`、`displayNames`）<!-- ref: formatjs/packages/intl/types.ts:262-270 -->。`createIntlCache()` 提供一个零值默认 cache，`createIntl` 接受外部 cache 注入以便服务端避免内存泄漏 <!-- ref: formatjs/packages/intl/utils.ts:81-91 -->。

关键事实：**formatjs 默认不淘汰**——单 cache 里塞多少 entry 就保留多少。它通过 React/Vue 集成层在组件卸载时丢弃整个 cache 对象来回收。`fast-memoize` 自带的 `variadic` 策略只是 string-key 哈希，无 LRU。

### 2.2 cache key 维度对比

| 维度 | formatjs | V8 (隐式) | translate-agent/intl | 推荐 |
|------|----------|-----------|---------------------|------|
| key 类型 | `string`（fast-memoize variadic） | 无（直接持有 ICU 对象） | 不缓存（每次解析 pattern） | `string` |
| 序列化方式 | `JSON.stringify`-like 拼接 | — | — | `locale.String() + "|" + canonicalOptionString` |
| 容器 | `Record<string, formatter>` | — | — | `sync.Map` |
| 淘汰 | 无（cache 生命周期等于调用方） | 无 | — | 默认无；可注入 `Cache` 接口 |
| 命名空间 | 按 formatter 类型 7 桶 | — | — | 按 formatter 类型 4 桶（Phase 1） |

### 2.3 Go 实现细节

**Key 生成**：locale 由 `locale.Locale.String()` 给出（`language.Tag` 已经做过 BCP47 规范化），options 部分需要在 `numberformat`、`datetimeformat`、`pluralrules` 各自 `Options.canonicalKey() string` 内部完成（按字段排序后写 buffer）。和 ResolvedOptions 不同，缓存 key 必须包含的是「输入选项」而非「resolved」，否则会出现两个不同输入但等价 resolved 的实例被错认。formatjs 把 cache 放在「输入参数序列化」也是同样原因 <!-- ref: formatjs/packages/intl/utils.ts:127-134 -->。

**容器选择**：`sync.Map` 适合 read-heavy、key-set 接近稳定的场景，正好对应 i18n 中「一旦预热，hit 率极高」的场景。`map+sync.RWMutex` 在写多场景更优，但 i18n 的写操作（构造新 formatter）只在冷启动出现。

**淘汰策略**：

- **推荐**：默认不淘汰。Phase 1 中可缓存的 formatter 总数 = `|locales| * |distinct option-set|`。对一个典型 SaaS 应用（3 语言 × 5 选项组合 × 4 formatter 种类 = 60 实例），整个 cache < 100KB，远低于淘汰带来的复杂度收益。
- **备选**：暴露 `Cache` 接口供调用方注入 LRU（例如 `hashicorp/golang-lru/v2`）。formatjs 的 `IntlCache` 同样是「接口而非实现」。
- **否决**：内置 LRU 默认值——在不知道调用方负载的情况下选中间值（如 1024）容易出现「单进程多租户场景，租户切换间隙就被淘汰」的反模式。

```go
package intl

type Cache interface {
    LoadOrStore(key string, fn func() any) any
    // 可选淘汰
    Reset()
}

type Option func(*Config)

func WithCache(c Cache) Option         // 注入自定义实现
func WithoutCache() Option              // 关闭缓存（每次新建）
```

`sync.Map` 包一层零依赖默认实现即可。

---

## 3. 错误模型：onError vs error

### 3.1 formatjs 的错误分层

`IntlErrorCode` 五类：`FORMAT_ERROR`、`UNSUPPORTED_FORMATTER`、`INVALID_CONFIG`、`MISSING_DATA`、`MISSING_TRANSLATION` <!-- ref: formatjs/packages/intl/error.ts:3-9 -->。每个 `formatXxx` 内部 `try/catch`，错误用 `IntlFormatError` 包裹后调用 `config.onError(err)`，并返回降级字符串（如 `String(value)` 或空数组）<!-- ref: formatjs/packages/intl/number.ts:79-88, dateTime.ts:90-99 -->。

这与 ECMA-402 规范一致：规范要求 `Intl.NumberFormat.prototype.format` 在合法 formatter 上不抛出（输入异常如 `BigInt` + 非整数选项除外）；formatjs 的 polyfill 把「数据缺失」「locale 不支持」也按相同契约降级到 fallback。

### 3.2 V8/Node 的错误形态

V8 实现里，构造期的 `RangeError`/`TypeError` 直接抛 JS 异常；运行期沿规范不抛。`node_i18n.cc` 不暴露 onError 钩子，错误经由 V8 异常对象传递。

### 3.3 messageformat-go 的混合模型

`MessageFunctionContext.OnError(err error)` 是回调式（与 formatjs 同款）；同时函数返回 `MessageValue`，错误时返回 `FallbackValue` <!-- ref: messageformat-go/pkg/functions/types.go:36-119, offset.go:55-91 -->。这表明 Go 生态接受「回调 + fallback 双轨」。

### 3.4 Go 端的调和方案

ECMA-402 的「构造期可错、运行期不错」契约在 Go 中能两边兼顾：

| 时机 | 错误形态 | 调用方姿势 |
|------|---------|-----------|
| `numberformat.New(loc, opts...)` | `(*NumberFormat, error)` | `if err != nil { ... }` |
| `(*NumberFormat).Format(v)` | `string`（不返回 error） | 直接拿结果；异常输入返回 ECMA-402 fallback |
| `intl.New(IntlConfig)` | `(*Intl, error)` | 同上 |
| `(*Intl).FormatNumber(v, opts...)` | `string`，options 错误经 `onError` 钩子上报 | `WithOnError` 注入观察者 |

`onError` 钩子保留是因为 façade 在「per-call 选项」上仍可能遇到不合法 option（例如 unknown currency）：构造期 `formatter` 已经合法，但 per-call options 是 façade 自行 merge 的，不该让其抛 panic 也不该改为 `(string, error)`（字节级模仿 formatjs）。

```go
package intl

type ErrorCode string

const (
    ErrCodeFormat              ErrorCode = "FORMAT_ERROR"
    ErrCodeUnsupportedFormatter ErrorCode = "UNSUPPORTED_FORMATTER"
    ErrCodeInvalidConfig       ErrorCode = "INVALID_CONFIG"
    ErrCodeMissingData         ErrorCode = "MISSING_DATA"
)

type Error struct {
    Code    ErrorCode
    Locale  string
    Message string
    Wrapped error
}

func (e *Error) Error() string { return /* "[go-intl FORMAT_ERROR] ..." */ "" }
func (e *Error) Unwrap() error { return e.Wrapped }

type Option func(*Config)
func WithOnError(fn func(error)) Option
```

子包则给出真正可返回的构造错误：

```go
package numberformat

var (
    ErrInvalidOption     = errors.New("numberformat: invalid option")
    ErrUnsupportedLocale = errors.New("numberformat: unsupported locale")
    ErrUnsupportedCurrency = errors.New("numberformat: unsupported currency")
)

func New(loc locale.Locale, opts ...Option) (*NumberFormat, error)
```

Sentinel 错误在 `errors.go`，wrap 时用 `fmt.Errorf("numberformat: invalid currency %q: %w", code, ErrInvalidOption)`，调用方用 `errors.Is` 判定。这与 CLAUDE.md 的错误处理规约（构造期可返回 error、运行期 format 不返回 error）一致。

---

## 4. messageformat-go 集成边界

### 4.1 现状盘点

`messageformat-go/pkg/functions/` 当前自带七个内建 function：`:integer`、`:number`、`:string`、`:offset`（DefaultFunctions）以及 `:currency`、`:date`、`:datetime`、`:math`、`:percent`、`:time`、`:unit`（DraftFunctions）<!-- ref: messageformat-go/pkg/functions/registry.go:28-69 -->。每个 function 都是同款形态：

```go
type MessageFunction func(
    ctx MessageFunctionContext,
    options map[string]any,
    operand any,
) messagevalue.MessageValue
```

<!-- ref: messageformat-go/pkg/functions/types.go:18-22 -->

实现细节：
- `number.go`（477 行）—— 自行解析数字操作数、merge 选项 map、最终走某个内部 ICU 桥；
- `datetime.go`（324 行）—— 自行解析 `dateFields`/`timePrecision` 等 LDML 48 选项；
- `currency.go`（191 行）、`unit.go`（120 行）、`percent.go`（118 行）—— 三者都把 LDML `:currency`/`:unit`/`:percent` 选项 map 后转 `Intl.NumberFormat` 等价调用；
- `offset.go`（134 行）—— 数值偏移后递归调用 `NumberFunction`；
- `math.go`（159 行）、`string.go`（71 行）—— 与 ICU 无关，不在迁移面内。

`MessageFunctionContext` 暴露 `Locales() []string`、`LocaleMatcher() string`、`OnError(error)` <!-- ref: messageformat-go/pkg/functions/types.go:99-114 -->。SPEC §6.1 已写明：context 的 `[]string` 在 adapter 层 `locale.Parse` 一次。

### 4.2 上下游关系定位

可能性：

| 选项 | 描述 | 评估 |
|------|------|------|
| 上下游（推荐） | go-intl 是底层；messageformat-go import go-intl，把 7 个内建 function 改写成 adapter | SPEC §6.1 已选；语义最干净，messageformat-go 不再持有重复的 ECMA-402 实现 |
| 平级（共用 context 类型） | 两个包共享一个 `Context` interface | 否决：会形成循环或共同上游包，无收益 |
| 反向（go-intl 依赖 messageformat-go） | go-intl 借用 `MessageValue` | 否决：go-intl 不做消息层，引入此依赖污染 façade 表面 |

### 4.3 迁移边界（concrete）

按 function 划分迁移工作量：

| messageformat-go function | go-intl 对应 API | 迁移类型 | 备注 |
|---------------------------|-----------------|---------|------|
| `:integer` | `numberformat.New(loc, WithMaximumFractionDigits(0), ...)` | thin adapter | 选项 map → typed options |
| `:number` | `numberformat.New(loc, ...)` | thin adapter | 与 `:integer` 共享 helper |
| `:currency` | `numberformat.New(loc, WithStyle(StyleCurrency), WithCurrency(code))` | thin adapter | currency code 校验委托给 go-intl |
| `:percent` | `numberformat.New(loc, WithStyle(StylePercent))` | thin adapter | `* 100` 由 go-intl 处理（与 ECMA-402 对齐）|
| `:unit` | `numberformat.New(loc, WithStyle(StyleUnit), WithUnit(id))` | thin adapter | unit identifier 校验委托给 go-intl |
| `:offset` | `numberformat.New(...)` + 偏移在 adapter 层完成 | partial | 偏移是 messageformat 自身语义 |
| `:date` / `:datetime` / `:time` | `datetimeformat.New(loc, ...)` | thin adapter | LDML 48 的 `dateFields`/`timePrecision` 在 adapter 层翻译为 ECMA-402 字段（year/month/day/...）|
| `:string` | 无 | 不迁移 | 与 ECMA-402 无关 |
| `:math` | 无 | 不迁移 | 算术，不涉及 locale-aware 格式化 |

迁移契约（API 形态）：

```go
// messageformat-go 侧的 adapter 形态（仅签名，非实现）
package functions

import (
    "github.com/agentable/go-intl/locale"
    "github.com/agentable/go-intl/numberformat"
)

func NumberFunction(ctx MessageFunctionContext, opts map[string]any, operand any) messagevalue.MessageValue {
    loc, _ := locale.Parse(GetFirstLocale(ctx.Locales())) // [] string -> Locale
    // map[string]any -> []numberformat.Option：在 adapter 层完成 LDML 48 -> ECMA-402 命名映射
    nf, err := numberformat.New(loc, mapNumberOptions(opts)...)
    if err != nil {
        ctx.OnError(err)
        return messagevalue.NewFallbackValue(ctx.Source(), loc.String())
    }
    // ...转 MessageValue
}
```

API 稳定性约束（SPEC §6.1 已定）：go-intl 1.0 之后不允许 `numberformat.Option` / `numberformat.ResolvedOptions` 字段重命名或删除。

### 4.4 共享类型表面

唯一跨包共享的类型 = `locale.Locale`。`numberformat.Options` / `datetimeformat.Options` 不应被 `messageformat-go` 直接持有（否则签名表面绑死）；adapter 层负责 `map[string]any` → typed options 的翻译。

---

## 5. host binding（go-typescript）的 Phase 1/4 决策

### 5.1 Node 是怎么暴露 Intl 的

`node_i18n.cc` 实际承担三件事：
- ICU common data 路径加载与切换（`InitializeICUDirectory`、`SMALL_ICUDATA_ENTRY_POINT`）<!-- ref: node/src/node_i18n.cc:42-86 -->；
- `UConverter`/`UConverter` Buffer 与 V8 String 互转 <!-- ref: node/src/node_i18n.h:56-75 -->；
- 时区设置 (`SetDefaultTimeZone`)、IDNA 转换、字符串宽度等周边能力。

它**不是** Intl 对象的暴露层——`Intl.NumberFormat` 等以 V8 的 `JSObject` 形式直接装配在 `globalThis.Intl` 上，由 `deps/v8/src/objects/js-*.cc` 实现，`node_i18n.cc` 只是初始化 ICU 全局数据。

### 5.2 含义

如果 `go-typescript` 未来想暴露 `globalThis.Intl`，它需要的是「一组 host function 表」把 JS 调用桥到 Go 的 formatter 实例：

```
globalThis.Intl.NumberFormat → go-typescript host fn → numberformat.New(...)
```

这层桥由 `go-typescript` 自己写更合适，因为：
- 它知道自己的 JS runtime 边界（V8 embedder API 还是其它）；
- go-intl 不应在公共表面引入「JS 类型」依赖。

### 5.3 推荐

- **Phase 1 不交付 host binding**——保持 go-intl 公共表面只用 Go 类型（`time.Time`、`language.Tag`、`locale.Locale`），不暴露任何 V8/JS 兼容形态。
- **Phase 4 重新评估**——当 `go-typescript` 真的要 `globalThis.Intl` 时，由 `go-typescript` 直接 import go-intl 的子包，写薄桥。
- **保留架构兼容**：façade 的 per-method 签名应可机械映射（`(loc, value, opts) -> string` 与 JS `(loc, options).format(value)` 形态一一对应）。这条由 SPEC §6.4 已要求。

---

## 6. defaultRichTextElements 与 Go 的对应

formatjs `intl/types.ts:75` 的 `defaultRichTextElements?: Record<string, FormatXMLElementFn<T>>` 仅在 `formatMessage` 中被消费——把 ICU MessageFormat 中的 `<bold>...</bold>` tag 替换成 React/Vue 的元素 <!-- ref: formatjs/packages/intl/types.ts:65-78 -->。

这是 **消息格式化** 范畴，不是 ECMA-402。SPEC §1.1 已声明 go-intl 不做消息格式化。结论：

- go-intl 不保留 `defaultRichTextElements` 钩子；
- 由 messageformat-go 自己暴露（其 `MessageValue` 已支持富类型 fallback）；
- `intl.IntlConfig` 不引入 `RichTextElements` 字段。

---

## 7. 对本项目的落地建议

### 7.1 包结构（intl.go）

```go
// /Users/lincheng/work/golang/go-intl/intl.go
package intl

import (
    "time"

    "github.com/agentable/go-intl/datetimeformat"
    "github.com/agentable/go-intl/locale"
    "github.com/agentable/go-intl/numberformat"
    "github.com/agentable/go-intl/pluralrules"
)

type Config struct {
    Locale         locale.Locale
    DefaultLocale  locale.Locale
    TimeZone       string
    Cache          Cache
    OnError        func(error)
}

type Option func(*Config)

func WithLocale(l locale.Locale) Option
func WithDefaultLocale(l locale.Locale) Option
func WithTimeZone(tz string) Option
func WithCache(c Cache) Option
func WithoutCache() Option
func WithOnError(fn func(error)) Option

type Intl struct { /* 持有 resolved Config + formatter cache */ }

func New(opts ...Option) (*Intl, error)

// Per-call methods
func (i *Intl) FormatNumber(v any, opts ...numberformat.Option) string
func (i *Intl) FormatNumberToParts(v any, opts ...numberformat.Option) []numberformat.Part
func (i *Intl) FormatDate(t time.Time, opts ...datetimeformat.Option) string
func (i *Intl) FormatDateToParts(t time.Time, opts ...datetimeformat.Option) []datetimeformat.Part
func (i *Intl) FormatTime(t time.Time, opts ...datetimeformat.Option) string
func (i *Intl) FormatRange(from, to time.Time, opts ...datetimeformat.Option) string
func (i *Intl) SelectPlural(n any, opts ...pluralrules.Option) pluralrules.Form

// 一行 helper（共享一个进程级 sync.Map cache）
func FormatNumber(loc locale.Locale, v any, opts ...numberformat.Option) string
func FormatDate(loc locale.Locale, t time.Time, opts ...datetimeformat.Option) string
func FormatTime(loc locale.Locale, t time.Time, opts ...datetimeformat.Option) string
func FormatRange(loc locale.Locale, from, to time.Time, opts ...datetimeformat.Option) string
func SelectPlural(loc locale.Locale, n any, opts ...pluralrules.Option) pluralrules.Form
```

### 7.2 Cache 接口

```go
package intl

type Cache interface {
    // GetOrCreate 命中则返回已存的 entry；未命中调用 build 创建并存入。
    // build 失败时返回 (nil, err) 并不缓存。
    GetOrCreate(key string, build func() (any, error)) (any, error)
    // Reset 清空整个缓存。供热升级、配置重载场景使用。
    Reset()
}

// 默认实现：sync.Map，无淘汰
type defaultCache struct { /* sync.Map */ }
```

子包 formatter 暴露 `canonicalKey() string`：

```go
package numberformat

type Options struct { /* 全部字段含 zero-value 语义 */ }

func (o Options) canonicalKey() string // 内部用
```

### 7.3 错误模型

错误类型在根包 `errors.go`：

```go
package intl

import "errors"

var (
    ErrInvalidOption       = errors.New("intl: invalid option")
    ErrUnsupportedLocale   = errors.New("intl: unsupported locale")
    ErrMissingLocaleData   = errors.New("intl: missing locale data")
)

type Error struct {
    Code    string
    Locale  string
    Message string
    Wrapped error
}
```

子包各自有 sentinel：`ErrInvalidCurrency`、`ErrInvalidUnit`、`ErrInvalidTimeZone` 等等。

---

## 8. 决策矩阵

| 决策 | 推荐 | 备选 | 否决 | 依据章节 |
|------|------|------|------|---------|
| 双轨 façade | `intl.New` 持久 + `intl.FormatXxx` 一行 helper | 仅持久 | 仅一行 helper | §1.3 |
| Cache 容器 | `sync.Map` 默认 + `Cache` 接口可注入 | `map+RWMutex` | 内置 LRU | §2.3 |
| Cache key | `string(loc + canonical(opts))` | typed struct key | 指针地址 | §2.3 |
| Cache 淘汰 | 默认无 | LRU 注入 | 内置默认 LRU 大小 | §2.3 |
| 错误模型 | 构造期 error + 运行期 onError + fallback | 全 error 返回 | 全 panic / 全静默 | §3.4 |
| messageformat-go 关系 | go-intl 下游被依赖 | 平级共享 | 反向依赖 | §4.2 |
| Adapter 层位置 | messageformat-go 自带 | go-intl 提供 messageformat 子包 | 共享中间包 | §4.3 |
| host binding | Phase 1 不做 | Phase 1 提供契约文档 | Phase 1 实现 host binding | §5.3 |
| defaultRichTextElements | 不收纳 | 在 façade 选项里保留 hook | 在 numberformat/datetimeformat 收纳 | §6 |

---

## 9. 代码块索引

| 章节 | 块 | 内容 |
|------|---|------|
| §1.3 | Go signatures | `intl.Intl` / `intl.New` / `FormatNumber` |
| §1.3 | call example | 3 行 `intl.New + FormatNumber` 调用 |
| §2.3 | Go signatures | `intl.Cache` / `Option` |
| §3.4 | Go signatures | `intl.ErrorCode` / `intl.Error` / `numberformat` sentinels |
| §4.3 | adapter signature | messageformat-go `NumberFunction` adapter |
| §7.1 | Go signatures | 完整 `intl` 包 API 汇总 |
| §7.2 | Go signatures | `Cache` interface + `Options.canonicalKey()` |
| §7.3 | Go signatures | 错误类型与 sentinel |

---

## 10. 引用清单

### formatjs
- `formatjs/packages/intl/create-intl.ts:24-30` — `CreateIntlFn` 类型签名
- `formatjs/packages/intl/create-intl.ts:57-162` — `createIntl` 主体与 bind 形态
- `formatjs/packages/intl/utils.ts:24-41` — `filterProps` 选项白名单过滤
- `formatjs/packages/intl/utils.ts:43-79` — `defaultErrorHandler` / `DEFAULT_INTL_CONFIG`
- `formatjs/packages/intl/utils.ts:81-91` — `createIntlCache` 默认 cache 工厂
- `formatjs/packages/intl/utils.ts:93-170` — `createFormatters` + fast-memoize variadic 策略
- `formatjs/packages/intl/utils.ts:172-192` — `getNamedFormat`：custom format 字典
- `formatjs/packages/intl/types.ts:48-58` — `OnErrorFn` / `OnWarnFn` 错误回调签名
- `formatjs/packages/intl/types.ts:65-78` — `ResolvedIntlConfig` / `defaultRichTextElements`
- `formatjs/packages/intl/types.ts:225-260` — `Formatters` / `IntlShape`
- `formatjs/packages/intl/types.ts:262-270` — `IntlCache` 7 桶结构
- `formatjs/packages/intl/error.ts:3-9` — `IntlErrorCode` 五分类
- `formatjs/packages/intl/error.ts:11-110` — 错误类层级（IntlError / Format / MissingData / MissingTranslation）
- `formatjs/packages/intl/number.ts:11-39` — `NUMBER_FORMAT_OPTIONS` 白名单
- `formatjs/packages/intl/number.ts:41-66` — `getFormatter` filterProps + 缓存桥
- `formatjs/packages/intl/number.ts:68-111` — `formatNumber` / `formatNumberToParts` try/catch 模式
- `formatjs/packages/intl/dateTime.ts:11-33` — `DATE_TIME_FORMAT_OPTIONS` 白名单
- `formatjs/packages/intl/dateTime.ts:35-76` — `getFormatter` 含 timeZone/format 默认值合并
- `formatjs/packages/intl/dateTime.ts:78-211` — formatDate / formatTime / formatDateTimeRange 等
- `formatjs/packages/intl/plural.ts:1-49` — `formatPlural` 仅返回 LDMLPluralRule
- `formatjs/packages/intl/tests/create-intl.test.ts:4-62` — createIntl 与 onWarn 用例

### Node / V8
- `node/src/node_i18n.h:38-130` — `i18n` namespace API、Converter、idna_mode
- `node/src/node_i18n.cc:1-120` — ICU 数据加载、ICUDirectory、SMALL_ICUDATA_ENTRY_POINT
- `node/src/node_i18n.cc:42-86` — `InitializeICUDirectory` 行为说明（注释）
- `node/test/parallel/test-intl.js:22-80` — Node-side Intl 烟雾测试形态

### messageformat-go (sibling)
- `messageformat-go/pkg/functions/types.go:18-22` — `MessageFunction` 签名
- `messageformat-go/pkg/functions/types.go:36-119` — `MessageFunctionContext` 字段、构造、accessor
- `messageformat-go/pkg/functions/registry.go:28-69` — `DefaultFunctions` / `DraftFunctions` 表
- `messageformat-go/pkg/functions/registry.go:72-141` — `FunctionRegistry` 并发安全注册
- `messageformat-go/pkg/functions/number.go:59-130` — `readNumericOperand` 数字操作数解析
- `messageformat-go/pkg/functions/number.go:300-370` — `IntegerFunction` 选项 merge 流水线
- `messageformat-go/pkg/functions/number.go:373-419` — `getMessageNumber` selectable 校验
- `messageformat-go/pkg/functions/datetime.go:11-65` — datetime 选项白名单与 readStringOption
- `messageformat-go/pkg/functions/datetime.go:67-199` — `dateTimeImplementation` 选项流水线
- `messageformat-go/pkg/functions/currency.go:1-100` — `:currency` adapter 形态（带原 TS 注释）
- `messageformat-go/pkg/functions/percent.go:69-118` — `PercentFunction`
- `messageformat-go/pkg/functions/unit.go:67-120` — `UnitFunction`
- `messageformat-go/pkg/functions/offset.go:49-134` — `OffsetFunction` 数值偏移 + 调用 `NumberFunction`
- `messageformat-go/pkg/functions/utils.go:33-131` — `asBoolean` / `asPositiveInteger` / `asString` 操作数转换

### go-intl 项目本身
- `SPECS/00-vision-and-scope.md:97-115` — Phase 1 公共表面
- `SPECS/00-vision-and-scope.md:209-227` — 消费者契约（messageformat-go / go-test / go-typescript）
- `CLAUDE.md` — 错误处理与设计原则约束
