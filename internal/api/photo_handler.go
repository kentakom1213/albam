package api

import (
	"net/http"
	"strings"

	"github.com/kentakom1213/albam/internal/storage"
)

func (s *Server) handleListAlbumPhotos(w http.ResponseWriter, r *http.Request, albumID string) {
	limit := parseIntQuery(r, "limit", 100)
	offset := parseIntQuery(r, "offset", 0)

	if limit <= 0 {
		limit = 100
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	album, err := s.store.GetAlbumBySlug(albumID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get album")
		return
	}
	if album == nil {
		writeError(w, http.StatusNotFound, "album_not_found", "album not found")
		return
	}

	rows, total, err := s.store.ListAssetsByAlbumSlug(albumID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list photos")
		return
	}

	photos := make([]Photo, 0, len(rows))
	for _, row := range rows {
		photos = append(photos, photoFromRow(row, albumID))
	}

	writeJSON(w, http.StatusOK, PhotosResponse{
		Photos: photos,
		Pagination: Pagination{
			Limit:   limit,
			Offset:  offset,
			Total:   total,
			HasNext: offset+limit < total,
		},
	})
}

func (s *Server) handlePhotoSubroutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}

	rest := strings.TrimPrefix(r.URL.Path, "/api/photos/")
	rest = strings.Trim(rest, "/")
	if rest == "" || strings.Contains(rest, "/") {
		writeError(w, http.StatusNotFound, "not_found", "not found")
		return
	}

	s.handleGetPhoto(w, r, rest)
}

func (s *Server) handleGetPhoto(w http.ResponseWriter, r *http.Request, photoID string) {
	row, err := s.store.GetAssetBySlug(photoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to get photo")
		return
	}
	if row == nil {
		writeError(w, http.StatusNotFound, "photo_not_found", "photo not found")
		return
	}

	writeJSON(w, http.StatusOK, PhotoResponse{
		Photo: photoFromRow(*row, row.AlbumSlug),
	})
}

func photoFromRow(row storage.AssetRow, albumID string) Photo {
	photoID := row.Slug

	return Photo{
		ID:          photoID,
		AlbumID:     albumID,
		Filename:    row.Filename,
		Title:       nil,
		Description: nil,
		TakenAt:     nil,
		Width:       nil,
		Height:      nil,
		AspectRatio: nil,
		Favorite:    false,
		Links: PhotoLinks{
			Self:     "/api/photos/" + photoID,
			Thumb:    "/media/photos/" + photoID + "/thumb",
			Preview:  "/media/photos/" + photoID + "/preview",
			Original: "/media/photos/" + photoID + "/original",
		},
	}
}
