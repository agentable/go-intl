package gointl

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBenchmarkLayoutRequiresFiles(t *testing.T) {
	t.Parallel()

	for _, file := range []string{
		"numberformat/benchmark_test.go",
		"numberformat/benchmark_baseline_test.go",
		"datetimeformat/benchmark_test.go",
		"pluralrules/benchmark_test.go",
		"pluralrules/benchmark_baseline_test.go",
		"locale/benchmark_test.go",
	} {
		if _, err := os.Stat(file); err != nil {
			t.Fatalf("missing benchmark file %s", file)
		}
	}
}

func TestBenchmarkLayoutRequiresSpecBenchmarks(t *testing.T) {
	t.Parallel()

	required := map[string][]string{
		"numberformat/benchmark_test.go": {
			"BenchmarkNumberFormat_Decimal_PerCall",
			"BenchmarkNumberFormat_Decimal_Cached",
			"BenchmarkNumberFormat_Percent_Cached",
			"BenchmarkNumberFormat_Currency_Cached",
			"BenchmarkNumberFormat_Compact_Cached",
			"BenchmarkNumberFormat_FormatToParts_Cached",
			"BenchmarkNumberFormat_New",
		},
		"datetimeformat/benchmark_test.go": {
			"BenchmarkDateTimeFormat_DateStyleShort_PerCall",
			"BenchmarkDateTimeFormat_DateStyleShort_Cached",
			"BenchmarkDateTimeFormat_DateTimeRange_Cached",
			"BenchmarkDateTimeFormat_FormatToParts_Cached",
			"BenchmarkDateTimeFormat_New",
		},
		"pluralrules/benchmark_test.go": {
			"BenchmarkPluralRules_Cardinal_Cached",
			"BenchmarkPluralRules_Ordinal_Cached",
			"BenchmarkPluralRules_SelectRange_Cached",
		},
		"numberformat/benchmark_baseline_test.go": {
			"BenchmarkBaseline_XText_Decimal",
			"BenchmarkBaseline_XText_Percent",
		},
		"pluralrules/benchmark_baseline_test.go": {
			"BenchmarkBaseline_XText_Plural_Cardinal",
			"BenchmarkBaseline_XText_Plural_Ordinal",
		},
		"locale/benchmark_test.go": {
			"BenchmarkLocale_Parse",
			"BenchmarkLocale_New",
		},
	}
	for file, names := range required {
		benchmarks, err := benchmarkNames(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range names {
			if !benchmarks[name] {
				t.Fatalf("%s missing %s", file, name)
			}
		}
	}
}

func benchmarkNames(file string) (map[string]bool, error) {
	f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		return nil, err
	}
	benchmarks := map[string]bool{}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if ok && strings.HasPrefix(fn.Name.Name, "Benchmark") {
			benchmarks[fn.Name.Name] = true
		}
	}
	return benchmarks, nil
}

func TestBenchmarkLayoutUsesSpecNames(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*/benchmark_test.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		f, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || !strings.HasPrefix(fn.Name.Name, "Benchmark") {
				continue
			}
			parts := strings.Split(fn.Name.Name, "_")
			if len(parts) == 2 && (parts[1] == "New" || parts[1] == "Parse") {
				continue
			}
			if len(parts) != 3 || parts[0] == "BenchmarkBaseline" {
				t.Fatalf("%s: benchmark %s must use Benchmark<Type>_<Feature>_<Layer>", file, fn.Name.Name)
			}
		}
	}
}
