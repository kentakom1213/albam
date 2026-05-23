package main

import (
	"fmt"
	"os"

	"github.com/kentakom1213/go-webapp-tutorial/internal/indexer"
	"github.com/kentakom1213/go-webapp-tutorial/internal/scanner"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: albam <command> [args]")
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "scan":
		if err := runScan(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func runScan(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: albam scan <dir>")
	}

	root := args[0]

	files, err := scanner.Scan(root)
	if err != nil {
		return err
	}

	library, err := indexer.BuildLibrary(files)
	if err != nil {
		return err
	}

	fmt.Println("albums:")
	for _, album := range library.Albums {
		parent := "none"
		if album.ParentID != nil {
			parent = fmt.Sprint(*album.ParentID)
		}

		fmt.Printf("- id=%d path=%q parent=%s title=%q\n",
			album.ID,
			album.Path,
			parent,
			album.Title,
		)
	}

	fmt.Println("assets:")
	for _, asset := range library.Assets {
		fmt.Printf("- id=%d album_id=%d path=%q\n",
			asset.ID,
			asset.AlbumID,
			asset.Path,
		)
	}

	return nil
}
