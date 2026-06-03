package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Title string `toml:"title"`

	Server   ServerConfig   `toml:"server"`
	Media    MediaConfig    `toml:"media"`
	Database DatabaseConfig `toml:"database"`
	Build    BuildConfig    `toml:"build"`
	Theme    ThemeConfig    `toml:"theme"`
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

type DatabaseConfig struct {
	Path string `toml:"path"`
}

type BuildConfig struct {
	OutDir string `toml:"out_dir"`
}

type ThemeConfig struct {
	Name string `toml:"name"`
	Dir  string `toml:"dir"`
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
		Database: DatabaseConfig{
			Path: ".albam/db.sqlite",
		},
		Build: BuildConfig{
			OutDir: ".albam/public",
		},
		Theme: ThemeConfig{
			Dir: "theme/default",
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
