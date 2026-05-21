package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: albam <command> [args]")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "scan":
		fmt.Println("scan command is not implemented yet")
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}
}
