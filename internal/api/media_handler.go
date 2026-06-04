package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/kentakom1213/albam/internal/imageproc"
)

type PhotoMediaStore interface {
	GetPhotoForMediaByID(ctx context.Context, slug string) (string, error)
}

var ErrOriginalDownloadDisabled = errors.New("original download is disabled")

type MediaHandler struct {
	Store                 PhotoMediaStore
	MediaRoot             string
	CacheRoot             string
	AllowOriginalDownload bool
}

type VariantKind string

const (
	VariantThumb    VariantKind = "thumb"
	VariantPreview  VariantKind = "preview"
	VariantOriginal VariantKind = "original"
)

func NewMediaHandler(store PhotoMediaStore, mediaRoot, cacheRoot string, allowOriginalDownload bool) *MediaHandler {
	return &MediaHandler{
		Store:                 store,
		MediaRoot:             mediaRoot,
		CacheRoot:             cacheRoot,
		AllowOriginalDownload: allowOriginalDownload,
	}
}

func (h *MediaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	photoID, kind, ok := parseMediaPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}

	if kind == VariantOriginal {
		if err := h.serveOriginal(w, r, photoID); err != nil {
			status := http.StatusInternalServerError
			code := "media_error"

			switch {
			case errors.Is(err, ErrOriginalDownloadDisabled):
				status = http.StatusForbidden
				code = "original_download_disabled"
			case errors.Is(err, sql.ErrNoRows):
				status = http.StatusNotFound
				code = "photo_not_found"
			}

			writeError(w, status, code, err.Error())
			return
		}
		return
	}

	if kind != VariantThumb && kind != VariantPreview {
		writeError(w, http.StatusNotFound, "media_variant_not_found", "media variant not found")
		return
	}

	if err := h.serveVariant(w, r, photoID, kind); err != nil {
		status := http.StatusInternalServerError
		code := "media_error"

		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
			code = "photo_not_found"
		}

		writeError(w, status, code, err.Error())
		return
	}
}

func parseMediaPath(path string) (string, VariantKind, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) == 3 {
		if parts[0] != "media" {
			return "", "", false
		}
		if parts[1] == "" || parts[2] == "" {
			return "", "", false
		}

		return parts[1], VariantKind(parts[2]), true
	}

	if len(parts) != 4 {
		return "", "", false
	}
	if parts[0] != "media" || parts[1] != "photos" {
		return "", "", false
	}
	if parts[2] == "" || parts[3] == "" {
		return "", "", false
	}

	return parts[2], VariantKind(parts[3]), true
}

func (h *MediaHandler) serveVariant(w http.ResponseWriter, r *http.Request, photoID string, kind VariantKind) error {
	if h.Store == nil {
		return fmt.Errorf("photo media store is nil")
	}
	if !SafeCacheID(photoID) {
		return fmt.Errorf("invalid photo id")
	}

	photoPath, err := h.Store.GetPhotoForMediaByID(r.Context(), photoID)
	if err != nil {
		return err
	}

	srcPath, err := ResolveUnderRoot(h.MediaRoot, photoPath)
	if err != nil {
		return err
	}

	cacheRel := filepath.Join("photos", string(kind), photoID+".jpg")
	cachePath, err := ResolveUnderRoot(h.CacheRoot, cacheRel)
	if err != nil {
		return err
	}

	opts := optionsForVariant(kind)

	if _, err := imageproc.EnsureJPEGVariant(srcPath, cachePath, opts); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	http.ServeFile(w, r, cachePath)
	return nil
}

func (h *MediaHandler) serveOriginal(w http.ResponseWriter, r *http.Request, photoID string) error {
	if !h.AllowOriginalDownload {
		return ErrOriginalDownloadDisabled
	}

	if h.Store == nil {
		return fmt.Errorf("photo media store is nil")
	}
	if !SafeCacheID(photoID) {
		return fmt.Errorf("invalid photo id")
	}

	photoPath, err := h.Store.GetPhotoForMediaByID(r.Context(), photoID)
	if err != nil {
		return err
	}

	srcPath, err := ResolveUnderRoot(h.MediaRoot, photoPath)
	if err != nil {
		return err
	}

	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFile(w, r, srcPath)
	return nil
}

func optionsForVariant(kind VariantKind) imageproc.Options {
	switch kind {
	case VariantThumb:
		return imageproc.Options{
			MaxWidth:  512,
			MaxHeight: 512,
			Quality:   82,
		}
	case VariantPreview:
		return imageproc.Options{
			MaxWidth:  1600,
			MaxHeight: 1600,
			Quality:   86,
		}
	default:
		return imageproc.Options{
			MaxWidth:  512,
			MaxHeight: 512,
			Quality:   82,
		}
	}
}
