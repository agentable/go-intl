package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: data-preflight")
	}
	goUpdateFile, err := resolveGoUpdateFile(context.Background())
	if err != nil {
		return err
	}
	return checkDataPins(preflightConfig{
		versionFile:  "internal/cldr/VERSION",
		packageFile:  "tools/gen-cldr/.cldr-json/package.json",
		tzLockFile:   "tools/gen-cldr/tzdata.json",
		goUpdateFile: goUpdateFile,
	})
}

func resolveGoUpdateFile(ctx context.Context) (string, error) {
	output, err := exec.CommandContext(ctx, "go", "env", "GOROOT").Output()
	if err != nil {
		return "", fmt.Errorf("resolve GOROOT with go env: %w", err)
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", fmt.Errorf("resolve GOROOT with go env: empty output")
	}
	return filepath.Join(root, "lib", "time", "update.bash"), nil
}
