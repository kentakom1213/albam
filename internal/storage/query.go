package storage

import (
	"database/sql"
	"errors"
)

type AlbumRow struct {
	ID         int64
	Path       string
	Slug       string
	Title      string
	CreatedAt  string
	UpdatedAt  string
	PhotoCount int
}

type AssetRow struct {
	ID        int64
	AlbumID   int64
	Path      string
	Filename  string
	Ext       string
	Size      string
	ModTime   string
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
	COUNT(assets.id) AS photo_count
FROM albums
LEFT JOIN assets ON assets.album_id = albums.id
GROUP BY albums.id
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
    COUNT(assets.id) AS photo_count
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
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	return &album, nil
}

func (s *Storage) ListAssetsByAlbumSlug(slug string, limit, offset int) ([]AssetRow, int, error) {
	total, err := s.countAssetsByAlbumSlug(slug)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(`
SELECT
    assets.id,
    assets.album_id,
    assets.path,
    assets.filename,
    assets.ext,
    assets.size_bytes,
    assets.file_mtime,
    assets.created_at,
    assets.updated_at
FROM assets
JOIN albums ON albums.id = assets.album_id
WHERE albums.slug = ?
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
			&asset.AlbumID,
			&asset.Path,
			&asset.Filename,
			&asset.Ext,
			&asset.Size,
			&asset.ModTime,
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

	err := s.db.QueryRow(`SELECT COUNT(*) FROM albums`).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *Storage) countAssetsByAlbumSlug(slug string) (int, error) {
	var total int

	err := s.db.QueryRow(`
SELECT COUNT(*)
FROM assets
JOIN albums ON albums.id = assets.album_id
WHERE albums.slug = ?
`, slug).Scan(&total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
