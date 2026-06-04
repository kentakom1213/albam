package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/kentakom1213/albam/internal/storage"
)

func (s *Server) handleListAlbums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	limit := parseIntQuery(r, "limit", 50)
	offset := parseIntQuery(r, "offset", 0)

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, total, err := s.store.ListAlbums(limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list albums")
		return
	}

	albums := make([]Album, 0, len(rows))
	for _, row := range rows {
		albums = append(albums, albumFromRow(row))
	}

	writeJSON(w, http.StatusOK, AlbumsResponse{
		Albums: albums,
		Pagination: Pagination{
			Limit:   limit,
			Offset:  offset,
			Total:   total,
			HasNext: offset+limit < total,
		},
	})
}

func (s *Server) handleAlbumSubroutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	rest := strings.TrimPrefix(r.URL.Path, "/api/albums/")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}

	parts := strings.Split(rest, "/")
	albumID := parts[0]

	switch {
	case len(parts) == 1:
		s.handleGetAlbum(w, r, albumID)
	case len(parts) == 2 && parts[1] == "photos":
		s.handleListAlbumPhotos(w, r, albumID)
	default:
		writeError(w, http.StatusNotFound, "not_found", "not found")
	}
}

func (s *Server) handleGetAlbum(w http.ResponseWriter, r *http.Request, albumID string) {
	row, err := s.store.GetAlbumBySlug(albumID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get album")
		return
	}
	if row == nil {
		writeError(w, http.StatusNotFound, "album_not_found", "album not found")
		return
	}

	writeJSON(w, http.StatusOK, AlbumResponse{
		Album: albumFromRow(*row),
	})
}

func albumFromRow(row storage.AlbumRow) Album {
	var coverPhotoID *string
	var coverURL *string

	if row.CoverPhotoID.Valid {
		id := row.CoverPhotoID.String
		coverPhotoID = &id

		url := "/media/photos/" + id + "/thumb"
		coverURL = &url
	}

	return Album{
		ID:           row.Slug,
		Title:        row.Title,
		Description:  "",
		Date:         nil,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
		PhotoCount:   row.PhotoCount,
		CoverPhotoID: coverPhotoID,
		Visibility:   "private",
		Tags:         []Tag{},
		Links: AlbumLinks{
			Self:   "/api/albums/" + row.Slug,
			Photos: "/api/albums/" + row.Slug + "/photos",
			Cover:  coverURL,
		},
	}
}

func parseIntQuery(r *http.Request, name string, fallback int) int {
	value := r.URL.Query().Get(name)
	if value == "" {
		return fallback
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return n
}
