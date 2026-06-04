package main

import (
	"fmt"

	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/indexer"
	"github.com/kentakom1213/albam/internal/scanner"
)

func runScan(args []string) error {
	cfg, err := config.Load("albam.toml")
	if err != nil {
		return err
	}

	root := cfg.Media.SourceDir
	if len(args) == 1 {
		root = args[0]
	} else if len(args) > 1 {
		return fmt.Errorf("usage: albam scan [dir]")
	}

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
