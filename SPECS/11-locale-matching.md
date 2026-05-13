# SPEC 11 — Locale Matching

> **Status:** Draft (2026-05-08)
> **Priority:** High(所有 formatter 必经的 locale 解析层;阻塞 SPEC 20 / 30 / 40 / 60)
> **Authority:** 本 SPEC 是 `internal/localematcher` 包、`CanonicalizeLocaleList`、`LookupMatchingLocaleByPrefix`、`LookupMatchingLocaleByBestFit`、`ResolveOptions`、`ResolveLocale`、`FilterLocales`、UTS #35 距离表与三层匹配算法的 SSOT。Normative source: `.references/ecma402/spec/negotiation.html`.

---

## Overview

ECMA-402 locale negotiation 是所有 formatter(`NumberFormat` / `DateTimeFormat` / `PluralRules`)在初始化阶段必然执行的 locale 选择算法。它把 `locales` 参数规范化为 Language Priority List,用 `options.localeMatcher` 在实现支持的 locale 集合中选择 locale,并把 `-u-` 扩展键(`ca` / `nu` / `hc` / ...)与 options override 合并为最终 resolved 字段。

本 SPEC 决定:

1. **不**复用 `golang.org/x/text/language.Matcher`(输出 `Confidence` 而非 CLDR distance,与 ECMA-402 conformance 不可比)。
2. 在 `internal/localematcher/` **自实现**ECMA-402 lookup,并在 best-fit 的 implementation-defined 部分移植 formatjs `intl-localematcher` 的三层匹配算法 + UTS #35 distance 表。
3. `LookupMatchingLocaleByBestFit` 是 `LookupMatchingLocaleByPrefix` 的增强;两者共享 subtag truncation 与 canonicalization 子例程。
4. `ResolveOptions` 读取 `locales` 和 `options.localeMatcher`,调用 `ResolveLocale`,并返回 options object、resolved locale record、resolution option record 的 Go 等价物。
5. `FilterLocales` 是 constructor `supportedLocalesOf` 的唯一语义源。

本 SPEC **不**定义 `Locale` 类型本身(SPEC 10)、`-u-` 扩展键的字段映射(SPEC 10 §1)、CLDR 数据格式与生成器(SPEC 50)、formatter 的 internal slots(SPEC 12)。

## 0. ECMA-402 Alignment

Go APIs do not accept arbitrary JavaScript values, but the semantic pipeline must match ECMA-402:

| ECMA-402 operation | Go responsibility |
|--------------------|-------------------|
| `CanonicalizeLocaleList(locales)` | Parse/canonicalize/deduplicate requested `locale.Locale` values while preserving order |
| `ResolveOptions(constructor, localeData, locales, options, ...)` | Resolve locale/options for formatter constructors from typed `Options` |
| `LookupMatchingLocaleByPrefix` | Implement RFC 4647 lookup ignoring Unicode extension sequences |
| `LookupMatchingLocaleByBestFit` | Implementation-defined best-fit, using the selected formatjs/UTS #35 distance algorithm |
| `ResolveLocale` | Merge matched locale data, relevant extension keys, Unicode extension requests, and explicit option overrides |
| `FilterLocales` | Implement `Intl.<Constructor>.supportedLocalesOf` |

Rules:

1. `localeMatcher` accepts exactly `"lookup"` and `"best fit"`, defaulting to `"best fit"`.
2. Unicode extension keys influence resolved locale only when they are in the constructor's relevant extension key list.
3. Explicit options override Unicode extension values when both are supported.
4. Returned supported locale lists preserve requested order.
5. Unrecognized but well-formed requested locales are ignored for `supportedLocalesOf` and fall back through `ResolveLocale` for constructors.

---

## 1. Package Layout

### 1.1 决策:私有包 `internal/localematcher`

```text
internal/localematcher/
├── match.go        ← 公共入口 Match / ResolveLocale 与 Algorithm 枚举
├── lookup.go       ← LookupMatcher(ECMA-402 §9.2.6) + BestAvailableLocale
├── best_fit.go     ← BestFitMatcher 三层算法(Tier 1 / Tier 2 / Tier 3)
├── distance.go     ← findMatchingDistance + UTS #35 表查询 + sync.Map memoize
├── resolve.go      ← ResolveLocale(ECMA-402 §9.2.7)+ relevantExtensionKeys 处理
├── ucanonicalize.go← UnicodeExtensionValue / InsertUnicodeExtensionAndCanonicalize
├── canonicalize.go ← CanonicalizeLocaleList(ECMA-402 §6.2.1)
└── data.go         ← languageMatching / paradigmLocales / matchVariables 数据 accessor 包装
```

> **Why**:
> 1. **`internal/`** —— matcher 是 formatter 入口的实现细节,不向用户暴露;SPEC 00 §3 已定 `internal/localematcher`。
> 2. **文件按 ECMA-402 abstract operation 切分** —— 一个文件对应 spec 的一节(`LookupMatcher.ts` / `BestFitMatcher.ts` / `ResolveLocale.ts`),与 formatjs 1:1 镜像,降低移植成本。
> 3. **数据通过 `data.go` 间接访问 `internal/cldr`** —— matcher 不直接 `import internal/cldr`,而通过 `data.go` 提供的接口(SPEC 50 §6 Data Access API);避免 SPEC 11 与 SPEC 50 的循环依赖。
>
> **Rejected**:
> - **公开 `localematcher` 包** —— 用户从不直接构造 matcher;`internal/` 强制隐藏实现。
> - **单文件 `match.go` 全塞** —— spec 三个核心算子(Lookup / BestFit / Resolve)各自有移植成本与测试 fixture;按文件切分便于 fixture 对齐。

### 1.2 公共入口签名

```go
// internal/localematcher/match.go(签名)
package localematcher

// Algorithm 是 ECMA-402 §9.2.5 的 LocaleMatcher 选项。
type Algorithm int

const (
    AlgorithmLookup  Algorithm = iota // sec-LookupMatcher(精确 + subtag truncation)
    AlgorithmBestFit                  // sec-BestFitMatcher(三层 + UTS #35 distance)
)

// Result 是 matcher 内部产物,不是 ResolvedLocale;后者由 ResolveLocale 组合。
type Result struct {
    Locale     string // 命中的 supported locale 串(canonical BCP 47)
    DataLocale string // 用于查 CLDR 数据的 locale(去 -u- 扩展)
    Distance   int    // 0 = 完全等价;>= threshold 则视作未命中
}

// Match 选择 supported locales 中与 requested 列表最匹配的一个。
// alg = AlgorithmLookup 走 LookupMatcher;alg = AlgorithmBestFit 走 BestFitMatcher。
// 任一 algorithm 都不返回错误(无匹配时 Result{Locale: defaultLocale})。
func Match(requested, supported []string, defaultLocale string, alg Algorithm) Result

// DefaultMatchingThreshold 是 BestFitMatcher Tier 3 距离阈值(formatjs DEFAULT_MATCHING_THRESHOLD)。
// 距离 >= 该阈值则视作"超出可接受范围",回退到 defaultLocale。
const DefaultMatchingThreshold = 838
```

调用示例:

```go
res := localematcher.Match(
    []string{"zh-TW"},
    []string{"zh-Hans", "zh-Hant", "en"},
    "en", localematcher.AlgorithmBestFit,
)
fmt.Println(res.Locale)   // "zh-Hant"(zh-TW maximize 到 zh-Hant-TW;subtag 截断回退)
fmt.Println(res.Distance) // 0
```

Formatter callers **必须**传入与 formatter payload 对齐的 supported list:

| Formatter | Supported list source |
|-----------|-----------------------|
| NumberFormat / PluralRules | `internal/cldr.NumberSupportedLocales()` |
| DateTimeFormat | `internal/cldr.DateSupportedLocales()` |
| Locale-only operations | `internal/cldr.AvailableLocales()` |

`internal/cldrmatch` 是该映射的唯一封装点;公共 formatter 包禁止维护自己的 supported-locale 列表。

> **Why `Algorithm` 是 int 枚举而非 string**:Go iota 比字符串校验更廉价;ECMA-402 spec 文本是 `"lookup"` / `"best fit"`(含空格),边界处用 `parseAlgorithm(string) (Algorithm, error)` 转换,内部一律 int。
>
> **Why `DefaultMatchingThreshold = 838` 而不是 spec verbatim 数值**:formatjs `intl-localematcher/abstract/utils.ts` 定义 `DEFAULT_MATCHING_THRESHOLD = 838`;sec-BestFitMatcher 不规定具体数值。我们沿用 formatjs 数值确保 conformance 测试一致。
>
> **Why formatter-specific supported lists**:CLDR `availableLocales` 是数据全集,不是 formatter payload 集。用 actual generated payload 派生 supported list 可以避免 matcher 命中一个没有 numbers/date 数据的 locale。

---

## 2. LookupMatcher(ECMA-402 §9.2.6)

### 2.1 算法

ECMA-402 §9.2.6 `LookupMatcher`:

```text
1. 对 requestedLocales 列表中的每个 locale L:
   a. noExtensionsLocale := 移除 L 的 -u- / -t- 扩展
   b. availableLocale    := BestAvailableLocale(supportedLocales, noExtensionsLocale)
   c. 若 availableLocale 非 undefined,返回 { locale: availableLocale, extension: L 中的 -u- 段 }
2. 否则返回 { locale: defaultLocale, extension: undefined }
```

`BestAvailableLocale(availableLocales, locale)`(§9.2.4):

```text
1. candidate := locale
2. 循环:
   a. 若 candidate ∈ availableLocales,返回 candidate
   b. pos := candidate 中最后一个 '-' 的下标;若不存在,返回 undefined
   c. 若 pos >= 2 且 candidate[pos-2] = '-',pos -= 2(去掉单字符 subtag,如 'en-x')
   d. candidate := candidate[:pos]
```

### 2.2 Go 签名

```go
// internal/localematcher/lookup.go(签名)

// LookupMatcher 实现 ECMA-402 §9.2.6。
func LookupMatcher(requested, supported []string, defaultLocale string) Result

// BestAvailableLocale 实现 ECMA-402 §9.2.4。
// 返回空串表示无匹配。
func BestAvailableLocale(supported []string, locale string) string
```

> **Why subtag truncation 是循环而非递归**:ECMA-402 用伪代码描述循环;Go 风格也偏好循环;且 BCP 47 标签深度有界(实际 ≤ 5 subtag),无栈溢出风险。
>
> **Why 单独导出 `BestAvailableLocale`**:`LookupSupportedLocales`(§9.2.8)、`ResolveLocale`(§9.2.7)、`BestFitMatcher` Tier 2 都需要 subtag truncation 子例程;DRY。

### 2.3 边界情形

| 输入 | 期望行为 |
|------|---------|
| `requested = []` | 返回 `{Locale: defaultLocale, Distance: 0}` |
| `requested = ["xx-INVALID"]`,`supported = ["en"]` | 截断后无匹配,回退 `defaultLocale` |
| `requested = ["en-US-u-ca-buddhist"]`,`supported = ["en-US"]` | `noExtensionsLocale = "en-US"` 命中;extension `-u-ca-buddhist` 由 `ResolveLocale` 后续处理 |
| `requested = ["en-x-private"]`,`supported = ["en"]` | subtag truncation 跳过单字符 `x` 直接去掉 `x-private`,命中 `en` |

---

## 3. BestFitMatcher(ECMA-402 §9.2.5)

### 3.1 ECMA-402 文本

ECMA-402 §9.2.5 `BestFitMatcher`:

> The algorithm used to determine the best fit is implementation-defined. Conforming implementations are recommended to satisfy the following criteria: [...]

即 spec **不强制**具体算法。这给三个选项,本 SPEC 的决策与拒绝清单:

| 候选 | 决策 | 理由 |
|------|------|------|
| 移植 formatjs 三层算法(Tier 1 精确 / Tier 2 maximize+truncation / Tier 3 UTS #35) | ✅ 选定 | 已有公开测试基线;与 conformance 目标对齐 |
| 复用 `golang.org/x/text/language.Matcher` | ❌ 拒绝 | 输出 `Confidence`(No/Low/High/Exact)不是 CLDR distance;tie-breaking 由 `x/text` 决定,与 formatjs 不一致;`x/text` 的 CLDR 数据版本与 `internal/cldr` 不同步 |
| ICU-only 简化启发式 | ❌ 拒绝 | 没有公开 fixture;无法从 conformance 测试反向验证 |

> **Why formatjs 算法**:
> 1. **conformance 是硬约束** —— SPEC 70 要求 byte-equality 通过 formatjs `intl-localematcher/tests/conformance.test.ts`;只有移植同算法才能通过。
> 2. **CLDR distance 是有意义的标量** —— `Confidence` 是 4 级离散,不能区分 "es-MX vs es-419" 之类细粒度差异;CLDR distance(0-840+)能。
> 3. **可独立维护** —— `internal/localematcher` 的三层是 ~500 LOC,fixture 全在 formatjs;升级 CLDR 时只要重跑 fixture 即可发现回归。
>
> **Rejected `language.Matcher`**:
> - ❌ `language.Matcher` 的 tie-breaking 用 `Confidence` + `x/text` 内部实现细节,**不可见**给调用方,无法 byte-match formatjs。
> - ❌ `x/text` 的数据版本与 ECMA-402 / formatjs 钉版不构成同一 conformance 基线(SPEC 50 §1 锁定 CLDR 48.1.0)。
> - ❌ 引入第二个 CLDR 数据源,违反 SSOT。
>
> **退路**:consumer-driven expansion 可暴露 `localematcher.WithLanguageMatcher()` 选项让用户切到 `x/text`(适用于"我只想要 `x/text` 的行为"场景);active scope 不实现。

### 3.2 三层算法

`findBestMatch(requested, supported, threshold)`(formatjs `intl-localematcher/abstract/utils.ts` §findBestMatch):

```text
let lowestDistance = +Inf
let result = { matchedDesired: "", matchedSupported: "", distances: {} }

# === TIER 1 — 精确匹配快路径 ===
for i, desired in requested:
    if desired ∈ supportedSet:
        distance = 0 + i*40
        result.distances[desired] = { desired: distance }
        if distance < lowestDistance:
            lowestDistance = distance
            result.matchedDesired = desired
            result.matchedSupported = desired
        if i == 0:
            return result   # 第一个 requested 命中,立即返回

# === TIER 2 — maximize + 后缀截断 ===
for i, desired in requested:
    maximized = Locale(desired).maximize().toString()
    if maximized != desired:
        candidates = getFallbackCandidates(maximized)  # ["zh-Hant-TW","zh-Hant","zh"]
        for j, candidate in candidates:
            if candidate == desired: continue          # Tier 1 已查
            if candidate ∈ supportedSet:
                # 比较 candidate 与 desired 的 maximize 结果是否一致
                candidateMax = Locale(candidate).maximize().toString()
                distance = (candidateMax == maximized) ? (0 + i*40) : (j*10 + i*40)
                if distance < lowestDistance:
                    lowestDistance = distance
                    result.matchedDesired   = desired
                    result.matchedSupported = candidate
                break

if result.matchedSupported != "" and lowestDistance == 0:
    return result   # Tier 2 找到 distance=0,直接返回

# === TIER 3 — 完整 UTS #35 距离计算 ===
lowestDistance = +Inf  # 重置:Tier 2 的"位置惩罚"与 CLDR distance 不可比
for i, desired in requested:
    for k, candidate in supported:
        d = findMatchingDistance(desired, candidate)
        finalDistance = d + i*40
        result.distances[desired][candidate] = finalDistance
        if finalDistance < lowestDistance:
            lowestDistance = finalDistance
            result.matchedDesired   = desired
            result.matchedSupported = candidate

if lowestDistance >= threshold:
    result.matchedDesired   = ""
    result.matchedSupported = ""
return result
```

> **Why Tier 3 reset `lowestDistance`**:Tier 2 的距离是"subtag 移除位置 × 10 + 请求顺序 × 40"启发式,与 Tier 3 的 CLDR distance(基于 paradigmLocales / matchVariables 表)是不同标度。混用会导致 "Tier 2 给 es→es-MX = 20,Tier 3 给 es-419 = 39",误判 es 优于 es-419(后者在 CLDR 表里更近)。
>
> **Why position penalty `i*40`**:formatjs verbatim;ECMA-402 不支持 `Accept-Language` 的 `q=0.1` 加权,但保留请求顺序作为弱优先级。

### 3.3 Go 签名

```go
// internal/localematcher/best_fit.go(签名)

// BestFitMatcher 实现 ECMA-402 §9.2.5(formatjs 三层算法)。
func BestFitMatcher(requested, supported []string, defaultLocale string) Result

// findBestMatch 是核心三层入口,被 BestFitMatcher 包装。
type bestMatchResult struct {
    matchedDesired   string
    matchedSupported string
    distances        map[string]map[string]int
}

func findBestMatch(requested, supported []string, threshold int) bestMatchResult

// getFallbackCandidates 把 maximize 结果按 right-to-left subtag 截断生成候选。
// 例:"zh-Hant-TW" → ["zh-Hant-TW", "zh-Hant", "zh"]
func getFallbackCandidates(maximized string) []string

// findMatchingDistance 走 UTS #35 distance 表(memoized 通过 sync.Map)。
func findMatchingDistance(desired, supported string) int
```

> **Why `findBestMatch` 不导出**:它返回 `distances` map 用于调试 / fixture 比对,但生产代码只用 `BestFitMatcher` 的 `Result`;不导出避免 ABI 暴露。
>
> **Why `findMatchingDistance` memoize 用 `sync.Map`**:在 formatRange / 大批量请求场景重复计算同一对 (desired, supported) 距离;`sync.Map` 在并发读多写少下零锁开销。

### 3.4 UTS #35 Distance Table

`findMatchingDistance` 依赖 CLDR `languageMatching.json` 派生表:

| 数据项 | 值范围 | 来源 |
|-------|-------|------|
| `paradigmLocales` | 集合: `{en, en-GB, es, es-419, pt-BR, pt-PT}` | CLDR `cldr-core/supplemental/languageMatching.json` |
| `matchVariables` | `$enUS` / `$cnsar` / `$americas` / `$maghreb` 等区域宏 | 同上 |
| 距离权重 | language=80 / script=20-50 / region=4-50 | 同上(取决于 paradigm 与 variable 命中) |

数据由 SPEC 50 §6 codegen 输出到 `internal/cldr/locale_matching.go`,本 SPEC 通过 `data.go` accessor 间接访问。

```go
// internal/localematcher/data.go(签名)

// 间接 accessor —— 不直接 import internal/cldr,避免 SPEC 11 ↔ SPEC 50 循环引用。
type LanguageMatchingData interface {
    ParadigmLocales() map[string]struct{}
    MatchVariables(name string) []string
    DistanceFor(desiredLSR, supportedLSR string) int
}

func data() LanguageMatchingData  // 单例;由 internal/cldr 注入
```

> **Why interface boundary**:SPEC 50 codegen 直接生成 `internal/cldr` 实现 `LanguageMatchingData` 的具体类型;`internal/localematcher` 仅依赖接口。这样 SPEC 50 的数据更新(CLDR 升级)不强制 SPEC 11 改代码。

### 3.5 Performance Targets

| 场景 | 目标 |
|------|------|
| Tier 1 命中 | < 200 ns(supportedSet 是 `map[string]struct{}`) |
| Tier 2 命中 | < 5 µs / 单次匹配(包括 maximize + 截断 + lookup) |
| Tier 3 完整(supported 列表 = 100) | < 100 µs(memoize 命中后的稳态) |
| Tier 3 cold(memoize 未命中) | < 1 ms / 单次匹配 |

> **Why 这些数字**:SPEC 71 §benchmark 把 matcher 列为 NumberFormat / DateTimeFormat 构造时的固定开销;构造一次 < 10 µs 是合理目标。

---

## 4. ResolveLocale(ECMA-402 §9.2.7)

### 4.1 算法

ECMA-402 §9.2.7 `ResolveLocale(availableLocales, requestedLocales, options, relevantExtensionKeys, localeData)`:

```text
1. matcher := options["localeMatcher"]               # "lookup" | "best fit"
2. r := (matcher == "lookup") ? LookupMatcher(...) : BestFitMatcher(...)
3. foundLocale := r.locale
4. result := { dataLocale: foundLocale }
5. supportedExtension := "-u"
6. for each key ∈ relevantExtensionKeys:
   a. foundLocaleData := localeData[foundLocale][key]
   b. value           := foundLocaleData[0]   # locale 默认
   c. supportedExtensionAddition := ""
   d. 若 r.extension 非 undefined:
        i.  requestedValue := UnicodeExtensionValue(r.extension, key)
        ii. 若 requestedValue ∈ keyLocaleData,且 (key 不需要规范化或 requestedValue 已规范),
            value := requestedValue;supportedExtensionAddition := "-" + key + "-" + value
   e. 若 options[key] 不为 undefined 且 ∈ keyLocaleData,
        覆盖 value(options 优先级 > extension)
   f. result[key] := value
   g. supportedExtension += supportedExtensionAddition
7. 若 supportedExtension != "-u":
     foundLocale := InsertUnicodeExtensionAndCanonicalize(foundLocale, supportedExtension)
8. result.locale := foundLocale
9. return result
```

### 4.2 Go 签名

```go
// internal/localematcher/resolve.go(签名)

// ResolveOptions 是 ResolveLocale 的输入(options + 上下文)。
type ResolveOptions struct {
    Algorithm             Algorithm                  // "lookup" | "best fit"
    Requested             []string                   // CanonicalizeLocaleList 输出
    Supported             []string                   // available locales
    DefaultLocale         string                     // matcher fallback
    RelevantExtensionKeys []string                   // 例:`["nu"]`(NumberFormat)
    Options               map[string]string          // user-provided -u- overrides
    LocaleData            LocaleDataLookup           // 从 internal/cldr 注入
}

// LocaleDataLookup 是 SPEC 12 §5 / SPEC 50 §6 数据访问接口的子集。
type LocaleDataLookup interface {
    // For 返回 locale 在 key 上的合法值列表,首项为该 locale 的默认。
    For(locale, key string) []string
}

// ResolvedLocale 是 ResolveLocale 的最终产物;formatter 把它写入 internal slot。
type ResolvedLocale struct {
    Locale     string            // 含 supportedExtension 的 canonical BCP 47
    DataLocale string            // 用于查 CLDR 数据(无 -u-)
    Extensions map[string]string // relevantExtensionKeys 的最终值(option > extension > default)
}

// ResolveLocale 实现 ECMA-402 §9.2.7。
func ResolveLocale(opts ResolveOptions) ResolvedLocale
```

调用示例:

```go
res := localematcher.ResolveLocale(localematcher.ResolveOptions{
    Algorithm:             localematcher.AlgorithmBestFit,
    Requested:             []string{"zh-TW-u-nu-hanidec"},
    Supported:             []string{"zh-Hant", "en"},
    DefaultLocale:         "en",
    RelevantExtensionKeys: []string{"nu"},
    LocaleData:            cldr.NumberFormatLocaleData(),
})
fmt.Println(res.Locale)            // "zh-Hant-u-nu-hanidec"
fmt.Println(res.DataLocale)        // "zh-Hant"
fmt.Println(res.Extensions["nu"])  // "hanidec"
```

> **Why `Extensions` 是 `map[string]string`**:`relevantExtensionKeys` 是动态(NumberFormat 用 `["nu"]`,DateTimeFormat 用 `["ca","nu","hc"]`,PluralRules 用 `["nu"]`);硬编码字段会重复定义。
>
> **Why `LocaleData` 是接口而不是具体 type**:同 §3.4 — 解耦 SPEC 11 与 SPEC 50 的实现细节。formatter 在 `New()` 时把 `internal/cldr.NumberFormatLocaleData()` 等具体类型传入。

### 4.3 InsertUnicodeExtensionAndCanonicalize

`InsertUnicodeExtensionAndCanonicalize(locale, extension)`(formatjs `intl-localematcher/abstract/InsertUnicodeExtensionAndCanonicalize.ts`):

```go
// internal/localematcher/ucanonicalize.go(签名)

// InsertUnicodeExtensionAndCanonicalize 把 supportedExtension(如 "-u-nu-hanidec")
// 插入 locale,并执行 ECMA-402 §6.2.3 CanonicalizeUnicodeLocaleId。
//
// 关键点:
//   - 如果 locale 已有 -u-,合并(option 优先;后续 -u-key- 替换前面的)
//   - 如果 locale 有 -t- 等其他 extension,-u- 的位置必须在它们之后(BCP 47 排序)
//   - canonicalize 后,-u- 内部 keys 按字典序(ca < co < fw < hc < kf < kn < nu)
func InsertUnicodeExtensionAndCanonicalize(locale, extension string) string

// UnicodeExtensionValue 从 -u- 段读取 key 的 type 值(ECMA-402 §6.2.2)。
// 例:UnicodeExtensionValue("-u-ca-buddhist-hc-h23", "ca") = "buddhist"
// 缺省 type(`-u-kn` 单独写)返回 "true"。
func UnicodeExtensionValue(extension, key string) string
```

> **Why 包装 `language.Tag.SetTypeForKey`**:`x/text` 的 `SetTypeForKey` 接受单 key,不做 dict-order 强制;BCP 47 spec 与 ECMA-402 都要求 keys 按字典序;我们在 `InsertUnicodeExtensionAndCanonicalize` 内做最终排序。
>
> **Rejected**:直接调 `language.Tag.SetTypeForKey` 多次拼接 —— 不保证 dict-order,违反 spec。

---

## 5. CanonicalizeLocaleList(ECMA-402 §6.2.1)

### 5.1 算法

ECMA-402 §6.2.1 `CanonicalizeLocaleList(locales)`:

```text
1. 若 locales = undefined,返回 []
2. seen := []
3. 若 typeof locales = "string" 或 Locale 实例,O := [locales];否则 O := ToObject(locales)
4. for k 从 0 到 O.length-1:
   a. tag := O[k]
   b. 若 tag 不是 string / Locale,RangeError
   c. 若 tag 是 Locale,canonicalizedTag := tag.toString();否则 canonicalizedTag := CanonicalizeUnicodeLocaleId(tag)
   d. 若 canonicalizedTag ∉ seen,seen.push(canonicalizedTag)
5. 返回 seen
```

### 5.2 Go 签名

```go
// internal/localematcher/canonicalize.go(签名)

// CanonicalizeLocaleList 实现 ECMA-402 §6.2.1。
// 接受 []string / []locale.Locale / locale.Locale / string 等多形式输入,
// 返回去重后的 canonical BCP 47 字符串切片。
func CanonicalizeLocaleList(locales any) ([]string, error)
```

> **Why `any`**:ECMA-402 接受 polymorphic 输入;Go 无 union types,`any` + 反射 / 类型 switch 是唯一方案。
>
> **Why 返回 error 而非 panic**:无效 BCP 47 标签是用户错误,边界处统一返回 `ErrInvalidLocale`;CLAUDE.md 红线"no panic in production"。

### 5.3 输入规范化

| 输入类型 | 处理 |
|---------|------|
| `nil` | 返回空切片 |
| `string` | 单元素切片,经 `language.Parse` 规范化 |
| `locale.Locale` | 单元素切片,使用 `loc.String()` |
| `[]string` | 逐项 `language.Parse` 规范化 |
| `[]locale.Locale` | 逐项 `loc.String()` |
| 其他 | 返回 `ErrInvalidLocale` |

> **Why 不接受 `[]any`**:Go 风格倾向于具名 slice 类型;混合类型 slice 在 Go 极罕见。要混合,先转 `[]string`。

---

## 6. FilterLocales and LookupSupportedLocales

### 6.1 FilterLocales 用途

`Intl.NumberFormat.supportedLocalesOf(locales, options)` 等 spec 静态方法的语义源。Go typed bridge 接收已经 parse 成 `locale.Locale` 的值,但仍必须按 canonical locale string 去重,再按 `options.localeMatcher` 在 `lookup` 与 `best fit` 之间选择。返回能被支持列表匹配的 requested locale 子集,并保留 requested locale 本身、相对顺序和 Unicode 扩展。

### 6.2 Go 签名

```go
// internal/localematcher/filter.go(签名)

// FilterLocales canonical-deduplicates requested locales and implements
// ECMA-402 FilterLocales.
func FilterLocales(supported []string, requested []locale.Locale, matcher Algorithm) []locale.Locale
```

```go
// 调用示例(由 formatter 包装暴露)
out := localematcher.FilterLocales(
    cldr.NumberSupportedLocales(),
    []locale.Locale{locale.MustParse("en-US-u-nu-latn"), locale.MustParse("fr-FR")},
    localematcher.AlgorithmBestFit,
)
fmt.Println(out)  // 例如 [en-US-u-nu-latn fr-FR]
```

> **Why 返回 requested locale**:ECMA-402 `FilterLocales` 明确把命中的 `_locale_` 追加到结果,而不是追加 data locale。构造器 `supportedLocalesOf` 是能力探测 API,调用方关心的是自己请求的哪些 locale 可以被当前实现满足。

### 6.3 LookupSupportedLocales(ECMA-402 legacy helper)

低层 lookup-only helper。用于测试与算法分解;constructor `SupportedLocalesOf` **不得**绕过 `FilterLocales` 直接调用它。

```go
// internal/localematcher/lookup.go(签名)

// LookupSupportedLocales 实现 ECMA-402 §9.2.8。
// 对每个 requestedLocale 走 BestAvailableLocale,命中即保留(去 -u- 扩展)。
func LookupSupportedLocales(supported, requested []string) []string
```

```go
// 调用示例(由 formatter 包装暴露)
out := localematcher.LookupSupportedLocales(
    cldr.NumberSupportedLocales(),
    []string{"zh-TW", "fr-FR", "xx-INVALID"},
)
fmt.Println(out)  // 例如 ["zh-TW", "fr-FR"]
```

---

## 7. 错误处理

### 7.1 哨兵错误

```go
// internal/localematcher/errors.go(签名)
package localematcher

import "errors"

// 边界错误(从 CanonicalizeLocaleList 等公共入口返回)。
// 业务错误经 fmt.Errorf("%w", ...) wrap 这些哨兵。
var (
    ErrInvalidLocale = errors.New("localematcher: invalid locale")
)
```

> **Why 仅 1 个哨兵**:matcher 内部 `Match` / `ResolveLocale` **不返回错误**(spec 要求始终回退到 `defaultLocale`);只有 `CanonicalizeLocaleList` 在输入解析阶段会失败。

### 7.2 与 `locale` 包错误协调

`locale.ErrInvalidLocale`(SPEC 10)与 `localematcher.ErrInvalidLocale` 是**同一**哨兵的 re-export;`locale` 包是 SSOT。`internal/localematcher` 通过 `var ErrInvalidLocale = locale.ErrInvalidLocale` 重导出。

> **Why re-export 而非定义两个**:用户通过 formatter 包 catch 错误时,`errors.Is(err, locale.ErrInvalidLocale)` 应该匹配两侧;两个独立哨兵会破坏这个等价。

---

## 8. 与 SPEC 12 / SPEC 50 的边界

### 8.1 与 SPEC 12 的关系

| SPEC | 提供 | 消费 |
|------|------|------|
| SPEC 11(本) | `ResolveLocale` 返回 `ResolvedLocale` | 调用 SPEC 12 §3 `CanonicalizeLocaleList` 等 abstract op? **No** —— `CanonicalizeLocaleList` 由本 SPEC 拥有(ECMA-402 §6.2.1) |
| SPEC 12 | 生产路径复用的 validators / pattern / decimal boundary | 不消费 SPEC 11 |

> **决定**:`CanonicalizeLocaleList` 与 `BestAvailableLocale` 等"locale-shape abstract operations" SSOT 在 **SPEC 11**(本)。typed formatter constructors own option validation; SPEC 12 no longer carries the JS options-object pipeline.

### 8.2 与 SPEC 50 的关系

| SPEC | 提供 | 消费 |
|------|------|------|
| SPEC 11(本) | `LanguageMatchingData` / `LocaleDataLookup` 接口 | 通过 `data.go` accessor 间接读 |
| SPEC 50 | 实现接口的具体类型(由 codegen 输出到 `internal/cldr/locale_matching.go` / `internal/cldr/likely_subtags.go`) | 不消费 SPEC 11 |

依赖方向:**SPEC 11 → SPEC 50 接口** + **SPEC 50 实现注入**。无循环。

---

## 9. Forbidden

### 9.1 ❌ 不要复用 `language.Matcher`

```go
// ❌ 错误:输出 Confidence 不是 CLDR distance
matcher := language.NewMatcher([]language.Tag{language.AmericanEnglish, language.Japanese})
tag, _, conf := matcher.Match(language.MustParse("en-GB"))

// ✅ 正确:走 internal/localematcher 三层算法
res := localematcher.Match(
    []string{"en-GB"}, []string{"en-US", "ja"}, "en-US",
    localematcher.AlgorithmBestFit,
)
```

> **Why**:`Confidence`(No/Low/High/Exact)4 级离散,无法 byte-match formatjs `distance` 数值;tie-breaking 由 `x/text` 内部决定不可见。

### 9.2 ❌ 不要在 matcher 中直接 `import internal/cldr`

```go
// ❌ 错误:循环依赖风险 + SPEC 11 与 SPEC 50 紧耦合
import "github.com/agentable/go-intl/internal/cldr"
func findMatchingDistance(...) int {
    return cldr.LanguageMatching.Distance(...)
}

// ✅ 正确:通过 LanguageMatchingData 接口注入
func findMatchingDistance(d, s string) int {
    return data().DistanceFor(toLSR(d), toLSR(s))
}
```

> **Why**:CLDR 数据更新(SPEC 50)不应强制 matcher 改代码;接口隔离允许 SPEC 50 独立演进。

### 9.3 ❌ 不要在 `Match` / `ResolveLocale` 返回 error

```go
// ❌ 错误:违反 ECMA-402 spec("matcher 始终成功")
func Match(...) (Result, error)

// ✅ 正确:无匹配时返回 defaultLocale
func Match(requested, supported []string, defaultLocale string, alg Algorithm) Result
```

> **Why**:ECMA-402 §9.2.5 / §9.2.6 都规定"无匹配返回 defaultLocale";让调用方处理 error 增加 cognitive load 且与 spec 不符。`CanonicalizeLocaleList` 是唯一可能 error 的边界(用户输入解析)。

### 9.4 ❌ 不要 panic

```go
// ❌ 错误:违反 CLAUDE.md "no panic in production"
func BestAvailableLocale(supported []string, locale string) string {
    if locale == "" {
        panic("empty locale")
    }
    // ...
}

// ✅ 正确:返回空串(spec 语义"无匹配")
func BestAvailableLocale(supported []string, locale string) string {
    if locale == "" {
        return ""
    }
    // ...
}
```

### 9.5 ❌ 不要在 BestFitMatcher 内重复实现 maximize

```go
// ❌ 错误:在 best_fit.go 内手写 likelySubtags 表查询
func tier2Maximize(loc string) string { /* 自己查表 */ }

// ✅ 正确:复用 locale.Locale.Maximize()(SPEC 10 §4)
func tier2Maximize(loc string) string {
    l, err := locale.Parse(loc)
    if err != nil { return loc }
    return l.Maximize().String()
}
```

> **Why**:maximize 算法 SSOT 在 SPEC 10 / SPEC 50(`MaximizeSubtags` 表);matcher 复用避免数据不同步。

### 9.6 ❌ 不要把 Tier 2 与 Tier 3 距离混合比较

```go
// ❌ 错误:Tier 2 跳过 Tier 3 时 lowestDistance 未重置
if tier2Distance < threshold { return tier2Result }
// Tier 3 用 tier2 的 lowestDistance 作初值 —— 误判

// ✅ 正确:Tier 3 入口重置 lowestDistance = +Inf
lowestDistance = math.MaxInt
for ... { /* Tier 3 */ }
```

> **Why**:Tier 2 距离是启发式(subtag 位置 × 10),Tier 3 是 CLDR 实测距离(0-840);两者不在同一标度。formatjs `utils.ts` 注释明确指出。

### 9.7 ❌ 不要在 `findMatchingDistance` 不 memoize

```go
// ❌ 错误:每次匹配重新查 CLDR 表
func findMatchingDistance(d, s string) int {
    return data().DistanceFor(toLSR(d), toLSR(s))  // 调用 100 次 = 查 100 次
}

// ✅ 正确:sync.Map memoize
var distanceCache sync.Map  // map[[2]string]int
func findMatchingDistance(d, s string) int {
    key := [2]string{d, s}
    if v, ok := distanceCache.Load(key); ok { return v.(int) }
    dist := data().DistanceFor(toLSR(d), toLSR(s))
    distanceCache.Store(key, dist)
    return dist
}
```

> **Why**:Tier 3 在 100-locale 场景需要 10000 次距离计算;memoize 把稳态降到 0 µs/次。

### 9.8 ❌ 不要通过比较字符串等价 `Result`

```go
// ❌ 错误:Result.Distance 字段也参与比较,但用户只关心 Locale
if res1 == res2 { /* ... */ }

// ✅ 正确:用 Result.Locale 字符串比较
if res1.Locale == res2.Locale { /* ... */ }
```

> **Why**:`Distance` 在 fixture 测试中可能不同(不同 CLDR 版本);用 `Locale` 比较更稳定。

---

## 10. Acceptance Criteria

### 包结构

- [ ] `internal/localematcher/` 子目录下分文件 `match.go` / `lookup.go` / `best_fit.go` / `distance.go` / `resolve.go` / `ucanonicalize.go` / `canonicalize.go` / `data.go` / `errors.go`。
- [ ] `internal/localematcher` 不直接 `import "github.com/agentable/go-intl/internal/cldr"`;通过 `data.go` 接口间接访问。
- [ ] `internal/localematcher` 不直接 `import "github.com/agentable/go-intl/internal/ecma402"`(SPEC 12 是 option-shape;本 SPEC 是 locale-shape;两者无环)。

### LookupMatcher

- [ ] `LookupMatcher(requested, supported, defaultLocale) Result` 实现 ECMA-402 §9.2.6 verbatim。
- [ ] `BestAvailableLocale(supported, locale) string` 实现 ECMA-402 §9.2.4(单字符 subtag 跳过 -2 位置)。
- [ ] formatjs `intl-localematcher/tests/LookupMatcher.test.ts` 全部 fixture 在 `internal/localematcher/lookup_test.go` 通过。

### BestFitMatcher

- [ ] `BestFitMatcher(requested, supported, defaultLocale) Result` 实现三层算法(Tier 1 精确 / Tier 2 maximize+truncation / Tier 3 UTS #35)。
- [ ] `findBestMatch` 在 Tier 3 入口重置 `lowestDistance = +Inf`(不与 Tier 2 启发式混合)。
- [ ] `getFallbackCandidates(maximized)` 输出 `["zh-Hant-TW","zh-Hant","zh"]`(right-to-left subtag 截断)。
- [ ] `findMatchingDistance` 用 `sync.Map` memoize(同一 (desired, supported) 对仅查表一次)。
- [ ] `DefaultMatchingThreshold = 838`(formatjs verbatim)。
- [ ] formatjs `intl-localematcher/tests/BestFitMatcher.test.ts` 与 `tests/conformance.test.ts` 全部 fixture 在 `internal/localematcher/best_fit_test.go` 通过。

### ResolveLocale

- [ ] `ResolveLocale(opts) ResolvedLocale` 实现 ECMA-402 §9.2.7(option > extension > localeData default 优先级)。
- [ ] `relevantExtensionKeys` 是动态参数(NumberFormat 用 `["nu"]`,DateTimeFormat 用 `["ca","nu","hc"]`)。
- [ ] `InsertUnicodeExtensionAndCanonicalize` 输出 `-u-` keys 按字典序(ca < co < fw < hc < kf < kn < nu)。
- [ ] `UnicodeExtensionValue` 在缺省 type 时返回 `"true"`(`-u-kn` 单独写场景)。
- [ ] formatjs `intl-localematcher/tests/ResolveLocale.test.ts` 全部 fixture 通过。

### CanonicalizeLocaleList

- [ ] `CanonicalizeLocaleList(any) ([]string, error)` 接受 `nil` / `string` / `locale.Locale` / `[]string` / `[]locale.Locale`。
- [ ] 输入是无效 BCP 47 标签 → 返回 `ErrInvalidLocale` wrap 的错误,**不**panic。
- [ ] 去重后保留首次出现顺序。

### LookupSupportedLocales

- [ ] `FilterLocales(supported, requested, matcher) []locale.Locale` implements ECMA-402 `FilterLocales`.
- [ ] `FilterLocales` canonical-deduplicates requested locale values before filtering.
- [ ] `FilterLocales` preserves requested locale order and Unicode extensions for matched locales.
- [ ] Constructor package `SupportedLocalesOf` methods call `FilterLocales` rather than duplicating matcher loops.
- [ ] `LookupSupportedLocales(supported, requested) []string` 实现 ECMA-402 §9.2.8。
- [ ] 输出去除 `-u-` 扩展。

### 错误

- [ ] `errors.Is(err, locale.ErrInvalidLocale)` 在 `CanonicalizeLocaleList` 失败下返回 true。
- [ ] `Match` / `ResolveLocale` 不返回 error(始终回退 `defaultLocale`)。
- [ ] 包内**无** `panic` 调用(测试覆盖各种异常输入)。

### 性能

- [ ] Tier 1 命中路径 benchmark < 200 ns(`go test -bench=BenchmarkTier1Match`)。
- [ ] Tier 3 稳态(memoize 命中)100-locale supported 列表 benchmark < 100 µs。
- [ ] `sync.Map` memoize 在并发 10 goroutine 下无 race(`-race` 通过)。

### 与 `language.Matcher` 边界

- [ ] `internal/localematcher` 包内**不**导入 `golang.org/x/text/language.Matcher` / `MatchStrings`(`grep -r "language.Matcher" internal/localematcher/` 应为空)。
- [ ] `Locale.Maximize` / `Minimize`(SPEC 10)在 Tier 2 内被复用(不重复实现 likelySubtags 查询)。

### 测试

- [ ] formatjs `intl-localematcher/tests/locale-match-fixtures.json` 移植到 `internal/localematcher/testdata/match-fixtures.json` 并在 `match_test.go` 表驱动消费。
- [ ] 所有测试用 `t.Parallel()`。
- [ ] 至少 1 个 `Example*` 函数演示 `Match` + `ResolveLocale` 串联。

---

## References

### Specification

- [ECMA-402 §6.2 — Language Tags](https://tc39.es/ecma402/#sec-language-tags)
- [ECMA-402 §9.2 — Locale Resolution](https://tc39.es/ecma402/#sec-locale-resolution)
- [ECMA-402 §9.2.4 — BestAvailableLocale](https://tc39.es/ecma402/#sec-bestavailablelocale)
- [ECMA-402 §9.2.5 — BestFitMatcher](https://tc39.es/ecma402/#sec-bestfitmatcher)
- [ECMA-402 §9.2.6 — LookupMatcher](https://tc39.es/ecma402/#sec-lookupmatcher)
- [ECMA-402 §9.2.7 — ResolveLocale](https://tc39.es/ecma402/#sec-resolvelocale)
- [ECMA-402 §9.2.8 — LookupSupportedLocales](https://tc39.es/ecma402/#sec-lookupsupportedlocales)
- [UTS #35 §EnhancedLanguageMatching](https://unicode.org/reports/tr35/#EnhancedLanguageMatching)

### Reference implementations

- `.references/formatjs/packages/intl-localematcher/abstract/utils.ts` —— 三层 `findBestMatch` 实现 + `findMatchingDistance` + `DEFAULT_MATCHING_THRESHOLD = 838`(关键文件)
- `.references/formatjs/packages/intl-localematcher/abstract/BestFitMatcher.ts` —— `BestFitMatcher` 包装
- `.references/formatjs/packages/intl-localematcher/abstract/LookupMatcher.ts` —— ECMA-402 §9.2.6 实现
- `.references/formatjs/packages/intl-localematcher/abstract/BestAvailableLocale.ts` —— subtag truncation
- `.references/formatjs/packages/intl-localematcher/abstract/ResolveLocale.ts` —— 完整 resolution + relevantExtensionKeys 处理
- `.references/formatjs/packages/intl-localematcher/abstract/InsertUnicodeExtensionAndCanonicalize.ts`
- `.references/formatjs/packages/intl-localematcher/abstract/UnicodeExtensionValue.ts`
- `.references/formatjs/packages/intl-localematcher/abstract/CanonicalizeLocaleList.ts`
- `.references/formatjs/packages/intl-localematcher/abstract/LookupSupportedLocales.ts`
- `.references/formatjs/packages/intl-localematcher/abstract/languageMatching.ts` —— CLDR matching data 表(paradigmLocales: en, en-GB, es, es-419, pt-BR, pt-PT;matchVariables: $enUS, $cnsar, $americas, $maghreb;距离权重)
- `.references/formatjs/packages/intl-localematcher/tests/conformance.test.ts` —— ICU4J 对齐 fixture
- `.references/formatjs/packages/intl-localematcher/tests/locale-match-fixtures.json` —— 表驱动 fixture
- `.references/ext/src/ecma402/locale.cpp` —— PHP `ecma402_bestAvailableLocale` 经 ICU(同 spec 但路径不同)

### Cross-SPEC

- [SPEC 00 §8 Q5 — BestFit Matcher 实现选择](./00-vision-and-scope.md#8-open-questions)(本 SPEC 关闭)
- [SPEC 10 §4 — Maximize & Minimize](./10-locale.md#4-maximize--minimize) —— Tier 2 复用
- [SPEC 10 §1 — Locale 结构](./10-locale.md#1-locale-structure) —— `Locale.String()` 是 matcher 输入
- [SPEC 12 §3 — Option Validation](./12-abstract-operations.md#3-option-validation) —— matcher 不重复实现 formatter option validation
- [SPEC 12 §5 — Internal Slots](./12-abstract-operations.md#5-internal-slots) —— `[[Locale]]` slot 持有 `ResolvedLocale.Locale`
- [SPEC 50 §6 — Data Access API](./50-cldr-data.md#6-data-access-api) —— `LanguageMatchingData` / `LocaleDataLookup` 实现注入
- [SPEC 70 §Conformance](./70-conformance.md) —— matcher fixture 是 conformance 测试一部分

### Research

- `.research/R02-locale-and-matcher.md` §4 —— BestFitMatcher 决策、`language.Matcher` 拒绝理由、UTS #35 数据需求

---

> 本 SPEC 是 `internal/localematcher` 的 SSOT。新增 ECMA-402 matcher 子例程(罕见)或 CLDR `languageMatching.json` 数据结构变化触发本 SPEC 修订;UTS #35 距离表的具体数据更新由 SPEC 50 codegen 完成。
