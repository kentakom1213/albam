package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kentakom1213/albam/internal/config"
	"github.com/kentakom1213/albam/internal/indexer"
	"github.com/kentakom1213/albam/internal/model"
)

func TestHandleListAlbumsIncludesLatestMonth(t *testing.T) {
	store := openTestStorage(t)
	saveTestLibrary(t, store)

	server := NewServer(store, config.Config{})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/albums", nil)

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body AlbumsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Albums) != 1 {
		t.Fatalf("album count = %d, want 1", len(body.Albums))
	}
	if body.Albums[0].LatestMonth == nil || *body.Albums[0].LatestMonth != "2026/05" {
		t.Fatalf("latest month = %v, want 2026/05", body.Albums[0].LatestMonth)
	}
	if body.Albums[0].Date == nil || *body.Albums[0].Date != "2026-05-02" {
		t.Fatalf("date = %v, want 2026-05-02", body.Albums[0].Date)
	}
	if body.Albums[0].OldestTakenAt == nil || *body.Albums[0].OldestTakenAt != "2026-04-02T12:34:56Z" {
		t.Fatalf("oldest taken_at = %v, want 2026-04-02T12:34:56Z", body.Albums[0].OldestTakenAt)
	}
	if body.Albums[0].NewestTakenAt == nil || *body.Albums[0].NewestTakenAt != "2026-05-02T12:34:56Z" {
		t.Fatalf("newest taken_at = %v, want 2026-05-02T12:34:56Z", body.Albums[0].NewestTakenAt)
	}
}

func TestHandleListAlbumsSortsByTakenAt(t *testing.T) {
	store := openTestStorage(t)
	olderTakenAt := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	newerTakenAt := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)

	if err := store.SaveLibrary(&indexer.Library{
		Albums: []model.Album{
			{Path: "older", Title: "Older"},
			{Path: "newer", Title: "Newer"},
		},
		Assets: []model.Asset{
			{
				Path:     "older/old.jpg",
				Filename: "old.jpg",
				Ext:      ".jpg",
				Size:     123,
				ModTime:  olderTakenAt,
				Width:    100,
				Height:   100,
				TakenAt:  &olderTakenAt,
			},
			{
				Path:     "newer/new.jpg",
				Filename: "new.jpg",
				Ext:      ".jpg",
				Size:     123,
				ModTime:  newerTakenAt,
				Width:    100,
				Height:   100,
				TakenAt:  &newerTakenAt,
			},
		},
	}); err != nil {
		t.Fatal(err)
	}

	server := NewServer(store, config.Config{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/albums?sort=date_desc", nil)
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var descBody AlbumsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &descBody); err != nil {
		t.Fatal(err)
	}
	if got := descBody.Albums[0].Title; got != "Newer" {
		t.Fatalf("newest album = %q, want Newer", got)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/albums?sort=date_asc", nil)
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var ascBody AlbumsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &ascBody); err != nil {
		t.Fatal(err)
	}
	if got := ascBody.Albums[0].Title; got != "Older" {
		t.Fatalf("oldest album = %q, want Older", got)
	}

	recorder = httptest.NewRecorder()
	request = httptest.NewRequest(http.MethodGet, "/api/albums", nil)
	server.Routes().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var defaultBody AlbumsResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &defaultBody); err != nil {
		t.Fatal(err)
	}
	if got := defaultBody.Albums[0].Title; got != "Older" {
		t.Fatalf("default album = %q, want Older", got)
	}
}
