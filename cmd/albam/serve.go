package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kentakom1213/albam/internal/api"
	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/storage"
)

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
