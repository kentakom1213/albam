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
	sort := parseAlbumSortQuery(r)

	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	rows, total, err := s.store.ListAlbums(limit, offset, sort)
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

	breadcrumbRows, err := s.store.ListAlbumBreadcrumbsBySlug(albumID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", "failed to list album breadcrumbs")
		return
	}

	album := albumFromRow(*row)
	album.Breadcrumbs = breadcrumbsFromRows(breadcrumbRows)

	writeJSON(w, http.StatusOK, AlbumResponse{
		Album: album,
	})
}

func albumFromRow(row storage.AlbumRow) Album {
	var coverPhotoID *string
	var coverURL *string
	var date *string
	var latestMonth *string
	var oldestTakenAt *string
	var newestTakenAt *string

	if row.CoverPhotoID.Valid {
		id := row.CoverPhotoID.String
		coverPhotoID = &id

		url := "/media/photos/" + id + "/thumb"
		coverURL = &url
	}

	if row.LatestMonth.Valid {
		month := row.LatestMonth.String
		latestMonth = &month
	}

	if row.Date.Valid {
		value := row.Date.String
		date = &value
	}

	if row.OldestTakenAt.Valid {
		value := row.OldestTakenAt.String
		oldestTakenAt = &value
	}

	if row.NewestTakenAt.Valid {
		value := row.NewestTakenAt.String
		newestTakenAt = &value
	}

	return Album{
		ID:            row.Slug,
		Title:         row.Title,
		Description:   "",
		Date:          date,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
		PhotoCount:    row.PhotoCount,
		LatestMonth:   latestMonth,
		OldestTakenAt: oldestTakenAt,
		NewestTakenAt: newestTakenAt,
		CoverPhotoID:  coverPhotoID,
		Visibility:    "private",
		Breadcrumbs:   []Breadcrumb{},
		Links: AlbumLinks{
			Self:   "/api/albums/" + row.Slug,
			Photos: "/api/albums/" + row.Slug + "/photos",
			Cover:  coverURL,
		},
	}
}

func parseAlbumSortQuery(r *http.Request) storage.AlbumSort {
	switch r.URL.Query().Get("sort") {
	case string(storage.AlbumSortDateAsc):
		return storage.AlbumSortDateAsc
	default:
		return storage.AlbumSortDateDesc
	}
}

func breadcrumbsFromRows(rows []storage.AlbumBreadcrumbRow) []Breadcrumb {
	breadcrumbs := make([]Breadcrumb, 0, len(rows))

	for _, row := range rows {
		breadcrumbs = append(breadcrumbs, Breadcrumb{
			ID:    row.Slug,
			Title: row.Title,
			Path:  row.Path,
			Links: BreadcrumbLinks{
				Self: "/albums/" + row.Slug + "/",
			},
		})
	}

	return breadcrumbs
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
