package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/kentakom1213/albam/internal/cache"
	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/indexer"
	"github.com/kentakom1213/albam/internal/scanner"
	"github.com/kentakom1213/albam/internal/storage"
)

func runIndex(args []string) error {
	cfg, err := config.Load("albam.toml")
	if err != nil {
		return err
	}

	root := cfg.Media.SourceDir
	if len(args) == 1 {
		root = args[0]
	} else if len(args) > 1 {
		return fmt.Errorf("usage: albam index [dir]")
	}

	files, err := scanner.Scan(root)
	if err != nil {
		return err
	}

	library, err := indexer.BuildLibrary(files)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Database.Path), 0o755); err != nil {
		return err
	}

	store, err := storage.Open(cfg.Database.Path)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		return err
	}

	if err := store.SaveLibrary(library); err != nil {
		return err
	}

	// 削除処理
	keepPaths := make([]string, 0, len(files))
	for _, file := range files {
		keepPaths = append(keepPaths, file.RelPath)
	}

	pruneResult, err := store.PruneAssetsByPaths(keepPaths)
	if err != nil {
		return err
	}

	removedPhotoIDs := make([]int64, 0, len(pruneResult.Removed))
	for _, asset := range pruneResult.Removed {
		removedPhotoIDs = append(removedPhotoIDs, asset.ID)
	}

	if err := cache.RemovePhotoVariantCaches(cfg.Media.CacheDir, removedPhotoIDs); err != nil {
		return err
	}

	fmt.Printf(
		"indexed %d albums and %d assets, removed %d assets\n",
		len(library.Albums),
		len(library.Assets),
		len(pruneResult.Removed),
	)
	return nil
}
