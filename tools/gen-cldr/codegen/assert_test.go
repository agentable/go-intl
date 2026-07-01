package codegen

import (
	"bytes"
	"strings"
	"testing"
)

func assertBytesEqual(t *testing.T, name string, got, want []byte) {
	t.Helper()
	if !bytes.Equal(got, want) {
		t.Fatalf("%s = %v, want %v", name, got, want)
	}
}

func assertSourceContains(t *testing.T, name, src, want string) {
	t.Helper()
	if !strings.Contains(src, want) {
		t.Fatalf("%s missing %q:\n%s", name, want, src)
	}
}

func assertSourceContainsAll(t *testing.T, name, src string, wants ...string) {
	t.Helper()
	for _, want := range wants {
		assertSourceContains(t, name, src, want)
	}
}

func assertErrorContains(t *testing.T, name string, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s error = nil, want containing %q", name, want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("%s error = %q, want containing %q", name, err, want)
	}
}

func assertErrorContainsAll(t *testing.T, name string, err error, wants ...string) {
	t.Helper()
	for _, want := range wants {
		assertErrorContains(t, name, err, want)
	}
}
