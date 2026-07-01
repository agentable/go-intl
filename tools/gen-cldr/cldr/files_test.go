package cldr

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestReadRequiredFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "data.json")
	mustWriteFile(t, path, `{"ok":true}`)

	raw, err := readRequiredFile(path)
	if err != nil {
		t.Fatalf("readRequiredFile(existing) error = %v", err)
	}
	if string(raw) != `{"ok":true}` {
		t.Fatalf("readRequiredFile(existing) raw = %q, want %q", raw, `{"ok":true}`)
	}
}

func TestReadRequiredFileRejectsMissingFile(t *testing.T) {
	t.Parallel()

	_, err := readRequiredFile(filepath.Join(t.TempDir(), "missing.json"))
	if err == nil {
		t.Fatal("readRequiredFile(missing) succeeded, want error")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("readRequiredFile(missing) error = %v, want os.ErrNotExist", err)
	}
}

func TestReadRequiredFileRejectsReadError(t *testing.T) {
	t.Parallel()

	if _, err := readRequiredFile(t.TempDir()); err == nil {
		t.Fatal("readRequiredFile(directory) succeeded, want error")
	}
}

func TestReadOptionalFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, "data.json")
	mustWriteFile(t, path, `{"ok":true}`)

	raw, ok, err := readOptionalFile(path)
	if err != nil {
		t.Fatalf("readOptionalFile(existing) error = %v", err)
	}
	if !ok {
		t.Fatal("readOptionalFile(existing) ok = false, want true")
	}
	if string(raw) != `{"ok":true}` {
		t.Fatalf("readOptionalFile(existing) raw = %q, want %q", raw, `{"ok":true}`)
	}
}

func TestReadOptionalFileMissing(t *testing.T) {
	t.Parallel()

	raw, ok, err := readOptionalFile(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("readOptionalFile(missing) error = %v", err)
	}
	if ok {
		t.Fatal("readOptionalFile(missing) ok = true, want false")
	}
	if raw != nil {
		t.Fatalf("readOptionalFile(missing) raw = %q, want nil", raw)
	}
}

func TestReadOptionalFileRejectsReadError(t *testing.T) {
	t.Parallel()

	_, _, err := readOptionalFile(t.TempDir())
	if err == nil {
		t.Fatal("readOptionalFile(directory) succeeded, want error")
	}
}
