# golang.org/x/text：Locale 语法与别名后端的能力边界

以下第 1、2 节是 2026-10-04 的历史验证，当时 `go.mod` 是 x/text v0.42.0、Go 1.27.0；旧报告的 v0.41.0 版本号不能作为当前证据。本报告记录 go-intl 所需 ECMA-402 契约与现有后端的差异，没有判定 x/text 违背它自己的契约，也没有定位或修改依赖内部根因。

## 1. Well-formed 不应等同于后端已识别

`locale/locale.go:33` 通过 `internal/localeid/tag.go:10` 的 `ParseTag` 调用 `language.Parse`，再执行 `(language.Macro | language.CLDR).Canonicalize`。`language.Parse` 对以下合法但未知的 subtag 返回错误，go-intl 将其映射成 `ErrInvalidValue`。

| 输入 | 当前 Go 结果 | Node 26.10.0 / ICU 78.3 结果 |
|---|---|---|
| `und-Gara` | `ErrInvalidValue`；dependency cause 为 `language: subtag "Gara" is well-formed but unknown` | `und-Gara` |
| `sjc` | `ErrInvalidValue`；dependency cause 为 `language: subtag "sjc" is well-formed but unknown` | `sjc` |
| `xx-INVALID` | `ErrInvalidValue` | `xx-invalid` |
| `abcdefgh` | `ErrInvalidValue` | `abcdefgh` |
| `en-foobar` | `ErrInvalidValue` | `en-foobar` |
| `qaa` | `qaa` | `qaa` |

代表性的公共错误：

```text
locale: invalid value languageTag "und-Gara": expected a well-formed BCP 47 language tag; got "und-Gara"
```

预期依据是 `.references/ecma402/spec/locales-currencies-tz.html:34` 的 `IsWellFormedLanguageTag`：语法检查不考虑 tag 是否具有 meaningful semantics。`.references/formatjs/packages/intl-getcanonicallocales/parser.ts` 也以 ASCII subtag 语法检查，而不是识别表成员资格。

`und-Gara`、`sjc` 来自当前 pin 的 CLDR 48.1.0；来源为 `tools/gen-cldr/.cldr-json/node_modules/cldr-core/supplemental/likelySubtags.json`。旧报告列出的其他新 CLDR subtag 在本次没有逐一重测，不能把旧清单全部标为当前复现。

## 2. 地区别名的路径差异

已执行以下同版本探针：

```text
language.Parse("und-Armn-SU")                         => und-Armn-SU
localeid.ParseTag("und-Armn-SU")                      => und-Armn-SU
locale.Parse("und-Armn-SU")                           => und-Armn-AM
locale.New("und-Armn", Options{Region: new("SU")})     => und-Armn-SU
```

Node `new Intl.Locale("und-Armn-SU")` 和 `new Intl.Locale("und-Armn", {region: "SU"})` 都得到 `und-Armn-AM`；`new Intl.Locale("und-Armn-SU", {language: "ru"})` 得到 `ru-Armn-AM`，保留先规范化原 tag、再应用 override 的顺序。

当前 Parse 的 AM 结果来自仓库已有 `locale/validate.go:55` 的手写特例，constructor Region override 没有经过该特例。这是现有本地路径分裂，不能只归责于依赖，也不能把移动这一条 if 当作完整别名能力修复。

相关实现证据：`.references/formatjs/packages/intl-getcanonicallocales/canonicalizer.ts:205` 按 territory alias 与 likely region 选择替代地区；`.references/node/deps/icu-small/source/common/locid.cpp` 的 `AliasReplacer::replaceRegion` 用同样的数据选择；V8 `.references/node/deps/v8/src/objects/js-locale.cc:312` 的 `ApplyOptionsToTag` 在 override 前规范化，`JSLocale::New` 在构造后再规范化。

## 重现方式

在当前 module 的临时程序中调用上述公共 API，以及 `language.Parse` / `localeid.ParseTag`，打印 canonical tag 与错误 cause；Node 对照使用：

```javascript
for (const tag of ["und-Gara", "sjc", "xx-INVALID", "abcdefgh", "en-foobar", "qaa"]) {
  console.log(tag, new Intl.Locale(tag).toString());
}
console.log(new Intl.Locale("und-Armn", {region: "SU"}).toString());
```

本次 Go/Node 探针已执行，临时 `.tmp` 程序已删除。`go test -race ./locale ./internal/localeid ./internal/localematcher ./internal/ecma402` 也通过；这只说明当前测试通过，现有 `TestParseInvalidLocale` 仍把 `xx-INVALID` 归为非法，不能充当 ECMA 契约正确性的证明。

## 建议与停止条件

先为公开 Locale 边界建立原生/规范 case，再判断上游是否有符合所需语法和别名契约的 API 或修复。只有后端能够保留合法未知标识、提供规范别名语义并保持 variant/extension/private-use 后缀时，才更新统一 `ParseTag` owner 并删除现有私有特例。

本次没有实施 workaround、替代 parser、dependency fork、版本升级或生产代码修改；不要继续增加语言/脚本/地区特例来绕过能力限制。不存在可用后端修复时，依赖于该能力的实现项保持阻塞，其他 Locale 语法和 preference 工作可以独立进行。


## 3. v0.43.0：合法未知地区触发 getter panic

2026-10-15，Go 1.27.2、x/text v0.43.0。`go list -m -json golang.org/x/text@latest` 同样返回 v0.43.0。本节未重新认证上面所有历史案例。

`locale.Parse("en-u-rg-000zzzz")` 与 `locale.Parse("en-u-sd-000abc")` 成功，但 GetCalendars、GetHourCycles、GetWeekInfo 均 panic。调用路径是 `locale/info.go:canonicalUnicodeSubdivision` → `locale/canonical.go:mustLanguageTag` → `internal/localeid/tag.go:ParseTag`；x/text 拒绝 `und-000`，调用者却把这个由用户输入生成的 tag 当成必定可解析的内部值。

| 依赖输入 | ParseRegion / Canonicalize | ParseTag 路径 |
|---|---|---|
| `000` | 返回 ZZ 与 ValueError；原地区已丢失 | Parse 返回 und 与 error，继续 Canonicalize 无法找回 000 |
| `ZZ` | ZZ，无错误 | und-ZZ |
| `001` | 001，无错误 | und-001 |
| `BU` | BU → MM | und-MM |
| `UK` | UK → GB | und-GB |
| `SU` | SU → SU | und-SU |

依赖源码位于 `$(go env GOMODCACHE)/golang.org/x/text@v0.43.0/`：

- `language/language.go:ParseRegion` 明确对 well-formed but unknown 地区返回 ValueError，不能忽略错误并使用其返回值。
- `language/language.go:Region.Canonicalize` 明确不替换拆分为多个地区的废弃地区，不能用它代替完整的 Unicode alias 选择。
- `language/parse.go:CanonType.Compose` 接受已构造的 Region；它不提供保留未知 000 的另一条入口。

本库所需合同由 `.references/ecma402/spec/locale.html:CanonicalUnicodeSubdivision/RegionPreference` 定义：先规范化合法地区，再按地区偏好查数据。`.references/formatjs/packages/intl-locale/index.ts:regionPreference` 经 Locale 规范化；`.references/node/deps/v8/src/objects/js-locale.cc:CanonicalUnicodeSubdivision` 此处只大写化地区，不能充当完整 alias 语义的替代算法。

主机 Node v26.10.0 / ICU 78.3 / CLDR 48.0 对两个 000 输入的三个 getter 都正常返回：calendars 为 [gregory]，hour cycles 为 [h12]，week 为 firstDay=7、weekend=[6,7]。Go 的 BU/001 控制例正常返回，但 hour-cycle 候选和其他实现定义数据不要求逐字照搬主机 Node。主机观察不是仓库 pinned Node v26.0.0 witness。

### 可复制复现

在仓库根目录将以下程序保存为临时目录下的 `main.go`，运行 `go run ./该目录`。这里的 recover 仅用于诊断程序逐个观察 getter；不得进入生产路径。

```go
package main

import (
	"fmt"
	"github.com/agentable/go-intl/locale"
	"golang.org/x/text/language"
)

func inspect(tag, name string, getter func(locale.Locale) any) {
	defer func() {
		if e := recover(); e != nil {
			fmt.Printf("%s %s panic=%v\n", tag, name, e)
		}
	}()
	l, e := locale.Parse(tag)
	if e != nil {
		fmt.Println(tag, e)
		return
	}
	fmt.Printf("%s %s=%v\n", tag, name, getter(l))
}
func main() {
	for _, r := range []string{"000", "ZZ", "SU", "BU", "UK", "GB", "001"} {
		region, err := language.ParseRegion(r)
		tag, terr := language.Parse("und-" + r)
		canon, cerr := (language.Macro | language.CLDR).Canonicalize(tag)
		fmt.Printf("region=%s parsed=%s canonical=%s err=%T:%v tag=%s err=%v canonicalTag=%s err=%v\n", r, region, region.Canonicalize(), err, err, tag, terr, canon, cerr)
	}
	for _, tag := range []string{"en-u-rg-000zzzz", "en-u-sd-000abc", "en-u-rg-buzzzz", "en-u-sd-buabc", "en-u-rg-zzzzzz", "en-u-rg-001zzzz"} {
		inspect(tag, "calendars", func(l locale.Locale) any { return l.GetCalendars() })
		inspect(tag, "hours", func(l locale.Locale) any { return l.GetHourCycles() })
		inspect(tag, "week", func(l locale.Locale) any { return l.GetWeekInfo() })
	}
}

```

Node 对照：

```sh
node - <<'JS'
console.log(process.version, process.versions.icu, process.versions.cldr);
for (const tag of ['en-u-rg-000zzzz', 'en-u-sd-000abc']) {
  const l = new Intl.Locale(tag);
  console.log(tag, l.getCalendars(), l.getHourCycles(), l.getWeekInfo());
}
for (const region of ['000', 'ZZ', '001', 'BU', 'UK', 'SU']) {
  console.log(region, new Intl.Locale('und-' + region).region);
}
JS
```

**裁决：DEFER。** 当前后端尚无已验证解法，getter 缺陷未修复。这是 x/text 能力与本库 ECMA-402 合同不一致，不是已证明 x/text 违反自身合同。保持公开 getter 形状，不添加 recover、自制 alias 表、忽略错误后的默认值或依赖 fork。规范规定的无数据 fallback 与解析失败后伪装成功是不同边界。只有具体上游能力或经过证明的局部方案才能重新打开修复；该限制不阻塞原始语法拒绝与显式身份保留。
