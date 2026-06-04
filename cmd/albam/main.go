package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kentakom1213/albam/internal/api"
	"github.com/kentakom1213/albam/internal/cache"
	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/indexer"
	"github.com/kentakom1213/albam/internal/scanner"
	"github.com/kentakom1213/albam/internal/storage"
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
	case "index":
		if err := runIndex(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "serve":
		if err := runServe(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

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

func runServe(args []string) error {
	cfg, err := config.Load("albam.toml")
	if err != nil {
		return err
	}

	apiOnly := false
	for _, arg := range args {
		switch arg {
		case "--api-only":
			apiOnly = true
		default:
			return fmt.Errorf("usage: albam serve [--api-only]")
		}
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

	server := api.NewServer(store, cfg)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	if apiOnly {
		fmt.Printf("albam API server started\n\n  API: http://%s/api\n", addr)
	} else {
		fmt.Printf("albam server started\n\n  Local: http://%s\n  API: http://%s/api\n", addr, addr)
	}

	return http.ListenAndServe(addr, server.Routes())
}
