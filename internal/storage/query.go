package storage

import (
	"database/sql"
	"errors"
	"strings"
)

type AlbumRow struct {
	ID            int64
	Path          string
	Slug          string
	Title         string
	Date          sql.NullString
	CreatedAt     string
	UpdatedAt     string
	PhotoCount    int
	CoverPhotoID  sql.NullString
	LatestMonth   sql.NullString
	OldestTakenAt sql.NullString
	NewestTakenAt sql.NullString
}

type AssetRow struct {
	ID        int64
	Slug      string
	AlbumID   int64
	AlbumSlug string
	Path      string
	Filename  string
	Ext       string
	Size      int64
	ModTime   string
	Width     sql.NullInt64
	Height    sql.NullInt64
	TakenAt   sql.NullString

	GPSLatitude         sql.NullFloat64
	GPSLongitude        sql.NullFloat64
	CameraMake          sql.NullString
	CameraModel         sql.NullString
	LensMake            sql.NullString
	LensModel           sql.NullString
	FocalLengthMM       sql.NullFloat64
	FocalLength35mm     sql.NullInt64
	ApertureFNumber     sql.NullFloat64
	ExposureTimeSeconds sql.NullFloat64
	ISO                 sql.NullInt64
	Orientation         sql.NullInt64

	CreatedAt string
	UpdatedAt string
}

type AssetSort string
type AlbumSort string

const (
	AssetSortTakenAtDesc AssetSort = "taken_at_desc"
	AssetSortTakenAtAsc  AssetSort = "taken_at_asc"

	AlbumSortDateDesc AlbumSort = "date_desc"
	AlbumSortDateAsc  AlbumSort = "date_asc"
)

func (s *Storage) ListAlbums(limit, offset int, sort AlbumSort) ([]AlbumRow, int, error) {
	total, err := s.countAlbums()
	if err != nil {
		return nil, 0, err
	}

	orderBy := albumOrderBy(sort)
	rows, err := s.db.Query(`
SELECT
	albums.id,
	albums.path,
	albums.slug,
	albums.title,
	albums.created_at,
	albums.updated_at,
	COUNT(assets.id) AS photo_count,
	(
		SELECT a.slug
		FROM assets AS a
		WHERE a.album_id = albums.id
		ORDER BY a.path
		LIMIT 1
	) AS cover_photo_id
FROM albums
LEFT JOIN assets ON assets.album_id = albums.id
GROUP BY albums.id
HAVING COUNT(assets.id) > 0
`+orderBy+`
LIMIT ? OFFSET ?
`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	albums := make([]AlbumRow, 0)
	for rows.Next() {
		var album AlbumRow
		if err := rows.Scan(
			&album.ID,
			&album.Path,
			&album.Slug,
			&album.Title,
			&album.CreatedAt,
			&album.UpdatedAt,
			&album.PhotoCount,
			&album.CoverPhotoID,
		); err != nil {
			return nil, 0, err
		}

		albums = append(albums, album)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	for i := range albums {
		dateRange, err := s.GetAssetDateRangeByAlbumSlug(albums[i].Slug)
		if err != nil {
			return nil, 0, err
		}
		applyAlbumDateRange(&albums[i], dateRange)
	}

	return albums, total, nil
}

func (s *Storage) GetAlbumBySlug(slug string) (*AlbumRow, error) {
	var album AlbumRow

	err := s.db.QueryRow(`
SELECT
    albums.id,
    albums.path,
    albums.slug,
    albums.title,
    albums.created_at,
    albums.updated_at,
    COUNT(assets.id) AS photo_count,
	(
		SELECT a.slug
		FROM assets AS a
		WHERE a.album_id = albums.id
		ORDER BY a.path
		LIMIT 1
	) AS cover_photo_id
FROM albums
LEFT JOIN assets ON assets.album_id = albums.id
WHERE albums.slug = ?
GROUP BY albums.id
`, slug).Scan(
		&album.ID,
		&album.Path,
		&album.Slug,
		&album.Title,
		&album.CreatedAt,
		&album.UpdatedAt,
		&album.PhotoCount,
		&album.CoverPhotoID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	dateRange, err := s.GetAssetDateRangeByAlbumSlug(album.Slug)
	if err != nil {
		return nil, err
	}
	applyAlbumDateRange(&album, dateRange)

	return &album, nil
}

type AssetDateRange struct {
	Oldest sql.NullString
	Newest sql.NullString
}

func (s *Storage) GetAssetDateRangeByAlbumSlug(slug string) (AssetDateRange, error) {
	var dateRange AssetDateRange

	err := s.db.QueryRow(`
SELECT
	MIN(taken_at),
	MAX(taken_at)
FROM albums AS root
JOIN albums AS child
	ON child.path = root.path
	OR (
		child.path COLLATE BINARY >= root.path || '/'
		AND child.path COLLATE BINARY < root.path || '0'
	)
JOIN assets ON assets.album_id = child.id
WHERE root.slug = ?
	AND assets.taken_at IS NOT NULL
`, slug).Scan(&dateRange.Oldest, &dateRange.Newest)
	if err != nil {
		return AssetDateRange{}, err
	}

	return dateRange, nil
}

func applyAlbumDateRange(album *AlbumRow, dateRange AssetDateRange) {
	album.OldestTakenAt = dateRange.Oldest
	album.NewestTakenAt = dateRange.Newest

	if dateRange.Newest.Valid {
		album.Date = dateFromTimestamp(dateRange.Newest.String)
		album.LatestMonth = monthFromTimestamp(dateRange.Newest.String)
	}
}

func dateFromTimestamp(value string) sql.NullString {
	if len(value) < 10 {
		return sql.NullString{}
	}

	return sql.NullString{String: value[:10], Valid: true}
}

func monthFromTimestamp(value string) sql.NullString {
	if len(value) < 7 {
		return sql.NullString{}
	}

	return sql.NullString{String: strings.ReplaceAll(value[:7], "-", "/"), Valid: true}
}

func albumOrderBy(sort AlbumSort) string {
	switch sort {
	case AlbumSortDateAsc:
		return `ORDER BY COALESCE((
	SELECT MIN(a.taken_at)
	FROM albums AS child
	JOIN assets AS a ON a.album_id = child.id
	WHERE (
		child.path = albums.path
		OR (
			child.path COLLATE BINARY >= albums.path || '/'
			AND child.path COLLATE BINARY < albums.path || '0'
		)
	)
	AND a.taken_at IS NOT NULL
), albums.updated_at) ASC, albums.path ASC`
	default:
		return `ORDER BY COALESCE((
	SELECT MAX(a.taken_at)
	FROM albums AS child
	JOIN assets AS a ON a.album_id = child.id
	WHERE (
		child.path = albums.path
		OR (
			child.path COLLATE BINARY >= albums.path || '/'
			AND child.path COLLATE BINARY < albums.path || '0'
		)
	)
	AND a.taken_at IS NOT NULL
), albums.updated_at) DESC, albums.path DESC`
	}
}

func assetMonthFromPath(assetPath string) string {
	parts := strings.Split(assetPath, "/")
	for i := 0; i < len(parts)-1; i++ {
		if isYearPart(parts[i]) && hasMonthPrefix(parts[i+1]) {
			return parts[i] + "/" + parts[i+1][:2]
		}
	}

	for _, part := range parts {
		if len(part) < 10 {
			continue
		}

		for i := 0; i <= len(part)-10; i++ {
			if part[i:i+4] == "PXL_" && isYearPart(part[i+4:i+8]) && isMonth(part[i+8:i+10]) {
				return part[i+4:i+8] + "/" + part[i+8:i+10]
			}
		}
	}

	return ""
}

func isYearPart(value string) bool {
	if len(value) != 4 {
		return false
	}

	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}

func hasMonthPrefix(value string) bool {
	if len(value) < 2 {
		return false
	}

	return isMonth(value[:2])
}

func isMonth(value string) bool {
	return value >= "01" && value <= "12"
}

func (s *Storage) GetAssetByID(id int64) (*AssetRow, error) {
	var asset AssetRow

	err := s.db.QueryRow(`
SELECT
    id,
	slug,
    album_id,
    path,
    filename,
    ext,
    size_bytes,
    file_mtime,
    width,
    height,
	taken_at,
	gps_latitude,
	gps_longitude,
	camera_make,
	camera_model,
	lens_make,
	lens_model,
	focal_length_mm,
	focal_length_35mm,
	aperture_f_number,
	exposure_time_seconds,
	iso,
	orientation,
    created_at,
    updated_at
FROM assets
WHERE id = ?
`, id).Scan(
		&asset.ID,
		&asset.Slug,
		&asset.AlbumID,
		&asset.Path,
		&asset.Filename,
		&asset.Ext,
		&asset.Size,
		&asset.ModTime,
		&asset.Width,
		&asset.Height,
		&asset.TakenAt,
		&asset.GPSLatitude,
		&asset.GPSLongitude,
		&asset.CameraMake,
		&asset.CameraModel,
		&asset.LensMake,
		&asset.LensModel,
		&asset.FocalLengthMM,
		&asset.FocalLength35mm,
		&asset.ApertureFNumber,
		&asset.ExposureTimeSeconds,
		&asset.ISO,
		&asset.Orientation,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &asset, nil
}

func (s *Storage) GetAssetBySlug(slug string) (*AssetRow, error) {
	var asset AssetRow

	err := s.db.QueryRow(`
SELECT
    assets.id,
	assets.slug,
    assets.album_id,
	albums.slug,
    assets.path,
    assets.filename,
    assets.ext,
    assets.size_bytes,
    assets.file_mtime,
    assets.width,
    assets.height,
	assets.taken_at,
	assets.gps_latitude,
	assets.gps_longitude,
	assets.camera_make,
	assets.camera_model,
	assets.lens_make,
	assets.lens_model,
	assets.focal_length_mm,
	assets.focal_length_35mm,
	assets.aperture_f_number,
	assets.exposure_time_seconds,
	assets.iso,
	assets.orientation,
    assets.created_at,
    assets.updated_at
FROM assets
JOIN albums ON albums.id = assets.album_id
WHERE assets.slug = ?
`, slug).Scan(
		&asset.ID,
		&asset.Slug,
		&asset.AlbumID,
		&asset.AlbumSlug,
		&asset.Path,
		&asset.Filename,
		&asset.Ext,
		&asset.Size,
		&asset.ModTime,
		&asset.Width,
		&asset.Height,
		&asset.TakenAt,
		&asset.GPSLatitude,
		&asset.GPSLongitude,
		&asset.CameraMake,
		&asset.CameraModel,
		&asset.LensMake,
		&asset.LensModel,
		&asset.FocalLengthMM,
		&asset.FocalLength35mm,
		&asset.ApertureFNumber,
		&asset.ExposureTimeSeconds,
		&asset.ISO,
		&asset.Orientation,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &asset, nil
}

func (s *Storage) ListAssetsByAlbumSlug(slug string, limit, offset int, sort AssetSort) ([]AssetRow, int, error) {
	total, err := s.countAssetsByAlbumSlug(slug)
	if err != nil {
		return nil, 0, err
	}

	orderBy := assetOrderBy(sort)
	rows, err := s.db.Query(`
SELECT
    assets.id,
	assets.slug,
    assets.album_id,
    assets.path,
    assets.filename,
    assets.ext,
    assets.size_bytes,
    assets.file_mtime,
    assets.width,
    assets.height,
	assets.taken_at,
	assets.gps_latitude,
	assets.gps_longitude,
	assets.camera_make,
	assets.camera_model,
	assets.lens_make,
	assets.lens_model,
	assets.focal_length_mm,
	assets.focal_length_35mm,
	assets.aperture_f_number,
	assets.exposure_time_seconds,
	assets.iso,
	assets.orientation,
    assets.created_at,
    assets.updated_at
FROM albums AS root
JOIN albums AS child
	ON child.path = root.path
	OR (
		child.path COLLATE BINARY >= root.path || '/'
		AND child.path COLLATE BINARY < root.path || '0'
	)
JOIN assets ON assets.album_id = child.id
WHERE root.slug = ?
`+orderBy+`
LIMIT ? OFFSET ?
`, slug, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	assets := make([]AssetRow, 0)
	for rows.Next() {
		var asset AssetRow
		if err := rows.Scan(
			&asset.ID,
			&asset.Slug,
			&asset.AlbumID,
			&asset.Path,
			&asset.Filename,
			&asset.Ext,
			&asset.Size,
			&asset.ModTime,
			&asset.Width,
			&asset.Height,
			&asset.TakenAt,
			&asset.GPSLatitude,
			&asset.GPSLongitude,
			&asset.CameraMake,
			&asset.CameraModel,
			&asset.LensMake,
			&asset.LensModel,
			&asset.FocalLengthMM,
			&asset.FocalLength35mm,
			&asset.ApertureFNumber,
			&asset.ExposureTimeSeconds,
			&asset.ISO,
			&asset.Orientation,
			&asset.CreatedAt,
			&asset.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}

		assets = append(assets, asset)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return assets, total, nil
}

func assetOrderBy(sort AssetSort) string {
	switch sort {
	case AssetSortTakenAtAsc:
		return "ORDER BY COALESCE(assets.taken_at, assets.file_mtime) ASC, assets.path ASC"
	default:
		return "ORDER BY COALESCE(assets.taken_at, assets.file_mtime) DESC, assets.path DESC"
	}
}

func (s *Storage) countAlbums() (int, error) {
	var total int

	err := s.db.QueryRow(`
SELECT COUNT(*)
FROM albums
WHERE EXISTS (
	SELECT 1
	FROM assets
	WHERE assets.album_id = albums.id
)
`).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Storage) countAssetsByAlbumSlug(slug string) (int, error) {
	var total int

	err := s.db.QueryRow(`
SELECT COUNT(*)
FROM albums AS root
JOIN albums AS child
	ON child.path = root.path
	OR (
		child.path COLLATE BINARY >= root.path || '/'
		AND child.path COLLATE BINARY < root.path || '0'
	)
JOIN assets ON assets.album_id = child.id
WHERE root.slug = ?
`, slug).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
