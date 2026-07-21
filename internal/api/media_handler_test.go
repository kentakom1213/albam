package api

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/webp"
)

type stubPhotoMediaStore struct {
	path string
}

func (s stubPhotoMediaStore) GetPhotoForMediaByID(context.Context, string) (string, error) {
	return s.path, nil
}

func TestParseMediaPath(t *testing.T) {
	photoID, kind, ok := parseMediaPath("/media/photo-001/original")
	if !ok {
		t.Fatal("parseMediaPath did not match")
	}
	if photoID != "photo-001" {
		t.Fatalf("photoID = %q, want photo-001", photoID)
	}
	if kind != VariantOriginal {
		t.Fatalf("kind = %q, want %q", kind, VariantOriginal)
	}
}

func TestParseMediaPathRejectsLegacyPhotoPath(t *testing.T) {
	if _, _, ok := parseMediaPath("/media/photos/photo-001/original"); ok {
		t.Fatal("parseMediaPath matched legacy photo path")
	}
}

func TestMediaHandlerServesWebPVariant(t *testing.T) {
	mediaRoot := t.TempDir()
	cacheRoot := t.TempDir()
	sourceName := "source.png"
	sourcePath := filepath.Join(mediaRoot, sourceName)

	source, err := os.Create(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 32, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}
	if err := png.Encode(source, img); err != nil {
		_ = source.Close()
		t.Fatal(err)
	}
	if err := source.Close(); err != nil {
		t.Fatal(err)
	}

	handler := NewMediaHandler(stubPhotoMediaStore{path: sourceName}, mediaRoot, cacheRoot, false)
	request := httptest.NewRequest(http.MethodGet, "/media/photo-001/thumb", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if got := response.Header().Get("Content-Type"); got != "image/webp" {
		t.Fatalf("Content-Type = %q, want image/webp", got)
	}
	if _, err := webp.DecodeConfig(response.Body); err != nil {
		t.Fatalf("response is not valid webp: %v", err)
	}
	if _, err := os.Stat(filepath.Join(cacheRoot, "media", mediaVariantCacheVersion, "thumb", "photo-001.webp")); err != nil {
		t.Fatalf("stat cached webp: %v", err)
	}
}
