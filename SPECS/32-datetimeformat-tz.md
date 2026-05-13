# SPEC 32 — DateTimeFormat TimeZone & Calendar Data

> **Status:** Draft (2026-05-08)
> **Owner:** `internal/cldr/timezones.go` + `internal/cldr/metazones.go` + `internal/cldr/dates.go` + `internal/tz/`
> **Reference contract:** `formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` + `time/tzdata` + CLDR `cldr-core/supplemental/metaZones.json` + `cldr-dates-full/main/<locale>/timeZoneNames.json` + `cldr-dates-full/main/<locale>/ca-gregorian.json`

## Overview

定义 DateTimeFormat 所需的两类底层数据:

1. **TimeZone 数据**:IANA tzdata 注入策略、`*time.Location` 解析、UTC offset string 解析、CLDR `metaZones` → 显示名映射。
2. **Calendar 数据**(active scope 仅 Gregorian):era / month / weekday / dayPeriod 名称、formats、intervalFormats、availableFormats 的 schema 与访问。

本 SPEC 不重定义:
- `DateTimeFormat.New` 调用流 → 见 [SPEC 30 §公开 API](./30-datetimeformat.md#1-公开-api)
- Skeleton 字符表与 BestFitFormatMatcher → 见 [SPEC 31 §Skeleton 字符表](./31-datetimeformat-skeleton.md#2-skeleton-字符表)
- CLDR 全量数据生成流水线 → 见 [SPEC 50 §Codegen](./50-cldr-data.md#codegen)
- 版本钉化全局策略 → 见 [SPEC 50 §Version Pin](./50-cldr-data.md#version-pin)

---

## 1. TimeZone 数据策略 <a id="tz-data"></a>

### 1.1 数据源决策

**MUST** 规则:

1. IANA tzdata **必须**通过 Go 官方 `_ "time/tzdata"` 空白导入注入。
2. **禁止**复制 tzif 文件或 `cldr-core/supplemental/timezone.json` 的 transition 数据到仓库。
3. **禁止**依赖宿主机的 `/usr/share/zoneinfo`(`time.LoadLocation` 默认行为):部署到 Alpine / scratch 容器会失败。
4. **禁止**移植 formatjs 的 `tz_data.tar.gz` 流水线(zdump + Docker + base36 编码)—— Go 已有 `time/tzdata`,重复造轮子无收益。

> **Why**:
> 1. `time/tzdata` 由 Go 团队维护,与 IANA tzdata 同源,~450 KB 增量加到 binary 可接受(Go binary 通常 10–50 MB)。
> 2. ECMA-402 要求三类输入并存:`"Etc/UTC"` / `"America/New_York"` / `"+05:30"`;前两类走 `time.LoadLocation`,第三类自行解析。
> 3. SPEC 00 §5.4 已声明"我们要确定性输出"—— 系统 zoneinfo 在不同部署环境差异即不确定。
>
> **Rejected**:
> - **依赖系统 zoneinfo**:Alpine / scratch 容器无文件,prod / dev 失败模式不一致。
> - **移植 formatjs tz_data**:CI 复杂度高(需 Docker 跑 zic),且 packed 二进制格式难以 Go 端机械翻译。
> - **自行 zone-transition 表 codegen**:`time/tzdata` 已是最小可信源,无收益。

### 1.2 注入位置

**MUST** 规则:

1. 空白导入 **必须**位于 `internal/tz/tzdata.go`(单文件):
   ```go
   package tz

   import _ "time/tzdata"
   ```
2. **禁止**让顶层 `intl` 包或 formatter 包做该导入 —— 集中在 `internal/tz/` 便于审计。
3. `internal/tz/` **必须**作为唯一时区相关入口;`datetimeformat/` 包通过 `internal/tz` accessor 调用。

### 1.3 *time.Location 解析

**MUST** 规则:

1. `internal/tz.Resolve(name string) (*time.Location, error)` **必须**实现:
   - 支持 IANA canonical name(`"America/New_York"`)。
   - 支持 IANA link(`"US/Eastern"` → `"America/New_York"`)。
   - 支持 UTC offset string(`"+05:30"` / `"-08:00"` / `"+00:00"` / `"+14:00"`)。
2. 失败 **必须**返回 `ErrUnsupportedTimeZone` wrapped error,消息含输入字符串。
3. UTC offset string 解析 **必须**在 `internal/tz.ParseOffsetString(s string) (int64, error)` 实现,返回 milliseconds east of UTC;**禁止**通过正则 fallback 到 `time.FixedZone`(必须显式校验范围 `[-14:00, +14:00]`,与 ECMA-402 §13.1.3 一致)。<a id="parseoffsetstring"></a>

```go
// 接口形态(签名,无实现)
package tz

func Resolve(name string) (*time.Location, error)
func ParseOffsetString(s string) (int64, error)        // ms east of UTC
func CanonicalLink(name string) string                 // "US/Eastern" → "America/New_York"
func LookupAt(loc *time.Location, t time.Time) ZoneInfo

type ZoneInfo struct {
    Name     string  // canonical IANA name
    OffsetMs int64
    IsDST    bool
    Abbrv    string  // e.g. "EST" / "EDT"
    Metazone string  // e.g. "America_Eastern", from CLDR metazoneInfo
}
```

### 1.4 CanonicalLink 表 <a id="canonicallink"></a>

**MUST** 规则:

1. IANA link → canonical 映射 **必须**通过 codegen 从 `cldr-core/supplemental/metaZones.json` `metazoneInfo.timezone` 节点 + IANA `backward` 文件抽取,生成 Go literal。
2. **禁止**运行时解析 IANA `backward` 文件;**禁止** `//go:embed backward` 字符串文件 + 解析。
3. 表存放路径 **必须**是 `internal/cldr/timezones.go`(generated)。
4. 形态:
   ```go
   var canonicalTimeZoneLinks = map[string]string{
       "US/Eastern":    "America/New_York",
       "US/Pacific":    "America/Los_Angeles",
       "Asia/Calcutta": "Asia/Kolkata",
       // ...
   }
   ```
5. `CanonicalLink(name)` **必须**先查表;查不到则原样返回(假定输入已是 canonical)。

### 1.5 metaZones 三段映射 <a id="metazones"></a>

DateTimeFormat `timeZoneName: short | long | shortGeneric | longGeneric` 的本地化显示名走 CLDR `metaZones` 三段映射:

```text
zone (e.g. "America/New_York")
  → metazoneInfo (zone → metazone, time-bounded)
    → metaZoneNames (locale × metazone → display name strings)
      → exemplarCity (locale × zone → city name fallback)
```

**MUST** 规则:

1. 三段映射 **必须**由 codegen 输出 Go literal,生成位置 `internal/cldr/metazones.go`。
2. **禁止**运行时 JSON 解析(`//go:embed metaZones.json` + `encoding/json`);**禁止**启动时反序列化。
3. 三段表的 schema:
   ```go
   // generated_metazones.go (片段,签名)

   // 阶段 1:zone → []metazone 时段(metazone 可随时间变更)
   type MetazonePeriod struct {
       Metazone string
       Start    int64  // unix milliseconds, MinInt64 表示 -∞
       End      int64  //                  , MaxInt64 表示 +∞
   }
   var zoneToMetazones = map[string][]MetazonePeriod{
       "America/New_York": {{"America_Eastern", math.MinInt64, math.MaxInt64}},
       // ...
   }

   // 阶段 2:locale × metazone × form → display name
   type MetazoneNames struct {
       LongGeneric, LongStandard, LongDaylight string
       ShortGeneric, ShortStandard, ShortDaylight string
   }
   var metazoneNamesByLocale = map[Locale]map[string]MetazoneNames{
       localeIndex["en-US"]: {
           "America_Eastern": {
               LongGeneric: "Eastern Time", LongStandard: "Eastern Standard Time", LongDaylight: "Eastern Daylight Time",
               ShortGeneric: "ET", ShortStandard: "EST", ShortDaylight: "EDT",
           },
       },
       // ...
   }

   // 阶段 3:locale × zone → exemplar city
   var exemplarCitiesByLocale = map[Locale]map[string]string{
       localeIndex["en-US"]: {"America/New_York": "New York"},
       // ...
   }
   ```
4. accessor `internal/cldr.TimeZoneDisplayName(loc cldr.Locale, zone string, form TimeZoneName, isDST bool, instant int64, offsetMs int64) string` **必须**实现 ECMA-402 §13.1.5 fallback 链:
   ```text
   1. 查 metazoneNames[locale][metazone(zone, instant)][form] → 命中即返回
   2. 查 exemplarCities[locale][zone] → 命中即返回 generic fallback("Time in {city}")
   3. 回落 GMT offset 字符串 ("GMT-05:00")
   ```
5. metazone 选择 **必须**支持时间敏感(`instant` 决定取哪个 `MetazonePeriod`)—— 例:`Europe/Moscow` 在 2011 年前后 metazone 不同。

> **Why**: `metaZones` 三段映射是 conformance 测试的难点(formatjs `tests/timezone-name.test.ts` 30+ fixture);任何运行时解析都会引入冷启动延迟与 binary 体积开销。
>
> **Rejected**: 启动时反序列化 JSON —— 与 SPEC 50 §"no runtime file I/O for CLDR" 冲突。

### 1.6 tzdata 版本钉化 <a id="tzdata-version"></a>

**MUST** 规则:

1. tzdata 版本号 **必须**写入 `internal/cldr/VERSION` 单文件,作为 `tzdata=2025b` 行(与 `cldr=` / `icu=` 同文件)。
2. CI **必须**校验 `time/tzdata` 嵌入的 IANA 版本与 `VERSION` 文件一致;不一致即 block。
3. tzdata 版本与 CLDR / ICU 版本 **必须**同时 bump;独立 bump 任一即 PR block。
4. tzdata 与 metaZones 不一致 **必须**以 IANA tzdata 为权威(行为正确性 > 显示名一致性);分歧记录到 `divergences.md`。

> **Why**: tzdata 与 CLDR 独立发布(IANA 一年多次,CLDR 每年两次);版本不一致会让 `LookupAt(loc, t)` 与 metaZones 数据出现"该时刻应在 metazone X,但 tzdata 已变更"的边角 case。钉同一文件 + CI 校验避免漂移。
>
> **Rejected**: 让 tzdata 与 CLDR 独立 bump —— 时区数据漂移成本远高于"统一 bump 一次"成本。

---

## 2. Calendar 数据 <a id="calendar-data"></a>

### 2.1 active scope 范围

**MUST** 规则:

1. active scope calendar pattern 数据 **必须**仅生成 Gregorian。`SupportedCalendars()` 仍必须包含 ECMA-402 required `iso8601`;其它 well-formed calendar 请求通过 `ResolveLocale` 回落到支持值,不得构造期显式拒绝。
2. codegen 工具(`tools/gen-cldr/`)**必须**显式 skip 非 Gregorian pattern payload;若强制生成未消费 payload,build break。
3. Calendar 数据存放路径 **必须**是 `internal/cldr/dates.go`(generated)。
4. consumer-driven expansion calendar pattern 只能在有消费方需求时新增;新增时再引入对应 generated schema 和测试,不得在 active scope 预留 `//go:build future` 占位文件。

> **Why**:
> 1. SPEC 30 §3 已声明 active scope 仅 Gregorian。
> 2. active scope 占位文件没有运行时消费者,会扩大审计面并制造未使用代码。
> 3. 未来 calendar 的目录和 schema 应由实际实现反推,不要在当前阶段冻结。

### 2.2 Gregorian Schema

`internal/cldr/dates.go` 中的 Gregorian 数据 schema **必须**对齐 CLDR `cldr-dates-full/main/<locale>/ca-gregorian.json`:

```go
// 签名,无实现
package cldr

type Gregorian struct {
    // 来源:dates.calendars.gregorian.eras.eraNames / eraAbbr / eraNarrow
    Eras struct {
        Wide   [2]string // ["Before Christ", "Anno Domini"]
        Abbr   [2]string // ["BC", "AD"]
        Narrow [2]string // ["B", "A"]
    }

    // 来源:dates.calendars.gregorian.months.format.{wide,abbreviated,narrow}
    Months struct {
        Wide       [12]string // ["January", "February", ...]
        Abbr       [12]string // ["Jan", "Feb", ...]
        Narrow     [12]string // ["J", "F", ...]
        StandWide  [12]string // stand-alone form
        StandAbbr  [12]string
        StandNarrow [12]string
    }

    // 来源:dates.calendars.gregorian.days.format.{wide,abbreviated,narrow,short}
    Weekdays struct {
        Wide        [7]string // ["Sunday", ..., "Saturday"]
        Abbr        [7]string // ["Sun", ..., "Sat"]
        Narrow      [7]string // ["S", "M", ...]
        Short       [7]string // ["Su", "Mo", ...]
        StandWide   [7]string
        StandAbbr   [7]string
        StandNarrow [7]string
        StandShort  [7]string
    }

    // 来源:dates.calendars.gregorian.dayPeriods.format.{wide,abbreviated,narrow}
    DayPeriods struct {
        AM struct{ Wide, Abbr, Narrow string }
        PM struct{ Wide, Abbr, Narrow string }
        // flexible day periods(b / B 字符,§2.5)
        Flex map[string]struct{ Wide, Abbr, Narrow string }
        // 例如 keys: "morning1", "afternoon1", "evening1", "night1", "noon", "midnight"
    }

    // 来源:dates.calendars.gregorian.dateFormats / timeFormats / dateTimeFormats
    DateFormats     [4]string // [full, long, medium, short]
    TimeFormats     [4]string
    DateTimeFormats [4]string // 组合 pattern,如 "{1} {0}" 或 "{1}, {0}"

    // 来源:dates.calendars.gregorian.dateTimeFormats.availableFormats
    AvailableFormats map[string]string // skeleton → pattern

    // 来源:dates.calendars.gregorian.dateTimeFormats.intervalFormats
    IntervalFormats map[string]map[string]string // skeleton → field → pattern
    IntervalFallback string                      // intervalFormatFallback

    // 来源:dates.calendars.gregorian.dateTimeFormats.appendItems
    AppendItems map[string]string // field → "{0} ({1})" pattern
}
```

**MUST** 规则:

1. 全部字段 **必须**由 codegen 从 CLDR JSON 抽取生成;**禁止**手写。
2. 数据形态 **必须**是 Go 字面量(`var gregorian_en_US = Gregorian{...}`);**禁止** `//go:embed *.json` + 运行时反序列化。
3. accessor `internal/cldr.GregorianFor(loc cldr.Locale) Gregorian` **必须**接收已解析的 CLDR locale handle;public `locale.Locale` 到 `cldr.Locale` 的匹配由 `internal/cldrmatch` / `internal/cldr.ResolveLocale(language.Tag)` 在构造期完成。
4. `dayPeriods.format.flex` **必须**全量生成(`morning1` / `afternoon1` / `evening1` / `night1` / `noon` / `midnight`);**禁止**仅生成 `noon` / `midnight`(formatjs 当前限制)。

> **Why**:
> 1. Go 字面量 = 编译期校验 + 运行时零分配;JSON + `encoding/json` 增加 5–10 MB 内存 + 100–500 ms 冷启动。
> 2. flex day periods 是 `b` / `B` 字符正确输出的必要条件(`zh-CN` 凌晨/早上/上午/中午/下午/晚上),fallback 到 AM/PM 即破坏字节相等。

### 2.3 dayPeriodRules <a id="dayperiodrules"></a>

flex day periods 边界数据来自 CLDR `cldr-core/supplemental/dayPeriods.json` `dayPeriodRules` 节点(每 locale 不同):

**MUST** 规则:

1. `dayPeriodRules` **必须**全量内嵌到 `internal/cldr/dates.go`,生成 Go literal:
   ```go
   // internal/cldr/dates.go (片段)
   type DayPeriodRange struct {
       From, To time.Duration  // 自当日 00:00 起的时间偏移
       Type     string         // "morning1" / "afternoon1" / ...
   }
   var dayPeriodRules = map[string][]DayPeriodRange{
       "en": {
           {0 * time.Hour, 5 * time.Hour, "night1"},
           {5 * time.Hour, 12 * time.Hour, "morning1"},
           {12 * time.Hour, 12 * time.Hour, "noon"},     // exact 12:00
           {12 * time.Hour, 18 * time.Hour, "afternoon1"},
           {18 * time.Hour, 21 * time.Hour, "evening1"},
           {21 * time.Hour, 24 * time.Hour, "night1"},
       },
       "zh": { /* ... */ },
   }
   ```
2. `internal/cldr.DayPeriodFor(loc cldr.Locale, hour, minute int) string` **必须**返回 `"morning1"` / `"noon"` 等 type 名;调用方再查 `Gregorian.DayPeriods.Flex[type]`。
3. **禁止**仅硬编码 `en` 边界;**必须**全量 codegen。

---

## 3. ResolvedOptions.TimeZone 行为

**MUST** 规则(对应 [SPEC 30 §ResolvedOptions](./30-datetimeformat.md#23-resolvedoptions)):

1. `ResolvedOptions().TimeZone` 返回值 **必须**是 IANA canonical name(对 link 已转换)或 UTC offset string(对 `+HH:MM` 输入保留)。
2. 转换流程 **必须**是:
   ```text
   1. resolveTZ := internal/tz.Resolve(input)    // 接受 link / canonical / offset
   2. canonical := internal/tz.CanonicalLink(input)
   3. f.timeZoneCanonical := canonical            // 缓存到 DateTimeFormat slot
   4. ResolvedOptions().TimeZone == canonical
   ```
3. UTC offset string 输入 **必须**保留原 sign 与 padding(`"+05:30"` 不变成 `"05:30"`)。

---

## 4. Performance

**MUST** 规则:

1. `BenchmarkTZ_Resolve` **目标** ≤ 5 μs/op(P50),含 `time.LoadLocation`;若平台差异导致波动,以趋势回归和 SPEC 71 统一阈值为准。
2. `internal/cldr.GregorianFor(loc)` **必须**基于已解析的 `cldr.Locale` 常量级查表,不得在调用期解析 BCP 47 字符串。
3. `BenchmarkCLDR_TimeZoneDisplayName` **目标** ≤ 500 ns/op(典型 IANA name + form);阈值需要由 benchmark 持续验证。

---

## 5. Forbidden

- **禁止** 复制 tzif 文件到仓库 —— 必须通过 `_ "time/tzdata"` 注入。
- **禁止** 依赖系统 `/usr/share/zoneinfo`(`time.LoadLocation` 默认行为,Alpine 容器无文件)。
- **禁止** 移植 formatjs `tz_data.tar.gz` 流水线 —— Go 已有 `time/tzdata`。
- **禁止** 运行时 JSON 解析 `metaZones.json` —— 必须 codegen 输出 Go literal。
- **禁止** `//go:embed metaZones.json` + `encoding/json` 路径 —— 与 SPEC 50 "no runtime file I/O" 冲突。
- **禁止** active scope 实现 Gregorian 之外任何 calendar 数据(包括 Buddhist 占位)。
- **禁止** tzdata 版本与 CLDR / ICU 版本独立 bump —— 必须同 `internal/cldr/VERSION` 文件,CI 校验。
- **禁止** `dayPeriodRules` 仅生成 `en` —— 必须全量 codegen 所有 active scope locale。
- **禁止** `dayPeriodRules` 只覆盖 `noon` / `midnight` —— 必须全量 flex(morning1/afternoon1/evening1/night1)。
- **禁止** `Format` 路径调用 `internal/tz.Resolve` 或 `time.LoadLocation` —— 必须在 `New` 时缓存 `*time.Location`(SPEC 30 §4.2)。
- **禁止** UTC offset 输入跳过范围校验 —— 必须 `[-14:00, +14:00]`。
- **禁止** IANA link 表(`canonicalTimeZoneLinks`) 通过运行时解析 `backward` 文件生成 —— 必须 codegen。

---

## 6. Acceptance Criteria

- [ ] `internal/tz/tzdata.go` 仅含一行 `import _ "time/tzdata"`;无其他声明。
- [ ] `internal/tz.Resolve("America/New_York")` 返回非空 `*time.Location`,offset 含 DST(夏季 -04:00,冬季 -05:00)。
- [ ] `internal/tz.Resolve("US/Eastern")` 返回与 `Resolve("America/New_York")` 等价(canonical 化)。
- [ ] `internal/tz.Resolve("+05:30")` 返回 fixed-offset Location,DST 永远 false。
- [ ] `internal/tz.Resolve("+15:00")` 返回 `ErrUnsupportedTimeZone`(超出 ±14:00 范围)。
- [ ] `internal/tz.Resolve("Mars/Olympus")` 返回 `ErrUnsupportedTimeZone` wrapped error,消息含 `"Mars/Olympus"`。
- [ ] `internal/tz.CanonicalLink("US/Eastern") == "America/New_York"`。
- [ ] `internal/tz.CanonicalLink("America/New_York") == "America/New_York"`(canonical 原样返回)。
- [ ] `internal/tz.ParseOffsetString("+05:30") == 5*3600*1000 + 30*60*1000`(ms east of UTC)。
- [ ] `internal/tz.ParseOffsetString("-08:00") == -8*3600*1000`。
- [ ] `internal/cldr/VERSION` 文件含 `cldr=` / `icu=` / `tzdata=` 三行,CI 校验 hash 一致。
- [ ] `internal/cldr.ResolveLocale(language.MustParse("en-US"))` 成功后,`internal/cldr.GregorianFor(loc)` 返回完整 `Gregorian` 结构,所有字段非空。
- [ ] `internal/cldr.ResolveLocale(language.MustParse("zh-Hans-CN"))` 成功后,`internal/cldr.GregorianFor(loc).DayPeriods.Flex` 包含 `"morning1" / "afternoon1" / "evening1" / "night1" / "noon" / "midnight"` 全部 6 个 key。
- [ ] `internal/cldr.DayPeriodFor(enUSCLDRLocale, 5, 0) == "morning1"`。
- [ ] `internal/cldr.DayPeriodFor(zhHansCNCLDRLocale, 13, 0) == "afternoon1"`。
- [ ] `internal/cldr.TimeZoneDisplayName(enUSCLDRLocale, "America/New_York", TimeZoneNameLongGeneric, false, time.Now().UnixMilli(), -5*3600*1000) == "Eastern Time"`。
- [ ] `internal/cldr.TimeZoneDisplayName(enUSCLDRLocale, "America/New_York", TimeZoneNameShort, false, instant, -5*3600*1000) == "EST"`。
- [ ] `internal/cldr.TimeZoneDisplayName(enUSCLDRLocale, "Mars/Olympus", TimeZoneNameShortGeneric, false, instant, -5*3600*1000) == "GMT-05:00"`。
- [ ] `internal/cldr/metazones.go` 是 generated,文件头含 `// Code generated by tools/gen-cldr; DO NOT EDIT.`。
- [ ] `internal/cldr/dates.go` 是 generated,文件头含 `// Code generated by tools/gen-cldr; DO NOT EDIT.`;不存在 `internal/cldr/calendars/*` active scope 占位文件。
- [ ] codegen 工具只读取 `ca-gregorian.json`,且输出的 `datesByLocale` active scope 仅包含 `"gregorian"` calendar。
- [ ] `formatjs/packages/intl-datetimeformat/tests/timezone-name.test.ts` 全部 fixture 在 `datetimeformat/conformance_test.go` 通过。
- [ ] `formatjs/packages/intl-datetimeformat/tests/day-period.test.ts` 全部 fixture 通过(含 `b` / `B` flex day period)。
- [ ] benchmark `BenchmarkTZ_Resolve` 作为趋势守卫;目标阈值 ≤ 5 μs/op。
- [ ] benchmark `BenchmarkCLDR_TimeZoneDisplayName` 作为趋势守卫;目标阈值 ≤ 500 ns/op。

---

## 7. References

### Primary

- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` — CLDR `metaZones` / `timeZoneNames` 抽取路径(L137-200)
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts` — UTC offset string 解析(`OFFSET_TIMEZONE_FORMAT_REGEX`,L17-18,45-74)与 `getApplicableZoneData` 分支(L96-100)
- `cldr-core/supplemental/metaZones.json` — zone → metazone 映射、metazoneInfo
- `cldr-dates-full/main/<locale>/timeZoneNames.json` — locale × metazone display name + exemplarCity
- `cldr-dates-full/main/<locale>/ca-gregorian.json` — Gregorian calendar 数据全量
- `cldr-core/supplemental/dayPeriods.json` — `dayPeriodRules` 边界

### Secondary

- Go `time/tzdata` 包文档 — 嵌入式 IANA 数据
- Go `time.LoadLocation` 文档 — 解析路径
- IANA `backward` 文件 — link 列表权威源
- Upstream ICU tzdata — ICU tzdata(裁判用)
- `.references/intl/intl.go` — translate-agent/intl 不实现时区(反例)

### Project Cross-References

- [SPEC 30 §公开 API](./30-datetimeformat.md#1-公开-api) — `DateTimeFormat.New` 调用方
- [SPEC 30 §时区处理](./30-datetimeformat.md#4-时区处理) — Format 期不查 TZ 规则
- [SPEC 31 §Skeleton 字符表](./31-datetimeformat-skeleton.md#2-skeleton-字符表) — `z/Z/O/v/V/X` 字符 → `TimeZoneName` 映射
- [SPEC 50 §Codegen](./50-cldr-data.md#codegen) — `tools/gen-cldr` 生成器架构
- [SPEC 50 §Version Pin](./50-cldr-data.md#version-pin) — `internal/cldr/VERSION` 文件结构
