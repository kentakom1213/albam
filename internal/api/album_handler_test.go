package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kentakom1213/albam/internal/config"
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
}
