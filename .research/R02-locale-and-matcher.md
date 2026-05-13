---
id: R02
title: Locale 类型、规范化与最佳匹配调研 — language.Tag 覆盖、扩展字段、BestFitMatcher
task: r02
date: 2026-05-08
status: draft
scope:
  - language.Tag 与 ECMA-402 Locale 字段的覆盖差异
  - maximize / minimize 数据来源（language.Tag likelyFromLikelyTo vs CLDR likelySubtags.json）
  - weekInfo / calendars / hourCycles / numberingSystems / timeZones getter 的实例化时机
  - BestFitMatcher 实现选择：FormatJS 匹配算法 vs language.Matcher 复用
  - Locale 可比性（值类型 + Equal/String vs comparable struct 约束）
tags: [locale, matcher, bcp47, cldr, language-tag, best-fit, formatjs]
---

# R02 — Locale 类型、规范化与最佳匹配

## 执行摘要

| 决策点 | 推荐方案 | 置信度 | 依据 |
|--------|----------|--------|------|
| Locale shape | `Locale` struct 嵌入 `language.Tag`，加 7 个 ECMA-402 扩展字段 + 1 缓存字段，**值类型** | High | 与 SPEC 00 §4.1 已决；本报告把扩展字段集合定为 `ca/co/hc/kf/kn/nu/fw` 全集 |
| 最佳匹配实现 | 移植 formatjs 的 **三层匹配算法**（Tier 1 精确、Tier 2 maximize+回退、Tier 3 完整 UTS #35），不复用 `language.Matcher` | High | `language.Matcher` 输出 quality 而非 CLDR distance，与 ECMA-402 conformance 不可比；formatjs 算法已有公开测试基线 |
| getter 实例化时机 | 构造时**预解析** language/script/region/calendar/hourCycle 等"快"字段（已在 BCP 47 解析时得到）；`getCalendars`/`getCollations`/`getHourCycles`/`getNumberingSystems`/`getTimeZones`/`getWeekInfo` 作为**方法**惰性计算（每次调用读 CLDR 数据） | High | formatjs 方案；CLDR 数据访问是 O(1) 表查找；预解析增加构造开销且大多数调用方不读这些字段 |
| `weekInfo` 的预热 | Phase 1 不缓存，按 method 调用时计算；若性能基线显示瓶颈再加 sync.Once 缓存 | Medium | YAGNI；formatjs 也未缓存 |
| 可比性 | 通过 `(Locale).Equal(other) bool` + `(Locale).String() string` 暴露；不要求 Go `==` 可比 | High | language.Tag 不是 `comparable`（含 slice 内部）；嵌入后 Locale 也不是 |
| 字符串往返 | `Parse(s) (Locale, error)` 与 `(Locale).String()` 互逆，含完整 `-u-` 扩展 | High | spec 强制；与 SPEC 00 §4.3 一致 |

**前提**：本报告锚定 SPECS §8 Q3（getter 实例化时机）与 Q5（最佳匹配实现），未涉及 Q1/Q2/Q4。

## 1. ECMA-402 Locale 字段集

### 1.1 spec 字段清单

ECMA-402 `Intl.Locale` 通过 7 个 Unicode `-u-` 扩展键暴露 locale-aware 配置：<!-- ref: formatjs/packages/intl-locale/index.ts:52-60 -->

```typescript
const RELEVANT_EXTENSION_KEYS = ['ca', 'co', 'hc', 'kf', 'kn', 'nu', 'fw'] as const
```

| 扩展键 | 字段名 | 取值范围 | 默认 | 影响 formatter |
|-------|--------|---------|------|----------------|
| `ca` | calendar | gregory / buddhist / chinese / hebrew / islamic / islamic-civil / persian / japanese 等 | 区域决定 | DateTimeFormat |
| `co` | collation | standard / phonebk / pinyin 等 | 区域决定 | Collator（Phase 3） |
| `hc` | hourCycle | h11 \| h12 \| h23 \| h24 | 区域决定 | DateTimeFormat |
| `kf` | caseFirst | upper \| lower \| false | false | Collator |
| `kn` | numeric | bool（kn-true / kn 单独写表示 true） | false | Collator |
| `nu` | numberingSystem | latn / arab / arabext / beng / deva ... | 区域决定 | NumberFormat / DateTimeFormat |
| `fw` | firstDayOfWeek | sun / mon / tue / wed / thu / fri / sat | 区域决定 | DateTimeFormat（formatRange 跨周） |

所有字段皆可由调用者覆盖（`new Intl.Locale('en-US', {hourCycle: 'h23'})`）或在 BCP 47 标签里直接写（`'en-US-u-hc-h23'`），二者等价。

### 1.2 language.Tag 的覆盖范围

`golang.org/x/text/language.Tag` 暴露的访问器：

| 访问器 | 对应 spec 字段 | 状态 |
|--------|---------------|------|
| `Base()` | language | OK |
| `Script()` | script | OK |
| `Region()` | region | OK |
| `Variants()` | variants | OK |
| `Extensions()` / `TypeForKey("ca")` | 所有 `-u-` 扩展键 | **半透传**：Tag 保留 `-u-` 段，但只能用 `TypeForKey/SetTypeForKey` 取/设单键 |
| 无直接访问器 | calendar / hourCycle / caseFirst / numeric / numberingSystem / firstDayOfWeek / collation | 必须自己解析 |

**结论**：`language.Tag` **保留**了 7 个扩展字段的原始字符串值（通过 `TypeForKey("hc")` 等可读），但 **不暴露**为类型化字段，也不做 `numeric` 的字符串↔bool 转换、不做 `caseFirst` 的"false 字面量是合法值"识别。`go-intl` 的 `Locale` 必须在嵌入 `Tag` 之外，自己存这 7 个字段（pre-parsed 状态）。

PHP `ecma402_locale` 同样选择"全字段平铺"：<!-- ref: ext/src/ecma402/locale.h -->

```c
typedef struct ecma402_locale {
    char *baseName, *calendar, *canonical, *caseFirst, *collation, *currency,
         *hourCycle, *language, *numberingSystem, *original, *region, *script;
    bool numeric;
    ecma402_errorStatus *status;
} ecma402_locale;
```

`numeric` 用 `bool`，其它扩展字段用 `char*`——这是构造时全部 eager 解析、运行时零分配的极端方案。Go 走折中：扩展字段使用空串表示"未指定/默认"，`Numeric *bool` 用三态以区分"未设" / "false" / "true"（spec 要求 caseFirst 'false' 与 numeric=false 不重叠语义）。

### 1.3 推荐 Go shape

```go
// locale/locale.go（仅签名）
type Locale struct {
    Tag             language.Tag // 嵌入：language/script/region/variants 直接用 Tag 方法
    Calendar        string       // 默认 ""
    Collation       string
    HourCycle       string       // "h11" | "h12" | "h23" | "h24" | ""
    CaseFirst       string       // "upper" | "lower" | "false" | ""
    NumberingSystem string
    Numeric         bool
    FirstDayOfWeek  string       // "mon" | "tue" | ... | ""
}

// 构造与字符串往返
func Parse(s string) (Locale, error)
func MustParse(s string) Locale
func New(tag language.Tag, opts ...Option) Locale
func (l Locale) String() string

// 比较 / 拷贝
func (l Locale) Equal(other Locale) bool
func (l Locale) WithCalendar(cal string) Locale
// 等价物：WithHourCycle / WithCaseFirst / WithNumberingSystem / WithNumeric / WithFirstDayOfWeek / WithCollation
```

调用示例（3-5 行）：

```go
loc, err := locale.Parse("en-US-u-hc-h23-ca-buddhist")
fmt.Println(loc.HourCycle)               // "h23"
fmt.Println(loc.Calendar)                // "buddhist"
fmt.Println(loc.WithHourCycle("h12").String()) // "en-US-u-ca-buddhist-hc-h12"
```

## 2. 规范化：maximize / minimize

### 2.1 三方对比

| 项目 | maximize 数据源 | minimize 数据源 | 预计算 |
|------|----------------|----------------|-------|
| formatjs `intl-getcanonicallocales` | CLDR `likelySubtags.json` 直接生成 TS 表 | 同上反向（移除可推断的 subtags） | TS 表内联，运行时表查找 |
| `golang.org/x/text/language` | `Tag.LikelyScript()` / `LikelyRegion()` 单维度 + `Tag.Parent()` 链 | `Tag` 内部表（CLDR 派生） | 静态表 |
| translate-agent/intl | 完全不调 maximize/minimize；依赖 `language.MustParseBase` 的硬编码常量 | 不需要 | N/A |

### 2.2 关键差异

`language.Tag.LikelyScript()` 与 ECMA-402 `addLikelySubtags` 在 95% 案例上一致，但有边界差异——例如 `und` 推断、罕见组合（`zh-Hant-MO` 的回退路径）。formatjs 测试 `tests/likely-subtags.test.ts` 和 `tests/minimize.test.ts` 是真值表；移植到 Go 时应**先跑这些测试针对 `language.Tag` 的封装**，按测试失败的 case 决定是否需要补充 CLDR `likelySubtags.json` 表。

ICU 用的是从 CLDR 同源生成的表，输出与 formatjs 几乎完全一致；`x/text` 的表也来自 CLDR 但版本可能不一致（`x/text` 的更新滞后 CLDR 大版本）。

### 2.3 推荐

**Phase 1**：先用 `language.Tag` 的内置 maximize/minimize 实现（通过 `language.Matcher`/`language.MatchStrings` 内部用到的相同表），跑 formatjs `likely-subtags.test.ts` 全部用例。失败的 case：

- 若 < 5%：在 `internal/cldr/likelysubtags.go` 用 CLDR 数据补丁覆盖。
- 若 ≥ 5%：完全切到自带 CLDR 表（`internal/cldr/likelysubtags.go` 由 `tools/gen-cldr/` 从 `likelySubtags.json` 生成 Go 表）。

**置信度**：Medium。需要先跑测试再做最终决定，但默认走 `language.Tag` 内置 + 局部补丁是最低风险路径。

## 3. Getter 实例化时机（SPEC 00 §8 Q3）

### 3.1 三方对比

| 项目 | 字段类型 getter（calendar / hourCycle 等） | 集合方法（getCalendars / getHourCycles） | 复杂方法（getWeekInfo / getTextInfo） |
|------|-------------------------------------------|----------------------------------------|--------------------------------------|
| formatjs | 构造时解析进 internal slot；getter 直接读 slot | **每次调用**计算（基于 `maximize().region` + CLDR preference 表） | **每次调用**计算 |
| PHP ext/intl | char* 字段、构造时全部预解析 | helper 函数，每次调用 | 同上 |
| translate-agent/intl | 不区分（直接基于 `language.Tag`） | N/A（功能不全） | N/A |

formatjs `getCalendars()` 实现节选：

```typescript
public getCalendars(): string[] {
  return calendarsOfLocale(this)  // 内部走 maximize().region 查 cldr.locale 表
}
```

`calendarsOfLocale` 每次都会 `loc.maximize()` 取 region，再 `getCalendarPreferenceDataForRegion(region)` 查 generated CLDR 表。这两次都是 O(1)：maximize 是表查找，preference 是 region key 查 map。**没有缓存**。<!-- ref: formatjs/packages/intl-locale/index.ts:340, preference-data.ts -->

### 3.2 缓存 vs 不缓存

缓存的代价：

- 增加 Locale struct 大小（加 7 个 `[]string`/`map`/`*WeekInfo` 指针 ≈ 100+ bytes per Locale）。
- 必须使用 `*Locale`（指针）才能内部 mutate，否则失去值语义；或加 `sync.Once` 但增加同步成本。
- 大多数调用方只读 `loc.Calendar`、`loc.HourCycle` 这种 O(1) 字段，从不调 `getCalendars()`。

不缓存的代价：

- 每次 `loc.GetCalendars()` 重做 maximize + region 查表 ≈ 数十 ns。
- 调用频率极低（只有"想知道这个 locale 支持哪些日历"的元信息查询场景）。

### 3.3 推荐

**Phase 1**：

- **构造时预解析**：`Calendar`、`HourCycle`、`CaseFirst`、`Collation`、`NumberingSystem`、`Numeric`、`FirstDayOfWeek`——这些都是从 `-u-` 扩展键直接拿到，BCP 47 解析顺手完成，零额外成本。
- **method 惰性计算**：`GetCalendars()`、`GetCollations()`、`GetHourCycles()`、`GetNumberingSystems()`、`GetTimeZones()`、`GetWeekInfo()`、`GetTextInfo()`。每次调用走 `maximize() + CLDR preference 查表`。
- **不缓存**：保持 `Locale` 值类型；将来若 benchmark 显示某个 method 是热点，再加 `sync.Once` 缓存（YAGNI）。

```go
// locale/info.go（仅签名）
func (l Locale) GetCalendars() []string
func (l Locale) GetCollations() []string
func (l Locale) GetHourCycles() []string
func (l Locale) GetNumberingSystems() []string
func (l Locale) GetTimeZones() []string  // 仅当 region 已知；否则空切片
type WeekInfo struct {
    FirstDay    time.Weekday
    Weekend     []time.Weekday
    MinimalDays int
}
func (l Locale) GetWeekInfo() WeekInfo
type TextInfo struct {
    Direction string // "ltr" | "rtl"
}
func (l Locale) GetTextInfo() TextInfo
```

调用示例：

```go
loc := locale.MustParse("ar-SA")
fmt.Println(loc.GetWeekInfo().FirstDay)  // time.Sunday
fmt.Println(loc.GetTextInfo().Direction) // "rtl"
fmt.Println(loc.GetCalendars())          // ["islamic-umalqura", "gregory", "islamic", ...]
```

**置信度：High**。与 formatjs 完全对齐；CLDR 表查找成本可忽略；调用方少；保持值语义。

### 3.4 命名补充

ECMA-402 spec 把这些写为 `Intl.Locale.prototype.getCalendars()`（method）而不是 `get calendars`（getter），所以 Go 端用 `GetCalendars` 方法（不是字段）是 spec-faithful 的；`Calendar`（单数 string 字段）则对应 `Intl.Locale.prototype.calendar` getter。命名上能区分"单一选择 vs 候选列表"。

## 4. BestFitMatcher 决策（SPEC 00 §8 Q5）

### 4.1 ECMA-402 在 BestFit 上的开放性

ECMA-402 对 `BestFitMatcher` 的规定文本（sec-bestfitmatcher）：

> The algorithm used to determine the best fit is implementation-defined. **Conforming implementations are recommended to satisfy the following criteria:** [...]

即 spec 不强制具体算法。这给了 `go-intl` 三个选项：

1. 完全照抄 formatjs 算法（CLDR distance 表 + UTS #35）。
2. 复用 `language.Matcher`（Go 原生，输出 `Confidence`：No/Low/High/Exact）。
3. 自研启发式（简化版）。

### 4.2 三方实现对比

| 项目 | 算法 | 数据来源 | 输出指标 |
|------|------|---------|---------|
| formatjs `intl-localematcher` | 三层：Tier 1 精确/Tier 2 maximize+subtag truncation/Tier 3 完整 UTS #35 | CLDR `languageMatching.json`（含 paradigmLocales、matchVariables、distance 表） | 数值距离 0-840+，threshold 838 | <!-- ref: formatjs/packages/intl-localematcher/abstract/utils.ts:33-496 -->
| `language.Matcher` | 自家算法（基于 CLDR 但不是 UTS #35 完整版） | `x/text` 内部 CLDR 表（旧） | `Confidence`：No / Low / High / Exact + 选中 Tag |
| translate-agent/intl | 不做 best-fit；只用 `MustParseBase` 暴力查 | 硬编码 base 列表 | N/A |

### 4.3 算法细节（formatjs 三层）

`findBestMatch(requestedLocales, supportedLocales, threshold)`：<!-- ref: formatjs/packages/intl-localematcher/abstract/utils.ts:341-496 -->

**Tier 1（精确匹配）**：把 supportedLocales 装入 `Set`，对每个 requestedLocale 做 `Set.has`。命中即 distance = `0 + i*40`（`i` 是 requestedLocale 的位置惩罚）。第一个 requestedLocale 命中时立刻返回。

**Tier 2（maximize + 后缀截断）**：对每个 requestedLocale 调用 `Locale.maximize()`，对结果做 right-to-left subtag 截断生成候选（`zh-Hant-TW` → `["zh-Hant-TW", "zh-Hant", "zh"]`），逐个查 supportedSet。命中后比较 candidate 的 maximize 结果是否等于 requested 的 maximize 结果——是则 distance = 0，否则 `j*10 + i*40`。

**Tier 3（完整 UTS #35）**：当 Tier 2 没找到 distance=0 的匹配时，对每对 (requested × supported) 做 `findMatchingDistance(desired, supported)` 得 LSR-based distance（从 CLDR `languageMatching.json` 表查 paradigmLocales/matchVariables）。`findMatchingDistance` 走 `fast-memoize`。

性能（formatjs 文档基于 700+ supported locales）：

- Tier 1 ~12ms（无 maximize）
- Tier 2 ~13-15ms（maximize + 回退）
- Tier 3 ~100ms+（即使有 memoize，初始 cold cache 时较慢）

### 4.4 `language.Matcher` 复用的成本

```go
matcher := language.NewMatcher([]language.Tag{language.AmericanEnglish, language.Japanese})
tag, _, conf := matcher.Match(language.MustParse("en-GB"))
```

- 输出是 `Confidence` 枚举（No/Low/High/Exact），无法直接还原 ECMA-402 的"匹配距离"语义。
- 若 conformance 测试要求"`en-XZ` 应该回退到 `en` 还是 `ja`"，`language.Matcher` 的回答可能与 FormatJS 不一致——`language.Matcher` 在多个 supported 中选 confidence 最高的，但 distance tie-breaking 规则是 `x/text` 内部决定。
- `x/text` 的 CLDR 数据版本与 `internal/cldr/VERSION` 不一定一致（SPEC 00 §8 Q4 待决），会引入第二个数据源。

如果走 `language.Matcher`：

- ✅ 实现成本低（`language.NewMatcher` 一行调用）
- ✅ 与其他 `x/text` 用户（`messageformat-go`）行为一致
- ❌ formatjs conformance 测试很可能大量失败
- ❌ 不能在 Phase 0 conformance 测试中作为"通过/失败"的金标准

### 4.5 推荐

**Phase 1**：移植 formatjs 三层算法到 `internal/localematcher/`：

```go
// internal/localematcher/match.go（仅签名）
type Algorithm int
const (
    AlgorithmLookup   Algorithm = iota  // sec-LookupMatcher
    AlgorithmBestFit                    // sec-BestFitMatcher
)

type Result struct {
    Locale      string  // 选中的 supported locale
    DataLocale  string  // 用于查 CLDR 数据的 locale（去掉无关 -u- 扩展）
    Distance    int     // 0 = 完全等价，838 = 阈值，>= threshold 则 fallback
}

func Match(requested, supported []string, defaultLocale string, alg Algorithm) (Result, error)
func ResolveLocale(requested, supported []string, defaultLocale string,
    relevantExtensionKeys []string, localeData LocaleData) (Resolved, error)

// 三层内部函数（unexported）
func findBestMatch(requested, supported []string, threshold int) bestMatchResult
func findMatchingDistance(desired, supported string) int  // 走 sync.Map 缓存
```

调用示例：

```go
res, err := localematcher.Match(
    []string{"zh-TW"},
    []string{"zh-Hans", "zh-Hant", "en"},
    "en", localematcher.AlgorithmBestFit,
)
fmt.Println(res.Locale)   // "zh-Hant"
fmt.Println(res.Distance) // 0（zh-TW maximize 到 zh-Hant-TW，回退到 zh-Hant）
```

**为什么不复用 `language.Matcher`**：

1. ECMA-402 conformance 测试是 SPEC 00 §2 强制目标，无法因为 `language.Matcher` 输出不同而放弃。
2. formatjs 三层算法已有公开测试基线（`intl-localematcher/tests/`）可移植。
3. `internal/localematcher/` 私有，未来可接 `x/text` 切换器；用户不感知。
4. `findMatchingDistance` 的 memoize 在 Go 用 `sync.Map` 即可，性能可控。

**`language.Matcher` 的退路**：可以暴露 `localematcher.WithLanguageMatcher()` 选项让用户切到 `x/text` 实现（适用于"我只想要 x/text 的行为"的场景）；这是 Phase 4 优化项。

**置信度：High**。conformance 测试是硬约束；formatjs 算法已工程化。

### 4.6 数据需求

移植 formatjs 三层算法需要：

| 数据 | 来源 | 大小估算（CLDR 45+） |
|------|-----|--------------------|
| `languageMatching.json` | CLDR `cldr-core/supplemental/` | ~30KB JSON → ~15KB Go literal |
| `likelySubtags.json` | CLDR `cldr-core/supplemental/` | ~150KB JSON → ~80KB Go literal |
| `regions.json`（matchVariables 展开） | CLDR `cldr-core/supplemental/` | ~50KB |
| 总 | | ~150KB Go source 嵌入 |

由 `tools/gen-cldr/` 在生成时产出 `internal/cldr/locale_matching.go`、`internal/cldr/likely_subtags.go`、`internal/cldr/regions.go`。

## 5. Locale 可比性

### 5.1 约束

- `language.Tag` **不是** Go `comparable`（内部含 slice/字符串切片字段或类似），用 `==` 编译失败或行为未定义。
- ECMA-402 `Intl.Locale` 没有标准 equality method；JS 用 `loc1.toString() === loc2.toString()` 比较。
- `messageformat-go` 当前 `MessageFunctionContext.Locales()` 返回 `[]string`，不依赖 Locale 可比性。

### 5.2 推荐

提供显式比较 API，不要求 `comparable` 约束：

```go
// locale/locale.go（仅签名）
func (l Locale) Equal(other Locale) bool        // 比较所有扩展字段 + Tag.String()
func (l Locale) String() string                 // 含 -u- 扩展的 canonical BCP 47

// 排序需求：直接用 String() 比较或 sort.Slice + 自定义 less
```

```go
// 使用（5 行内）
a, _ := locale.Parse("en-US-u-hc-h23")
b, _ := locale.Parse("en-US-u-hc-h23-ca-gregory")
fmt.Println(a.Equal(b))             // false（calendar 不同）
fmt.Println(a.String() == b.String()) // 同上
fmt.Println(slices.Contains([]locale.Locale{a, b}, a)) // 编译错——locale.Locale 不可比
```

**警告**：用户若把 `Locale` 放进 `map[locale.Locale]V`，需要改用 `map[string]V` 配 `loc.String()`。文档中需明确说明。

**置信度：High**。

## 6. 对本项目的落地建议

### 6.1 包布局

```text
locale/
├── locale.go          ← Locale struct、Parse/MustParse/New、String/Equal、With* setters
├── info.go            ← GetCalendars/GetCollations/GetHourCycles/GetNumberingSystems/GetTimeZones/GetWeekInfo/GetTextInfo
├── canonical.go       ← Maximize/Minimize 包装 language.Tag + 必要的 CLDR 补丁
├── option.go          ← Option type for With*
└── internal_test.go   ← 同 package 单元测试
internal/
├── localematcher/
│   ├── match.go       ← Match / ResolveLocale 公共入口
│   ├── lookup.go      ← LookupMatcher（spec sec-LookupMatcher）+ BestAvailableLocale
│   ├── best_fit.go    ← BestFitMatcher 三层（findBestMatch / Tier 1-3）
│   ├── distance.go    ← findMatchingDistance + sync.Map 缓存
│   ├── ucanonicalize.go  ← UnicodeExtensionValue / InsertUnicodeExtensionAndCanonicalize
│   └── data.go        ← languageMatching / paradigmLocales / matchVariables 加载
└── cldr/
    ├── likely_subtags.go    ← likelySubtags 表（gen-cldr 生成）
    ├── locale_matching.go   ← languageMatching 表
    ├── regions.go           ← matchVariables 区域展开
    └── preference.go        ← getCalendarPreferenceDataForRegion / getHourCyclesPreferenceDataForLocaleOrRegion / getTimeZonePreferenceForRegion / getWeekDataForRegion
```

### 6.2 构造函数行为

```go
// 解析路径
Parse("en-US")                    // OK：base=en, region=US, 其余空
Parse("en-US-u-ca-buddhist-hc-h23") // OK：Calendar=buddhist, HourCycle=h23
Parse("en-US-u-ca-gregorian")     // 接受 spec 别名（gregorian → gregory）
Parse("xx-INVALID")               // RangeError → ErrInvalidLocale

// 构造路径
New(language.AmericanEnglish, locale.WithCalendar("buddhist"), locale.WithHourCycle("h23"))
```

构造时验证：

- `Calendar`、`Collation`、`NumberingSystem`、`HourCycle`、`CaseFirst`、`FirstDayOfWeek` 用对应 spec 枚举校验。
- `gregorian` 别名→ `gregory`，`islamic-civil` 别名→ `islamicc`（与 formatjs `index.ts` 同步）。
- `String()` 的字段顺序按 spec：`-u-` 扩展键按字典序（`ca` < `co` < `fw` < `hc` < `kf` < `kn` < `nu`）。

### 6.3 Maximize / Minimize

```go
// canonical.go（仅签名）
func (l Locale) Maximize() Locale  // 走 language.Tag.LikelyScript/Region + CLDR 表补丁
func (l Locale) Minimize() Locale  // 反向
```

实现策略：

1. 优先走 `language.Tag.LikelyScript()` + `LikelyRegion()`。
2. 如果结果与 formatjs `addLikelySubtags` 不一致（按 fixture 测试发现），用 `internal/cldr/likely_subtags.go` 表补丁覆盖。

## 7. 决策矩阵

| 决策 | 推荐 | 备选 | 否决 | 依据 |
|------|------|------|------|------|
| Locale 类型 | struct 嵌入 `language.Tag` + 7 个扩展字段 | type alias `language.Tag` | 全平铺无 Tag | SPEC 00 §4.1；扩展字段不在 Tag 类型化 API 内 |
| Numeric 字段类型 | `bool`（默认 false）+ `WithNumeric(true)` | `*bool` 三态 | spec 字符串 "true"/"false" | spec 默认就是 false；用户极少需"未设"vs"明确 false"区分 |
| getCalendars 等 | method 惰性计算（不缓存） | 构造预算 + 字段 | sync.Once 缓存 | YAGNI；formatjs 同方案；CLDR 表查表 O(1) |
| BestFitMatcher | 移植 formatjs 三层 | 复用 `language.Matcher` | 自研简化启发式 | conformance 测试是硬约束 |
| likelySubtags 表 | 默认走 `language.Tag`，按 fixture 失败 case 决定补丁 | 完全自带 CLDR 表 | 完全依赖 `language.Tag` | YAGNI + conformance 兜底 |
| Equality | `(Locale).Equal` + `(Locale).String()` | Go `==` | 全字段直接 `reflect.DeepEqual` | `language.Tag` 不可比；显式 API 更清晰 |
| 错误类型 | `ErrInvalidLocale`（packageroot） + wrap `language.Parse` 错误 | 直接 panic | 透传 `language` 包错误 | spec 要求 RangeError 语义 |
| 字段命名 | `Calendar`/`HourCycle`/`CaseFirst`/`NumberingSystem`/`FirstDayOfWeek`（PascalCase 即 spec 名 + Go 大写） | `Cal`/`Hc`（短缩写） | spec 全小写带破折号（`hour-cycle`） | spec verbatim；Go 习惯天然兼容 |

## 8. 代码块索引

| 位置 | 主题 | 类型 |
|------|------|------|
| §1.1 | formatjs `RELEVANT_EXTENSION_KEYS` | TypeScript const 数组 |
| §1.2 | PHP `ecma402_locale` struct | C struct |
| §1.3 | go-intl `Locale` struct + 关键方法签名 | Go 类型签名 |
| §1.3 | Parse 调用示例 | Go 调用片段 |
| §3.1 | formatjs `getCalendars()` 实现引用 | TypeScript 片段 |
| §3.3 | go-intl `GetCalendars`/`GetWeekInfo`/`GetTextInfo` 签名 + `WeekInfo`/`TextInfo` 类型 | Go 类型签名 |
| §3.3 | locale info 调用示例 | Go 调用片段 |
| §4.5 | `localematcher.Match` / `ResolveLocale` 签名 | Go 签名 |
| §4.5 | match 调用示例 | Go 调用片段 |
| §5.2 | `(Locale).Equal` / `(Locale).String` 签名 + 比较示例 | Go 签名 + 调用 |
| §6.2 | Parse / New 行为示例 | Go 调用片段 |
| §6.3 | `Maximize` / `Minimize` 签名 | Go 签名 |

## 9. 引用清单

### formatjs（主参考）

- `.references/formatjs/packages/intl-locale/index.ts` — Locale 类完整实现：`IntlLocaleOptions`、`RELEVANT_EXTENSION_KEYS`、`applyOptionsToTag`、`applyUnicodeExtensionToTag`、`addLikelySubtags`、`removeLikelySubtags`、`maximize/minimize/toString`、所有 getter（`baseName`/`calendar`/`collation`/`caseFirst`/`numeric`/`numberingSystem`/`language`/`script`/`region`/`variants`/`firstDayOfWeek`/`hourCycle`）、所有 method（`getCalendars`/`getCollations`/`getHourCycles`/`getNumberingSystems`/`getTimeZones`/`getTextInfo`/`getWeekInfo`）
- `.references/formatjs/packages/intl-locale/preference-data.ts` — `getCalendarPreferenceDataForRegion`/`getHourCyclesPreferenceDataForLocaleOrRegion`/`getTimeZonePreferenceForRegion`/`getWeekDataForRegion`，全部以 region 为键查 CLDR 表，不存在则回退 `001`（world）
- `.references/formatjs/packages/intl-locale/get_internal_slots.ts` — WeakMap-based slots，惰性创建
- `.references/formatjs/packages/intl-locale/tests/index.test.ts`、`likely-subtags.test.ts`、`minimize.test.ts` — fixture 来源
- `.references/formatjs/packages/intl-localematcher/index.ts` — `match(requested, supported, defaultLocale, opts)` 公共入口
- `.references/formatjs/packages/intl-localematcher/abstract/utils.ts` — 三层 `findBestMatch` 实现 + `findMatchingDistance` + `DEFAULT_MATCHING_THRESHOLD = 838`（关键文件）
- `.references/formatjs/packages/intl-localematcher/abstract/BestFitMatcher.ts` — `BestFitMatcher` 包装
- `.references/formatjs/packages/intl-localematcher/abstract/LookupMatcher.ts` — sec-LookupMatcher 实现
- `.references/formatjs/packages/intl-localematcher/abstract/BestAvailableLocale.ts` — subtag truncation
- `.references/formatjs/packages/intl-localematcher/abstract/ResolveLocale.ts` — 完整 resolution + relevantExtensionKeys 处理
- `.references/formatjs/packages/intl-localematcher/abstract/languageMatching.ts` — CLDR matching data 表（paradigmLocales: en, en_GB, es, es_419, pt_BR, pt_PT；matchVariables: $enUS, $cnsar, $americas, $maghreb；distance: 5/10/20/30/80）
- `.references/formatjs/packages/intl-localematcher/tests/` — 公开测试基线
- `.references/formatjs/packages/intl-getcanonicallocales/` — `likelySubtags`、`parseUnicodeLanguageId`、`emitUnicodeLocaleId`、`isStructurallyValidLanguageTag`

### PHP ext/intl（scope check）

- `.references/ext/src/ecma402/locale.{cpp,h}` — `ecma402_locale` struct 全字段平铺；`ecma402_getCalendar`、`ecma402_getCaseFirst`、`ecma402_maximize`、`ecma402_minimize`、`ecma402_bestAvailableLocale`、`ecma402_canonicalizeLocaleList`、`ecma402_canonicalizeUnicodeLocaleId`
- `.references/ext/src/ecma402/language_tag.{cpp,h}` — BCP 47 解析（ICU 调用）

### Go prior art

- `.references/intl/intl.go` — translate-agent/intl 直接基于 `language.Tag`，定义 `Era`/`Year`/`Month`/`Day` 等枚举类型，不实现 best-fit
- `.references/intl/internal/cldr/locale.go` — `language.MustParseBase` 硬编码常量列表（AF, AGQ, AK, AM, …），展示"嵌入 base 列表"的最简方案

### Go 生态

- `golang.org/x/text/language` — `Tag`、`Matcher`、`MatchStrings`、`LikelyScript`、`LikelyRegion`、`TypeForKey`/`SetTypeForKey`、`Parent`
- `golang.org/x/text/language/display` — Locale 展示名

### 内部文档

- `SPECS/00-vision-and-scope.md` §4（Locale 模型决议）、§8 Q3 / Q5（开放问题）
- `ANALYSIS.md` §2 — 调研方向（Locale + matcher）
- `task.md` §r02 — 调研任务定义
