package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/agentable/go-intl/internal/conformance"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	for _, root := range args {
		if err := conformance.ValidateDivergences(root, filepath.Join(root, "testdata", "divergences.md")); err != nil {
			return err
		}
	}
	return nil
}
