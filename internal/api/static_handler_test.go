package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestStaticHandlerServesAlbumShellForAlbumDetailPath(t *testing.T) {
	root := t.TempDir()
	albumsDir := filepath.Join(root, "albums")
	if err := os.MkdirAll(albumsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(albumsDir, "index.html"), []byte("album shell"), 0o644); err != nil {
		t.Fatal(err)
	}

	handler := NewStaticHandler(root)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/albums/weekend-trip/", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != "album shell" {
		t.Fatalf("body = %q, want album shell", recorder.Body.String())
	}
}

func TestStaticHandlerDoesNotFallbackAPIPaths(t *testing.T) {
	root := t.TempDir()
	albumsDir := filepath.Join(root, "albums")
	if err := os.MkdirAll(albumsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(albumsDir, "index.html"), []byte("album shell"), 0o644); err != nil {
		t.Fatal(err)
	}

	handler := NewStaticHandler(root)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/albums/weekend-trip", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}
