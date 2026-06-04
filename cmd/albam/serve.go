package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kentakom1213/albam/internal/api"
	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/storage"
)

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	var (
		configPath string
		host       string
		port       int
		publicDir  string
		apiOnly    bool
	)

	fs.StringVar(&configPath, "config", "albam.toml", "config file path")
	fs.StringVar(&host, "host", "", "listen host")
	fs.IntVar(&port, "port", 0, "listen port")
	fs.StringVar(&publicDir, "public-dir", "", "static public directory")
	fs.BoolVar(&apiOnly, "api-only", false, "serve only API and media routes")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: albam serve [--api-only] [--host host] [--port port] [--public-dir dir]")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	host = chooseString(host, cfg.Server.Host, "127.0.0.1")
	port = chooseInt(port, cfg.Server.Port, 8080)
	publicDir = chooseString(publicDir, cfg.Build.OutDir, ".albam/public")

	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = ".albam/db.sqlite"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return err
	}

	store, err := storage.Open(dbPath)
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		return err
	}

	server := api.NewServer(store, cfg)

	var handler http.Handler
	if apiOnly {
		handler = server.Routes()
	} else {
		handler, err = server.RoutesWithStatic(publicDir)
		if err != nil {
			return err
		}
	}

	addr := net.JoinHostPort(host, strconv.Itoa(port))

	printServeInfo(addr, apiOnly, publicDir)

	return http.ListenAndServe(addr, handler)
}

func chooseString(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}

	return ""
}

func chooseInt(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}

	return 0
}

func printServeInfo(addr string, apiOnly bool, publicDir string) {
	fmt.Println("albam server started")
	fmt.Println()
	fmt.Printf("  Local:  http://%s\n", addr)
	fmt.Printf("  API:    http://%s/api\n", addr)
	fmt.Printf("  Media:  http://%s/media\n", addr)

	if apiOnly {
		fmt.Println("  Mode:   api-only")
	} else {
		fmt.Printf("  Static: %s\n", publicDir)
	}
}
