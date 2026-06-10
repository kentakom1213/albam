package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kentakom1213/albam/internal/config"
)

func TestHandleGetConfig(t *testing.T) {
	themeDir := writeTestThemeManifest(t)
	server := NewServer(nil, config.Config{
		Title: "Test Albums",
		Media: config.MediaConfig{
			AllowOriginalDownload: true,
		},
		Theme: config.ThemeConfig{
			Name: "default",
			Dir:  themeDir,
			Params: map[string]any{
				"appearance": map[string]any{
					"accent": "mint",
				},
			},
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
	if body.Site.Title != "Test Albums" {
		t.Fatalf("site.title = %q, want %q", body.Site.Title, "Test Albums")
	}
	if body.Theme.Name != "default" {
		t.Fatalf("theme.name = %q, want %q", body.Theme.Name, "default")
	}
	appearance, ok := body.Theme.Params["appearance"].(map[string]any)
	if !ok {
		t.Fatalf("theme.params.appearance = %T, want map", body.Theme.Params["appearance"])
	}
	if appearance["accent"] != "mint" {
		t.Fatalf("theme.params.appearance.accent = %v, want %q", appearance["accent"], "mint")
	}
	features, ok := body.Theme.Params["features"].(map[string]any)
	if !ok {
		t.Fatalf("theme.params.features = %T, want map", body.Theme.Params["features"])
	}
	if features["show_header"] != true {
		t.Fatalf("theme.params.features.show_header = %v, want true", features["show_header"])
	}
}

func TestHandleGetConfigReloadsConfigFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "albam.toml")
	themeDir := writeTestThemeManifest(t)
	writeConfig := func(title string, accent string) {
		t.Helper()
		source := `title = "` + title + `"

[theme]
name = "default"
dir = "` + filepath.ToSlash(themeDir) + `"

[theme.params.appearance]
accent = "` + accent + `"
`
		if err := os.WriteFile(configPath, []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	requestConfig := func(server *Server) ConfigResponse {
		t.Helper()
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
		return body
	}

	writeConfig("First Albums", "coral")
	server := NewServerWithConfigPath(nil, config.Default(), configPath)

	first := requestConfig(server)
	firstAppearance := first.Theme.Params["appearance"].(map[string]any)
	if first.Site.Title != "First Albums" || firstAppearance["accent"] != "coral" {
		t.Fatalf("first config = %+v, want title %q and accent %q", first, "First Albums", "coral")
	}

	writeConfig("Second Albums", "mint")

	second := requestConfig(server)
	secondAppearance := second.Theme.Params["appearance"].(map[string]any)
	if second.Site.Title != "Second Albums" || secondAppearance["accent"] != "mint" {
		t.Fatalf("second config = %+v, want title %q and accent %q", second, "Second Albums", "mint")
	}
}

func writeTestThemeManifest(t *testing.T) string {
	t.Helper()

	themeDir := t.TempDir()
	source := `name = "default"
display_name = "Default"
version = "0.1.0"

[defaults.appearance]
accent = "coral"

[defaults.features]
show_header = true
show_footer = true
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.toml"), []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}

	return themeDir
}
