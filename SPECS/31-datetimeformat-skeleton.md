# SPEC 31 — DateTimeFormat Skeleton & BestFitFormatMatcher

> **Status:** Draft (2026-05-08)
> **Owner:** `internal/ecma402/datetimeformat/skeleton.go` + `internal/ecma402/datetimeformat/matcher.go`
> **Reference contract:** `formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts` + `BestFitFormatMatcher.ts` + `BasicFormatMatcher.ts`

## Overview

定义 ECMA-402 选项 → CLDR skeleton(如 `yMMMd`)→ pattern(如 `MMM d, y`)三段式流水线在 go-intl 的算法层实现:

1. **Skeleton 解析**(`internal/ecma402/datetimeformat/skeleton.go`):LDML TR35 字符 → `Intl.DateTimeFormatOptions` 字段映射,无状态算法。
2. **BestFitFormatMatcher / BasicFormatMatcher**(`internal/ecma402/datetimeformat/matcher.go`):评分函数 + ICU `adjustFieldTypes` 后处理。
3. **数据接入边界**:算法层禁 import `internal/cldr`,数据切片由 [SPEC 30 §Calendar 数据访问](./30-datetimeformat.md#32-calendar-数据访问) 调用方注入。

本 SPEC 不重定义:
- `DateTimeFormat` 公开 API → 见 [SPEC 30 §公开 API](./30-datetimeformat.md#1-公开-api)
- CLDR `availableFormats` / `intervalFormats` 数据 schema → 见 [SPEC 50 §Schema](./50-cldr-data.md#schema)
- `Locale.HourCycle()` 联动 → 见 [SPEC 30 §HourCycle 联动](./30-datetimeformat.md#22-hourcycle-联动132111)

---

## 1. 分层契约 <a id="skeleton"></a>

### 1.1 算法层与数据层边界

```text
internal/ecma402/datetimeformat/
├── skeleton.go    # 算法:DATE_TIME_REGEX、matchSkeletonPattern、parseDateTimeSkeleton
├── matcher.go     # 算法:BasicFormatMatcher、BestFitFormatMatcher、adjustFieldTypes
└── tokens.go      # 类型:TokenLabel、Formats、SkeletonField

internal/cldr/
├── dates.go       # 数据:per-locale availableFormats / intervalFormats / formats 表
└── ...
```

**MUST** 规则:

1. `internal/ecma402/datetimeformat/` 算法层 **必须**无状态(无包级可变 var,无 `init()` 副作用)。
2. `internal/ecma402/datetimeformat/` **必须不**直接 import `internal/cldr`;CLDR 数据切片由调用方(`datetimeformat/` 包)在 `New` 时通过参数注入。
3. `internal/cldr/dates.go` 是 per-locale 数据 SSOT,由 codegen 产出,**禁止**手写映射表。
4. `tokens.go` 中的 `Formats` 结构(per-skeleton pattern 表)**必须**与 formatjs `parseDateTimeSkeleton` 输出的 `Formats` 字段一一对齐。

> **Why**: 升级 LDML 不触发数据再生(算法稳定);升级 CLDR 不触发算法修改(数据稳定);conformance 测试可独立替换算法或数据做差分。
>
> **Rejected**: 算法与数据合并到 `internal/cldr` —— 升级路径耦合,算法测试需要 mock CLDR 数据。
>
> **Rejected**: translate-agent 风格"穷举分支"(16 个 `seqEraYearMonthDay` 函数)—— `Intl.DateTimeFormatOptions` 共 11 个互斥/可选字段,二项式展开 2^11 + dateStyle/timeStyle 5×5 不可行。

---

## 2. Skeleton 字符表 <a id="字符表"></a>

### 2.1 字符表来源

字符表来源 **必须**是 LDML TR35 Date Field Symbol Table。formatjs 的 `DATE_TIME_REGEX` 是权威移植入口:

```text
/(?:[Eec]{1,6}|G{1,5}|[Qq]{1,5}|(?:[yYur]+|U{1,5})|[ML]{1,5}|d{1,2}|D{1,3}|F{1}|[abB]{1,5}|[hkHK]{1,2}|w{1,2}|W{1}|m{1,2}|s{1,2}|[zZOvVxX]{1,4})(?=([^']*'[^']*')*[^']*$)/g
```

**MUST** 规则:

1. `internal/ecma402/datetimeformat/skeleton.go` **必须**移植上述正则及全部分支语义。
2. 字符 → 选项字段映射(`matchSkeletonPattern`)**必须**与 formatjs 1:1 对齐:

| LDML 字符 | 对应 Options 字段 | 字段值依字符长度 |
|-----------|------------------|------------------|
| `G G GG GGG` | `Era` | `short`(`G/GG/GGG`) / `long`(`GGGG`) / `narrow`(`GGGGG`) |
| `y yy yyyy ...` | `Year` | `2-digit`(`yy`) / `numeric`(其他) |
| `M MM MMM MMMM MMMMM` | `Month` | `numeric/2-digit/short/long/narrow` |
| `L LL LLL LLLL LLLLL` | `Month`(stand-alone) | 同上 |
| `d dd` | `Day` | `numeric/2-digit` |
| `E EE EEE EEEE EEEEE EEEEEE` | `Weekday` | `short` × 3 / `long` / `narrow` / `short` |
| `c cc ccc cccc ccccc cccccc` | `Weekday`(stand-alone) | 同 E |
| `h hh / H HH / k kk / K KK` | `Hour` + `HourCycle` | `numeric/2-digit`;hourCycle 由字符决定(h12/h23/h24/h11) |
| `m mm` | `Minute` | `numeric/2-digit` |
| `s ss` | `Second` | `numeric/2-digit` |
| `S SS SSS` | `FractionalSecondDigits` | 1/2/3 |
| `a aa aaa aaaa aaaaa` | `DayPeriod` | `short`/`long`/`narrow` |
| `b bb ... B BB ...` | `DayPeriod`(flexible)| 同 a |
| `z zz zzz zzzz` | `TimeZoneName` | `short` / `long`(`zzzz`) |
| `Z ZZ ZZZ ZZZZ` | `TimeZoneName` | `shortOffset` / `longOffset`(`ZZZZ`) |
| `O OOOO` | `TimeZoneName` | `shortOffset` / `longOffset` |
| `v vvvv` | `TimeZoneName` | `shortGeneric` / `longGeneric` |
| `V VV VVV VVVV` | `TimeZoneName` | IANA name / city / fallback |
| `X XX XXX XXXX XXXXX` | `TimeZoneName` | offset variants |

3. `parseDateTimeSkeleton(skeleton string, hour12 *bool, hourCycle string, ...) Formats` **必须**返回 `Formats{...}` 结构(对应 ECMA-402 §13.1.2):
   ```go
   type Formats struct {
       Pattern               string  // 解析后的 pattern 字符串(含 literal 引号)
       Pattern12             string  // 12 制版本(如可用)
       Skeleton              string  // 原 skeleton
       Era                   FieldStyle
       Year                  NumericStyle
       Month                 FieldStyle
       Day                   NumericStyle
       Weekday               FieldStyle
       Hour                  NumericStyle
       HourCycle             HourCycle
       Minute                NumericStyle
       Second                NumericStyle
       DayPeriod             FieldStyle
       FractionalSecondDigits int
       TimeZoneName          TimeZoneName
   }
   ```

> **Why**: LDML 字符表是 ICU/CLDR 跨实现一致点;formatjs 已完成 TS 端权威移植,Go 端机械翻译即可保证字节相等。
>
> **Rejected**: 重新设计字符 → 字段表 —— 与 formatjs 不一致即破坏字节相等。

### 2.2 字面量(literal)与转义

**MUST** 规则:

1. Skeleton 中单引号包围的子串 **必须**视为 literal,跳过字符表查找(对齐 LDML TR35 quote 语义)。例:`'Year' yyyy` 中 `'Year'` 是 literal `"Year"`。
2. 双单引号 `''` **必须**输出为单引号字符 `'`(对齐 LDML escape)。
3. `processDateTimePattern(rawPattern string, result *Formats)` **必须**正确处理 literal,不让 literal 字符触发 `matchSkeletonPattern`。

---

## 3. BestFitFormatMatcher 评分

### 3.1 评分算法

`BestFitFormatMatcher` 是 ECMA-402 默认 `formatMatcher`,实现 §13.1.1.2 算法:

```text
function bestFitFormatMatcherScore(options DateTimeFormatOptions, format Formats) int:
    score := 0
    if options.Hour12 != format.HourCycle 是否 12 制等价:
        score -= removalPenalty 或 additionPenalty
    for each field in DATE_TIME_PROPS:
        optVal := options[field]
        fmtVal := format[field]
        if optVal == nil and fmtVal != nil:
            score -= additionPenalty           // 多余字段 +20(扣分)
        else if optVal != nil and fmtVal == nil:
            score -= removalPenalty            // 缺失字段 +120(扣分)
        else if optVal != nil and fmtVal != nil:
            if optVal != fmtVal:
                if 数值/字母 不同类:
                    score -= differentNumericTypePenalty
                else:
                    delta := |index(optVal) - index(fmtVal)|  // [2-digit, numeric, narrow, short, long]
                    score -= delta == 2 ? longMore : shortMore  // 或 longLess / shortLess
    return score
```

**MUST** 规则:

1. 计分常量 **必须**与 §13.1.1.2 对齐:
   - `removalPenalty = 120`(缺失字段)
   - `additionPenalty = 20`(多余字段)
   - `longMore = 6`(长度差 2,方向"更长")
   - `shortMore = 3`(长度差 1,方向"更长")
   - `longLess = 8`(长度差 2,方向"更短")
   - `shortLess = 6`(长度差 1,方向"更短")
   - `differentNumericTypePenalty = 15`(数值/字母同字段但不同类型)
   - `offsetPenalty = 1`(timezone generic/specific 与 offset fallback 的差异)
2. `DATE_TIME_PROPS` **必须**是固定字段序列:`Weekday | Era | Year | Month | Day | DayPeriod | Hour | Minute | Second | FractionalSecondDigits | TimeZoneName`(对齐 formatjs `DateTimeFormat/utils.ts`)。
3. 选出最高分(分数最大,扣分最少)的 `Formats`;并列时取**第一个**(对齐 formatjs `formats[0]` 语义)。

> **Why**: 计分常量是 ICU `DateTimePatternGenerator` 的固化值,FormatJS 使用该值;任何偏差都改变 skeleton → pattern 的命中结果,破坏字节相等。

### 3.2 adjustFieldTypes 后处理

**MUST** 规则:

1. 选出最佳 `Formats` 后 **必须**调用 `adjustFieldTypes(format, options)`,根据 `options` 的字段值修改 `format.Pattern` 中对应字段的字符长度。例:`format.Year = numeric` 但 `options.Year = 2-digit`,则把 pattern 中 `y` 替换为 `yy`。
2. 替换规则 **必须**与 formatjs `BestFitFormatMatcher.ts` 一致:可以把 alphabetic pattern 调整到请求宽度,但不得把 numeric pattern 强行改成 alphabetic month/day-period 形态。例:中文 `yMEd` 的 pattern `y年M月d日E` 即使请求 `month: "long"` 也保持 `M月`,对齐 FormatJS。
3. `adjustFieldTypes` 的 pattern 扫描 **必须**保持 ASCII byte 级:LDML pattern field 字符均为 ASCII,字段 membership 可用 `strings.IndexByte` 等 stdlib byte helper;禁止改成 rune/regex 扫描。
4. pattern 扫描 loop **必须**保留显式 index 前进,因为 quoted literal 与重复字段宽度会一次跳过多字节段;禁止机械改成 `for range len(pattern)` 后在循环体内隐藏 index 跳跃。
5. `adjustFieldTypes` 后,`format.Pattern` 已是最终 pattern 字符串,可直接送入 `FormatDateTimePattern`。

> **Why**: `adjustFieldTypes` 是 ICU `DateTimePatternGenerator::adjustFieldTypes` 的等价物;不做后处理即不能让 best-fit pattern 适配用户的精确长度需求。
> ASCII byte scanning keeps the algorithm aligned with LDML field grammar while avoiding a private one-off membership helper.

### 3.3 BasicFormatMatcher

**MUST** 规则:

1. **必须**实现 `BasicFormatMatcher`(§13.1.1.1):同样的评分算法,但**不**做 `adjustFieldTypes` 后处理。
2. **必须**支持 `formatMatcher: "basic"` 选项;默认值是 `"best fit"`。
3. `BasicFormatMatcher` 的存在主要为 conformance 测试覆盖(§13.1.1.1 算法对照)。

> **Why**: §13.1.1.1 是 spec 强制条款,即使默认 best-fit 也必须实现 basic。

---

## 4. Skeleton → Pattern 数据接入

### 4.1 数据来源

**MUST** 规则:

1. Skeleton → Pattern 映射 **必须**走 CLDR `cldr-dates-full/main/<locale>/ca-gregorian.json` 的 `dateTimeFormats.availableFormats` 节点。
2. Range pattern 选择 **必须**走同文件的 `dateTimeFormats.intervalFormats` 节点。
3. `dateFormat / timeFormat / dateTimeFormat` 的 `full / long / medium / short` **必须**走 `dates.calendars.gregorian.dateFormats` / `timeFormats` / `dateTimeFormats`。
4. **禁止**手写 skeleton → pattern 映射表;**禁止**通过 `// hard-coded` 注释跳过 codegen。

> **Why**: CLDR 是 SSOT;手写表会在 CLDR 升级时漂移,且 ~100 locale × 几十种 skeleton 不可手维护。

### 4.2 数据切片注入

**MUST** 规则:

1. Skeleton 算法层 **必须**接受 `Formats[]` 切片作为输入参数;**禁止**算法层自行查 CLDR。
2. 调用流(`datetimeformat.New` 内部):
   ```text
   1. cldrFormats := internal/cldr.AvailableFormatsFor(loc, "gregorian")
   2. result      := skeleton.Match(options, cldrFormats)  // BestFitFormatMatcher
   3. f.pattern   := result.Pattern                       // 缓存到 DateTimeFormat slot
   ```
3. `Formats[]` 切片 **必须**已经过 `parseDateTimeSkeleton` 解析(在 codegen 时,或在 `internal/cldr` accessor 内 lazy 解析后缓存),算法层接收的是已解析的 `Formats` 结构。

```go
// 算法层签名(无实现)
package skeleton

func Match(opts Options, formats []Formats) Formats           // BestFitFormatMatcher
func MatchBasic(opts Options, formats []Formats) Formats      // BasicFormatMatcher
func Parse(skeleton string, hour12 *bool, hc HourCycle) Formats // 单独 skeleton 字符串解析
```

> **Why**: 算法层无状态、可独立测试;数据切片由调用方注入,符合 CLAUDE.md "no production Go code beyond signatures" 与 KISS 原则。

---

## 5. ICU Skeleton 字符串语法兼容

**MUST** 规则:

1. `parseDateTimeSkeleton` **必须**兼容 ICU skeleton 字符串语法(用户可显式传入 skeleton 字符串如 `"yMMMd"` `"hms"`)。
2. 兼容范围:LDML TR35 §2.6.1 Date Field Symbol Table 全部字符。
3. `parseDateTimeSkeleton` **必须不**改变字段字符顺序;输出 `Formats.Skeleton` 字段保留输入字符串原值。

> **Why**: messageformat-go ICU MessageFormat 解析器会传入 skeleton 字符串作为函数选项;不兼容即不能集成。

---

## 6. Performance

**MUST** 规则:

1. `Match(opts, formats[])` 算法层 **必须** ≤ 500 ns/op(典型 ~30 个 candidate formats)。
2. `Match` **必须**零分配(除返回值);candidate 切片不做拷贝。
3. `parseDateTimeSkeleton(s)` 单次解析 **必须** ≤ 1 μs/op(典型 skeleton ≤ 12 字符)。

> **Why**: `New` 期 P50 ≤ 10 μs/op(SPEC 30 §性能目标)的预算中,skeleton 解析 + matcher 占大头;500 ns 是预留预算。

---

## 7. Forbidden

- **禁止** `internal/ecma402/datetimeformat/` 算法层 import `internal/cldr` —— 必须由调用方注入数据。
- **禁止** 手写 skeleton → pattern 映射表(per-locale)—— 必须 codegen。
- **禁止** translate-agent 风格的"穷举 16 个 seq* 函数" —— 不支持任意 skeleton 字符串。
- **禁止** 调用 ICU `DateTimePatternGenerator`(无 cgo / 无 ICU)。
- **禁止** 修改 formatjs 当前计分常量(`removalPenalty=120 / additionPenalty=20 / longMore=6 / shortMore=3 / longLess=8 / shortLess=6 / differentNumericTypePenalty=15 / offsetPenalty=1`),除非先用参考实现 diff 证明 formatjs 已变更。
- **禁止** 跳过 `adjustFieldTypes` 后处理 —— 即使 best-fit 命中精确,长度调整仍要做。
- **禁止** 算法层使用 `init()` 或包级可变 var(必须无状态)。
- **禁止** `parseDateTimeSkeleton` 改变 LDML 字符语义(每字符长度的字段值必须与 formatjs 1:1)。
- **禁止** 在 `Format` 路径调用 `Match` 或 `parseDateTimeSkeleton` —— pattern 必须在 `New` 时缓存。

---

## 8. Acceptance Criteria

- [ ] `formatjs/packages/intl-datetimeformat/tests/skeleton-resolution.test.ts` 全部 fixture 在 `internal/ecma402/datetimeformat/skeleton_test.go` 通过(byte-equality)。
- [ ] `formatjs/packages/intl-datetimeformat/tests/best-fit-format-matcher.test.ts` 全部 fixture 在 `internal/ecma402/datetimeformat/matcher_test.go` 通过。
- [ ] `formatjs/packages/intl-datetimeformat/tests/basic-format-matcher.test.ts` 全部 fixture 通过。
- [ ] `parseDateTimeSkeleton("yMMMd")` 返回 `Formats{Year:numeric, Month:short, Day:numeric, Skeleton:"yMMMd"}`。
- [ ] `parseDateTimeSkeleton("hms")` 返回 `Formats{Hour:numeric, HourCycle:h12, Minute:numeric, Second:numeric}`。
- [ ] `parseDateTimeSkeleton("HHmmss")` 返回 `Formats{Hour:2-digit, HourCycle:h23, Minute:2-digit, Second:2-digit}`。
- [ ] `Match(opts={Year:numeric, Month:long}, formats=...)` 在 `en-US` 下命中 `"MMMM y"` 而非 `"y"` 或 `"MMMM"`(完整字段优先)。
- [ ] `Match` 算法层独立测试时,**不需要** import `internal/cldr`(import path 检查通过)。
- [ ] `internal/ecma402/datetimeformat/` 包级 `init()` 数 = 0(`grep -c "func init" *.go == 0`)。
- [ ] `adjustFieldTypes(format={Year:numeric, Pattern:"y"}, opts={Year:"2-digit"})` 修改 `format.Pattern` 为 `"yy"`。
- [ ] `BasicFormatMatcher` 调用路径与 `BestFitFormatMatcher` 区别仅在 `adjustFieldTypes` 调用(用 `formatMatcher: "basic"` 选项验证)。
- [ ] benchmark `BenchmarkSkeleton_Match` ≤ 500 ns/op。
- [ ] benchmark `BenchmarkSkeleton_Parse` ≤ 1 μs/op。
- [ ] LDML literal 测试:`parseDateTimeSkeleton("'Year:' y")` 输出 `Formats.Pattern` 包含 literal `"Year:"`,且 `Year:numeric`。
- [ ] LDML escape 测试:`parseDateTimeSkeleton("'it''s' y")` literal 部分输出 `"it's"`。

---

## 9. References

### Primary

- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts` — `DATE_TIME_REGEX`、`matchSkeletonPattern`、`parseDateTimeSkeleton`、`processDateTimePattern`
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/BestFitFormatMatcher.ts` — `bestFitFormatMatcherScore`、`adjustFieldTypes`、计分常量
- `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/BasicFormatMatcher.ts` — `BasicFormatMatcher` 算法
- `.references/formatjs/packages/intl-datetimeformat/scripts/extract-dates.ts` — CLDR `availableFormats` 抽取路径(数据生成参考)
- `.references/formatjs/packages/intl-datetimeformat/tests/skeleton-resolution.test.ts` — 主 conformance fixture
- `.references/formatjs/packages/intl-datetimeformat/tests/best-fit-format-matcher.test.ts` — matcher fixture
- `.references/formatjs/packages/intl-datetimeformat/tests/basic-format-matcher.test.ts` — basic matcher fixture

### Secondary

- LDML TR35 §2.6.1 Date Field Symbol Table — 字符表权威文档(Unicode 标准)
- ICU4J `com.ibm.icu.text.DateTimePatternGenerator` — `adjustFieldTypes` C++ / Java 等价

### Project Cross-References

- [SPEC 30 §公开 API](./30-datetimeformat.md#1-公开-api) — `DateTimeFormat.New` 调用方
- [SPEC 30 §HourCycle 联动](./30-datetimeformat.md#22-hourcycle-联动132111) — `Hour12 / HourCycle / Locale.HourCycle()` 解析
- [SPEC 32 §Calendar Data](./32-datetimeformat-tz.md#calendar-data) — `internal/cldr` Gregorian 数据访问层
- [SPEC 50 §Schema](./50-cldr-data.md#schema) — `availableFormats` / `intervalFormats` JSON schema
