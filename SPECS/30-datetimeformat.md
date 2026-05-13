# SPEC 30 — DateTimeFormat Core

> **Status:** Draft (2026-05-08)
> **Owner:** `datetimeformat/`
> **Reference contract:** `.references/ecma402/spec/datetimeformat.html` first, then `formatjs/packages/intl-datetimeformat/` + `formatjs/packages/ecma402-abstract/DateTimeFormat/`

## Overview

定义 `datetimeformat.DateTimeFormat` 公开 API、构造期校验、ECMA-402 locale/options negotiation、`time.Time` 的时区上下文处理、`Format`/`FormatToParts`/`FormatRange`/`FormatRangeToParts`/`ResolvedOptions` 行为。active generated pattern 数据仍以 Gregorian 为核心,但 well-formed unsupported calendar 请求必须通过 `ResolveLocale` 回落,不得作为构造期错误。

本 SPEC 不重定义:
- Skeleton 解析与 BestFitFormatMatcher → 见 [SPEC 31 §Skeleton](./31-datetimeformat-skeleton.md#skeleton)
- 时区数据注入与 metaZones → 见 [SPEC 32 §TZ Data](./32-datetimeformat-tz.md#tz-data)
- Calendar 数据 schema → 见 [SPEC 32 §Calendar Data](./32-datetimeformat-tz.md#calendar-data)
- `Locale` 结构 → 见 [SPEC 10 §Locale 结构](./10-locale.md#locale-结构)
- 抽象操作总入口 → 见 [SPEC 12 §Abstract Ops](./12-abstract-operations.md)

---

## 1. 公开 API

### 1.1 构造与 Option

```go
package datetimeformat

type DateTimeFormat struct{ /* 不可变;包含 resolved + 缓存的 *time.Location + selectedPattern + cldr 数据切片 */ }

type Options struct {
    Calendar               string
    NumberingSystem        string
    LocaleMatcher          LocaleMatcher
    FormatMatcher          FormatMatcher
    TimeZone               string
    TimeZoneName           TimeZoneName
    Weekday                FieldStyle
    Year                   NumericStyle
    Month                  MonthStyle
    Day                    NumericStyle
    DayPeriod              FieldStyle
    Hour                   NumericStyle
    Minute                 NumericStyle
    Second                 NumericStyle
    HourCycle              HourCycle
    Hour12                 Hour12Option
    DateStyle              Style
    TimeStyle              Style
    FractionalSecondDigits FractionalSecondDigitOptions
}

func New(loc locale.Locale, opts ...Options) (*DateTimeFormat, error)

func (f *DateTimeFormat) Format(t time.Time) string
func (f *DateTimeFormat) FormatToParts(t time.Time) []Part
func (f *DateTimeFormat) FormatRange(start, end time.Time) string
func (f *DateTimeFormat) FormatRangeToParts(start, end time.Time) []Part
func (f *DateTimeFormat) ResolvedOptions() ResolvedOptions
```

**MUST** 规则:

1. `New` **必须**在构造期完成全部选项语法校验、locale/options negotiation、`*time.Location` 解析,失败返回 `error`。
2. `New` **最多接受一个** `Options` 值。`New(loc)` 等价 JS 省略 options;`New(loc, Options{})` 等价 JS 传空 options object;传入多个 `Options` 必须返回 `ErrInvalidOption`。
3. `Format` / `FormatToParts` / `FormatRange` / `FormatRangeToParts` 的普通成功路径 **必须不返回** option errors;invalid time 等 JS `RangeError` 等价错误若无法用 `time.Time` 类型表达,不应发明额外错误。
4. `DateTimeFormat` 是不可变值;`*DateTimeFormat` 上的所有方法 **必须**并发安全。
5. formatter options **必须**采用 typed `Options` 值(同 SPEC 20),禁止 functional options 作为公共主路径。
6. `ResolvedOptions` **必须**返回不可变快照(值类型);多次调用结果相等。

> **Why**: 构造期校验集中错误处理;`Format` 不返 error 与 `formatjs` 字节相等。`*time.Location` 的解析(`time.LoadLocation`)在 `Format` 期不可重做(违反 hot path 零分配规则)。

### 1.2 输入类型

**MUST** 规则(对应 ECMA-402 §13.5.1 `time.Time` 入参):

1. `Format`/`FormatToParts`/`FormatRange`/`FormatRangeToParts` 入参 **必须**是 `time.Time`(非 `*time.Time`,非 `int64` 毫秒);**禁止**接受 `any` 入口。
2. **必须**在方法入口立即调用 `t = t.Round(0)` 剥离 monotonic clock,再走格式化主流程。
3. `time.Time` 自带的 `Location()` **不能**作为每次调用的显示时区。ECMA-402 在构造期把 `timeZone` 解析为 `[[TimeZone]]`;未传 `Options.TimeZone` 时必须 snapshot `SystemTimeZoneIdentifier()` 的 Go 等价物(`time.Local` 的可解释 canonical 名称)。
4. `time.Time{}` 零值是 Go 可表示的 instant,不是 JavaScript invalid Date。它必须按普通 `time.Time` 经 formatter `[[TimeZone]]` 格式化;禁止把它改写成 Unix epoch 或 error。

```go
// 入口归一化(签名,无实现)
func (f *DateTimeFormat) Format(t time.Time) string {
    t = t.Round(0)            // 剥离 monotonic
    loc := f.timeZone         // 已在 New 时缓存;永不来自单次输入的 Location()
    instant := t.UTC().UnixMilli()
    // 用 (instant, loc) 走 PartitionDateTimePattern
}
```

> **Why**:
> 1. `time.Time` 是 Go idiom,与 `time.Now()` / `time.Date()` / time package 互操作。
> 2. ECMA-402 模型:输入 = instant,formatter `[[TimeZone]]` = 输出时区;输入的 `Location` 不影响输出(formatjs `ToLocalTime` 行为)。Go idiom 把时区放在 `time.Time` 上,我们必须在构造期确定 formatter time zone。
> 3. `t.Round(0)` 是 Go pkg.go.dev 推荐做法;否则 `Format(t1) != Format(t1.Round(0))` 会出现可观测差异(monotonic clock 影响相等性测试)。
>
> **Rejected**:
> - **接受 `any` 入口**(模拟 JS `Date | number | string`)—— Go 类型安全弱化;且 messageformat-go 桥接已经持有 `time.Time`。
> - **不剥离 monotonic** —— 输出抖动,`Format(t) == Format(t)` 成立但 `t.Equal(t2) ⇒ Format(t) == Format(t2)` 不成立。
> - **未指定 TimeZone 时跟随每个输入的 `t.Location()`** —— 这把 formatter 从 ECMA-402 的 fixed `[[TimeZone]]` 对象变成 per-input display-zone helper,必须拒绝。
> - **沿用 translate-agent 模式(直接用 `t.Year()/Month()/Day()`,忽略选项 TimeZone)** —— 偏离 ECMA-402,与 formatjs 字节相等不兼容。

### 1.3 调用示例

```go
df, err := datetimeformat.New(locale.MustParse("zh-CN"),
    datetimeformat.Options{
        DateStyle: datetimeformat.FullDateTimeStyle,
        TimeZone:  "Asia/Shanghai",
    })
out := df.Format(time.Now())
// out == "2026年5月8日星期五"
```

```go
df, _ := datetimeformat.New(locale.MustParse("en-US"),
    datetimeformat.Options{
        Month: datetimeformat.ShortMonthStyle,
        Day:   datetimeformat.NumericFieldStyle,
        Year:  datetimeformat.NumericFieldStyle,
    })
// df.Format(t) == "May 8, 2026"
```

### 1.4 Option 参数类型 — typed ECMA-402 values

**MUST** 规则:

1. 所有取自 ECMA-402 字符串字面量并集的选项(`localeMatcher` / `hourCycle` / `formatMatcher` / `weekday` / `year` / `month` / `day` / `dayPeriod` / `hour` / `minute` / `second` / `timeZoneName` / `dateStyle` / `timeStyle`)**必须**以命名类型承载,底层 kind 保持 `string`,序列化形态与 JS resolvedOptions 字符串一致。
2. `calendar` 和 `timeZone` 仍是 `string`,因为它们是 Unicode extension / IANA or offset identifiers,不是小枚举。
3. 校验 **必须**在 `New` 内集中完成,失败包装 `ErrInvalidOption` 并显示用户传入值。
4. `Hour12` 是特例(JS `boolean`):Go 端 **必须** 使用 `Hour12(bool)` 子值,通过内部 `set` 标记区分 "未传"与"显式 false"(同 SPEC 20 §1.2)。
5. `FractionalSecondDigits` **必须**使用 `FractionalSecondDigits(n)` 子值,以区分未传与显式非法值。

> **Why**: typed value 仍保留 ECMA-402 字符串作为 wire/resolved 形态,但让 Go 调用点通过常量表达合法值。conformance fixture 和 messageformat-go adapter 在边界做一次映射,不把公共 API 降级成 JSON 形态。
>
> **Rejected**: functional options + 裸字符串 —— 隐藏状态、不可序列化、难以静态发现。

---

## 2. Options 与 ResolvedOptions

### 2.1 Options 字段

`Options`(内部 config struct)字段对应 ECMA-402 §13.4.1 `InitializeDateTimeFormat`:

```go
// 全部命名类型 underlying kind 都是 string(JS 字面量对齐;参见 §1.4)。
type (
    LocaleMatcher string  // "lookup" | "best fit"
    HourCycle     string  // "h11" | "h12" | "h23" | "h24"
    FormatMatcher string  // "basic" | "best fit"
    FieldStyle    string  // "narrow" | "short" | "long"
    NumericStyle  string  // "numeric" | "2-digit"
    MonthStyle    string  // "numeric" | "2-digit" | "narrow" | "short" | "long"
    TimeZoneName  string  // "short" | "long" | "shortOffset" | "longOffset" | "shortGeneric" | "longGeneric"
    Style         string  // "full" | "long" | "medium" | "short"
)

type Options struct {
    LocaleMatcher          LocaleMatcher
    Calendar               string          // BCP 47 -u-ca-* 字面量;well-formed unsupported values fall back through ResolveLocale
    NumberingSystem        string
    HourCycle              HourCycle
    TimeZone               string          // IANA name 或 "+HH:MM"
    FormatMatcher          FormatMatcher   // 默认 "best fit"
    Hour12                 Hour12Option    // 通过 Hour12(bool) 构造,区分"未传"与"显式 false"
    Weekday                FieldStyle
    Era                    FieldStyle
    Year                   NumericStyle
    Month                  MonthStyle
    Day                    NumericStyle
    DayPeriod              FieldStyle
    Hour                   NumericStyle
    Minute                 NumericStyle
    Second                 NumericStyle
    FractionalSecondDigits FractionalSecondDigitOptions // 1..3
    TimeZoneName           TimeZoneName
    DateStyle              Style
    TimeStyle              Style
}
```

`Options{}` 零值等价 ECMA-402 默认选项。`Hour12` 和 `FractionalSecondDigits` 使用小子值保存 set bit;其它命名类型字段以空字符串表示未传。

**MUST** 规则:

1. **必须**校验 `DateStyle` / `TimeStyle` 与具体字段(`Weekday`/`Year`/`Month`/`...`)互斥;若同时设置,返回 `ErrInvalidOption`(对齐 ECMA-402 §13.1.1.1 异常)。
2. `Hour12` **必须**通过 `Hour12(bool)` 子值表达,以区分"未传"与"显式 false"。
3. `FractionalSecondDigits` **必须**校验 ∈ `{1, 2, 3}`,其它返回 `ErrInvalidOption`。
4. `Calendar` 与 `NumberingSystem` **必须**只校验 Unicode type 语法;well-formed 但 unsupported 的值由 `ResolveLocale` 选择支持值或回落默认值。

### 2.2 HourCycle 联动(§13.1.1.1) <a id="22-hourcycle-联动132111"></a>

**MUST** 规则:

1. `ResolvedOptions().HourCycle` **必须**与 `Locale.HourCycle()`(BCP 47 `-u-hc-...`)+ 显式 `Options.HourCycle` + `Options.Hour12` 三者联动,按 ECMA-402 §13.1.1.1 步骤决议:
   ```text
   if Hour12 != nil:
       resolved.HourCycle := Hour12 ? "h11"|"h12"(locale 默认 12 制位置) : "h23"|"h24"
       (具体二选一由 dataLocale 默认决定)
   else if HourCycle 显式设置:
       resolved.HourCycle := HourCycle
   else if Locale.HourCycle() != "":
       resolved.HourCycle := Locale.HourCycle()
   else:
       resolved.HourCycle := dataLocale 默认(从 CLDR `timeData.json` `preferred` 取)
   ```
2. **禁止**让 `Hour12 = false` 默认覆盖 `Locale.HourCycle() = h11`(必须按上述优先级)。
3. `Options{HourCycle: H11HourCycle, Hour12: Hour12(true)}` 同时存在 **必须**以 `Hour12` 优先(ECMA-402 §13.1.1.1)。

> **Why**: HourCycle 是 conformance 测试中错误率最高的字段(formatjs `tests/hour-cycle.test.ts` 30+ fixture),必须严格按 §13.1.1.1。

### 2.3 ResolvedOptions

```go
type ResolvedOptions struct {
    Locale                 locale.Locale
    Calendar               string         // 总是 "gregory" in active scope
    NumberingSystem        string
    TimeZone               string         // IANA canonical name 或 "+HH:MM"
    HourCycle              HourCycle
    Hour12                 *bool
    Weekday                FieldStyle
    Era                    FieldStyle
    Year                   NumericStyle
    Month                  FieldStyle
    Day                    NumericStyle
    DayPeriod              FieldStyle
    Hour                   NumericStyle
    Minute                 NumericStyle
    Second                 NumericStyle
    FractionalSecondDigits int
    TimeZoneName           TimeZoneName
    DateStyle              Style
    TimeStyle              Style
}
```

**MUST** 规则:

1. 字段顺序 **必须**与 ECMA-402 §13.4.5 spec 顺序一致。
2. `TimeZone` **必须**返回 IANA canonical name(如 `"America/New_York"`),即使输入是 link(`"US/Eastern"`);通过 [SPEC 32 §CanonicalLink](./32-datetimeformat-tz.md#canonicallink) 取得。
3. `Hour12` **必须**仅在用户显式传入或选项 `HourCycle` 解析为 12 制位置时呈现非空。
4. **必须**返回值类型(非指针),并发安全。

---

## 3. Calendar Resolution(active scope)

### 3.1 决策

active generated pattern 数据仅 Gregorian。ECMA-402 可观测行为仍由 `ResolveLocale` 决定:`Options.Calendar` 或 locale `-u-ca-*` 中 well-formed 但 unsupported 的值必须回落到支持值,不得返回 `ErrUnsupportedCalendar`。

> **Why**:
> 1. ECMA-402 constructor 先做 Unicode type 语法校验,再通过 locale data negotiation 选择支持值;unsupported but well-formed calendar 不是 RangeError。
> 2. 非 Gregorian pattern/calculation 的核心难度是 ICU `Calendar::computeFields` 的状态机(~5000 行 C++);active scope 不生成对应 payload。
> 3. messageformat-go 当前 `:datetime` 函数只支持 Gregorian。
>
> **Rejected**:
> - **active scope 加入 Buddhist(年偏移 +543)** —— consumer-driven expansion 时再引入对应 generated schema 与 conformance。
> - **active scope 加入 Persian** —— 依赖 `yaa110/go-persian-calendar` 第三方,且 Solar Hijri 算法独立成 SPEC。
> - **active scope 加入 Japanese** —— 需"年号表"(Reiwa/Heisei/Showa/Taisho/Meiji),ICU 在 2019 年改元后多次发版修订,版本钉化复杂。

**MUST** 规则:

1. `Calendar` 字段类型 **必须**是 `string`(暴露 BCP 47 `-u-ca-...` 形态),**不**改成受控枚举(`type Calendar int`)—— 字符串语义直接对应 BCP 47,且为 consumer-driven expansion 扩展预留。
2. `Locale.Calendar()` / `Options.Calendar` 非空时 **必须**只做 Unicode type 语法校验;well-formed unsupported 值交给 `ResolveLocale` 回落。
3. `SupportedCalendars()` **必须**包含 `"gregory"` 与 ECMA-402 required `"iso8601"`。
4. `ResolvedOptions().Calendar` **必须**返回 resolved locale record 的 `ca` 值;当前 active data 通常回落到 `"gregory"`。

### 3.2 Calendar 数据访问

**MUST** 规则:

1. Calendar 数据(era / month / weekday / dayPeriod 名)**必须**通过 [SPEC 32 §Calendar Data](./32-datetimeformat-tz.md#calendar-data) 接入;`internal/cldr/dates.go` 是 generated SSOT。
2. `New` 期 **必须**先把 public `locale.Locale` 解析为 `cldr.Locale`,再通过 `internal/cldr.GregorianFor(cldrLoc)` 一次拉取并冻结到 `DateTimeFormat` 内部 slot。

---

## 4. 时区处理

### 4.1 TimeZone 选项解析

**MUST** 规则:

1. `Options.TimeZone` 入参 **必须**接受三种形态:
   - IANA canonical name:`"America/New_York"`
   - IANA link(向后兼容):`"US/Eastern"` —— **必须**通过 [SPEC 32 §CanonicalLink](./32-datetimeformat-tz.md#canonicallink) 解析为 canonical
   - UTC offset string:`"+05:30"` / `"-08:00"` / `"+00:00"` —— **必须**通过 [SPEC 32 §ParseOffsetString](./32-datetimeformat-tz.md#parseoffsetstring) 解析
2. **必须**在 `New` 时调用 `time.LoadLocation` 或 `internal/tz.Resolve`,缓存 `*time.Location` 到内部 slot;**禁止**在 `Format` 调用期解析时区名。
3. 解析失败 **必须**返回 `ErrInvalidOption` wrapped error,消息含 timezone 字符串。
4. `Options{TimeZone: ""}` 与未传选项语义相同:沿用宿主默认(对齐 ECMA-402 `DefaultTimeZone()`)。
5. UTC offset 形式的 `*time.Location` **必须**永远不 DST(直接 fixed-offset zone),与 formatjs `getApplicableZoneData` 行为对齐。

> **Why**: `Format` 路径每次重做 `time.LoadLocation` 是显著性能损失(每次 ~10 μs 文件查找,即使 `time/tzdata` 也要走解析路径)。active scope 性能阈值 ≤ 2 μs/op 要求时区在 `New` 时物化。
>
> **Rejected**:
> - 让 `Format(t)` 直接用 `t.Location()` 决定显示时区 —— 偏离 ECMA-402 fixed `[[TimeZone]]`。
> - `Format` 路径接受 `*time.Location` 参数 —— 违反 ECMA-402 单参数 `format(date)` 模型。

### 4.2 Format 期不查 TZ

**MUST** 规则:

1. `Format` / `FormatToParts` 路径 **必须不**调用 `time.LoadLocation` / `internal/tz.Resolve`。
2. `Format` 路径 **必须不**解析 `metaZones.json` 或做文件 I/O;时区显示名(`shortGeneric` / `longGeneric` 等)通过 generated `internal/cldr.TimeZoneDisplayName(cldrLoc, zone, form, isDST, instant, offsetMs)` 查表和 GMT fallback 得出。

> **Why**: hot path 不应做时区解析或运行时 JSON;CLDR metazone/exemplar city 数据已在 codegen 时编译为 Go 表,调用期只做内存查表与必要的 GMT fallback。

---

## 5. 格式化主流程(PartitionDateTimePattern)

`Format` 与 `FormatToParts` 共享主流程 `PartitionDateTimePattern`,定义于 `formatjs/.../PartitionDateTimePattern.ts` + `ToLocalTime.ts` + `FormatDateTimePattern.ts`:

```text
1. instant := t.UTC().UnixMilli()
2. localTime := ToLocalTime(instant, calendar, location)
   // localTime = {era, year, month, day, weekday, hour, minute, second, ms, dst, offset}
3. pattern := f.pattern  // 已在 New 时决议为 selectedPattern(SPEC 31)
4. parts := []Part{}
5. for token in pattern:
       if token is literal: parts += {Type: "literal", Value: token.text}
       else: parts += FormatField(localTime, token, dataLocale, numberingSystem)
6. return parts
```

**MUST** 规则:

1. `Format` **必须**等价于 `strings.Join(part.Value for part in FormatToParts(t), "")` —— 二者输出字符串字节相等,通过 conformance 测试断言。
2. `instant` **必须**通过 `t.UTC().UnixMilli()` 得到;**禁止**通过 `t.Unix() * 1000` 中转(精度损失)。
3. `ToLocalTime` 实现 **必须**在 `internal/ecma402/datetimeformat/tolocaltime.go`,接受 `(instant int64, calendar string, location *time.Location)`,返回结构化 `LocalTime`。
4. Pattern token → 字段格式化的查表 **必须**通过 [SPEC 31 §Skeleton 字符表](./31-datetimeformat-skeleton.md#字符表)。
5. `FormatField` 输出的 `Part.Type` **必须**限定 ECMA-402 §13.1.5 + formatjs 扩展共 16 种枚举字符串:`era | year | month | day | hour | minute | second | weekday | dayPeriod | timeZoneName | literal | ampm | relatedYear | yearName | unknown | fractionalSecondDigits`(与 `.references/formatjs/packages/ecma402-abstract/types/date-time.ts` `IntlDateTimeFormatPartType` 严格对齐;**禁止**使用非 spec 字符串如 `hour24` / `hour11` / `dayperiod`(小写))。
6. `DateTimeFormat` **必须**在 `New` 时缓存 `selectedPattern`,而不是在每次 `FormatToParts` 中重复选择 pattern。`selectedPattern` 至少包含 kind(`date | time | dateTime | none`)以及 date/time/dateTime pattern 字符串;style pattern、component skeleton pattern、date+time interpolation 都从这个结构进入格式化。
7. `localeMatcher` 与 `formatMatcher` **必须**保持分离:`localeMatcher` 只影响 CLDR data locale 选择;`formatMatcher` 只影响 component options → pattern 的选择。**禁止**用 locale fallback 结果替代 formatMatcher 决策。

```go
type Part struct {
    Type  string  // 严格枚举,不开放
    Value string
}
```

> **Why**: `Part.Type` 是 conformance fixture 的对齐键;开放枚举(允许任意字符串)会让 fixture 无法机械比较。`selectedPattern` 让 constructor-eager 的 pattern spine 成为显式状态,避免 style、component、fallback 分支在 hot path 里扩散。
>
> **Rejected**: `type PartType int` 枚举常量 —— 与 JSON fixture 的字符串字段比对时多一步映射,且失去 formatjs 字符串可读性。

---

## 6. FormatRange / FormatRangeToParts <a id="rangekind"></a>

**MUST** 规则:

1. `FormatRange(start, end)` **必须**实现 ECMA-402 §13.5.5 `FormatDateTimeRange` 三段式 fallback:
   ```text
   diff := largestFieldDiff(start, end)
   pattern := intervalFormats[skeleton][diff]
   if pattern == nil:
       pattern := "{0} – {1}"  // dateTimeFormats.intervalFormatFallback
   return apply(pattern, format(start), format(end))
   ```
2. `intervalFormats` 数据 **必须**通过 `internal/cldr.IntervalFormatsFor(loc, calendar)` 取得(数据来自 CLDR `ca-gregorian.json` `dateTimeFormats.intervalFormats`)。
3. `FormatRangeToParts` **必须**返回 `[]Part` 元素带 `Source` 字段(`"startRange" | "endRange" | "shared"`);`Part` 类型扩展为:
   ```go
   type RangePart struct {
       Type   string  // 同 Part.Type
       Value  string
       Source string  // "startRange" | "endRange" | "shared"
   }
   ```
4. `Source` 字符串常量 **必须**复用 `internal/ecma402.RangeKind`(同 [SPEC 20 §FormatRange](./20-numberformat.md#5-formatrange--formatrangetoparts) `RangeKind`),避免拼写漂移。
5. `start.After(end)` 的输入 **必须不**返回 error,也不得调换两端或加 `~`;按 ECMA-402 `PartitionDateTimeRangePattern` 使用入参顺序。
6. `start == end` 且 `intervalFormats` 命中"相同字段"时 **必须**回落到 `Format(start)` 单一格式化,**禁止**输出"X – X"。
7. 当前 generated range fixture gate 至少 **必须**覆盖 `yMMMd` skeleton 的 day-difference interval pattern(`intervalFormats["yMMMd"]["d"]`),并保留 CLDR literal spacing(例如 narrow no-break/thin spaces)而不是手写 `" – "`。

> **Why**: `intervalFormats` 三段式 fallback 是 conformance 测试中第二高错误率的算子(formatjs `tests/format-range.test.ts` 100+ fixture)。
>
> **Rejected**: 抽象 `CollapseRange[T]` 与 NumberFormat 共享 —— Part 域不同(详见 [SPEC 20 §FormatRange](./20-numberformat.md#5-formatrange--formatrangetoparts))。

---

## 7. 错误模型

**MUST** 规则:

1. **必须**定义 sentinel:
   - `ErrInvalidOption`(重导出 `internal/ecma402.ErrInvalidOption`)
   - `ErrUnsupportedTimeZone`(本包独有)
2. 构造期错误 **必须**用 `fmt.Errorf("datetimeformat: ...: %w", ErrXxx)` 包装,消息含字段名 + 用户值 + locale。
3. **禁止** `panic` 任何用户路径。
4. `time.Time` 无法表示 JavaScript invalid Date;不要为 Go 零值或普通可表示时间发明 invalid-date fallback。

```go
// 错误示例(签名)
err := fmt.Errorf("datetimeformat: unsupported timezone %q for locale %q: %w",
    tz, loc.String(), ErrUnsupportedTimeZone)
```

---

## 8. 性能目标

**MUST** 规则:

1. `New` 对常见 locale + `Options{DateStyle: MediumDateTimeStyle}` **必须** ≤ 10 μs/op(P50;含 `time.LoadLocation`)。
2. Cached `Format(time.Time)` 默认选项 **必须** ≤ 2 μs/op(对齐 [SPEC 71 §阈值](./71-benchmark.md#thresholds))。
3. `Format` 在 hot path **必须**零分配,除 string return value 外。

> **Why**: messageformat-go `:datetime` 函数对每条消息可能 N 次 `Format`;< 2 μs 才能保留消息层 SLA。

---

## 9. Forbidden

- **禁止** 在 `Format` / `FormatToParts` / `FormatRange` 路径返回 option error;Go typed input 无法表达 JavaScript invalid Date 时,不要发明额外错误。
- **禁止** 在 `Format` 路径调用 `time.LoadLocation` —— `*time.Location` 必须在 `New` 时缓存。
- **禁止** 在 `Format` 路径查 CLDR 时区显示名 —— `metaZones` 数据必须在 `New` 时物化。
- **禁止** active scope 生成 Gregorian pattern payload 之外的 calendar pattern 数据(包括 Buddhist 年偏移)。
- **禁止** 沿用 translate-agent 模式(直接用 `t.Location()` 决定显示时区)—— 偏离 ECMA-402。
- **禁止** 通过 `t.Unix() * 1000` 中转 instant —— 精度损失(秒级)。
- **禁止** 接受 `any` 入参或 `*time.Time` —— 类型安全弱化。
- **禁止** `Format` 期不剥离 monotonic clock —— `t.Round(0)` 必须在方法入口立即调用。
- **禁止** 把 `Part.Type` 改为 `int` 常量枚举 —— 与 JSON conformance fixture 字符串字段比对时多一步映射。
- **禁止** 与 NumberFormat 共享 `CollapseRange` —— Part 域不同。
- **禁止** 引入 ICU C++ 依赖或 cgo 路径(SPEC 00 §1.1 非目标)。

---

## 10. Acceptance Criteria

- [ ] `formatjs/packages/intl-datetimeformat/tests/format.test.ts` 全部 fixture 在 `datetimeformat/conformance_test.go` 通过(byte-equality)。
- [ ] `formatjs/packages/intl-datetimeformat/tests/format-range.test.ts` 全部 fixture 通过。
- [ ] `formatjs/packages/intl-datetimeformat/tests/offset-timezone.test.ts` 全部 fixture 通过(`+05:30` / `-08:00` 等输入)。
- [ ] `go test -race ./datetimeformat/...` 通过(含 `TestDateTimeFormat_TimezoneContextPreservation`:同一 `time.Time` 在不同选项 `TimeZone` 下输出不同)。
- [ ] `go test -race ./datetimeformat/...` 通过(含 `TestDateTimeFormat_MonotonicClockStripping`:`t.Round(0)` 后多次 `Format(t)` 字节相等)。
- [ ] `go test -race ./datetimeformat/...` 通过(含 `TestDateTimeFormat_ConcurrentFormat` 100 goroutine × 1000 调用)。
- [ ] `go vet ./datetimeformat/...` 干净。
- [ ] `New(loc, Options{Calendar: "buddhist"})` 成功并通过 `ResolveLocale` 回落到支持 calendar;`ResolvedOptions().Calendar == "gregory"`。
- [ ] `New(loc, Options{TimeZone: "Mars/Olympus"})` 返回 `ErrUnsupportedTimeZone` wrapped error。
- [ ] `New(loc, Options{TimeZone: "US/Eastern"})` 成功;`ResolvedOptions().TimeZone == "America/New_York"`(canonical name 转换)。
- [ ] `New(loc, Options{TimeZone: "+05:30"})` 成功;`ResolvedOptions().TimeZone == "+05:30"`(offset string 保留)。
- [ ] `New(loc-en-US, Options{Hour12: Hour12(false), Hour: NumericFieldStyle})`:`ResolvedOptions().HourCycle == "h23"`。
- [ ] `New(loc-fr-FR, Options{Hour: NumericFieldStyle})`:`ResolvedOptions().HourCycle == "h23"`(法语默认 h23)。
- [ ] benchmark `BenchmarkDateTimeFormat_Cached_DateStyle_Medium` ≤ 2 μs/op(SPEC 71 阈值)。
- [ ] benchmark `BenchmarkDateTimeFormat_New` ≤ 10 μs/op(P50)。

---

## 11. References

### Primary

- `.references/formatjs/packages/intl-datetimeformat/core.ts` — 公开 API、`tzData` 注入路径
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/InitializeDateTimeFormat.ts` — 选项管线
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/PartitionDateTimePattern.ts` — 主流程
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/ToLocalTime.ts` — UTC → 本地时(含 `+offset` 与 IANA 两路、calendar=gregory invariant)
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/FormatDateTimePattern.ts` — Part.Type 枚举
- `.references/formatjs/packages/intl-datetimeformat/tests/` — 主 conformance 来源

- `.references/intl/intl.go` — translate-agent/intl(Go DateTimeFormat 先例,但忽略 TimeZone,我们不学)
- `.references/ext/src/ecma402/{calendar,hour_cycle,time_zone}.{c,h}` — PHP/ICU 标识符校验路径

### Project Cross-References

- [SPEC 12 §Abstract Ops](./12-abstract-operations.md) — shared validators / pattern helpers / `ErrInvalidOption`
- [SPEC 10 §Locale 结构](./10-locale.md#locale-结构) — `Locale.HourCycle()` / `Locale.Calendar()`
- [SPEC 31 §Skeleton](./31-datetimeformat-skeleton.md#skeleton) — Skeleton 字符表 + BestFitFormatMatcher
- [SPEC 32 §TZ Data](./32-datetimeformat-tz.md#tz-data) — `time/tzdata` 注入 + metaZones
- [SPEC 32 §Calendar Data](./32-datetimeformat-tz.md#calendar-data) — Gregorian 数据 schema
- [SPEC 50 §Schema](./50-cldr-data.md#schema) — `ca-gregorian.json` 数据形态
- [SPEC 60](./60-facade.md) — root namespace ownership; root `intl.FormatDate` one-shot helpers are outside the long-term public surface.
- [SPEC 71 §阈值](./71-benchmark.md#thresholds) — 性能基线
