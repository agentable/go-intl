package codegen

import (
	"strings"
	"testing"
)

func TestFormatFile_AddsGeneratedHeaderAndFormats(t *testing.T) {
	t.Parallel()

	got, err := FormatFile([]byte("package cldr\n\nfunc X( ){return}\n"))
	if err != nil {
		t.Fatalf("FormatFile: %v", err)
	}
	src := string(got)
	if !strings.HasPrefix(src, generatedHeader) {
		t.Fatalf("missing generated header:\n%s", src)
	}
	assertSourceContains(t, "FormatFile output", src, "func X() { return }")
}
