---
id: R04
title: DateTimeFormat 算法、骨架与时区调研
task: r04
date: 2026-05-08
status: draft
scope:
  - skeleton 字符 → 选项字段映射的归属
  - BasicFormatMatcher / BestFitFormatMatcher 评分函数移植
  - 时区数据策略（系统 tzdata vs 嵌入式 tzdata）
  - Phase 1 日历支持范围
  - time.Time 与时区上下文保持
  - dayPeriod 边界数据来源
tags: [datetimeformat, skeleton, timezone, calendar, cldr]
---

# R04 — DateTimeFormat 算法、骨架与时区调研

## 执行摘要

| 决策点 | 推荐 | 置信度 | 关键依据 |
|--------|------|--------|----------|
| skeleton 字符表归属 | `internal/ecma402/datetimeformat/skeleton.go`（算法常量）+ `internal/cldr/dates.go`（per-locale 模式表） | 高 | formatjs 把正则与映射放在 `ecma402-abstract`（算法），把 per-locale 数据放在 `intl-datetimeformat/scripts/` 抽取产物 |
| FormatMatcher 移植 | 表驱动复刻 `BestFitFormatMatcher`（含 ICU `adjustFieldTypes` 后处理），不引入 V8 的 ICU `DateTimePatternGenerator` | 高 | formatjs 的 `BestFitFormatMatcher` 已是 ICU4J 等价实现，约 100 行 TS，无外部依赖 |
| 时区数据策略 | 嵌入 IANA tzdata，参考 formatjs 的 zone-transition 列表压缩格式（base36 编码 ts/abbrv/offset），版本钉在 `internal/tz/VERSION` | 高 | translate-agent/intl 不处理时区（未实现 hour/min/sec/timezone 选项），无 Go 先例；`time.LoadLocation` 在容器/macOS 部署不一致；formatjs 解包后约 10 MB，但仅取 transition 表后可压至 ~150 KB |
| Phase 1 日历范围 | 仅 Gregorian。Buddhist 通过"年偏移 +543"在 Phase 1 末作为可选扩展，Persian / Islamic / Japanese 推迟到 Phase 3 | 高 | formatjs 的 `ToLocalTime` 显式 `invariant(calendar === 'gregory')`；translate-agent 已经付出额外 codegen 来支持 Persian/Buddhist，但 Persian 依赖第三方包（`yaa110/go-persian-calendar`）且未通过 ECMA-402 conformance 验证 |
| time.Time 处理 | 在 `Format` 入口剥离 monotonic 部分（调用 `t.Round(0)`），保留 `Location`；时区选项独立解析为 IANA 名 / UTC offset 字符串两类 | 高 | ECMA-402 的输入是 timestamp（毫秒数），不携带时区上下文，与 Go `time.Time` 的 `Location` 是冲突的；Go idiom 是 `time.Time` 自带时区，但 ECMA-402 要求"选项里的 timeZone 优先于输入" |
| dayPeriod 边界 | 从 CLDR `dayPeriods.xml`（`cldr-core/supplemental/dayPeriods.json`）按 locale 抽取并嵌入 `internal/cldr/dates.go` | 中 | formatjs 通过 ICU 的 dayPeriodRules 支撑 `b/B` 格式；生成路径已在 `extract-dates.ts` 内但只走了 `noon/midnight`，更细的 `morning1/afternoon1` 需补齐 |

## 1. Skeleton 解析与字段映射

ECMA-402 选项 → CLDR skeleton（如 `yMMMd`）→ pattern（如 `MMM d, y`）的三段式流水线在每个参考实现中都存在，但分层位置不同。

### 1.1 三个参考的分层

`formatjs` 把"字符 → 选项字段"的字典完全写死在算法层 `ecma402-abstract/DateTimeFormat/skeleton.ts`（160+ 行），通过一个统一的 `DATE_TIME_REGEX` 正则匹配并填充 `Intl.DateTimeFormatOptions`：

```ts
// 提炼自 .references/formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts:14-15,38-150
const DATE_TIME_REGEX =
  /(?:[Eec]{1,6}|G{1,5}|[Qq]{1,5}|(?:[yYur]+|U{1,5})|[ML]{1,5}|d{1,2}|D{1,3}|F{1}|[abB]{1,5}|[hkHK]{1,2}|w{1,2}|W{1}|m{1,2}|s{1,2}|[zZOvVxX]{1,4})(?=([^']*'[^']*')*[^']*$)/g

function matchSkeletonPattern(match: string, result: Options): TokenLabel
function processDateTimePattern(rawPattern: string, result: Formats): void
function parseDateTimeSkeleton(skeleton: string, ...): Formats
```

字符表来源是 LDML TR35（Date Field Symbol Table）<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts:9-12 -->。每个 locale 的"哪些骨架可用、对应什么 pattern"则由 `extract-dates.ts` 生成脚本从 CLDR `cldr-dates-full/main/<locale>/ca-gregorian.json` 抽取后保存在每个 locale 的 `formats` map<!-- ref: formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts:170-200 -->。

`V8` 委托 ICU 的 `DateTimePatternGenerator`，在 `js-date-time-format.cc` 中只做选项→ICU UDateFormat 的转换，骨架字符表完全由 ICU 持有。

`translate-agent/intl` 走了一条折中路径：它**没有运行时 skeleton 解析器**，而是通过 codegen 把"option 组合 → 已选好的 pattern 序列"穷举到 `seqEraYearMonthDay / seqYearMonth / ...` 共 16 个分支函数<!-- ref: intl/intl.go:412-446 -->；每个 `seq*` 函数内部从 `internal/symbols/symbols.go` 取出 LDML 字符序列。这个方案省了运行时正则但放弃了"任意 skeleton 字符串输入"的能力——translate-agent 的 `Options` 只接收 4 个布尔（Era/Year/Month/Day），不接收 hour/minute/second/weekday/timeZoneName。

`PHP ext/intl` 走 ICU C++，与 V8 思路一致。

### 1.2 Go 落地：分层归属

把"字符 → 选项字段"的算法（即 `DATE_TIME_REGEX` 与 `matchSkeletonPattern` 的等价物）放在 `internal/ecma402/datetimeformat/skeleton.go`。把"locale 的 formats 表 / dateFormat / timeFormat / intervalFormats"作为生成数据放在 `internal/cldr/dates.go`。这样：

- 升级 LDML 不会触发数据再生（算法稳定）。
- 升级 CLDR 不会触发算法修改（数据稳定）。
- conformance 测试可以独立替换"算法"或"数据"做差分。

不应学 translate-agent 的"穷举分支"——`Intl.DateTimeFormatOptions` 共 11 个互斥/可选字段（weekday/era/year/month/day/dayPeriod/hour/minute/second/fractionalSecondDigits/timeZoneName），二项式展开即 2^11 = 2048 种，再叠加 `dateStyle/timeStyle` 的 5x5 矩阵，codegen 不可行。保留 formatjs 的运行时正则。

### 1.3 BestFitFormatMatcher 评分

formatjs 的 `BestFitFormatMatcher` 在 100 行 TS 内完成<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/BestFitFormatMatcher.ts:1-141 -->：

```ts
// 提炼自 .references/formatjs/packages/ecma402-abstract/DateTimeFormat/BestFitFormatMatcher.ts:30-75
function bestFitFormatMatcherScore(options: DateTimeFormatOptions, format: Formats): number
function BestFitFormatMatcher(options: DateTimeFormatOptions, formats: Formats[]): Formats
```

评分逻辑：
- `hour12` 不匹配扣 `removalPenalty` / `additionPenalty`。
- 每个 `DATE_TIME_PROPS` 字段缺失/多余分别扣 `additionPenalty=20` / `removalPenalty=120`。
- 数值/字母不同扣 `differentNumericTypePenalty`。
- 同类（都是字母或都是数字）按 `[2-digit, numeric, narrow, short, long]` 索引差扣 `longMore/shortMore/shortLess/longLess`。
- 选出最高分后，按 ICU `adjustFieldTypes`（注释链接 ICU4J 源）做一次 pattern 字段替换<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/BestFitFormatMatcher.ts:103-138 -->。

`BasicFormatMatcher` 是规范定义的标准评分（无 best-fit 后处理），逻辑只比 `BestFit` 多了 `timeZoneName` 的细分（short / long / shortOffset / longOffset / shortGeneric / longGeneric）<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/BasicFormatMatcher.ts:30-100 -->。

V8 / ICU 的 `BestFit` 是 `DateTimePatternGenerator::adjustFieldTypes`，与 formatjs 注释链接的算法是同一个；这意味着**直接表驱动复刻 formatjs 即可保持与浏览器对齐**，不需要 ICU。

> 落地建议：在 `internal/ecma402/datetimeformat/matcher.go` 同时实现 `BasicFormatMatcher` 和 `BestFitFormatMatcher`。`localeMatcher: "best fit"` 是规范默认（参考 `InitializeDateTimeFormat`），所以 `BestFit` 必须实现；conformance 测试需要 `Basic` 才能验证打分。

## 2. 时区数据策略

DateTimeFormat 需要时区数据完成两件事：(a) 把 UTC timestamp 转本地年/月/日/时/分/秒（通过 transition 表查出该时刻的 offset 与 DST 标志），(b) 输出 `timeZoneName`（短/长名、generic / specific、offset 形式）。

### 2.1 三种数据形态

formatjs 走的是"自带 zdump 输出 + 二次打包"的路线：

- `tz_data.tar.gz` 共 808 KB，解压后约 10 MB（428 个文件，427 个时区 + `backward` link 表）。
- 真正消费时只取每个时区的 transition 列表，通过 `pack()` 压成 `(zones, abbrvs, offsets)` 三元组，每个 transition 序列化成 `(ts_base36, abbrvIndex, offsetIndex, dst)`；`unpack()` 在运行时反序列化<!-- ref: formatjs/packages/intl-datetimeformat/unpack.ts:4-22, packer.ts:3-19 -->。
- 类型契约 `ZoneData = [ts_seconds, abbrvIndex, offsetIndex, dst_0_or_1]`<!-- ref: formatjs/packages/intl-datetimeformat/types.ts:22-31 -->。
- 实际 packed 后的 JSON 大小（参考 `@formatjs_generated/tz`）约 150 KB（只在 generated package 中体现，无法直接 du，但 transition 数 ~10 KB × ~430 zones × 4 字段，base36 编码后压缩至此量级）。

`translate-agent/intl` 不实现时区（其 `Options` 只有 Era/Year/Month/Day；其 `Format` 调用直接读 `time.Time.Year/Month/Day`，没有时区转换路径）<!-- ref: intl/intl.go:367-403 -->。这意味着 Go 生态在"嵌入式 tzdata + ECMA-402 选项"上没有现成先例。

Go 标准库的 `time.LoadLocation` 提供两条路：(a) 系统 zoneinfo 文件（`/usr/share/zoneinfo`，macOS / Linux 默认有，Alpine / scratch 容器没有），(b) `time/tzdata` 包嵌入 Go 内置 tzdata（默认 ~450 KB，加进 binary）。

### 2.2 选择推荐

**推荐：嵌入 IANA tzdata，复用 `time/tzdata` + 自维护的 transition 索引层。**

依据：
- `time/tzdata` 已经把 Go 1.15+ 的 IANA 数据嵌入进来，开销 ~450 KB，是最小可信源。
- ECMA-402 要求三类输入并存：`"Etc/UTC"` / `"America/New_York"` / `"+05:30"`<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts:97-123 -->。`time.LoadLocation` 处理前两类，`+05:30` 需自行解析（formatjs 用 `OFFSET_TIMEZONE_FORMAT_REGEX` 处理，已是参考实现<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts:17-18,45-74 -->）。
- 但 `time/tzdata` 不向下暴露 transition 表（只暴露 `Location.lookup` 等私有 API），无法直接做 `timeZoneName: "shortGeneric"` 这类需要"该时区在该时刻属于哪个 metazone"的查询。
- 这部分元数据（zone → metazone → 显示名）必须从 CLDR `metaZones.json` 抽取，与 tzdata 是两个独立来源。formatjs 已有完整的 `extractTimezoneToMetazoneMap` 实现可参照<!-- ref: formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts:137-163 -->。

**不推荐：照搬 formatjs 的 zdump 流水线。** 原因：
- 需要 Docker 跑 zic 编译<!-- ref: formatjs/packages/intl-datetimeformat/scripts/execute_tz_docker.sh -->，CI 复杂度高。
- Go 生态有更直接的 `time/tzdata`，重新造轮子不划算。
- `time/tzdata` 与 IANA tzdata 同源，更新由 Go 团队维护。

**不推荐：依赖 `time.LoadLocation` 的系统 zoneinfo。** 原因：
- 部署环境差异（Alpine / scratch）会产生"在 prod 报错但 dev 正常"的失败模式。
- SPEC 00 §5.4 已明确"我们要确定性输出"<!-- ref: SPECS/00-vision-and-scope.md:202-206 -->。

> 落地接口建议（`internal/tz/`）：
>
> ```go
> // 提炼自 formatjs ToLocalTime + time.Location
> type ZoneInfo struct {
>     Name        string  // canonical IANA name, e.g. "America/New_York"
>     OffsetMs    int64   // milliseconds east of UTC
>     IsDST       bool
>     Abbrv       string  // e.g. "EST" / "EDT"
>     Metazone    string  // e.g. "America_Eastern", from CLDR metazoneInfo
> }
>
> func Resolve(zone string) (*Location, error)            // IANA name → Location
> func ParseOffsetString(s string) (int64, error)         // "+05:30" → offsetMs
> func LookupAt(loc *Location, t time.Time) ZoneInfo      // transition lookup at instant t
> func CanonicalLink(s string) string                     // "US/Eastern" → "America/New_York"
> ```

### 2.3 体积估算（高置信度）

| 来源 | 形态 | 解压后体积 | 实际 binary 增量 |
|------|------|-----------|------------------|
| formatjs `tz_data.tar.gz` | gzipped zdump 文本 | 10 MB | ~150 KB（packed 后） |
| formatjs `@formatjs_generated/tz` | TS 源 + base36 编码 | — | ~150 KB（在 generated package 中） |
| Go `time/tzdata` | 嵌入式 IANA | — | ~450 KB（加到 binary） |
| CLDR metaZones JSON | per-locale 显示名 | ~150 KB（en）to ~3 MB（all locales） | 依嵌入策略而定 |

`time/tzdata` 的 450 KB 是 transition 表 + abbrv + offset 全量；formatjs 的 150 KB 更紧凑（只覆盖 1980 - 2099 / "golden zones"）。考虑到 Go binary 通常 10–50 MB，450 KB 增量可忽略；如果未来需要削减，再切换到 formatjs 风格的 packed 表即可（向后兼容）。

## 3. Phase 1 日历范围

### 3.1 三个参考的覆盖

`formatjs` 的 `ToLocalTime` 写死了 `invariant(calendar === 'gregory', 'We only support Gregory calendar right now')`<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts:155-158 -->。也就是说 polyfill 只支持 Gregory；当 `Intl.DateTimeFormat(loc, {calendar: "buddhist"})` 时，formatjs 抛 `RangeError`。

`translate-agent/intl` 实现了 3 种：Gregory / Persian（依赖 `yaa110/go-persian-calendar`）/ Buddhist（年偏移 +543）<!-- ref: intl/intl.go:387-394, 449-491, 494-535 -->。其 Buddhist 实现就是 `t.AddDate(543, 0, 0)`<!-- ref: intl/intl.go:533-535 -->。Persian 依赖外部库做 Solar Hijri 换算。

`PHP ext/intl` 通过 ICU 支持全部。

### 3.2 决策

**Phase 1 仅 Gregorian。** 理由：
1. formatjs 自身只支持 Gregorian，我们的"输出与 formatjs 字节级一致"目标允许这个范围。
2. 非 Gregorian 日历的核心难度不是格式化，而是"日期换算"——把 ICU `UCalendar` 的状态机搬到 Go 需要重写 ICU 的 `Calendar::computeFields`，其代码量约 5000 行 C++。
3. `messageformat-go` 的当前 `:datetime` 函数也只支持 Gregorian。

**Buddhist 作为 Phase 1 末的可选扩展。** 理由：
1. 增量代码 < 50 行（年偏移 +543）。
2. CLDR 提供 `cldr-dates-full/main/<locale>/ca-buddhist.json` 与 Gregorian 同结构。
3. `messageformat-go` 的 ICU MessageFormat parser 已识别 `calendar=buddhist`。

**Persian / Islamic / Japanese / Hebrew / Chinese 推迟到 Phase 3。** 理由：
1. Persian 需要 Solar Hijri 算法（translate-agent 用了 `yaa110/go-persian-calendar`），引入第三方依赖。
2. Islamic 有 4 个变体（civil / umalqura / rgsa / tbla），每个换算规则不同。
3. Japanese 需要"年号表"（Reiwa/Heisei/Showa/Taisho/Meiji），且 ICU 在 2019 年改元后多次发版修订。
4. Hebrew / Chinese 需要月相计算，复杂度独立成 SPEC。

> 落地建议：`SPEC 30 — DateTimeFormat` 第 1 版只列 Gregorian + Buddhist 双日历。`Calendar` 从字符串改为 `type Calendar int` 受控枚举，便于 Phase 3 扩展。

## 4. time.Time 与时区上下文

ECMA-402 的输入是 `t: number | Date`，其中 `number` 是 epoch 毫秒数，`Date` 内部也是 epoch 毫秒数，**都不携带时区**。`Intl.DateTimeFormat(loc, {timeZone: "X"}).format(date)` 中，`X` 是输出时区，与 `date` 自身的时区无关。

Go 的 `time.Time` 自带 `Location`，与 ECMA-402 模型不一致。三种参考都在边界上做了归一化：

- `formatjs` 通过 `TimeClip(x)` 把输入归一到 `Decimal`（毫秒数），随后 `ToLocalTime(t, calendar, timeZone, ...)` 显式查 `tzData` 拿 offset<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/PartitionDateTimePattern.ts:21-41, ToLocalTime.ts:135-183 -->。**输入的时区被丢弃，选项的 `timeZone` 决定输出。**
- `V8` 同样把 `Date` 视作 timestamp。
- `translate-agent/intl` 直接用 `time.Time.Year()/Month()/Day()`<!-- ref: intl/intl.go:401-403 -->，**不剥离 time.Time 的 Location** —— 这是 ECMA-402 偏离：选项里没有 timeZone 字段，输入的 Location 决定输出。

### 4.1 决策

**遵循 ECMA-402：选项的 `TimeZone` 优先；输入 `time.Time` 的 `Location` 仅在选项未指定时才使用。**

```go
// 提炼自落地建议
type Options struct {
    TimeZone string  // "" = 用 input.Location() 或宿主时区
    // ...
}

func (f *DateTimeFormat) Format(t time.Time) string {
    t = t.Round(0) // strip monotonic clock
    loc := f.timeZone   // 已在 New() 解析好
    if loc == nil {
        loc = t.Location()
    }
    instant := t.UTC().UnixMilli()
    // 用 instant + loc 走 ToLocalTime
}
```

**剥离 monotonic clock 是必须的。** 原因：`time.Time.Round(0)` 是 Go 文档推荐做法<!-- ref: pkg.go.dev/time#Time.Round -->；formatjs 走 `Decimal`（毫秒数）不存在这个问题；如果不剥离，`Format(t1) == Format(t1)` 会成立但 `t1.Equal(t2) && Format(t1) != Format(t2)` 不会，破坏可观测性。

**`+05:30` 偏移字符串作为合法 `TimeZone`。** 在 `New` 时调用 `internal/tz.ParseOffsetString` 解析成固定 offset，不走 transition 表（offset zones 永远不 DST，formatjs 的 `getApplicableZoneData` 已有此分支<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts:96-100 -->）。

## 5. dayPeriod 边界

`b/B` skeleton 字符表示"flexible day periods"，输出"morning1 / afternoon1 / evening1 / night1 / noon / midnight"等 locale 特定的口语化时段<!-- ref: formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts:107-111 -->。边界来自 CLDR `cldr-core/supplemental/dayPeriods.json`，按 locale 不同（zh: 凌晨 0–4 / 早上 5–8 / 上午 8–11 / 中午 11–13 / 下午 13–18 / 晚上 18–24；en: morning1 5–12 / afternoon1 12–18 / evening1 18–21 / night1 21–5）。

formatjs 的 `extract-dates.ts` 已在跑 CLDR 抽取，但聚焦于 `noon / midnight`<!-- ref: formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts:1-30 -->；更细的"1/2/3" 段需要从 `dayPeriodRules` 节点读取。

V8 / ICU 通过 `DateFormatSymbols::getNarrowDayPeriods` 等访问。

translate-agent 没实现 `b/B`<!-- ref: intl/intl.go 中无 dayPeriod 路径 -->。

> 落地建议：把 `dayPeriodRules`（每 locale 的边界表）作为 `internal/cldr/dates.go` 中的 `DayPeriodRules` map，键为 `language.Tag.String()`，值为 `[]DayPeriodRange{From, To, Type}`。生成器从 `cldr-core/supplemental/dayPeriods.json` 抽取。

## 6. 对本项目的落地建议

### 6.1 包布局

```
go-intl/
├── datetimeformat/
│   ├── datetimeformat.go        # 公开类型 DateTimeFormat、Options、ResolvedOptions
│   ├── format.go                # Format / FormatToParts / FormatRange / FormatRangeToParts
│   └── new.go                   # New(loc, opts...) 构造与选项校验
├── internal/
│   ├── ecma402/datetimeformat/
│   │   ├── initialize.go        # InitializeDateTimeFormat
│   │   ├── todatetimeoptions.go # ToDateTimeOptions（dateStyle/timeStyle 展开）
│   │   ├── skeleton.go          # parseDateTimeSkeleton + processDateTimePattern
│   │   ├── matcher.go           # BasicFormatMatcher + BestFitFormatMatcher
│   │   ├── partition.go         # PartitionDateTimePattern + RangePattern
│   │   ├── format.go            # FormatDateTime / FormatDateTimePattern / FormatDateTimeRange
│   │   └── tolocaltime.go       # ToLocalTime（计算 era/year/month/day/h/m/s/dst/offset）
│   ├── cldr/
│   │   ├── dates.go             # 生成的 era/month/weekday/dayPeriods/formats per locale
│   │   ├── intervals.go         # intervalFormats per locale
│   │   └── metazones.go         # zone → metazone → displayName per locale
│   └── tz/
│       ├── location.go          # Location wrapper, Resolve, CanonicalLink
│       ├── offset.go            # ParseOffsetString
│       └── lookup.go            # LookupAt(loc, t) → ZoneInfo
```

### 6.2 公开类型签名

```go
// 提炼自 formatjs IDateTimeFormat + ECMA-402 第 11 章
type DateTimeFormat struct {
    locale          locale.Locale
    pattern         []patternToken    // resolved skeleton-to-pattern result
    timeZone        *tz.Location      // resolved
    calendar        Calendar          // gregory only in Phase 1
    numberingSystem string
    dataLocale      string            // CLDR fallback
    // ... resolved fields
}

type Options struct {
    LocaleMatcher          LocaleMatcher
    Calendar               string
    NumberingSystem        string
    HourCycle              HourCycle
    TimeZone               string
    FormatMatcher          FormatMatcher
    Hour12                 *bool
    Weekday                FieldStyle      // narrow|short|long
    Era                    FieldStyle
    Year                   NumericStyle    // numeric|2-digit
    Month                  FieldStyle
    Day                    NumericStyle
    DayPeriod              FieldStyle
    Hour                   NumericStyle
    Minute                 NumericStyle
    Second                 NumericStyle
    FractionalSecondDigits int             // 1..3
    TimeZoneName           TimeZoneName
    DateStyle              Style           // full|long|medium|short
    TimeStyle              Style
}

func New(loc locale.Locale, opts ...Option) (*DateTimeFormat, error)
func (f *DateTimeFormat) Format(t time.Time) string
func (f *DateTimeFormat) FormatToParts(t time.Time) []Part
func (f *DateTimeFormat) FormatRange(start, end time.Time) string
func (f *DateTimeFormat) FormatRangeToParts(start, end time.Time) []Part
func (f *DateTimeFormat) ResolvedOptions() ResolvedOptions
```

### 6.3 调用示例

```go
// 5 行调用示例
fmt, err := datetimeformat.New(
    locale.MustParse("zh-CN"),
    datetimeformat.WithDateStyle(datetimeformat.StyleFull),
    datetimeformat.WithTimeZone("Asia/Shanghai"),
)
out := fmt.Format(time.Now())
```

### 6.4 测试与 conformance

- 把 formatjs 的 `intl-datetimeformat/tests/format.test.ts`、`format-range.test.ts`、`offset-timezone.test.ts`、`skeleton.test.ts` 的断言机械抽取为 `testdata/conformance/<file>.json`。
- 在 Phase 1 末，对 `time.Format` 输出做对比（不要求字节相等，但要求差异都登记进 `divergences.md`）。

## 7. 决策矩阵

| 主题 | 推荐 | 备选 | 否决 | 依据 |
|------|------|------|------|------|
| skeleton 字符表归属 | `internal/ecma402/datetimeformat/skeleton.go`（算法）+ `internal/cldr/dates.go`（数据） | 算法与数据合并到 `internal/cldr` | translate-agent 风格的 codegen 穷举 | §1.1, §1.2 |
| FormatMatcher 实现 | 表驱动复刻 formatjs（含 `adjustFieldTypes` 后处理） | 仅实现 `BasicFormatMatcher`（spec 默认是 best-fit，必须实现） | 调用 ICU C++ DateTimePatternGenerator | §1.3 |
| 时区数据来源 | `time/tzdata` + CLDR metaZones（自维护 metazone 索引） | formatjs 风格 packed transition 表 | `time.LoadLocation` 系统文件 | §2.2 |
| 时区数据嵌入 | `internal/tz/` 包 + `//go:embed` | 独立 module（如 `agentable/tz`） | runtime fetch | §2.2 |
| Phase 1 日历 | Gregorian only | Gregorian + Buddhist（年偏移）作为可选 | Gregorian + Persian + Buddhist + Islamic + Japanese | §3.2 |
| time.Time 处理 | `Round(0)` 剥离 monotonic；选项 `TimeZone` 优先 | 完全沿用 `t.Location()`（违规） | 不剥离 monotonic（输出抖动） | §4.1 |
| dayPeriod 数据 | 从 CLDR `dayPeriods.json` 抽 `dayPeriodRules` per locale | 仅 noon/midnight（与 formatjs 一致但功能受限） | 硬编码 en 边界 | §5 |

## 8. 代码块索引

| 章节 | 代码块 | 来源 |
|------|--------|------|
| §1.1 | `DATE_TIME_REGEX`、`matchSkeletonPattern`、`parseDateTimeSkeleton` 签名 | formatjs/skeleton.ts |
| §1.3 | `bestFitFormatMatcherScore`、`BestFitFormatMatcher` 签名 | formatjs/BestFitFormatMatcher.ts |
| §2.2 | `ZoneInfo` 与 `internal/tz` 接口建议 | 落地建议 |
| §6.2 | `DateTimeFormat`、`Options`、`New/Format/FormatToParts/FormatRange/ResolvedOptions` 签名 | 落地建议 |
| §6.3 | `datetimeformat.New(...).Format(...)` 调用示例 | 落地建议 |

## 9. 引用清单

### formatjs（主参考）
- `.references/formatjs/packages/intl-datetimeformat/core.ts` — 构造、`boundFormat`、`tzData` 注入<!-- L122-181 -->
- `.references/formatjs/packages/intl-datetimeformat/types.ts:7-31` — `PackedData`、`ZoneData` 序列化布局
- `.references/formatjs/packages/intl-datetimeformat/unpack.ts:4-22` — packed → unpacked 反序列化
- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts:137-200` — CLDR 抽取与 metazone 映射
- `.references/formatjs/packages/intl-datetimeformat/scripts/packer.ts:3-19` — pack 算法
- `.references/formatjs/packages/intl-datetimeformat/tz_data.tar.gz` — 808 KB packed / 10 MB 解压
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts:14-150` — 骨架字符 → 选项映射
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/BestFitFormatMatcher.ts:30-141` — 评分函数
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/BasicFormatMatcher.ts:19-110` — 标准评分
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/InitializeDateTimeFormat.ts:68-180` — 选项流水线
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/PartitionDateTimePattern.ts:21-41` — partition 入口
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts:14-183` — UTC → 本地时换算（含 `+offset` 与 IANA 两路、calendar=gregory invariant）

### translate-agent/intl（Go 先例）
- `.references/intl/intl.go:367-403` — 公开类型 `Options` 仅 Era/Year/Month/Day
- `.references/intl/intl.go:387-394` — 日历分派（Gregory/Persian/Buddhist）
- `.references/intl/intl.go:412-446` — 16 个 `seq*` 函数的 if-else 穷举
- `.references/intl/intl.go:494-535` — Buddhist 实现（年偏移 +543）
- `.references/intl/intl.go:449-491` — Persian 实现（依赖 `yaa110/go-persian-calendar`）

### PHP ext/intl（范围参考）
- `.references/ext/src/ecma402/calendar.{c,h}` — 标识符校验入口（参考列表）
- `.references/ext/src/ecma402/hour_cycle.{c,h}` — h11/h12/h23/h24
- `.references/ext/src/ecma402/time_zone.{c,h}` — IANA 名校验

### 项目内部
- `SPECS/00-vision-and-scope.md:202-206` — §5.4 时区策略已声明嵌入 tzdata
- `SPECS/00-vision-and-scope.md:249-258` — §8 开放问题（CLDR 版本钉、best-fit 匹配器）
