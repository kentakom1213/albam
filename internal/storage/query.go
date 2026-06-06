package storage

import (
	"database/sql"
	"errors"
)

type AlbumRow struct {
	ID           int64
	Path         string
	Slug         string
	Title        string
	CreatedAt    string
	UpdatedAt    string
	PhotoCount   int
	CoverPhotoID sql.NullString
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
	CreatedAt string
	UpdatedAt string
}

func (s *Storage) ListAlbums(limit, offset int) ([]AlbumRow, int, error) {
	total, err := s.countAlbums()
	if err != nil {
		return nil, 0, err
	}

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
ORDER BY albums.path
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

	return &album, nil
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

func (s *Storage) ListAssetsByAlbumSlug(slug string, limit, offset int) ([]AssetRow, int, error) {
	total, err := s.countAssetsByAlbumSlug(slug)
	if err != nil {
		return nil, 0, err
	}

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
ORDER BY assets.path
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
