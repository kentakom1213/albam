package api

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type StaticHandler struct {
	root string
}

func NewStaticHandler(root string) *StaticHandler {
	return &StaticHandler{
		root: root,
	}
}

func EnsurePublicDir(publicDir string) error {
	info, err := os.Stat(publicDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf(`public directory does not exist: %s

Run:

  albam build`, publicDir)
		}

		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("public path is not a directory: %s", publicDir)
	}

	return nil
}

func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlPath := path.Clean("/" + r.URL.Path)

	if urlPath == "/api" || strings.HasPrefix(urlPath, "/api/") {
		http.NotFound(w, r)
		return
	}
	if urlPath == "/media" || strings.HasPrefix(urlPath, "/media/") {
		http.NotFound(w, r)
		return
	}

	rel := strings.TrimPrefix(urlPath, "/")
	if rel == "" || rel == "." {
		rel = "index.html"
	}

	filePath, err := resolveStaticPath(h.root, rel)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if info.IsDir() {
		filePath = filepath.Join(filePath, "index.html")

		info, err = os.Stat(filePath)
		if err != nil || info.IsDir() {
			http.NotFound(w, r)
			return
		}
	}

	setStaticCacheHeader(w, urlPath)
	http.ServeFile(w, r, filePath)
}

func resolveStaticPath(root string, rel string) (string, error) {
	if root == "" {
		return "", fmt.Errorf("static root is empty")
	}
	if rel == "" {
		return "", fmt.Errorf("static path is empty")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute static path is not allowed: %s", rel)
	}

	cleanRel := filepath.Clean(filepath.FromSlash(rel))
	if cleanRel == "." || cleanRel == ".." || strings.HasPrefix(cleanRel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("static path escapes root: %s", rel)
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(filepath.Join(absRoot, cleanRel))
	if err != nil {
		return "", err
	}

	checkRel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", err
	}

	if checkRel == ".." || strings.HasPrefix(checkRel, ".."+string(filepath.Separator)) || filepath.IsAbs(checkRel) {
		return "", fmt.Errorf("static path escapes root: %s", rel)
	}

	return absPath, nil
}

func setStaticCacheHeader(w http.ResponseWriter, urlPath string) {
	if strings.HasPrefix(urlPath, "/_astro/") || strings.HasPrefix(urlPath, "/assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}

	w.Header().Set("Cache-Control", "no-cache")
}
