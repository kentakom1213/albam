package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildThemePayloadMergesManifestDefaultsAndConfigParams(t *testing.T) {
	themeDir := t.TempDir()
	manifest := `name = "default"
display_name = "Default"
version = "0.1.0"

[defaults.appearance]
accent = "coral"

[defaults.layout]
photo_grid = "justified"
album_grid_columns = 4

[defaults.features]
show_header = true
show_footer = true
show_tags = true
show_album_count = true
`
	if err := os.WriteFile(filepath.Join(themeDir, "theme.toml"), []byte(manifest), 0o644); err != nil {
		t.Fatal(err)
	}

	payload, err := BuildThemePayload(Config{
		Title: "Test Albums",
		Theme: ThemeConfig{
			Name: "default",
			Dir:  themeDir,
			Params: map[string]any{
				"appearance": map[string]any{
					"accent": "mint",
				},
				"features": map[string]any{
					"show_footer": false,
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if payload.Site.Title != "Test Albums" {
		t.Fatalf("site.title = %q, want %q", payload.Site.Title, "Test Albums")
	}

	appearance := payload.Theme.Params["appearance"].(map[string]any)
	if appearance["accent"] != "mint" {
		t.Fatalf("appearance.accent = %v, want %q", appearance["accent"], "mint")
	}

	layout := payload.Theme.Params["layout"].(map[string]any)
	if layout["photo_grid"] != "justified" {
		t.Fatalf("layout.photo_grid = %v, want %q", layout["photo_grid"], "justified")
	}

	features := payload.Theme.Params["features"].(map[string]any)
	if features["show_header"] != true {
		t.Fatalf("features.show_header = %v, want true", features["show_header"])
	}
	if features["show_footer"] != false {
		t.Fatalf("features.show_footer = %v, want false", features["show_footer"])
	}
}
