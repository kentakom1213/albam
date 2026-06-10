package api

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/kentakom1213/albam/internal/storage"
)

func (s *Server) handleListAlbumPhotos(w http.ResponseWriter, r *http.Request, albumID string) {
	limit := parseIntQuery(r, "limit", 100)
	offset := parseIntQuery(r, "offset", 0)
	sort := parsePhotoSortQuery(r)

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

	rows, total, err := s.store.ListAssetsByAlbumSlug(albumID, limit, offset, sort)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list photos")
		return
	}

	photos := make([]Photo, 0, len(rows))
	for _, row := range rows {
		photos = append(photos, photoFromRow(row, albumID, s.cfg.PrivacyConfig.ExposeGPS))
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

func parsePhotoSortQuery(r *http.Request) storage.AssetSort {
	switch r.URL.Query().Get("sort") {
	case string(storage.AssetSortTakenAtDesc):
		return storage.AssetSortTakenAtDesc
	default:
		return storage.AssetSortTakenAtAsc
	}
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
		Photo: photoFromRow(*row, row.AlbumSlug, s.cfg.PrivacyConfig.ExposeGPS),
	})
}

func photoFromRow(row storage.AssetRow, albumID string, exposeGPS bool) Photo {
	photoID := row.Slug
	width := intPtrFromNullInt(row.Width)
	height := intPtrFromNullInt(row.Height)
	gpsLatitude := floatPtrFromNullFloat(row.GPSLatitude)
	gpsLongitude := floatPtrFromNullFloat(row.GPSLongitude)
	if !exposeGPS {
		gpsLatitude = nil
		gpsLongitude = nil
	}

	return Photo{
		ID:                  photoID,
		AlbumID:             albumID,
		Filename:            row.Filename,
		Title:               nil,
		Description:         nil,
		TakenAt:             stringPtrFromNullString(row.TakenAt),
		Width:               width,
		Height:              height,
		AspectRatio:         aspectRatioPtr(width, height),
		GPSLatitude:         gpsLatitude,
		GPSLongitude:        gpsLongitude,
		CameraMake:          stringPtrFromNullString(row.CameraMake),
		CameraModel:         stringPtrFromNullString(row.CameraModel),
		LensMake:            stringPtrFromNullString(row.LensMake),
		LensModel:           stringPtrFromNullString(row.LensModel),
		FocalLengthMM:       floatPtrFromNullFloat(row.FocalLengthMM),
		FocalLength35mm:     intPtrFromNullInt(row.FocalLength35mm),
		ApertureFNumber:     floatPtrFromNullFloat(row.ApertureFNumber),
		ExposureTimeSeconds: floatPtrFromNullFloat(row.ExposureTimeSeconds),
		ISO:                 intPtrFromNullInt(row.ISO),
		Orientation:         intPtrFromNullInt(row.Orientation),
		Favorite:            false,
		Links: PhotoLinks{
			Self:     "/api/photos/" + photoID,
			Thumb:    "/media/photos/" + photoID + "/thumb",
			Preview:  "/media/photos/" + photoID + "/preview",
			Original: "/media/photos/" + photoID + "/original",
		},
	}
}

func intPtrFromNullInt(value sql.NullInt64) *int {
	if !value.Valid || value.Int64 <= 0 {
		return nil
	}

	intValue := int(value.Int64)
	return &intValue
}

func stringPtrFromNullString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

func floatPtrFromNullFloat(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
}

func aspectRatioPtr(width, height *int) *float64 {
	if width == nil || height == nil || *width <= 0 || *height <= 0 {
		return nil
	}

	aspectRatio := float64(*width) / float64(*height)
	return &aspectRatio
}
