package api

import (
	"fmt"
	"net/http"
)

var accentColors = map[string]struct {
	color string
	soft  string
}{
	"pink":     {color: "#ff6fae", soft: "#ffe3ef"},
	"coral":    {color: "#ff6b5f", soft: "#ffe6e2"},
	"mint":     {color: "#35c99b", soft: "#dff8ef"},
	"blue":     {color: "#4da3ff", soft: "#e3f1ff"},
	"lavender": {color: "#9b7cff", soft: "#eee8ff"},
	"lemon":    {color: "#f4c430", soft: "#fff5c7"},
	"red":      {color: "#f04438", soft: "#ffe4e0"},
	"sakura":   {color: "#ff6fae", soft: "#ffe3ef"},
}

func (s *Server) handleRuntimeCSS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, payload, err := s.currentThemePayload()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	accent := stringParam(payload.Theme.Params, "appearance", "accent")
	colors, ok := accentColors[accent]
	if !ok {
		colors = accentColors["coral"]
	}

	w.Header().Set("Content-Type", "text/css; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	if r.Method == http.MethodHead {
		return
	}

	_, _ = fmt.Fprintf(w, ":root, body {\n  --theme-current-accent: %s;\n  --theme-current-accent-soft: %s;\n}\n", colors.color, colors.soft)
}

func stringParam(params map[string]any, section string, key string) string {
	rawSection, ok := params[section].(map[string]any)
	if !ok {
		return ""
	}

	value, ok := rawSection[key].(string)
	if !ok {
		return ""
	}

	return value
}
