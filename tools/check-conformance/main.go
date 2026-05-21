package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/agentable/go-intl/tools/conformance"
)

var exitProcess = os.Exit

func main() {
	exitProcess(mainExit(os.Args[1:], os.Stdout, os.Stderr))
}

func mainExit(args []string, stdout, stderr io.Writer) int {
	if err := run(args, stdout); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("check-conformance", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	skipListPath := fs.String("skip-list", "", "root .skip-list.json path")
	coverage := fs.Bool("coverage", false, "print conformance coverage health")
	if err := fs.Parse(args); err != nil {
		return err
	}
	roots := fs.Args()
	if err := conformance.ValidateFixtureRoots(roots, time.Now()); err != nil {
		return err
	}
	for _, root := range roots {
		if err := conformance.ValidateDivergences(root); err != nil {
			return err
		}
	}
	if *skipListPath != "" {
		if err := conformance.ValidateSkipList(*skipListPath, roots); err != nil {
			return err
		}
	}
	if *coverage {
		report, err := conformance.CoverageReport(roots, *skipListPath)
		if err != nil {
			return err
		}
		fmt.Fprint(stdout, report)
	}
	return nil
}
