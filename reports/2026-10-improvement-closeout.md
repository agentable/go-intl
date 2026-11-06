# 2026-10 改进收口

本次逐项复核基线 `32411204e14887b2e4d9b90111f2d5a2845f6b0d` 的
`improve.md` 及 `improves/*.md`：18 项完成，1 项有理由移除，无剩余实施任务。
完成的计划文件删除；下表保留验收入口与处置依据，当前契约以链接的长期文档为准。
不把已有设计、推测性的重构或未验证的平台能力留作待办。

## 逐项结果

| 任务 | 处置与验收 | 当前代码 / 回归证据 | 长期记录 |
|---|---|---|---|
| L01 | 完成。合法未知 subdivision 不 panic；BU/UK aliases 与已知地区结果一致，未知地区按 CLDR 默认回退。 | `locale/info.go`；`locale/language_identity_test.go:TestSubdivisionGettersAcceptUnknownRegion` | [Locale](../SPECS/10-locale.md) |
| L02a | 完成。有限 corpus 证明 x/text 不能完整表达合法 identity；决定完整、可失败的 `Tag() (language.Tag, error)`，无双签名兼容层。 | `locale/locale.go:Tag`；`locale/language_identity_test.go:TestCompleteTagBridge` | [依赖裁决](golang.org-x-text.md)、[愿景](../SPECS/00-vision-and-scope.md) |
| L02b | 完成。私有 canonical identity 按语法接受未知 subtag；语言、地区、script、variant、t/u/subdivision aliases 从固定 CLDR 生成；大小写、幂等及后缀受测试保护。 | `internal/localeid/{tag,unicode,maximize}.go`；`internal/localeid/localeid_test.go`；`tools/gen-cldr/cldr/identity_alias_test.go` | [数据](../SPECS/50-cldr-data.md) |
| L02c | 完成。Parse/New、String、Equal、marshal、maximize/minimize 与桥接统一；未知身份接受与 formatter 数据可用性分开验证。 | `locale/{locale,options,canonical}.go`；`locale/language_identity_test.go`；`intl_test.go:TestUnknownLocaleIdentityAndFormatterData` | [Locale](../SPECS/10-locale.md)、[协商](../SPECS/11-locale-matching.md)、[操作](../SPECS/12-abstract-operations.md) |
| L02d | 完成。DisplayNames 保留窄语法，使用同一 canonical owner；移除后端失败后选择另一身份的分支。 | `internal/ecma402/displaynames.go`；`internal/ecma402/displaynames_test.go:TestDisplayNamesSharesLocaleIdentity` | [DisplayNames](../SPECS/44-displaynames.md) |
| L03 | 完成。`Variants *string` 只替换变体，复制指针值；omitted、大小写、未知 variant、非法/重复输入及 t/u/x 后缀均有回归。 | `locale/options.go`；`locale/construct_test.go:TestVariantsOption` | [README](../README.md)、[Locale](../SPECS/10-locale.md)、[操作归属](../SPECS/72-operation-ledger.md)、[JSON](../SPECS/73-json-records.md) |
| L04 | 完成。重查无消费者后删除 `RequireLookupMatch` 字段和分支；保留真实 localeMatcher 派发。 | `internal/ecma402/locale_list.go`；`internal/ecma402/locale_list_test.go`；`internal/localematcher/filter_test.go` | [协商](../SPECS/11-locale-matching.md) |
| N20 | 完成。只验证参与舍入的位数族；auto + significant 忽略未启用 fraction，more/lessPrecision 仍拒绝非法活跃族。 | `internal/ecma402/numberformat/options.go`；`numberformat/options_test.go` 与 `pluralrules/options_test.go:TestInactiveFractionOptions` | [数字](../SPECS/20-numberformat.md)、[复数](../SPECS/40-pluralrules.md) |
| N21 | 完成。compact 类别和指数只取自舍入后的未缩放字符串；999499.6/999500.4、负数、阈值跨越与双向 range 一致。此为明确实现政策，不宣称唯一规范算法。 | `pluralrules/pluralrules.go:resolveDecimal`；`pluralrules/range_test.go:TestCompactRoundedSourceDeterminesCategory` | [复数](../SPECS/40-pluralrules.md)、[compact 证据](../pluralrules/testdata/compact-review.md) |
| DTF-01 | 完成。独立双单引号解码为一个单引号；endpoint、connector、interval 的文本、parts 与 source 一致。 | `datetimeformat/pattern_literals.go`；`datetimeformat/internal_edges_test.go:TestEscapedQuotesInPrograms` | [日期程序](../SPECS/31-datetimeformat-skeleton.md) |
| DTF-02 | 移除。本项原为 Windows 实机调查，没有可用 Windows 运行环境，源码推论不足以验收自动发现实现。删除默认时区通用支持的暗示，明确 UTC fallback 与显式 IANA TimeZone 用法；不声称 Windows 已修复或实测通过。 | `internal/tz/default.go`；Go `src/time/zoneinfo_windows.go` 的 Local 命名 | [README](../README.md)、[时区资格边界](../SPECS/32-datetimeformat-tz.md) |
| B01 | 完成。未知语言本体在 fallback=code 时仍组合已知 script/region/variant；fallback=none 返回缺失。覆盖 zz-US、qaa-Latn-US、zz-rozaj 及两种 languageDisplay。 | `internal/cldr/displaynames/accessors.go:resolveLanguage`；`displaynames/language_composition_test.go:TestUnknownLanguageComponentFallback` | [DisplayNames](../SPECS/44-displaynames.md) |
| C01 | 完成。required topic、skip-list 和 divergence witnesses 都排除 runner 将跳过的候选；替代可执行候选有效，自证/互证及被跳过 witness 失败。 | `tools/conformance/{runner,node_witness,coverage,divergences}.go`；对应测试 | [一致性验证](../SPECS/70-conformance.md) |
| C02 | 完成。Node script 随发布工具模块分发；空目录、checkout、含异源 helper 的目录输出相同，且只调用指定 Node 一次；发布候选另验外部命令。 | `tools/gen-fixtures-from-formatjs/node_script.go`；`main_test.go:TestNodeCLIOutsideCheckout`；删除 `tools/node-witness` | [README](../README.md)、[一致性验证](../SPECS/70-conformance.md)、`Taskfile.yml` |
| C03 | 完成。DisplayNames callback 消费每个声明或明确拒绝；静态操作不先构造实例，parts/ranges/未知 feature/混合错误观察不能静默通过。 | `displaynames/conformance_unified_test.go`；`displaynames/conformance_observations_test.go` | [一致性验证](../SPECS/70-conformance.md) |
| C04a | 完成。语义 key 排除期望值和断言序号，保留数字精度、观察 presence、operation/options/input；冲突失败，完全重复可合并。两次同源提取字节相同。 | `tools/gen-fixtures-from-formatjs/main.go:identifyFormatJSFixtures`；`main_test.go:TestFormatJSSemanticIdentity/TestSemanticIdentityObservationsAndPrecision` | [稳定身份](../SPECS/70-conformance.md) |
| C04b | 完成。226 条旧 generated fixtures 按完整非 ID payload 迁移并更新 ledger；只合并下列两组完全相同重复。全部正式生成文件及 skip-list 与新生成器重复输出一致。 | 各包 `testdata/conformance/formatjs/`；`datetimeformat/numberformat/pluralrules` 的 `testdata/divergences.md`；实际 formatter callbacks | [一致性验证](../SPECS/70-conformance.md) |
| D1 | 完成。两个真实生成器先写候选；auxiliary、plural 缺失/语法失败不改变正式生成路径和字节。安装完整生成集合，保留手写文件；安装失败非事务，保留候选并报具体位置。 | `tools/generate-data/{main,main_test}.go`；`tools/internal/generateddata/files.go`；`tools/check-generated-data/compare_test.go` | [生成边界](../SPECS/50-cldr-data.md) |
| DOC01 | 完成。纠正 Duration literal unit 来源、DateTimeFormat.RangeSource、best-fit 实现自由度和共享结构化错误归属。 | `durationformat/format.go`；`datetimeformat/resolved_options.go`；`internal/ecma402/datetimeformat/matcher.go`；`internal/intlerr/errors.go` | [SPEC 43](../SPECS/43-durationformat.md)、[30](../SPECS/30-datetimeformat.md)、[31](../SPECS/31-datetimeformat-skeleton.md)、[72](../SPECS/72-operation-ledger.md) |

迁移合并项：PluralRules 原 `index` 009/043 都是 `fr`、空 options、1000000 → `many`；
NumberFormat 原 `misc` 050/051 都是 `en`、signDisplay=exceptZero、0 → `0`。
其余非 ID payload 未改变；manual/Node 身份未变，未通过改期望掩盖行为差异。

## 直接参考证据

以下路径相对仓库根目录，参考代码只作行为证据，不要求复制其架构。

- L01–L03：`.references/ecma402/spec/locale.html`、`locales-currencies-tz.html`；
  `.references/formatjs/packages/intl-getcanonicallocales/{parser,canonicalizer}.ts`、
  `scripts/aliases.ts`；`.references/formatjs/packages/intl-locale/index.ts`；
  `.references/node/deps/icu-small/source/common/locid.cpp`；
  `.references/node/deps/v8/src/objects/js-locale.cc`。
- L02d/B01：`.references/ecma402/spec/displaynames.html`；
  `.references/formatjs/packages/intl-displaynames/index.ts`；
  `.references/node/deps/v8/src/objects/js-display-names.cc`；
  `.references/node/deps/icu-small/source/common/locdspnm.cpp`。
- L04：`.references/ecma402/spec/negotiation.html`；
  `.references/formatjs/packages/intl-localematcher/abstract/LookupSupportedLocales.ts`。
- N20/N21：`.references/ecma402/spec/numberformat.html`、`pluralrules.html`；
  `.references/formatjs/packages/ecma402-abstract/NumberFormat/SetNumberFormatDigitOptions.ts`；
  `.references/formatjs/packages/ecma402-abstract/PluralRules/ResolvePlural.ts`；
  `.references/node/deps/icu-small/source/i18n/number_compact.cpp`。
- DTF-01/DOC01：`.references/ecma402/spec/datetimeformat.html`、`durationformat.html`；
  `.references/formatjs/packages/ecma402-abstract/DateTimeFormat/skeleton.ts`；
  `.references/formatjs/packages/intl-durationformat/abstract/PartitionDurationFormatPattern.ts`。
- C01/C04：`.references/formatjs/tools/test262/runner/{validate,results,results-checks}.ts`；
  `.references/node/deps/v8/tools/testrunner/objects/testcase.py`。引用支持执行与身份边界，
  语义指纹为本库设计，参考项目并未提供相同 hash 算法。
- C02/D1：`.references/formatjs/packages/cli/BUILD.bazel`、
  `.references/formatjs/packages/intl-numberformat/BUILD.bazel`；
  `.references/node/tools/icu/icu-generic.gyp`。借鉴完整产物边界，不引入构建框架。
- C03：`.references/ecma402/spec/displaynames.html`；
  `.references/formatjs/packages/intl-displaynames/tests/index.test.ts`。

参考提交：ECMA-402 `730ca2b88c305862db7be2f83c510dcca33ab48c`，FormatJS
`aa3f90b53b157481515a0f7a7f0d76fcbd7c3274`，Node
`5a566ac6b971a9bcbcff1d338f391e0ae6dbd678`。宿主 Node 26.10 不替代 active
Node 26.0 witness，也未用于伪造新 Variants 行为。CLDR 仍固定为 48.1.0。

## 审查与交付

独立只读审查发现的 transformed/subdivision alias 漏洞、混合错误观察遗漏及
subdivision 回归缺少结果断言均已补齐；再次静态复核没有剩余运行时问题。
可读性、职责归属、安全输入边界与性能检查未发现需进一步扩张范围的问题。
旧 identity 文档和 compact 摘要也已同步。AGENTS.md 仍为 CLAUDE.md 的符号链接。

已通过完整 `task verify`（lint、vet、race、四模块 test/tidy、conformance、数据契约、漏洞扫描）、
真实 `task data` 与 `task data:check`。生成数据只增加预期的 canonical alias 数据，
其余生成文件与固定来源一致。发布前还对最终候选提交运行完整 gate、四模块
`GOWORK=off` 候选图及仓外发布 CLI 验证；发布后直接解析模块并核对 checksum 与提交。
Windows 实机验证不在本次通过声明中。`Locale.Tag` 是破坏性 API 变更；按用户明确
要求仍选择协调 patch 版本，不以 patch 标签暗示源码兼容。

参考依赖一并固定：`.agents/skills`、carbon、FormatJS、intl、Node 保留用户更新的
提交。ext 的嵌套 test262 更新提交发布到 `kaptinlin/ext` 的
`codex/test262-reference-20261021`，`.gitmodules` 使用该可获取 fork，避免引用只存在
本机的父仓库提交。未修改参考实现源码，也未向上游发送 PR。

CI 后续发现并修正：`v0.4.16` 的 Test 作业只安装 CLDR，缺少 IANA archive，
使 D1 故障测试在预期故障之前失败。`test` / `modules:verify` 现在通过共享
`data:fetch` 前置任务准备两类固定输入，并在最终隔离验证中从空 IANA cache
验证下载与测试顺序；未删断言或跳过测试。原标签保持不可变，修正采用下一组
协调 patch `v0.4.17`。发布命令在已安装的 Go 1.27.2 上禁用隐式工具链切换。
