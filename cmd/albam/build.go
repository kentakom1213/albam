package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kentakom1213/albam/internal/config"
)

func runBuild(args []string) error {
	fs := newFlagSet("build", "usage: albam build [--config path]")

	var configPath string
	fs.StringVar(&configPath, "config", "albam.toml", "config file path")

	if err := parseFlags(fs, args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: albam build [--config path]")
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	themeDir := config.ResolveThemeDir(cfg)
	outDir := cfg.Build.OutDir

	if outDir == "" {
		outDir = ".albam/public"
	}

	if err := ensureThemeReady(themeDir); err != nil {
		return err
	}

	payload, err := config.BuildThemePayload(cfg)
	if err != nil {
		return err
	}

	themeConfigFile := filepath.Join(".albam", "generated", "theme.json")
	if err := config.WriteThemePayload(themeConfigFile, payload); err != nil {
		return err
	}

	themeConfigFile, err = filepath.Abs(themeConfigFile)
	if err != nil {
		return err
	}

	if err := runPnpmBuild(themeDir, themeConfigFile); err != nil {
		return err
	}

	distDir := filepath.Join(themeDir, "dist")

	if err := replaceDir(outDir, distDir); err != nil {
		return err
	}

	fmt.Printf("built theme: %s -> %s\n", themeDir, outDir)
	return nil
}

func runPnpmBuild(themeDir string, themeConfigFile string) error {
	cmd := exec.Command("pnpm", "build")
	cmd.Dir = themeDir
	cmd.Env = append(os.Environ(), "ALBAM_THEME_CONFIG_FILE="+themeConfigFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pnpm build failed: %w", err)
	}

	return nil
}

func ensureThemeReady(themeDir string) error {
	packageJSON := filepath.Join(themeDir, "package.json")
	if _, err := os.Stat(packageJSON); err != nil {
		return fmt.Errorf("theme package.json not found: %s", packageJSON)
	}

	themeManifest := filepath.Join(themeDir, "theme.toml")
	if _, err := os.Stat(themeManifest); err != nil {
		return fmt.Errorf("theme manifest not found: %s", themeManifest)
	}

	nodeModules := filepath.Join(themeDir, "node_modules")
	if _, err := os.Stat(nodeModules); err != nil {
		return fmt.Errorf(`theme dependencies are not installed.

Run:

  cd %s
  pnpm install`, themeDir)
	}

	return nil
}

func replaceDir(dst string, src string) error {
	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}

	return copyDir(src, dst)
}

func copyDir(src string, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				return err
			}
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if err := copyFile(srcPath, dstPath, info.Mode()); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src string, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}
