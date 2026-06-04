package api

import "net/http"

func (s *Server) handleListTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, TagsResponse{
		Tags: []Tag{},
	})
}
