package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kentakom1213/albam/internal/config"
)

func runBuild(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: albam build")
	}

	cfg, err := config.Load("albam.toml")
	if err != nil {
		return err
	}

	themeDir := cfg.Theme.Dir
	outDir := cfg.Build.OutDir

	if themeDir == "" {
		themeDir = "themes/default"
	}
	if outDir == "" {
		outDir = ".albam/public"
	}

	if err := ensureThemeReady(themeDir); err != nil {
		return err
	}

	if err := runPnpmBuild(themeDir); err != nil {
		return err
	}

	distDir := filepath.Join(themeDir, "dist")

	if err := replaceDir(outDir, distDir); err != nil {
		return err
	}

	fmt.Printf("built theme: %s -> %s\n", themeDir, outDir)
	return nil
}

func runPnpmBuild(themeDir string) error {
	cmd := exec.Command("pnpm", "build")
	cmd.Dir = themeDir
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
