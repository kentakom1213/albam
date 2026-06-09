package api

import "net/http"

type ConfigResponse struct {
	EnableOriginalDownload bool   `json:"enable_original_download"`
	MapEnabled             bool   `json:"map_enabled"`
	ExposeGPS              bool   `json:"expose_gps"`
	LocationPrecision      string `json:"location_precision"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, ConfigResponse{
		EnableOriginalDownload: s.cfg.Media.AllowOriginalDownload,
		MapEnabled:             s.cfg.PrivacyConfig.MapEnabled,
		ExposeGPS:              s.cfg.PrivacyConfig.ExposeGPS,
		LocationPrecision:      s.cfg.PrivacyConfig.LocationPrecision,
	})
}
