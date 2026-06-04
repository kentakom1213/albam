package api

import (
	"net/http"

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
			Original: "/media/" + photoID + "/original",
		},
	}
}
