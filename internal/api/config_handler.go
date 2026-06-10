package api

import (
	"net/http"

	"github.com/kentakom1213/albam/internal/config"
)

type ConfigResponse struct {
	EnableOriginalDownload bool                    `json:"enable_original_download"`
	MapEnabled             bool                    `json:"map_enabled"`
	ExposeGPS              bool                    `json:"expose_gps"`
	LocationPrecision      string                  `json:"location_precision"`
	Title                  string                  `json:"title"`
	Site                   config.ThemePayloadSite `json:"site"`
	Theme                  ThemeConfigResponse     `json:"theme"`
}

type ThemeConfigResponse struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"params"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	cfg, payload, err := s.currentThemePayload()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "config_load_failed", err.Error())
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, ConfigResponse{
		EnableOriginalDownload: cfg.Media.AllowOriginalDownload,
		MapEnabled:             cfg.PrivacyConfig.MapEnabled,
		ExposeGPS:              cfg.PrivacyConfig.ExposeGPS,
		LocationPrecision:      cfg.PrivacyConfig.LocationPrecision,
		Title:                  cfg.Title,
		Site:                   payload.Site,
		Theme: ThemeConfigResponse{
			Name:   payload.Theme.Name,
			Params: payload.Theme.Params,
		},
	})
}

func (s *Server) currentConfig() (config.Config, error) {
	if s.configPath == "" {
		return s.cfg, nil
	}

	return config.Load(s.configPath)
}

func (s *Server) currentThemePayload() (config.Config, config.ThemePayload, error) {
	cfg, err := s.currentConfig()
	if err != nil {
		return config.Config{}, config.ThemePayload{}, err
	}

	payload, err := config.BuildThemePayload(cfg)
	if err != nil {
		return config.Config{}, config.ThemePayload{}, err
	}

	return cfg, payload, nil
}
