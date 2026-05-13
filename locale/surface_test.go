package locale_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocaleIsNotComparable(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), `module localecompare

go 1.26

require github.com/agentable/go-intl v0.0.0

replace github.com/agentable/go-intl => `+moduleRoot(t)+`
`)
	writeFile(t, filepath.Join(dir, "main.go"), `package main

import "github.com/agentable/go-intl/locale"

func main() {
	_ = locale.MustParse("en") == locale.MustParse("en")
}
`)
	cmd := exec.Command("/opt/homebrew/bin/go", "test", "-mod=mod", ".")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("go test unexpectedly succeeded:\n%s", out)
	}
	if !strings.Contains(string(out), "struct containing [0]func() cannot be compared") {
		t.Fatalf("go test error = %s, want non-comparable Locale", out)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o666); err != nil {
		t.Fatal(err)
	}
}
