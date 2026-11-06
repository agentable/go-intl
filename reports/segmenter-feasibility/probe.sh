#!/usr/bin/env bash
# Diagnostic only: dependencies and generated data stay outside the root module.
set -euo pipefail
probe_dir=$(mktemp -d "${TMPDIR:-/tmp}/go-intl-segmenter.XXXXXX")
trap 'rm -rf "$probe_dir"' EXIT
probe_go=${GO_BIN:-go}
probe_node=${NODE_BIN:-node}
if [[ $("$probe_node" -p 'process.version') != v26.0.0 ]]; then
    echo 'Set NODE_BIN to Node v26.0.0 (the active repository witness).' >&2
    exit 1
fi
cat > "$probe_dir/go.mod" <<'MOD'
module example.com/segmenter-probe

go 1.27.3

require (
 github.com/clipperhouse/uax29/v2 v2.7.0
 github.com/rivo/uniseg v0.4.7
)
MOD
cat > "$probe_dir/cases.json" <<'CASES'
["ไทยภาษา","ภาษาไทยx","中文测试","日本語の文章","ខ្មែរ","ພາສາລາວ","မြန်မာစာ","😀x","áb","👩‍💻x","ภาษาไทยภาษาไทย","研究生命起源","東京都に行く","English ไทยภาษา 中文测试","ไทยxyzภาษา","กำลังทดสอบระบบ"]
CASES
cat > "$probe_dir/main.go" <<'GO'
package main

import (
 "encoding/json/v2"
 "fmt"
 "os"
 "strings"
 "testing"
 "time"
 "unicode/utf16"

 "github.com/clipperhouse/uax29/v2/words"
 "github.com/rivo/uniseg"
)

type part struct {
 Segment string `json:"segment"`
 Index int `json:"index"`
 ByteIndex int `json:"byteIndex"`
}
func split(s, engine string) []part {
 result := []part{}
 byteIndex, index := 0, 0
 appendPart := func(word string) {
  result = append(result, part{word, index, byteIndex})
  byteIndex += len(word)
  index += len(utf16.Encode([]rune(word)))
 }
 if engine == "uax29" {
  it := words.FromString(s)
  for it.Next() { appendPart(it.Value()) }
 } else {
  state := -1
  for len(s) > 0 {
   var word string
   word, s, state = uniseg.FirstWordInString(s, state)
   appendPart(word)
  }
 }
 return result
}
var sink []part
func main() {
 data, err := os.ReadFile("cases.json"); if err != nil { panic(err) }
 var cases []string
 if err := json.Unmarshal(data, &cases); err != nil { panic(err) }
 results := map[string]map[string][]part{}
 for _, engine := range []string{"uax29", "uniseg"} {
  results[engine] = map[string][]part{}
  for _, input := range cases { results[engine][input] = split(input, engine) }
  long := strings.Repeat("English ไทยภาษา 中文测试 😀x ", 4096)
  started := time.Now()
  for range 20 { sink = split(long, engine) }
  elapsed := time.Since(started) / 20
  allocations := testing.AllocsPerRun(10, func() { sink = split(long, engine) })
  fmt.Fprintf(os.Stderr, "%s: bytes=%d segments=%d ns/op=%d allocs/op=%.0f (includes materializing records)\n", engine, len(long), len(sink), elapsed.Nanoseconds(), allocations)
 }
 output, err := json.Marshal(results); if err != nil { panic(err) }
 fmt.Println(string(output))
}
GO
cat > "$probe_dir/native.mjs" <<'JS'
import fs from 'node:fs';
const cases = JSON.parse(fs.readFileSync('cases.json', 'utf8'));
const go = JSON.parse(fs.readFileSync('go.json', 'utf8'));
console.log(JSON.stringify({versions: process.versions}));
for (const locale of ['en', 'th', 'zh', 'ja', 'km', 'lo', 'my']) {
 const segmenter = new Intl.Segmenter(locale, {granularity: 'word'});
 let mismatches = {uax29: 0, uniseg: 0};
 for (const input of cases) {
  const native = [...segmenter.segment(input)].map(p => ({segment: p.segment, index: p.index, byteIndex: Buffer.byteLength(input.slice(0, p.index)), isWordLike: p.isWordLike}));
  const boundaries = native.map(({isWordLike, ...p}) => p);
  for (const engine of Object.keys(mismatches)) {
   if (JSON.stringify(go[engine][input]) !== JSON.stringify(boundaries)) mismatches[engine]++;
  }
  // Detailed output once; other locales still compare the full records.
  if (locale === 'en') console.log(JSON.stringify({input, native, uax29: go.uax29[input], uniseg: go.uniseg[input]}));
 }
 console.log(JSON.stringify({locale, resolved: segmenter.resolvedOptions(), cases: cases.length, mismatches}));
}
JS
cd "$probe_dir"
GOWORK=off "$probe_go" mod tidy
GOWORK=off "$probe_go" build -o probe .
wc -c probe
./probe > go.json
"$probe_node" native.mjs
