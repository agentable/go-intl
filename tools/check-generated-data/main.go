package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("check-generated-data", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	committed := fs.String("committed", "", "committed generated-data root")
	generated := fs.String("generated", "", "fresh generated-data root")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *committed == "" || *generated == "" || fs.NArg() != 0 {
		return fmt.Errorf("usage: check-generated-data -committed <root> -generated <root>")
	}
	return compareGeneratedData(*committed, *generated)
}
