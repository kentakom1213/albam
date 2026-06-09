package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/indexer"
	"github.com/kentakom1213/albam/internal/model"
	"github.com/kentakom1213/albam/internal/storage"
)

func TestHandleGetPhoto(t *testing.T) {
	store := openTestStorage(t)
	saveTestLibrary(t, store)

	albums, _, err := store.ListAlbums(10, 0)
	if err != nil {
		t.Fatal(err)
	}
	assets, _, err := store.ListAssetsByAlbumSlug(albums[0].Slug, 10, 0)
	if err != nil {
		t.Fatal(err)
	}

	server := NewServer(store, config.Config{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/photos/"+assets[0].Slug, nil)

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body PhotoResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Photo.ID != assets[0].Slug {
		t.Fatalf("photo id = %q, want %q", body.Photo.ID, assets[0].Slug)
	}
	if body.Photo.AlbumID != albums[0].Slug {
		t.Fatalf("album id = %q, want %q", body.Photo.AlbumID, albums[0].Slug)
	}
	if body.Photo.Links.Original != "/media/photos/"+assets[0].Slug+"/original" {
		t.Fatalf("original link = %q", body.Photo.Links.Original)
	}
	if body.Photo.Width == nil || *body.Photo.Width != 1600 {
		t.Fatalf("photo width = %v, want 1600", body.Photo.Width)
	}
	if body.Photo.Height == nil || *body.Photo.Height != 900 {
		t.Fatalf("photo height = %v, want 900", body.Photo.Height)
	}
	if body.Photo.AspectRatio == nil || *body.Photo.AspectRatio != float64(1600)/float64(900) {
		t.Fatalf("photo aspect ratio = %v, want %v", body.Photo.AspectRatio, float64(1600)/float64(900))
	}
}

func TestHandleGetPhotoNotFound(t *testing.T) {
	store := openTestStorage(t)

	server := NewServer(store, config.Config{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/photos/missing-photo", nil)

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

func openTestStorage(t *testing.T) *storage.Storage {
	t.Helper()

	store, err := storage.Open(filepath.Join(t.TempDir(), "db.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Fatal(err)
		}
	})

	if err := store.Migrate(); err != nil {
		t.Fatal(err)
	}

	return store
}

func saveTestLibrary(t *testing.T, store *storage.Storage) {
	t.Helper()

	if err := store.SaveLibrary(&indexer.Library{
		Albums: []model.Album{
			{Path: "weekend-trip", Title: "Weekend Trip"},
		},
		Assets: []model.Asset{
			{
				Path:     "weekend-trip/PXL_20260402_030405000.jpg",
				Filename: "PXL_20260402_030405000.jpg",
				Ext:      ".jpg",
				Size:     123,
				ModTime:  time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
				Width:    1600,
				Height:   900,
			},
			{
				Path:     "weekend-trip/PXL_20260502_030405000.jpg",
				Filename: "PXL_20260502_030405000.jpg",
				Ext:      ".jpg",
				Size:     456,
				ModTime:  time.Date(2026, 1, 3, 3, 4, 5, 0, time.UTC),
				Width:    900,
				Height:   1600,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}
}
