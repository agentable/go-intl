package main

import (
	"context"
	"flag"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/agentable/go-intl/tools/internal/generateddata"
)

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	if err := run(context.Background(), *root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, root string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(root, ".tmp"), 0o750); err != nil {
		return err
	}
	stage, err := os.MkdirTemp(filepath.Join(root, ".tmp"), "generated-")
	if err != nil {
		return err
	}
	if err := generate(ctx, root, stage, filepath.Join(root, "tools", "gen-cldr", ".cldr-json", "node_modules", "cldr-core", "supplemental", "plurals.json")); err != nil {
		return fmt.Errorf("generation failed; working tree unchanged; candidate retained at %s: %w", stage, err)
	}
	if err := install(filepath.Join(root, "internal"), filepath.Join(stage, "internal")); err != nil {
		return fmt.Errorf("installation failed (may be partial); candidate retained at %s: %w", stage, err)
	}
	return os.RemoveAll(stage)
}

// generate writes only to the candidate tree. Each generator validates its inputs.
func generate(ctx context.Context, root, stage, plurals string) error {
	cldr := filepath.Join(root, "tools", "gen-cldr")
	commands := [][]string{
		{"-C", cldr, "run", ".", "-cldr-dir", filepath.Join(cldr, ".cldr-json", "node_modules"),
			"-out", filepath.Join(stage, "internal", "cldr"), "-version-file", filepath.Join(root, "internal", "cldr", "VERSION"),
			"-profile", filepath.Join(root, "tools", "locale-profile.json"),
			"-localeid-out", filepath.Join(stage, "internal", "localeid", "unicode_alias_data.go"),
			"-localematcher-out", filepath.Join(stage, "internal", "localematcher", "profile_data.go"),
			"-tzdata-lock", filepath.Join(cldr, "tzdata.json"), "-timezone-out", filepath.Join(stage, "internal", "tz", "registry_data.go")},
		{"-C", filepath.Join(root, "tools", "gen-plural-rules"), "run", ".", "-plurals", plurals,
			"-ranges", filepath.Join(cldr, ".cldr-json", "node_modules", "cldr-core", "supplemental", "pluralRanges.json"),
			"-out", filepath.Join(stage, "internal", "cldr", "plural"), "-profile", filepath.Join(root, "tools", "locale-profile.json")},
	}
	for _, args := range commands {
		// Arguments are repository-owned generator commands; no shell is used.
		cmd := exec.CommandContext(ctx, "go", args...) // #nosec G204
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("%s: %w", args[1], err)
		}
	}
	return nil
}

func install(target, candidate string) error {
	current, err := generateddata.Read(target)
	if err != nil {
		return err
	}
	fresh, err := generateddata.Read(candidate)
	if err != nil {
		return err
	}
	if len(fresh) == 0 {
		return fmt.Errorf("empty candidate %s", candidate)
	}
	dir, err := os.OpenRoot(target)
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()
	// Check the entire destination set before replacing anything handwritten.
	for path := range fresh {
		if _, owned := current[path]; owned {
			continue
		}
		if _, err := dir.Stat(path); err == nil {
			return fmt.Errorf("refuse to replace handwritten file %s", path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	for _, path := range slices.Sorted(maps.Keys(fresh)) {
		if err := dir.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return fmt.Errorf("mkdir %s: %w", path, err)
		}
		if err := dir.WriteFile(path, fresh[path], 0o644); err != nil {
			return fmt.Errorf("install %s: %w", path, err)
		}
	}
	for _, path := range slices.Sorted(maps.Keys(current)) {
		if _, ok := fresh[path]; !ok {
			if err := dir.Remove(path); err != nil {
				return fmt.Errorf("remove obsolete %s: %w", path, err)
			}
		}
	}
	return nil
}
