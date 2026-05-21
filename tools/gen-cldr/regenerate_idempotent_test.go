package main

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestRunGeneratesIdempotentOutput(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	root := filepath.Join(dir, "node_modules")
	writeRuntimeCLDRFixtures(t, root)
	writeListPatternCLDRFixture(t, root)
	writeRelativeTimeCLDRFixture(t, root)

	versionPath := filepath.Join(dir, "VERSION")
	if err := os.WriteFile(versionPath, []byte("cldr=48.1.0\nicu=78\ntzdata=2025b\n"), 0o666); err != nil {
		t.Fatalf("write VERSION: %v", err)
	}
	profilePath := writeLocaleProfileFixture(t, dir)

	first := filepath.Join(dir, "first")
	second := filepath.Join(dir, "second")
	runCLDRGenerator(t, root, first, versionPath, profilePath)
	runCLDRGenerator(t, root, second, versionPath, profilePath)

	requireGeneratedOutputEqual(t, first, second)
}

func runCLDRGenerator(t *testing.T, root, out, versionPath, profilePath string) {
	t.Helper()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := Config{
		CLDRDir:     root,
		OutDir:      out,
		VersionFile: versionPath,
		ProfileFile: profilePath,
	}
	if err := Run(context.Background(), cfg, log); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func requireGeneratedOutputEqual(t *testing.T, first, second string) {
	t.Helper()

	want := readGeneratedOutput(t, first)
	got := readGeneratedOutput(t, second)
	if !slices.Equal(want.names, got.names) {
		t.Fatalf("generated files differ:\nfirst:  %v\nsecond: %v", want.names, got.names)
	}
	for _, name := range want.names {
		if !bytes.Equal(want.files[name], got.files[name]) {
			t.Fatalf("generated file %s is not idempotent", filepath.ToSlash(name))
		}
	}
}

type generatedOutput struct {
	names []string
	files map[string][]byte
}

func readGeneratedOutput(t *testing.T, root string) generatedOutput {
	t.Helper()

	out := generatedOutput{files: make(map[string][]byte)}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		name := filepath.ToSlash(rel)
		out.names = append(out.names, name)
		out.files[name] = data
		return nil
	})
	if err != nil {
		t.Fatalf("walk generated output: %v", err)
	}
	slices.Sort(out.names)
	return out
}
