package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kentakom1213/albam/internal/config"
)

func TestHandleGetConfig(t *testing.T) {
	server := NewServer(nil, config.Config{
		Media: config.MediaConfig{
			AllowOriginalDownload: true,
		},
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/config", nil)

	server.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body ConfigResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if !body.EnableOriginalDownload {
		t.Fatal("enable_original_download = false, want true")
	}
}
