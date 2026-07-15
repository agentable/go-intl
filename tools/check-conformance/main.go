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
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}

func run(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("check-conformance", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	skipListPath := fs.String("skip-list", "", "root .skip-list.json path")
	coverage := fs.Bool("coverage", false, "print conformance coverage health")
	nodeWitness := fs.Bool("node-witness", false, "validate required Node witness coverage")
	if err := fs.Parse(args); err != nil {
		return err
	}
	roots := fs.Args()
	if err := conformance.ValidateFixtureRoots(roots, time.Now()); err != nil {
		return err
	}
	if *skipListPath != "" {
		if err := conformance.ValidateSkipList(*skipListPath, roots); err != nil {
			return err
		}
	}
	if *nodeWitness {
		if err := conformance.ValidateNodeWitnessCoverage(roots); err != nil {
			return err
		}
	}
	if *coverage {
		report, err := conformance.CoverageReport(roots, *skipListPath)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprint(stdout, report); err != nil {
			return err
		}
	}
	return nil
}
