package api

import "net/http"

type ConfigResponse struct {
	EnableOriginalDownload bool `json:"enable_original_download"`
}

func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, ConfigResponse{
		EnableOriginalDownload: s.cfg.Media.AllowOriginalDownload,
	})
}
