package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Title string `toml:"title"`

	Server        ServerConfig   `toml:"server"`
	Media         MediaConfig    `toml:"media"`
	PrivacyConfig PrivacyConfig  `toml:"privacy"`
	Database      DatabaseConfig `toml:"database"`
	Build         BuildConfig    `toml:"build"`
	Theme         ThemeConfig    `toml:"theme"`
}

type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type MediaConfig struct {
	SourceDir             string `toml:"source_dir"`
	CacheDir              string `toml:"cache_dir"`
	AllowOriginalDownload bool   `toml:"allow_original_download"`
}

type PrivacyConfig struct {
	MapEnabled        bool   `toml:"map_enabled"`
	ExposeGPS         bool   `toml:"expose_gps"`
	LocationPrecision string `toml:"location_precision"`
}

type DatabaseConfig struct {
	Path string `toml:"path"`
}

type BuildConfig struct {
	OutDir string `toml:"out_dir"`
}

type ThemeConfig struct {
	Name   string         `toml:"name"`
	Dir    string         `toml:"dir"`
	Params map[string]any `toml:"params"`
}

type ThemeManifest struct {
	Name        string         `toml:"name"`
	DisplayName string         `toml:"display_name"`
	Version     string         `toml:"version"`
	Author      string         `toml:"author"`
	Description string         `toml:"description"`
	Defaults    map[string]any `toml:"defaults"`
}

type ThemePayload struct {
	Site  ThemePayloadSite  `json:"site"`
	Theme ThemePayloadTheme `json:"theme"`
}

type ThemePayloadSite struct {
	Title string `json:"title"`
}

type ThemePayloadTheme struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"params"`
}

func Default() Config {
	return Config{
		Title: "My Albums",
		Server: ServerConfig{
			Host: "127.0.0.1",
			Port: 8080,
		},
		Media: MediaConfig{
			SourceDir:             "albums",
			CacheDir:              ".albam/cache",
			AllowOriginalDownload: false,
		},
		PrivacyConfig: PrivacyConfig{
			MapEnabled:        false,
			ExposeGPS:         false,
			LocationPrecision: "hidden",
		},
		Database: DatabaseConfig{
			Path: ".albam/db.sqlite",
		},
		Build: BuildConfig{
			OutDir: ".albam/public",
		},
		Theme: ThemeConfig{
			Name: "default",
			Dir:  "themes/default",
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return Config{}, err
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func ResolveThemeDir(cfg Config) string {
	if cfg.Theme.Dir != "" {
		return cfg.Theme.Dir
	}
	if cfg.Theme.Name != "" {
		return filepath.Join("themes", cfg.Theme.Name)
	}
	return filepath.Join("themes", "default")
}

func LoadThemeManifest(themeDir string) (ThemeManifest, error) {
	var manifest ThemeManifest
	path := filepath.Join(themeDir, "theme.toml")

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return manifest, nil
		}
		return ThemeManifest{}, err
	}

	if _, err := toml.DecodeFile(path, &manifest); err != nil {
		return ThemeManifest{}, err
	}

	return manifest, nil
}

func BuildThemePayload(cfg Config) (ThemePayload, error) {
	manifest, err := LoadThemeManifest(ResolveThemeDir(cfg))
	if err != nil {
		return ThemePayload{}, err
	}

	params := cloneMap(manifest.Defaults)
	mergeMap(params, cfg.Theme.Params)

	themeName := cfg.Theme.Name
	if themeName == "" {
		themeName = manifest.Name
	}
	if themeName == "" {
		themeName = "default"
	}

	return ThemePayload{
		Site: ThemePayloadSite{
			Title: cfg.Title,
		},
		Theme: ThemePayloadTheme{
			Name:   themeName,
			Params: params,
		},
	}, nil
}

func WriteThemePayload(path string, payload ThemePayload) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	return os.WriteFile(path, data, 0o644)
}

func cloneMap(source map[string]any) map[string]any {
	target := map[string]any{}
	for key, value := range source {
		if nested, ok := asMap(value); ok {
			target[key] = cloneMap(nested)
			continue
		}
		target[key] = value
	}
	return target
}

func mergeMap(target map[string]any, source map[string]any) {
	for key, value := range source {
		sourceNested, sourceIsMap := asMap(value)
		targetNested, targetIsMap := asMap(target[key])
		if sourceIsMap && targetIsMap {
			mergeMap(targetNested, sourceNested)
			target[key] = targetNested
			continue
		}
		if sourceIsMap {
			target[key] = cloneMap(sourceNested)
			continue
		}
		target[key] = value
	}
}

func asMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	default:
		return nil, false
	}
}
