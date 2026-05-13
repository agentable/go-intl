# SPEC 10 — Locale

> **Status:** Draft (2026-05-08)
> **Priority:** High(所有 formatter 入参类型;阻塞 SPEC 11 / 20 / 30 / 40 / 60)
> **Authority:** 本 SPEC 是 `locale.Locale` 类型、ECMA-402 `Intl.Locale` 对齐、解析与规范化、`Maximize` / `Minimize`、只读属性 getter、`String` / `Equal` / `MarshalText` / `UnmarshalText` 的 SSOT。Normative source: `.references/ecma402/spec/locale.html`.

---

## Overview

`locale.Locale` 是 go-intl 公开 API 中所有 locale-aware 操作的入参类型。它是 `Intl.Locale` 的 Go 表示,包装 [`golang.org/x/text/language.Tag`](https://pkg.go.dev/golang.org/x/text/language#Tag) 的 BCP 47 解析能力,叠加 ECMA-402 / UTS #35 Unicode 扩展状态,并通过只读 getter 暴露 spec 要求的属性。

本 SPEC 定义 `Locale` 结构、构造函数(`New` / `Parse` / `MustParse`)、规范化(`Maximize` / `Minimize`)、getter 物化策略、字符串往返、可比性。**不**定义 best-fit 匹配算法(SPEC 11)、CLDR 数据格式(SPEC 50)、formatter 内部 slot(SPEC 12 / 20 / 30 / 40)。

---

## 1. Locale Structure <a id="locale-结构"></a>

### 1.1 决策:不可变 `language.Tag` + extension state

```go
// locale/locale.go(签名)
package locale

import "golang.org/x/text/language"

// Locale 是 ECMA-402 Intl.Locale 的 Go 表示。
// 字段不导出;调用方必须通过 Parse/New 构造,通过 getter 读取。
type Locale struct {
    tag language.Tag
    ext extensions
}

func (l Locale) BaseName() string
func (l Locale) Tag() language.Tag
func (l Locale) Calendar() string
func (l Locale) Collation() string
func (l Locale) HourCycle() string
func (l Locale) CaseFirst() string
func (l Locale) Numeric() bool
func (l Locale) NumberingSystem() string
func (l Locale) Language() string
func (l Locale) Script() string
func (l Locale) Region() string
func (l Locale) Variants() []string
```

> **Why**:
> 1. **复用** —— `language.Tag` 已实现 BCP 47 解析、规范化、`Base` / `Script` / `Region` / `Variants` 访问;Go 生态(`x/text/message` / `x/text/currency` / `messageformat-go`)统一以它为底层句柄。
> 2. **只读模型** —— ECMA-402 `Intl.Locale.prototype` 属性是 accessor property 且无 setter。Go 端必须用未导出字段 + getter 表达同一不可变模型。
> 3. **半透传** —— `language.Tag` 通过 `TypeForKey("ca")` 等可读取 `-u-` 扩展键的字符串值,**但不暴露**为类型化字段、不做 `numeric` 字符串↔bool 转换、不识别 `caseFirst="false"` 的字面量合法性。go-intl 必须自己持有 extension state。
> 4. **值类型** —— struct(非 `*Locale`)保持值语义;调用方按值传递,`Locale` 是不可变快照。
>
> **Rejected**:
> - **`type Locale = language.Tag`(类型别名)**:扩展字段无处放置,违反 spec `Intl.Locale.calendar` 等独立 getter 语义。
> - **平铺无 Tag**(把 `language` / `script` / `region` 都自己存):重新发明 BCP 47 解析,违反 CLAUDE.md "no reinventing locale parsing"。
> - **导出字段 struct**:调用方可以绕过 constructor 制造非法状态,违反 `Intl.Locale` 只读属性模型。
> - **`*Locale` 指针类型**:违反 SPEC 00 §1 值类型偏好;`Equal` / `String` 用值语义更清晰。
> - **PHP 风格全字段平铺**(把 `language` / `script` / `region` / `calendar` / `hourCycle` 全部 `string`):放弃 `language.Tag` 的解析能力。

### 1.2 字段类型选择

| 字段 | 类型 | 默认零值含义 |
|------|------|-------------|
| `Calendar` | `string` | `""` = 未指定,formatter 走 region 默认 |
| `Collation` | `string` | `""` = 未指定 |
| `HourCycle` | `string` | `""` = 未指定,formatter 走 region 默认 |
| `CaseFirst` | `string` | `""` = 未指定;`"false"` 字面量是合法值(spec 强制) |
| `Numeric` | `bool` | `false` = spec 默认(不开启 numeric collation) |
| `NumberingSystem` | `string` | `""` = 未指定,formatter 走 locale 默认 |
| `FirstDayOfWeek` | `string` | `""` = 未指定,formatter 走 region 默认 |

> **Why `Numeric bool` not `*bool`**:spec 默认即 `false`;用户极少需"未设" vs "明确 false"区分。`bool` 免去 `nil` 检查,值语义自然。
>
> **Rejected `Numeric *bool`**:三态(`nil` / `false` / `true`)增加调用方负担;spec sec-applyunicodeextensiontotag 把缺省视同 `kn=false`。

---

## 2. Construction & Parsing

### 2.1 入口签名

```go
// locale/locale.go(签名)
package locale

// Parse 从 BCP 47 字符串解析 Locale。支持 -u- 扩展键 ca/co/hc/kf/kn/nu/fw。
// 解析失败返回 ErrInvalidLocale 包装的错误;边界统一在此一次性发生。
func Parse(s string) (Locale, error)

// MustParse 等价 Parse 但失败 panic。仅用于 const-like 场景(测试 / 示例)。
func MustParse(s string) Locale

// New 从 language.Tag 构造 Locale,通过 Options 对齐 Intl.Locale(tag, options)。
// 已设的扩展字段会进入返回 Locale 的 String() canonical 输出。
func New(tag language.Tag, options ...Options) (Locale, error)

type Options struct {
    Calendar        string
    Collation       string
    HourCycle       string
    CaseFirst       string
    Numeric         *bool
    NumberingSystem string
    FirstDayOfWeek  string
}
```

调用示例:

```go
loc, err := locale.Parse("en-US-u-hc-h23-ca-buddhist")
fmt.Println(loc.HourCycle())   // "h23"
fmt.Println(loc.Calendar())    // "buddhist"

loc2, err := locale.New(language.Japanese, locale.Options{Calendar: "japanese"})
fmt.Println(loc2.String())   // "ja-u-ca-japanese"
```

### 2.2 解析行为契约

| 输入 | 行为 |
|------|------|
| `"en-US"` | OK;`Tag = en-US`,所有扩展字段空 |
| `"en-US-u-hc-h23"` | OK;`HourCycle="h23"`,Tag 保留 `-u-hc-h23` |
| `"en-US-u-ca-buddhist-hc-h23"` | OK;字段顺序按 spec 规范化(见 §3.2) |
| `"en-US-u-ca-gregorian"` | OK;接受 spec 别名,内部归一为 `Calendar="gregory"` |
| `"en-US-u-ca-islamic-civil"` | OK;接受 spec 别名,内部归一为 `Calendar="islamicc"` |
| `"xx-INVALID"` | 返回 `fmt.Errorf("locale: %q: %w", s, ErrInvalidLocale)` |
| `""` | 返回 ErrInvalidLocale(空串非法) |
| `"en-US-u-kn"`(无值表 true) | OK;`Numeric=true` |
| `"en-US-u-kn-false"` | OK;`Numeric=false` |
| `"en-US-u-kf-false"` | OK;`CaseFirst="false"` (spec 字面量合法) |

> **Why**:
> 1. **边界唯一** —— Parse 是 BCP 47 边界,公共 API 不再接收 raw `string` locale(SPEC 11 / 20 / 30 / 40 入参全部 `Locale`)。
> 2. **spec 别名** —— `gregorian` → `gregory`、`islamic-civil` → `islamicc` 是 spec sec-applyunicodeextensiontotag 必须遵守的归一(formatjs `index.ts:62-90` 实现);`Parse` 内部完成归一,避免下游逻辑分支。
> 3. **`MustParse` 仅限 const-like 场景** —— 测试 / Example;生产代码必须用 `Parse` 检查 error。
>
> **Rejected**:
> - **接收 `string` 在 formatter 层**(`numberformat.New("en-US", ...)`):违反 CLAUDE.md "Locale arguments are typed (`locale.Locale` or `language.Tag`) — never raw `string`. Parsing happens at the boundary, once."
> - **`Parse(s) Locale`(panic on error)**:spec RangeError 必须返回 error,违反 CLAUDE.md "no panic in production code"。

### 2.3 构造期验证

`Parse` / `New` **必须**对 7 个扩展字段做 spec 校验:

| 字段 | 校验规则 |
|------|---------|
| `Calendar` | 非空时:必须匹配 BCP 47 type 子标签语法(2–8 字符 alphanumeric);**不**校验是否在 CLDR 中存在(formatter 层负责) |
| `Collation` | 非空时:同上 |
| `HourCycle` | 非空时:必须 ∈ {`"h11"`, `"h12"`, `"h23"`, `"h24"`} |
| `CaseFirst` | 非空时:必须 ∈ {`"upper"`, `"lower"`, `"false"`} |
| `Numeric` | bool,无校验 |
| `NumberingSystem` | 非空时:BCP 47 type 子标签语法 |
| `FirstDayOfWeek` | 非空时:必须 ∈ {`"mon"`, `"tue"`, `"wed"`, `"thu"`, `"fri"`, `"sat"`, `"sun"`, `"0"`, `"1"`, `"2"`, `"3"`, `"4"`, `"5"`, `"6"`, `"7"`};数字按 ECMA-402 `WeekdayToUValue` 规范化 |

校验失败返回 `fmt.Errorf("locale: invalid <field> %q: %w", value, ErrInvalidLocale)`。

> **Why**: 构造期一次性校验,formatter 层不再二次验证语法;但是否在 CLDR 中可解析(如 `Calendar="vulcan"`)由 formatter 在数据查找时报错(`numberformat`/`datetimeformat` 自行 sentinel)。
>
> **Rejected**: 构造期就检查 CLDR 数据存在性(违反层次:构造期不依赖 CLDR;formatter 层才需要 CLDR 数据)。

---

## 3. String Round-trip & Canonicalization

### 3.1 `String()` 契约

```go
// (Locale).String() 返回 canonical BCP 47 表示,含 -u- 扩展。
// 与 Parse 互逆:locale.MustParse(loc.String()).Equal(loc) == true。
func (l Locale) String() string
```

### 3.2 Canonicalization 规则

`String()` 输出 **必须**:

1. **subtag 顺序**:`language[-script][-region][-variants...]`(由 `language.Tag.String()` 给出)。
2. **`-u-` 扩展键按字典序**:`ca` < `co` < `fw` < `hc` < `kf` < `kn` < `nu`。
3. **`-u-` 扩展值小写**(`Calendar="GREGORY"` 输出 `-u-ca-gregory`)。
4. **空字段不输出**(`Calendar=""` 不写 `-u-ca-`)。
5. **Numeric=true 输出 `-u-kn`**(无值表 true,与 spec 一致);Numeric=false 不输出(默认值省略)。
6. **CaseFirst=`"false"` 输出 `-u-kf-false`**(字面量合法值)。
7. **Calendar / NumberingSystem / Collation 别名规范化**(`gregorian` → `gregory`、`islamic-civil` → `islamicc`)。

示例:

| 构造 | `String()` 输出 |
|------|-----------------|
| `Parse("en-US")` | `"en-US"` |
| `Parse("en-US-u-hc-h23")` | `"en-US-u-hc-h23"` |
| `Parse("en-us-U-Ca-Gregorian-Hc-H23")` | `"en-US-u-ca-gregory-hc-h23"` |
| `Parse("en-US-u-kn")` | `"en-US-u-kn"` |
| `Parse("en-US-u-kn-false")` | `"en-US"`(默认值省略) |
| `New(language.Japanese, Options{Calendar: "japanese", HourCycle: "h23"})` | `"ja-u-ca-japanese-hc-h23"` |

### 3.3 `MarshalText` / `UnmarshalText`

```go
// (Locale).MarshalText() / UnmarshalText() 实现 encoding.TextMarshaler / TextUnmarshaler。
// 等价 String() / Parse(),用于 JSON / YAML / config 文件。
func (l Locale) MarshalText() ([]byte, error)
func (l *Locale) UnmarshalText(text []byte) error
```

> **Why**: `encoding.TextMarshaler` 是 Go 标准接口;实现后 `json.Marshal(loc)` / `yaml.Marshal(loc)` 自动走 canonical String(),无需消费方写自定义 marshaler。
>
> **Rejected**: `MarshalJSON` / `UnmarshalJSON` 二者并存(冗余,`encoding/json` 自动 fallback 到 TextMarshaler)。

---

## 4. Maximize & Minimize <a id="4-maximize--minimize"></a>

### 4.1 入口签名

```go
// (Locale).Maximize 添加 likely subtags(language → script + region 推断)。
// 对应 ECMA-402 sec-addlikelysubtags + spec sec-intl.locale.prototype.maximize。
func (l Locale) Maximize() Locale

// (Locale).Minimize 移除可推断的 subtags,与 Maximize 互逆。
// 对应 sec-removelikelysubtags + sec-intl.locale.prototype.minimize。
func (l Locale) Minimize() Locale
```

调用示例:

```go
loc := locale.MustParse("zh-Hant")
fmt.Println(loc.Maximize().String())  // "zh-Hant-TW"
fmt.Println(loc.Maximize().Minimize().String())  // "zh-Hant"
```

### 4.2 实现策略

`Maximize` / `Minimize` **必须**底层走 `language.Tag.LikelyScript()` / `language.Tag.LikelyRegion()` + `internal/cldr` 的 `MaximizeSubtags` / `MinimizeSubtags` 表(见 [SPEC 50 §6](./50-cldr-data.md#6-data-access-api))。

策略:

1. 先用 `language.Tag.LikelyScript()` + `LikelyRegion()` 计算。
2. 跑 formatjs `tests/likely-subtags.test.ts` 全部 fixture。
3. 失败 case 用 `internal/cldr/likely_subtags.go` 的 CLDR 数据补丁覆盖。

> **Why**:
> 1. **`language.Tag` 95% 命中** —— `x/text` 内置 likelySubtags 表覆盖大部分常见 case;使用它免去 100% 数据嵌入。
> 2. **CLDR 表兜底** —— `x/text` 数据可能滞后于 CLDR 48;按 fixture 失败 case 补丁,精度可控。
> 3. **conformance 优先** —— SPEC 00 §2 要求与 formatjs 字节级一致,fixture 是真值表。
>
> **Rejected**:
> - **完全自带 CLDR `likelySubtags.json`** 替代 `language.Tag`:重写 5 万行表数据,开发期收益不抵成本。
> - **完全依赖 `language.Tag`** 不补丁:fixture 失败强制接受 divergence,违反 SPEC 00 §2。
> - **自实现 likelySubtags 算法**:违反 CLAUDE.md "no reinventing locale parsing"。

### 4.3 扩展字段保留

`Maximize` / `Minimize` **必须**保留所有 7 个扩展字段(`Calendar` / `HourCycle` / ...);仅修改 `Tag` 部分。

```go
loc := locale.MustParse("zh-u-hc-h23-ca-chinese")
m := loc.Maximize()
// m.HourCycle() == "h23"       // 保留
// m.Calendar() == "chinese"    // 保留
// m.Tag()       == zh-Hans-CN  // 推断
```

> **Why**: 与 formatjs `intl-locale/index.ts` `maximize()` 实现一致;扩展字段是用户显式选择,Maximize 不应静默丢弃。

---

## 5. Getter Materialization (关闭 SPEC 00 §8 Q3)

### 5.1 决策:简单字段预解析,候选列表方法惰性

| 类型 | 字段 / 方法 | 物化时机 |
|------|------------|---------|
| **简单 getter**(spec `Intl.Locale.prototype.<name>` getter) | `Calendar()` / `Collation()` / `HourCycle()` / `CaseFirst()` / `Numeric()` / `NumberingSystem()` / `FirstDayOfWeek()` | **构造时预解析**(从 BCP 47 `-u-` 扩展键直接读;零额外成本) |
| **候选列表方法**(spec `Intl.Locale.prototype.get<Name>s()` method) | `GetCalendars()` / `GetCollations()` / `GetHourCycles()` / `GetNumberingSystems()` / `GetTimeZones()` / `GetWeekInfo()` / `GetTextInfo()` | **每次调用时计算**(走 `Maximize()` + CLDR preference 表;不缓存) |

### 5.2 候选列表方法签名

```go
// locale/info.go(签名)

// GetCalendars 返回此 locale 偏好使用的 calendar 列表(按优先级排序)。
// 对应 ECMA-402 sec-intl.locale.prototype.getCalendars。
func (l Locale) GetCalendars() []string

// GetCollations 返回 generated supported collation 列表;显式 Locale.Collation() 非空时只返回该值。
// Collator 不在 active surface 内,ECMA-402 AvailableCanonicalCollations 又是 implementation-defined;
// 因此本方法只暴露 root supported sort collation 集合,不试图实现 locale-specific Collator 偏好。
func (l Locale) GetCollations() []string

// GetHourCycles 返回 hour cycle 偏好(按优先级)。
func (l Locale) GetHourCycles() []string

// GetNumberingSystems 返回 numbering system 偏好。
func (l Locale) GetNumberingSystems() []string

// GetTimeZones 返回 region subtag 对应的 canonical IANA time zone 列表;无 region 时返回 nil。
func (l Locale) GetTimeZones() []string

// WeekInfo 对应 ECMA-402 sec-week-info-of-locale。
type WeekInfo struct {
    FirstDay    time.Weekday  // 首日(默认 time.Monday)
    Weekend     []time.Weekday // 周末(常见 [Saturday, Sunday])
}
func (l Locale) GetWeekInfo() WeekInfo

// TextInfo 对应 ECMA-402 sec-text-info-of-locale。
type TextInfo struct {
    Direction string // "ltr" | "rtl"
}
func (l Locale) GetTextInfo() TextInfo
```

调用示例:

```go
loc := locale.MustParse("ar-SA")
fmt.Println(loc.GetWeekInfo().FirstDay)    // time.Sunday
fmt.Println(loc.GetTextInfo().Direction)   // "rtl"
fmt.Println(loc.GetCalendars())            // ["islamic-umalqura", "gregory", "islamic", ...]
```

> **Why**(关闭 SPEC 00 §8 Q3):
> 1. **构造期成本最小** —— 简单字段在 BCP 47 解析时顺手得到,零开销 + 字段直接读取(O(1))。
> 2. **候选列表低频** —— `getCalendars()` 等只在元信息查询场景使用(如"显示本 locale 支持的日历");大多数 formatter 调用 `loc.Calendar()`(单数 getter)而不调 `getCalendars()`。
> 3. **保持值语义** —— `Locale` struct 不含 lazy cache 字段;不需要 `*Locale` 指针、`sync.Once`、内部 mutex。
> 4. **YAGNI** —— 若 benchmark 显示某 method 是热点,再加 `sync.Once` 缓存。
> 5. **与 formatjs 完全一致** —— `intl-locale/index.ts` 的 `getCalendars()` 同样不缓存,每次调用 `calendarsOfLocale(this)` 走 `maximize() + region 查表`。
>
> **Rejected**:
> - **全部预解析(候选列表也存进 struct)**:增加 struct 大小约 100+ bytes(7 个 `[]string`/`map`/`*WeekInfo` 指针);大多数调用方不读这些字段,空间浪费。
> - **全部惰性(简单 getter 每次重新解析 BCP 47)**:hot path 性能损失,且 `loc.Calendar()` 是高频读取。
> - **`sync.Once` 缓存候选列表**:必须用 `*Locale` 才能内部 mutate;丢失值语义;且 CLDR 表查找本身 O(1),缓存收益微乎其微。

### 5.3 候选列表实现路径

候选列表方法 **必须**走 `internal/cldr` 的 region preference 数据(SPEC 50 §6):

```go
// locale/info.go(片段;非完整实现)
func (l Locale) GetCalendars() []string {
    region := l.Maximize().Tag.Region().String()  // 例:"SA"
    cldrLoc, _ := cldr.ResolveLocale(l.Tag)
    return cldrLoc.CalendarPreference()  // CLDR calendarPreferenceData
}
```

### 5.4 显式 Calendar 字段优先

如果 `Locale.Calendar()` 非空,`GetCalendars()` **必须**返回单元素列表(spec 行为)。

```go
loc := locale.MustParse("en-US-u-ca-buddhist")
fmt.Println(loc.GetCalendars())  // ["buddhist"]
```

---

## 6. Equality & Comparability

### 6.1 决策:显式 `Equal` + `String` 比较

```go
// (Locale).Equal 比较 Tag.String() 与 7 个扩展字段。
func (l Locale) Equal(other Locale) bool

// (Locale).String 返回 canonical BCP 47;同一 Locale 多次 String() 结果一致。
func (l Locale) String() string
```

### 6.2 `==` 不可比

`Locale` **不**保证 Go `==` 可比;`language.Tag` 内部含 `[]language.Variant` 等不可比字段,嵌入后 `Locale` 也不可比。

```go
a, _ := locale.Parse("en-US")
b, _ := locale.Parse("en-US")
// a == b  // 编译错误或行为未定义,禁止使用
fmt.Println(a.Equal(b))         // true
fmt.Println(a.String() == b.String())  // true(替代方案)
```

> **Why**:
> 1. **`language.Tag` 不是 comparable** —— Go 类型系统已强制。
> 2. **`Equal` 显式**比 `==` 安全清晰;调用方一眼知道是按字段语义比较。
> 3. **`String()` 比较**作为 map key 的退路:用户若需 `map[Locale]V` 必须改 `map[string]V` + `loc.String()`。
>
> **Rejected**:
> - **要求 `Locale` 是 `comparable`**:破坏对 `language.Tag` 的嵌入;丢失 Tag 的全部能力。
> - **`reflect.DeepEqual`**:慢、可能误判(slice 顺序、零值字段)。

### 6.3 排序

排序 **应当**按 `Locale.String()` 字典序;调用方用 `slices.SortFunc(locs, func(a, b Locale) int { return strings.Compare(a.String(), b.String()) })`。本 SPEC 不提供 `Less` 方法。

---

## 7. Options Object and Read-Only Locale

### 7.1 Constructor Options

`Options` 是 `Intl.Locale(tag, options)` 的 Go typed bridge。它只在 constructor boundary 生效;`Locale` 返回后是 read-only value。

```go
// locale/locale.go(签名)
type Options struct {
    Calendar        string
    Collation       string
    HourCycle       string
    CaseFirst       string
    Numeric         *bool
    NumberingSystem string
    FirstDayOfWeek  string
}
```

调用示例:

```go
numeric := true
loc, err := locale.New(language.English, locale.Options{
    Calendar:        "gregory",
    HourCycle:       "h23",
    Numeric:         &numeric,
    NumberingSystem: "latn",
})
```

### 7.2 校验时机

`New` **最多接受一个** `Options` 值。传入多个 `Options` 必须返回 wrapped `ErrInvalidOption`;不能 merge、覆盖或按顺序执行。

`Options` 校验必须在 `New` 中重用 §2.3 的校验逻辑,失败时返回 wrapped `ErrInvalidOption` 或 `ErrInvalidLocale`。除 `MustParse` 这种显式 Must wrapper 外,生产路径不新增 panic API。

> **Why**:
> 1. **统一边界** —— `Parse` / `New` 是 locale 用户输入进入系统的边界,错误应在这里集中返回。
> 2. **ECMA-402 对齐** —— JavaScript `Intl.Locale(tag, options)` 只有一个 options object;Go typed `Options` 保留同一形状。
> 3. **无隐式 panic** —— `MustParse` 已经覆盖测试和常量场景;普通 mutator/option 不应该再引入第二套 panic 语义。
>
> **Rejected**:
> - `With*` constructor options:把一个 JS options object 拆成一串 Go closures,并引入执行顺序语义。
> - `(Locale).WithCalendar(...)` immutable setters:JavaScript `Intl.Locale` 是 read-only accessor object;setter API 会让 Go surface 比原生 Intl 更宽。
> - `Options` 静默丢弃非法值:难以排查,与 spec 行为不一致。

---

## 8. Errors

```go
// locale/errors.go
package locale

import (
    "errors"
    "github.com/agentable/go-intl/internal/ecma402"
)

// ErrInvalidLocale 包装解析失败 / 字段值不合法 / language.Parse 错误。
// 等价 ECMA-402 RangeError(包装 ErrInvalidOption)。
var ErrInvalidLocale = errors.New("locale: invalid locale")

// 重导出抽象层 sentinel(便于消费方统一 errors.Is 匹配)
var ErrInvalidOption = ecma402.ErrInvalidOption
```

错误消息约定:

```go
return Locale{}, fmt.Errorf("locale: %q: %w", input, ErrInvalidLocale)
return Locale{}, fmt.Errorf("locale: invalid hourCycle %q: %w", hc, ErrInvalidLocale)
```

`errors.Is(err, locale.ErrInvalidLocale)` 应在所有解析 / 校验失败下返回 true。

---

## Forbidden

- **公开 API 接收 raw `string` locale**:破坏类型边界,让解析失败到 `Format` 调用层才暴露。
  - ✅ Do: `numberformat.New(loc locale.Locale, options ...numberformat.Options)`;调用方先 `locale.Parse("en-US")`。
  - ❌ Don't: `numberformat.New(s string, options ...numberformat.Options)`。

- **重写 BCP 47 解析**:违反 CLAUDE.md "no reinventing locale parsing"。
  - ✅ Do: 内部走 `language.Parse(s)` 之上叠加 `-u-` 扩展处理。
  - ❌ Don't: 自己写 `parseBCP47(s string)`。

- **自带 likelySubtags 完整版本**:重写 `x/text` 已有的 5 万行表。
  - ✅ Do: `language.Tag.LikelyScript()` + 必要 fixture 失败 case 补丁。
  - ❌ Don't: 在 `internal/cldr/likely_subtags.go` 内嵌完整 5 万条 mapping。

- **构造期 panic**:违反 CLAUDE.md "no panic in production code"。
  - ✅ Do: `Parse` / `New` 返回 `(Locale, error)`。
  - ❌ Don't: `Parse` 在 invalid 输入上 panic(`MustParse` 例外:仅限 const-like 场景)。

- **`Locale` 设计为 `*Locale` 指针类型**:破坏值语义。
  - ✅ Do: `Locale` 是 struct 值类型;需要改变 locale 扩展时重新调用 `locale.New(tag, Options{...})` 或 `locale.Parse(tag)`。
  - ❌ Don't: `func (l *Locale) WithCalendar(c string)`(in-place mutate)。

- **`Locale` 实现 Go `==` comparable**:`language.Tag` 内部不是 comparable,嵌入后嵌入类型也不是。
  - ✅ Do: `loc1.Equal(loc2)` 或 `loc1.String() == loc2.String()`。
  - ❌ Don't: `loc1 == loc2`(编译错或行为未定义)。

- **构造期检查 CLDR 数据存在性**(如 `Calendar="vulcan"` 是否在 CLDR 中):违反层次。
  - ✅ Do: 构造期只校验 BCP 47 语法;数据存在性由 formatter 在数据查找时报错。
  - ❌ Don't: `Parse` 内 `import "internal/cldr"` 检查 calendar 名。

- **候选列表方法返回固定 `[]string` literal**(忽略 region preference):破坏 spec 一致性。
  - ✅ Do: `GetCalendars()` 走 `Maximize().region` + `internal/cldr` preference 表。
  - ❌ Don't: `func (l Locale) GetCalendars() []string { return []string{"gregory"} }`。

- **`getCalendars` / `getWeekInfo` 等候选列表方法缓存到 struct 字段**:破坏值语义。
  - ✅ Do: 每次调用计算(O(1) CLDR 表查找)。
  - ❌ Don't: 在 `Locale` struct 中加 `cachedCalendars []string` + `sync.Once`。

- **`String()` 输出非 canonical**(不按字典序、未做别名归一):破坏 round-trip。
  - ✅ Do: `gregorian` → `gregory`、`-u-` 键按字典序、空字段省略。
  - ❌ Don't: 直接 `fmt.Sprintf("%s-u-ca-%s-hc-%s", base, cal, hc)`。

---

## Acceptance Criteria

### 结构

- [ ] `locale.Locale` 使用未导出 `language.Tag` + extension state 表达 `Intl.Locale` read-only object,与 §1.1 getter 一致。
- [ ] `Numeric()` getter 返回 `bool`;`Options.Numeric` 可用 `*bool` 表达 constructor boundary 的 omitted vs explicit value。

### 解析

- [ ] `Parse(s string) (Locale, error)` 接受 BCP 47 字符串(含 `-u-` 扩展)。
- [ ] `MustParse(s string) Locale` panic on error。
- [ ] `New(tag language.Tag, options ...Options) (Locale, error)` 最多接受一个 `Options` 值。
- [ ] 解析失败返回 `errors.Is(err, ErrInvalidLocale)` 为 true 的错误。
- [ ] 接受 spec 别名:`gregorian` → `gregory`、`islamic-civil` → `islamicc`(测试覆盖)。
- [ ] 7 个扩展字段在构造期 spec 校验(§2.3 表)。

### 字符串往返

- [ ] `(Locale).String()` 输出 canonical BCP 47:
  - `-u-` 扩展键按字典序(`ca` < `co` < `fw` < `hc` < `kf` < `kn` < `nu`)。
  - 空字段省略;`Numeric=false` 不输出;`Numeric=true` 输出 `-u-kn`。
  - 别名归一(`Calendar="GREGORIAN"` 输出 `-u-ca-gregory`)。
- [ ] `Parse(loc.String()).Equal(loc) == true`(round-trip)。
- [ ] `MarshalText` / `UnmarshalText` 实现 `encoding.TextMarshaler` / `TextUnmarshaler`。

### Maximize / Minimize

- [ ] `(Locale).Maximize() Locale` 走 `language.Tag.LikelyScript/Region` + `internal/cldr` 补丁。
- [ ] `(Locale).Minimize() Locale` 与 Maximize 互逆。
- [ ] formatjs `tests/likely-subtags.test.ts` 与 `tests/minimize.test.ts` 全部 fixture 在 `locale/canonical_test.go` 通过。
- [ ] `Maximize` / `Minimize` 保留 7 个扩展字段。

### Getter

- [ ] 简单字段(7 个扩展字段)在 `Parse` / `New` 内构造期预解析(§5.1 表)。
- [ ] 候选列表方法 `GetCalendars` / `GetCollations` / `GetHourCycles` / `GetNumberingSystems` / `GetTimeZones` / `GetWeekInfo` / `GetTextInfo` 每次调用走 CLDR preference 数据,**不**缓存到 struct 字段。
- [ ] 显式 `Calendar` 非空时,`GetCalendars()` 返回单元素列表。
- [ ] `WeekInfo` / `TextInfo` 类型签名与 §5.2 一致。

### Equality

- [ ] `(Locale).Equal(other) bool` 按字段语义比较(`Tag.String()` + 7 扩展字段)。
- [ ] `Locale` **不**实现 Go `==` comparable(嵌入的 `language.Tag` 不可比)。

### Options

- [ ] `Options` 字段与 §7.1 一致,通过 `New(tag, options)` 应用。
- [ ] 传入多个 `Options` 返回 wrapped `ErrInvalidOption`。
- [ ] 不提供 `With*` constructor options 或 `(Locale).With*` setters;locale object 保持 read-only。

### 错误

- [ ] `errors.Is(err, ErrInvalidLocale)` 在所有解析 / 校验失败下返回 true。
- [ ] `locale` 包重导出 `ecma402.ErrInvalidOption`。
- [ ] `Parse` / `New` **不** panic(测试覆盖各种异常输入)。

### 测试

- [ ] formatjs `intl-locale/tests/index.test.ts` 全部 case 移植到 `locale/locale_test.go` 并通过。
- [ ] 所有测试用 `t.Parallel()`。
- [ ] 至少 1 个 `Example*` 函数演示 `Parse` + `String()` round-trip。

---

## References

### Specification

- [ECMA-402 §14 — Intl.Locale Objects](https://tc39.es/ecma402/#locale-objects)
- [ECMA-402 §6.2.3 — CanonicalizeUnicodeLocaleId](https://tc39.es/ecma402/#sec-canonicalize-unicode-locale-id)
- [BCP 47 — Tags for Identifying Languages](https://www.rfc-editor.org/rfc/rfc5646)

### Reference implementations

- `.references/formatjs/packages/intl-locale/index.ts` —— `IntlLocaleOptions` / `RELEVANT_EXTENSION_KEYS` / `applyOptionsToTag` / `applyUnicodeExtensionToTag`
- `.references/formatjs/packages/intl-locale/preference-data.ts` —— region 映射 calendar/hourCycle/firstDayOfWeek 偏好
- `.references/formatjs/packages/intl-locale/tests/index.test.ts` —— fixture
- `.references/ext/src/ecma402/locale.h` —— PHP `ecma402_locale` struct 全字段平铺先例
- `.references/intl/intl.go` —— translate-agent/intl 直接基于 `language.Tag` 的 Go 先例

### Cross-SPEC

- [SPEC 00 §4 — Locale Model](./00-vision-and-scope.md#4-locale-model)
- [SPEC 11 §ResolveLocale](./11-locale-matching.md) —— 消费 `Locale.String()` 与扩展字段
- [SPEC 12 §Internal Slots](./12-abstract-operations.md#5-internal-slots) —— `[[Locale]]` slot 在抽象层的表达
- [SPEC 50 §6 — Data Access API](./50-cldr-data.md#6-data-access-api) —— `MaximizeSubtags` / `CalendarPreference` / `WeekData` 数据
- [SPEC 60](./60-facade.md) —— root `GetCanonicalLocales` consumes canonical `Locale` values without adding locale availability matching

### Research

- `.research/R02-locale-and-matcher.md` —— Locale 类型、扩展字段、getter 物化、`language.Matcher` 拒绝

---

> 本 SPEC 是 `locale.Locale` 的 SSOT。新增 ECMA-402 扩展键(spec 罕见地添加新 `-u-` 键)触发本 SPEC 修订;`x/text/language` 行为变化通过 fixture 失败暴露。
