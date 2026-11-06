# PR #12、Issue #10 / #11 评估与改进方案

评估日期：2026-11-06。本文记录审查依据、实施任务和验证结果。用户已授权实施、提交、推送及简短回复相关 PR / issue；只有验收后的行为才能标记完成。

## 决策

| 对象 | 建议 | 理由与交付边界 |
| --- | --- | --- |
| [PR #12](https://github.com/agentable/go-intl/pull/12) | **接受贡献方向，在原 PR 基础上修正后合并** | 内部构造器 matcher 的可用 locale 扩展不适用于应用资源集合；公开输入缺少校验；提案对齐声明已过时；与当前 workspace 和发布约定未集成。 |
| [Issue #11](https://github.com/agentable/go-intl/issues/11) | **接受需求，修改方案 A** | 为应用提供精确资源集合的匹配能力有真实用途。在主模块下的显式实验包 `x/localematcher` 中提供收敛后的接口，共享内部算法和数据；不直接导出内部类型，不将其称为已标准化 ECMA-402 API。 |
| [Issue #10](https://github.com/agentable/go-intl/issues/10) | **接受需求，先做可行性验证，再决定正式交付** | Intl.Segmenter 已有标准所有者，方向符合项目目标；但当前仓库没有完整分词后端，不能把 UAX #29 或几个样例通过等同于 Node/ICU 兼容。完整构造器通过验收后才进入主模块。 |

标准符合性、实现定义行为、Go 类型桥接分别说明：ECMA-402 约束可观察契约；best-fit 距离策略、分词边界和 `isWordLike` 的具体决策存在实现自由；UTF-8 字节索引属于 Go 桥接。用户确认：额外 API 可以依据 TC39 官方 Stage 1 草案交付；草案阶段本身不是拒绝理由。LocaleMatcher 在 x 实验包明确固定官方草案 revision，与现有正式 ECMA-402 surface 分开记录。不得改变既有格式器的公共接口、默认值、locale 协商、resolved options 或输出；共享内部执行核心的改动必须通过原有回归验证。

本次先更新本文，再沿用贡献者的 PR 实施。代码任务统一使用 `tdd-implementing`，各路线最终完成后做一次 `code-review`。#11 与 #10 可独立推进，#10 不依赖公开 matcher。保持既有格式器行为、共享 CLDR 数据所有权和内部边界；不为这两个请求建立通用插件、后端选择或未来提案准入框架。

审查基线：PR head `04c9ce4bcc4d2aaf8683c6881ae9cb46aebaa2b2`；本地 main `5979992440911c754bd069b3da2aa9b1c85afca6`。GitHub 查询时 base 为 `2ce67fc723fc3bc0cc15a28279b8c47ff41543f8`，合并状态 `CONFLICTING`，未返回检查结果；本文不把本地 main、PR 原 head 和线上 base 当作同一份代码。

## 原 PR head 的审查发现（历史基线）

以下 R1–R4 描述原提交 `04c9ce4`，不是修正后的当前状态；本次已分别修正其资源集合、输入校验、官方草案 API 和模块交付问题。最终验收见文末实施记录。

### R1 · P1：匹配结果可以落在调用者并未提供的资源上

PR `x/localematcher/localematcher.go:40–48` 原样调用构造器专用 matcher。该 matcher 将原始数据 locale 扩展成父级及 language-region 别名；这是为 Intl 服务构造器准备能力集合的规则，不是应用目录的资源声明。

在原 PR 的独立探针中：

```text
requested=[de-AT], available=[de-DE], algorithm=lookup, default=fr
=> Locale=de, DataLocale=de-DE
```

RFC 4647 lookup 对请求逐级截断，并在传入集合中查找。这里只有 `de-DE`，`de-AT → de` 均未命中，应当无匹配，按 PR 的旧接口则返回显式 default `fr`。返回 `de` 不仅使目录读取落空，使用 `DataLocale` 读取 `de-DE` 也改变了 lookup 的语义。best-fit 可以选 `de-DE`，但仍必须返回实际可用的条目。

**修正：**把服务构造器的可用集扩展和应用精确集合的编译入口分开，共享距离、最大化和匹配执行逻辑。不能仅将结果改写成 `DataLocale`，也不能通过关闭既有构造器扩展来修复。

**证据：**Current：[PR 源码](https://github.com/agentable/go-intl/blob/04c9ce4bcc4d2aaf8683c6881ae9cb46aebaa2b2/x/localematcher/localematcher.go#L40)，`internal/localematcher/available.go:availableLocalesFor`、`compiled.go:NewMatcher`。规范：`.references/ecma402/spec/negotiation.html#sec-internal-slots` 与 `#sec-lookupmatchinglocalebyprefix` / `#sec-lookupmatchinglocalebybestfit` 分别规定服务能力集和集合选择。Existing proof：PR `localematcher_test.go` 的 `lookup truncates without maximizing` 反而固定了返回合成 `zh` 的错误资源契约，必须替换。

### R2 · P1：内部前置条件丢失，公开入口静默接受错误输入

同一公开入口接受原始 `[]string`，却没有原构造器负责的解析、规范化和算法校验；也没有返回错误的通道。探针结果：

```text
lookup [en-us] against [en-US]       => en，而非规范化后的 en-US
lookup [en-US, en_US] against [en-US] => 成功，未处理后续非法 tag
lookup [en_US] against [en_US]       => en_US，返回非法 locale
Algorithm(99), request zh-HK        => 静默按 best-fit 匹配 zh-Hant
```

**修正：**使用现有 `locale.Parse` / `ParseList` 的类型边界，在公开层以 `locale.List` 接收已验证身份，按共享规则去重；构造时验证算法和 available 中不允许的 Unicode `u` 扩展。内部无错误 matcher 不等于公开原始字符串入口可以吞掉错误。

**证据：**Current：上述 PR 入口与 `internal/localematcher/compiled.go:Matcher.Match` 的非 lookup 分支；`locale/list.go:ParseList`、`internal/ecma402/locale_list.go` 已拥有输入处理规则。规范：`negotiation.html#sec-canonicalizelocalelist`；仓库 `SPECS/11-locale-matching.md` §5 / §7 将无错误内部匹配建立在先解析的前提上。

### R3 · P2：对提案的公开承诺已经过时

PR 声称映射静态 `Intl.LocaleMatcher.match`，返回含 `DataLocale`、`Extension`、`Distance` 的内部记录，并接受隐式匹配后的 default。当前提案已改为构造器固定 available/options、实例 `match(requested)` 返回可用 tag 或 `undefined`；不自动应用 default，也不把请求的 Unicode 扩展附加到结果。

**修正：**以本次核对的草案版本设计实例接口；未命中交给应用处理。隐藏内部结果、阈值和构造器数据 locale。不要给内部 `Matcher` 做公共类型别名；私有持有它或共享执行核心即可。当前别名尚未泄漏额外导出方法，问题是错误契约及后续耦合，不能虚构额外方法泄漏。

**证据：**Current：[PR doc.go](https://github.com/agentable/go-intl/blob/04c9ce4bcc4d2aaf8683c6881ae9cb46aebaa2b2/x/localematcher/doc.go#L5)。Design precedent：TC39 草案 [`spec.emu@c35e7b7`](https://github.com/tc39/proposal-intl-localematcher/blob/c35e7b759f698b16e9d73d8e98647207e35a14ce/spec.emu) 与同版本 README；本地 `.references/formatjs/packages/intl-localematcher/index.ts` 仍是旧版静态 ponyfill。采用最新草案的实例/未命中边界，保留本项目的 CLDR best-fit 策略；不把 ponyfill 的旧 API 或草案本身说成已完成标准。

### R4 · P2：必须补齐当前仓库的模块集成和实际消费验证

PR 原 head 没有 go.work，其独立 x 模块检查可通过；当前 main 已提交 go.work，未包含 x。把原 PR 的 x 目录叠加到本地 main 副本后：

```text
go -C x test ./...
=> directory prefix . does not contain modules listed in go.work ...
GOWORK=off go -C x mod tidy -diff
=> 失败：需要更新 Go 指令及依赖/校验和
```

PR SPEC 80 所称“go.work 是被忽略的个人选项”也不再符合当前事实。`x/go.mod:9` 的本地 `replace => ..` 违反当前已发布 manifest 不含本地替换的约定。注意：**依赖模块的 replace 会被下游忽略**，不能据此断言这个 library 无法被下载；真正缺少的是其声明版本与候选发布图的消费证据。

**修正：**先同步当前主线并解决冲突，再采用仓库已有 workspace 与协调发布方式；源码开发通过 go.work 连接，发布 manifest 不含本地路径 replace；在 workspace 关闭、无 replace 的外部消费者中验证真正候选模块图。

**证据：**Current：[PR Taskfile](https://github.com/agentable/go-intl/blob/04c9ce4bcc4d2aaf8683c6881ae9cb46aebaa2b2/Taskfile.yml#L151)、[PR x/go.mod](https://github.com/agentable/go-intl/blob/04c9ce4bcc4d2aaf8683c6881ae9cb46aebaa2b2/x/go.mod#L1)、当前 `go.work`、`CLAUDE.md` Release Inventory、`Taskfile.yml:modules:verify`。Existing proof：上述隔离副本检查，只证明与该本地主线版本的集成缺口，不声称已测试线上 merge commit。

### 审查覆盖与已执行检查

| 质量轴 | 结论 |
| --- | --- |
| 正确性 / 需求匹配 | R1、R2 阻断；目录按 key 选择真实资源的目标尚未成立。 |
| 可读性 | 49 行 wrapper 本身直观；R3 的文档和实际所有者不一致会误导调用者。 |
| 架构 | 共享 CLDR 数据和距离算法正确；服务数据 fallback 泄漏到应用资源选择错误，R4 的发布图还未证明。 |
| 安全 | 发现的是输入契约缺失，未发现单独的注入、权限或秘密泄漏路径；不扩大为安全漏洞。 |
| 性能 | 未建立新的性能回归；已有 compiled 路径和有界请求缓存应保留。未来按 available 子集复用实例，避免每次消息查找重新编译。 |

改动为 13 个文件，640 行新增、14 行删除，以文档和 198 行测试为主，单个功能可审查；这不是降低复杂度的重构。删除公开桥会把 path-squat/复制数据的负担推回调用者，因此桥有价值，但只需解决 #11，不需同时准入所有未来 TC39 提案。

已在原 PR 的隔离源码副本运行 x 模块的 `go test -race ./...`、`go vet ./...`、`GOWORK=off go mod tidy -diff`，均通过；另运行上面的独立行为探针及主线集成检查。使用本机 Go 1.27.3 二进制并设置 `GOTOOLCHAIN=local`：默认自动选择的 1.27.0 缓存缺失，初次失败属于工具链问题。初次审查阶段没有运行整个仓库的 `task verify`，也没有把 PR 的绿色单元测试视为规范符合性证明。

## 建议执行项

### M01 · #11：以精确资源集合实现实验性实例 matcher

优先级 P1，证据可信度高。先完成 R1–R3；路由 `api-surface-designing`，代码通过 `tdd-implementing` 实现。

**本次实现的调用形态：**

```go
available, err := locale.ParseList("de-DE")
if err != nil { return err }
matcher, err := localematcher.New(available, localematcher.Options{
    Algorithm: gointl.String("lookup"),
})
if err != nil { return err }
requested, err := locale.ParseList("de-AT")
if err != nil { return err }
matched, ok := matcher.Match(requested)
// ok == false；由应用决定缺译、默认文案或 fallback，不凭空选择 de-DE。
_ = matched
```

目标类型边界为 `New(locale.List, Options) (*Matcher, error)`、`Match(locale.List) (locale.Locale, bool)`。`Options.Algorithm *string` 的 nil 默认 best-fit；非法字符串为可识别的选项错误。available 中的 `u` 扩展属于非法 available 输入；复用既有错误分类并附上下文。原始字符串错误由 ParseList 报告，类型化 Match 不需要虚构解析错误。候选集合与选项在 New 中复制、冻结；请求列表按共享规则规范去重。`ok=false` 表示无匹配，与合法 `und` 区分。

内部精确集入口不添加父级或派生别名；输出是该规范化集合中的条目。现有服务构造器继续使用扩展集。CLDR `written-new`、likely subtags、阈值及有界缓存保持一个事实来源；不公开可调阈值、数据访问器或 `Result`。每种实际资源子集构造并复用实例，应用拥有 key→资源的映射与 fallback 顺序。

**验收：**lookup `de-AT / [de-DE]` 无匹配；`de-AT / [de]` 命中 de；zh-Hant/zh-Hans 用例分别验证 lookup 与 best-fit。任意成功结果属于传入的规范化集合，空请求/空 available 无匹配；大小写、别名、重复 tag、请求 `u` 扩展、available `u` 扩展、非法算法、合法未知 subtags 均有明确结果。改变原 available/options 不影响实例；独立请求并发正确。用真实按 key 读取文案的测试证明选到的资源存在，不能只断言 locale 字段。聚焦运行 `go test -race ./internal/localematcher ./internal/ecma402 ./locale` 和 `x/localematcher` 包测试，并保护原构造器派生 locale 的回归测试。

**证据：**采用 R1–R3 的源证据与草案；Existing proof：`internal/localematcher/available_test.go`、`compiled_test.go` 保护服务集合语义，`locale/language_identity_test.go` 保护身份规范化。新增精确集用例证明新用途，旧构造器测试不代替新用途验收。

### M02 · #11：作为主模块下的实验包交付

保留 `github.com/agentable/go-intl/x/localematcher` 导入路径，取消尚未发布的 `x/go.mod` 和独立模块 tag。用户允许官方草案作为额外 API 的依据，因此只为这一实验包记录主模块的明确例外；root Intl namespace 不增加 LocaleMatcher 别名，现有构造器及其能力集保持不变。

**变更理由：**独立 x 模块需要固定包含新内部入口的 root 版本，而协调发布尚未发生；同一模块直接共享内部代码和数据，消除内部 API 的跨模块版本轴，也让现有 root race/lint/test 命令自然覆盖新包。保留官方草案 revision、实验说明和同 root 版本交付的事实，不建立泛化准入框架。

**验收：**root `./...` 覆盖 x；workspace 和 `GOWORK=off` 均通过；无本地 replace、无新运行依赖、无复制 CLDR 表。通过临时 GOPROXY 的真实 root module zip/mod，验证无 replace 的外部 `example.com/...` 消费者可导入并使用新包。README、SPEC 00 / 11 / 72 / 80、开发指南同步这个单包例外。完成 M01、M02 后可以接受贡献，不再以 Stage 1 为拒绝理由。

**证据：**R4 的发布集成风险与用户最新授权；Go import path 相同，不要求 `x/` 必须是独立模块。本选择是对原提案包装方式的简化，尚未发布的模块没有既有版本迁移义务。

### M03 · #10：先证明纯 Go 分词后端与固定兼容目标可行

优先级 P1，可独立于 M01 推进。当前没有公开 `segmenter` 包；`internal/ecma402/segmenter.go` 只有 UTF-16 长度辅助函数，不能视作分词引擎。Issue 所说的“现有 API”不是本仓库已承诺的 API，不为假定的历史接口添加兼容层。

先定义并验证两套目标：一是标准的构造器/记录/索引/错误契约；二是固定 Node 构建与 ICU 数据上的边界兼容。ECMA-402 没有要求全球所有实现使用同一本词典、相同词边界，不能把所有 Node 差异都称为规范违规；但若承诺替代 pi 的 word navigation，就必须通过约定的 native 兼容验收。

**验证工作：**在私有原型中评估可维护的纯 Go 引擎或可行移植，覆盖 Thai、Khmer、Lao、Myanmar 和中日文；审计词典、模型、规范化、规则的版本、来源、许可证与再分发方式。不能只选一个 UAX #29 依赖就宣布完成。ICU 有词典及 LSTM 引擎选择，CJK 还有词典/规范化路径；“最长词匹配”并非已证明等价的替代。FormatJS 主要提供 API/规则参考，不能替代 native 词典行为的证明。

**验收/停止条件：**用 issue 的全部输入加歧义词、未知词、混合脚本、无空格长文本，记录边界、word-like 状态、构建大小、分配与长输入耗时；不仅测各自语言 locale，还测 en 下的 Thai/CJK 混合文本。运行时不得调用 Node、加载外部数据文件或依赖 cgo/ICU4C。明确所需数据是否具备可重复生成和合法再分发条件。若成本、数据或兼容性无法成立，保留 #10 为有证据的未完成需求，提交原型结果；不先公开一个残缺的构造器。

后端选择尚未完成，可信度为 MED；这是调查任务，不是已批准某个供应商/算法的实现计划。路由 `dependency-selecting` / `dependency-boundary-designing`，原型中的代码同样用 `tdd-implementing`。进入 M04 前必须产出选型结论、固定版本清单和可执行差异结果。

**证据：**Current：上述内部辅助文件与当前 `SPECS/00` §2.1 的完整构造器准入条件。规范：`.references/ecma402/spec/segmenter.html#sec-findboundary`、`#sec-createsegmentdataobject` 明确实现定义与 tailoring 建议。Design precedent：`.references/formatjs/packages/intl-segmenter/segmenter.ts` 的构造、规则编译及 Segments；`.references/node/deps/v8/src/objects/js-segmenter.cc` 的 ICU BreakIterator、`js-segments.cc:CurrentSegmentIsWordLike`；`.references/node/deps/icu-small/source/common/brkeng.cpp:ICULanguageBreakFactory::loadEngineFor` 与 `dictbe.cpp` 的专用引擎。采用 native 可观察行为与数据需求，不机械复制整个 ICU 架构。

### M04 · #10：后端成立后完整实现 Segmenter 与索引桥接

依赖 M03 通过；路由 `api-surface-designing` → `tdd-implementing`。先在新增的 Segmenter SPEC 中冻结类型契约；以下是供评审的目标形态，尚不代表已承诺的兼容 API：

```go
requested, err := locale.ParseList("th")
if err != nil { return err }
s, err := segmenter.New(requested, segmenter.Options{
    Granularity: gointl.String("word"),
})
if err != nil { return err }
segments, err := s.Segment("😀x")
if err != nil { return err }
for part := range segments.All() { // iter.Seq[Segment]
    fmt.Println(part.Segment, part.Index, part.ByteIndex)
}
part, ok := segments.Containing(1)     // UTF-16，仍在 😀 内
part, ok = segments.ContainingByte(3)  // UTF-8，仍在 😀 内
_, _ = part, ok
```

类型目标：`New(locale.List, Options) (*Segmenter, error)`；`Options` 仅有 `LocaleMatcher *string`、`Granularity *string`；默认 grapheme，另外支持 word、sentence。`Segment(string) (*Segments, error)` 接收有效 UTF-8，非法编码返回带上下文的既有 invalid-value 分类，不能默默替换字节。Go string 桥接限定 Unicode 标量文本，不宣称覆盖 JS 孤立 surrogate 的完整字符串域；如需要无损 UTF-16 输入，在 API 设计阶段另行论证必要的窄桥，不能用错误 UTF-8 冒充 surrogate。此决定必须在 SPEC 和使用文档中明确。

记录包含原生 `segment`、`index`、`input`，`isWordLike` 用 `*bool`：word 下 false 必须存在，其他粒度省略。`ByteIndex` 为 Go-only 字段，JSON 使用 `json:"-"`。原文与索引保持一致，不规范化后把结果切回错误的原始位置；CJK 内部规范化必须保留原文映射。

`Containing(float64) (Segment, bool)` 对应 JS Number 的 ToIntegerOrInfinity：NaN/±0 到 0、有限小数向零截断、负数/无穷/到达末尾均按转换后的边界返回未命中；负小数如 -0.5 转为 -0 后可命中起始段。代理对内部的任意 UTF-16 单位返回同段。`ContainingByte(int)` 接收原串字节位置，落在多字节字符内部也返回所在段，负值及末尾未命中。区间统一为半开；空文本没有片段。`All` 每次迭代独立，可提前终止；实例不共享可变迭代游标。

构造器通过现有共享 locale/options 管线解析，遵循先 localeMatcher 再 granularity 的顺序，相关 Unicode extension keys 为空。SupportedLocalesOf 使用同一个 formatter-owned matcher；可用 locale 从真实分词规则/数据能力生成，不能照搬 NumberFormat 或整个 CLDR locale 清单，也不能只按 locale 决定能否处理输入中的其他脚本。冻结构造器状态，文本迭代时不重新协商 locale、编译规则或构造引擎。

**验收：**grapheme/word/sentence 均完整实现；通过对应固定 Unicode 数据版本的 GraphemeBreakTest、WordBreakTest、SentenceBreakTest 默认算法测试，并区分已经声明的 tailoring；全部查询边界、独立迭代、空输入、word-like 字段存在性和 JSON 投影可观察测试通过。数据生成器对缺失/错误版本/损坏输入失败，产物确定且在运行时内嵌；字典覆盖不靠逐字拆分掩盖缺失。聚焦运行 `go test -race ./segmenter/...` 及修改过的生成器测试/vet。同步 root namespace 构造器别名、README、SPEC 00 / 50 / 60 / 72 / 73；未达到完整准入条件时留在私有实现，不发布功能占位或公开后端选择器。

**证据：**M03 的规范和两份实现参考拥有分词契约；Current：`internal/ecma402/constructor_locale.go`、`locale_list.go` 提供构造与 SupportedLocalesOf 共用边界；`SPECS/73-json-records.md` 拥有缺省属性与 Go-only JSON 字段规则。新数据接入现有结构化生成验证，不能新增手维护文件 diff 清单。

### M05 · #10：扩展 conformance 记录并以固定 native witness 关闭需求

与 M03 一起确定 witness 版本和数据记录，M04 完成后关闭。采用现有 fixture/extractor/coverage/divergence 管线，新增 Segmenter 的类型化观察字段；不把 segment 伪装成 formatter Part，也不再建一套独立 conformance 基础设施。

**目标 fixture 形态（示意）：**

```json
{
  "id": "segmenter-word-emoji",
  "source": "node:<pinned-version>:segmenter",
  "locale": "en",
  "options": {"granularity": "word"},
  "input": "😀x",
  "expectedSegments": [
    {"segment": "😀", "index": 0, "input": "😀x", "isWordLike": false},
    {"segment": "x", "index": 2, "input": "😀x", "isWordLike": true}
  ]
}
```

`expectedSegments: []` 是对空结果的断言，缺省才表示未观察。另以明确的 containing 查询记录保存 Node 直接调用的索引/命中/记录结果；失败用显式 found=false 表示，不能把“未观察”与“未命中”混为一谈。UTF-8 byte offset 与 ContainingByte 属于 Go 桥接测试，用原串前缀的 UTF-8 编码长度和对应 native 段进行独立交叉验证，不把字节字段加到 native JSON 记录中。实施时一并更新严格 schema、加载校验、覆盖统计和 skip/XFAIL/divergence 规则。

现有 active witness 为 **Node v26.0.0**。默认沿用它并记录完整 Node/V8/ICU/Unicode/CLDR 版本及构建信息；若 M03 发现必须升级，需明确修改基线并重验受影响的旧 witness，不能静默用最新 Node 重写期待值。可重复生成并保存含哈希的数据清单；CI 消费固定 fixtures，刷新 witness 的流程核对实际版本。

**验收：**issue 的全部十个输入、locale resolution、SupportedLocalesOf、所有粒度、构造器错误、UTF-16 查询（含代理对内部）、byte 查询（含 UTF-8 内部）、空文本和边界索引都有 native/Go 观察证明；加入能区分词典/模型与逐字分割的歧义及跨脚本用例。`go test ./tools/check-conformance ./tools/conformance`、fixtures 生成器模块测试/vet，以及 `task conformance:verify` 通过。不能以新增 skip、只检查段数或十个样例全部通过宣称完整兼容。

**证据：**Current：`tools/conformance/fixtures.go:Fixture` 目前没有 segment 记录；`tools/conformance/node_witness.go:ActiveNodeWitnessVersion` 定义基线；`tools/gen-fixtures-from-formatjs/main.go:publishNodeWitness` 验证版本、`node_script.go` 收集引擎版本。`SPECS/70-conformance.md` 拥有 fixture/证据流程。Design precedent：`.references/formatjs/packages/intl-segmenter/tests/{word,grapheme,sentence,isWordLike}.test.ts` 提供可移植观察，native 确认实现定义结果。

## 本次 Segmenter 探查的边界

本机 Node **v26.13.0 / ICU 78.3 / Unicode 17.0 / CLDR 48.0**，默认 locale en-US；它不等于仓库固定的 v26.0.0，以下只是可行性探查，不是新 golden baseline。对 en、th、zh、ja、km、lo、my 显式 locale 运行 issue 全部输入，观察到开头三个例子分别是 `ไทย | ภาษา`、`ภาษา | ไทยx`、`中文 | 测试`，与请求一致。

`😀x` 中 x 的 UTF-16/UTF-8 起点是 2/4；`👩‍💻x` 是 5/11，emoji 段均为 non-word-like。这直接证明 rune、UTF-16 unit、UTF-8 byte 不能共用一个索引单位。后续已用固定 v26.0.0 对两个 Go 候选完成差异探针，见下方实施记录；不声称 #10 已实现。

本地参考版本：ECMA-402 `730ca2b88c305862db7be2f83c510dcca33ab48c`；FormatJS `aa3f90b53b157481515a0f7a7f0d76fcbd7c3274`；Node `5a566ac6b971a9bcbcff1d338f391e0ae6dbd678`。正文的参考文件均以这些版本为背景，外部 LocaleMatcher 草案独立固定在 R3 的 revision。

## 最终门槛与不采纳方案

M01 → M02 完成 #11；M03 → M04 → M05 完成 #10。各路线只在自身行为、模块图、生成产物和文档一起验收后关闭对应 issue，不让 Segmenter 的调查阻塞 matcher 的正确修复。

- 每条交付路线完成后，执行当前仓库适用的 `task modules:verify`、`task lint`、`task test`、`task conformance:verify`；涉及生成数据再跑 `task data:check` 与 `task data:contract`，最终 `task verify`。x 作为 root 下的实验包由这些检查覆盖，另需无 workspace / 无 replace 的实际模块消费证明。
- 最终 `code-review` 必须重新检查标准/草案所有者、真实资源命中、SupportedLocalesOf 与可执行数据能力、Go-only 桥接边界和发布依赖图；保留未解决项，不用“实验性”免除正确性。
- 不采纳 #11 方案 B 的 CLDR 独立项目：当前需求是 locale 选择，未证明拆分全体数据值得新增版本轴。一个二进制导入相同版本的同一数据包即可共享，不需复制表或新增 CLDRVersion 配对协议。
- 不把内部类型别名、公开阈值、defaultLocale 参数及 DataLocale 暴露当作“薄就正确”；不以末端重写返回值修补错误候选集。
- 接受已核实的 TC39 官方草案作为 x 实验包额外 API 的契约依据，固定 revision 并标明阶段；不自行发明语义，不把它误称为已完成标准，不把标准建议的 tailoring 夸大为规定某组固定词典结果。
- 不以普通 UAX #29 库、手写 Thai/CJK 单词表、逐字 fallback、静默忽略非支持脚本、运行时 Node/cgo 依赖满足 #10。纯 Go 后端若不成立，停在有证据的调查结论，不承诺交付日期或虚假 complete 状态。

## 实施记录

- **M01 / M02 已实现并通过最终验证：**按固定官方草案提供 `New` / `Matcher.Match`，精确资源集合编译入口共享既有 matcher 核心；移除内部类型和阈值 re-export。实验包位于 root module，删除原 PR 的 `x/go.mod` / `x/go.sum`，同步 README、开发指南和 SPEC 00 / 11 / 72 / 80。
- **TDD 证据：**先观察新公开调用缺失的编译 RED；原构造器 compiler 的 `de-AT → de` 行为 RED；只过滤输出仍导致 `[de-AT,en]` 找不到 `en` 的 RED；精确集合实现后 GREEN。算法非法值和 available `u` 扩展校验先 RED 后 GREEN。额外身份/并发/翻译文本用例作为回归证据，不冒充全部经过 RED。
- **M03 已完成本次有界探查，后端门槛未通过：**[可复现报告](reports/segmenter-feasibility/README.md) 与 [探针](reports/segmenter-feasibility/probe.sh) 固定 uax29 v2.7.0 / uniseg v0.4.7，用仓库 witness Node v26.0.0 / ICU78.3 在七个 locale 下比较 issue 全部输入及六个补充输入；两个候选各自均有 13/16 边界/索引差异。候选的 UAX #29 合规不等于 issue 所需的 ICU 分词兼容。
- **M04 / M05 未进入：**没有经过验收的词典/模型后端，不公开残缺 Segmenter，不新增 root 依赖，不关闭 #10。报告同时记录版本、许可证、索引样例、长输入成本和下一步后端验收条件；该结果不证明纯 Go 的所有方案都不可行。

- **外部模块产物验证通过：**把候选 root 源码打包为临时 GOPROXY 的 `v0.0.0-candidate.zip/.mod/.info`，使用独立 `example.com/consumer`、空模块缓存、`GOWORK=off`、无 replace 实际下载编译，输出 `zh-Hant`。本机 `GOPRIVATE=github.com/agentable/*` 最初绕过代理导致不存在的候选 revision 错误；仅为该探针设置 `GOPRIVATE=none / GONOPROXY=none` 后成功，不改全局配置或发布版本。

- **最终验证通过：**`task verify` 全部成功（fmt、vet、0 lint issues、完整 race 测试、四模块 test/vet/tidy、conformance、数据契约、漏洞扫描无发现）；另有 `GOWORK=off go test -race ./x/localematcher ./internal/localematcher`、外部 module zip 消费者以及固定 Node 分词探针成功执行。首次 lint 提示测试切片预分配，修正后完整 verify 再跑通过。
- **最终独立审查：**`code-review` 五轴复核未发现运行时阻断项；原有构造器路径保持原样，只抽取共享编译函数。审查指出的本文“待实现”状态及 SPEC 11 公共包拒绝描述已同步修正。验证范围未包含新 release/tag 或完整 Segmenter 后端；#10 保持开放。
- **提交拆分：**先在原贡献者提交上合入当前 main 并提交 LocaleMatcher 修正，保留其原始提交历史；评估方案和可复现 Segmenter 报告作为后续独立文档提交，避免将后端调查混成已交付功能。推送目标为原 PR 分支，不发布 release/tag。
