package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func newFlagSet(name string, usage string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
	}

	return fs
}

func parseFlags(fs *flag.FlagSet, args []string) error {
	if err := validateLongOptions(args); err != nil {
		return err
	}

	if hasHelpFlag(args) {
		fs.Usage()
		return flag.ErrHelp
	}

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
		}
		return err
	}

	return nil
}

func validateLongOptions(args []string) error {
	for _, arg := range args {
		if arg == "-" || arg == "--" || !strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "--") {
			continue
		}

		return fmt.Errorf("invalid option %q: use --%s", arg, strings.TrimLeft(arg, "-"))
	}

	return nil
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "--help" {
			return true
		}
	}

	return false
}
