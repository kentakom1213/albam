package main

import (
	"flag"
	"fmt"
	"runtime"
)

var (
	version = "dev"
	commit  = "unknown"
	builtAt = "unknown"
)

func runVersion(args []string) error {
	fs := newFlagSet("version", "usage: albam version [--verbose]")

	var verbose bool
	fs.BoolVar(&verbose, "verbose", false, "print detailed version information")

	if err := parseFlags(fs, args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: albam version [--verbose]")
	}

	fmt.Printf("albam %s\n", version)
	if verbose {
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", builtAt)
		fmt.Printf("go: %s\n", runtime.Version())
	}

	return nil
}
