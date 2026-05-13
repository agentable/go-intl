package main

import (
	"fmt"
	"os"
	"time"

	"github.com/agentable/go-intl/internal/conformance"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return conformance.ValidateFixtureRoots(args, time.Now())
}
