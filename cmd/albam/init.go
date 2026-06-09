package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const releaseAPIBaseURL = "https://api.github.com/repos/kentakom1213/albam/releases"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type ThemeSpec struct {
	Name    string
	Version string
	URL     string
	SHA256  string
}

func runInit(args []string) error {
	fs := newFlagSet("init", "usage: albam init [--force] [--theme default] [--no-theme] [dir]")

	var force bool
	var noTheme bool
	var themeName string
	fs.BoolVar(&force, "force", false, "overwrite existing project files")
	fs.StringVar(&themeName, "theme", "default", "theme name")
	fs.BoolVar(&noTheme, "no-theme", false, "skip downloading themes/default")

	if err := parseFlags(fs, args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}
	if fs.NArg() > 1 {
		return fmt.Errorf("usage: albam init [--force] [--theme default] [--no-theme] [dir]")
	}
	if themeName != "default" {
		return fmt.Errorf("unsupported theme %q: only default is supported", themeName)
	}

	targetDir := "."
	if fs.NArg() == 1 {
		targetDir = fs.Arg(0)
	}

	if err := initProject(targetDir, initOptions{
		force:     force,
		noTheme:   noTheme,
		themeName: themeName,
	}); err != nil {
		return err
	}

	fmt.Printf("initialized albam project: %s\n", targetDir)
	return nil
}

type initOptions struct {
	force     bool
	noTheme   bool
	themeName string
}

func initProject(targetDir string, options initOptions) error {
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return err
	}

	if err := ensureInitTarget(targetDir, options.force); err != nil {
		return err
	}

	if err := writeDefaultConfig(filepath.Join(targetDir, "albam.toml"), options.force); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(targetDir, ".albam"), 0o755); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(targetDir, "albums/example"), 0o755); err != nil {
		return err
	}

	if err := writeDummyImage(filepath.Join(targetDir, "albums/example", "sample.png"), options.force); err != nil {
		return err
	}

	if !options.noTheme {
		if err := installTheme(options.themeName, filepath.Join(targetDir, "themes", options.themeName), options.force); err != nil {
			return fmt.Errorf("%w\n\nCreate a GitHub release for kentakom1213/albam, or run with --no-theme and add themes/default later.", err)
		}
	}

	return nil
}

func ensureInitTarget(targetDir string, force bool) error {
	paths := []string{
		filepath.Join(targetDir, "albam.toml"),
		filepath.Join(targetDir, ".albam"),
		filepath.Join(targetDir, "albums"),
		filepath.Join(targetDir, "themes", "default"),
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if force {
				continue
			}
			return fmt.Errorf("%s already exists: use --force to overwrite project files", path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}

	return nil
}

func writeDefaultConfig(path string, force bool) error {
	if err := ensureWritablePath(path, force); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(defaultConfigTOML), 0o644)
}

func writeDummyImage(path string, force bool) error {
	if err := ensureWritablePath(path, force); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	img := image.NewRGBA(image.Rect(0, 0, 640, 420))
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(255 - x/4),
				G: uint8(210 - y/6),
				B: uint8(180 + y/8),
				A: 255,
			})
		}
	}

	return png.Encode(file, img)
}

func ensureWritablePath(path string, force bool) error {
	if _, err := os.Stat(path); err == nil {
		if force {
			return nil
		}
		return fmt.Errorf("%s already exists: use --force to overwrite", path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return nil
}

func installTheme(themeName string, dst string, force bool) error {
	if _, err := os.Stat(dst); err == nil {
		if !force {
			return fmt.Errorf("%s already exists: use --force to overwrite", dst)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	release, err := fetchLatestRelease()
	if err != nil {
		return err
	}
	spec, err := themeSpecFromRelease(release, themeName)
	if err != nil {
		return err
	}

	return InstallTheme(spec, dst)
}

func fetchLatestRelease() (githubRelease, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest(http.MethodGet, releaseURLForVersion(version), nil)
	if err != nil {
		return githubRelease{}, err
	}
	request.Header.Set("Accept", "application/vnd.github+json")
	request.Header.Set("User-Agent", "albam")

	response, err := client.Do(request)
	if err != nil {
		return githubRelease{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("fetch latest release: %s", response.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(response.Body).Decode(&release); err != nil {
		return githubRelease{}, err
	}

	return release, nil
}

func releaseURLForVersion(value string) string {
	if value == "" || value == "dev" {
		return releaseAPIBaseURL + "/latest"
	}

	return releaseAPIBaseURL + "/tags/" + value
}

func themeSpecFromRelease(release githubRelease, themeName string) (ThemeSpec, error) {
	asset, ok := findThemeAsset(release.Assets, themeName, release.TagName)
	if !ok {
		return ThemeSpec{}, fmt.Errorf("release %s does not include albam-theme-%s_%s.tar.gz", release.TagName, themeName, release.TagName)
	}

	spec := ThemeSpec{
		Name:    themeName,
		Version: release.TagName,
		URL:     asset.BrowserDownloadURL,
	}

	if checksumAsset, ok := findAsset(release.Assets, "checksums.txt"); ok {
		checksums, err := downloadText(checksumAsset.BrowserDownloadURL)
		if err != nil {
			return ThemeSpec{}, err
		}
		spec.SHA256 = checksumForAsset(checksums, asset.Name)
	}

	return spec, nil
}

func findThemeAsset(assets []githubAsset, themeName string, version string) (githubAsset, bool) {
	exactName := fmt.Sprintf("albam-theme-%s_%s.tar.gz", themeName, version)
	if asset, ok := findAsset(assets, exactName); ok {
		return asset, true
	}

	prefix := fmt.Sprintf("albam-theme-%s_", themeName)
	for _, asset := range assets {
		if strings.HasPrefix(asset.Name, prefix) && strings.HasSuffix(asset.Name, ".tar.gz") {
			return asset, true
		}
	}

	return githubAsset{}, false
}

func findAsset(assets []githubAsset, name string) (githubAsset, bool) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, true
		}
	}

	return githubAsset{}, false
}

func downloadText(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", "albam")

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: %s", url, response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func checksumForAsset(checksums string, assetName string) string {
	for _, line := range strings.Split(checksums, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 2 && strings.TrimPrefix(fields[1], "*") == assetName && isSHA256(fields[0]) {
			return fields[0]
		}

		if strings.HasPrefix(line, "SHA256 ("+assetName+") = ") {
			value := strings.TrimPrefix(line, "SHA256 ("+assetName+") = ")
			if isSHA256(value) {
				return value
			}
		}
	}

	return ""
}

func isSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func InstallTheme(spec ThemeSpec, destDir string) error {
	if spec.Name == "" || spec.URL == "" {
		return fmt.Errorf("theme spec requires name and URL")
	}

	parentDir := filepath.Dir(destDir)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		return err
	}

	archiveFile, err := os.CreateTemp("", "albam-theme-*.tar.gz")
	if err != nil {
		return err
	}
	archivePath := archiveFile.Name()
	defer os.Remove(archivePath)

	if err := downloadThemeArchive(spec.URL, archiveFile); err != nil {
		archiveFile.Close()
		return err
	}
	if err := archiveFile.Close(); err != nil {
		return err
	}

	if spec.SHA256 != "" {
		if err := verifySHA256(archivePath, spec.SHA256); err != nil {
			return err
		}
	}

	tmpDir, err := os.MkdirTemp(parentDir, ".default-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	if err := extractThemeArchive(archivePath, tmpDir); err != nil {
		return err
	}

	if err := os.RemoveAll(destDir); err != nil {
		return err
	}
	if err := os.Rename(tmpDir, destDir); err != nil {
		return err
	}

	return nil
}

func downloadThemeArchive(url string, writer io.Writer) error {
	client := &http.Client{Timeout: 2 * time.Minute}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "albam")

	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("download theme archive: %s", response.Status)
	}

	_, err = io.Copy(writer, response.Body)
	return err
}

func verifySHA256(path string, expected string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actual := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actual, expected) {
		return fmt.Errorf("theme archive checksum mismatch: got %s, want %s", actual, expected)
	}

	return nil
}

func extractThemeArchive(archivePath string, dst string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	found := false
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		relativePath, ok := themeAssetRelativePath(header.Name)
		if !ok {
			continue
		}
		found = true

		targetPath, err := safeJoin(dst, relativePath)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
				return err
			}
			if err := writeTarFile(targetPath, tarReader, os.FileMode(header.Mode)); err != nil {
				return err
			}
		}
	}

	if !found {
		return fmt.Errorf("theme archive does not include theme files")
	}

	return nil
}

func themeAssetRelativePath(path string) (string, bool) {
	cleanPath := strings.Trim(filepath.ToSlash(filepath.Clean(path)), "/")
	if cleanPath == "." || cleanPath == "" {
		return "", false
	}

	parts := strings.Split(cleanPath, "/")
	for i := 0; i < len(parts)-1; i++ {
		if parts[i] == "themes" && parts[i+1] == "default" {
			if i+2 >= len(parts) {
				return "", false
			}
			return filepath.Join(parts[i+2:]...), true
		}
	}

	if len(parts) > 1 {
		return filepath.Join(parts[1:]...), true
	}

	return filepath.Join(parts...), true
}

func safeJoin(root string, relativePath string) (string, error) {
	targetPath := filepath.Join(root, relativePath)
	cleanRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}
	cleanTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return "", err
	}

	if cleanTarget != cleanRoot && !strings.HasPrefix(cleanTarget, cleanRoot+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid archive path: %s", relativePath)
	}

	return targetPath, nil
}

func writeTarFile(path string, reader io.Reader, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

const defaultConfigTOML = `title = "My Albums"

[server]
host = "127.0.0.1"
port = 8080

[media]
source_dir = "albums"
cache_dir = ".albam/cache"
allow_original_download = false

[privacy]
map_enabled = false
expose_gps = false
location_precision = "hidden"

[database]
path = ".albam/db.sqlite"

[build]
out_dir = ".albam/public"

[theme]
name = "default"
dir = "themes/default"
`
