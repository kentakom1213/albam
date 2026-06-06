package storage

import (
	"database/sql"
	"errors"

	"github.com/kentakom1213/albam/internal/indexer"
	"github.com/kentakom1213/albam/internal/model"
)

func (s *Storage) SaveLibrary(library *indexer.Library) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	albumIDByPath := make(map[string]int64, len(library.Albums))

	for _, album := range library.Albums {
		id, err := s.upsertAlbum(tx, album, nil)
		if err != nil {
			return err
		}
		albumIDByPath[album.Path] = id
	}

	for _, album := range library.Albums {
		parentID, ok := resolveParentID(album.Path, albumIDByPath)
		if !ok {
			continue
		}

		if err := updateAlbumParentID(tx, album.Path, parentID); err != nil {
			return err
		}
	}

	for _, asset := range library.Assets {
		albumPath := albumPathFromAssetPath(asset.Path)
		albumID := albumIDByPath[albumPath]

		if err := s.upsertAsset(tx, asset, albumID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) upsertAlbum(tx *sql.Tx, album model.Album, parentID *int64) (int64, error) {
	var id int64

	err := tx.QueryRow(`SELECT id FROM albums WHERE path = ?`, album.Path).Scan(&id)
	if err == nil {
		_, err := tx.Exec(`
UPDATE albums
SET title = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?
`, album.Title, id)
		if err != nil {
			return 0, err
		}

		return id, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	slug, err := s.generateUniqueAlbumSlug(tx)
	if err != nil {
		return 0, err
	}

	res, err := tx.Exec(`
INSERT INTO albums (
	parent_id,
	path,
	slug,
	title,
	updated_at
)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
`, parentID, album.Path, slug, album.Title)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func updateAlbumParentID(tx *sql.Tx, albumPath string, parentID int64) error {
	_, err := tx.Exec(`
UPDATE albums
SET parent_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE path = ?
`, parentID, albumPath)

	return err
}

func (s *Storage) upsertAsset(tx *sql.Tx, asset model.Asset, albumID int64) error {
	slug, err := s.generateUniquePhotoSlug(tx)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
INSERT INTO assets (
	album_id,
	slug,
	path,
	filename,
	ext,
	size_bytes,
	file_mtime,
	width,
	height,
	updated_at
)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(path) DO UPDATE SET
	album_id = excluded.album_id,
	filename = excluded.filename,
	ext = excluded.ext,
	size_bytes = excluded.size_bytes,
	file_mtime = excluded.file_mtime,
	width = excluded.width,
	height = excluded.height,
	updated_at = CURRENT_TIMESTAMP
`,
		albumID,
		slug,
		asset.Path,
		asset.Filename,
		asset.Ext,
		asset.Size,
		asset.ModTime,
		nullInt(asset.Width),
		nullInt(asset.Height),
	)

	return err
}

func nullInt(value int) any {
	if value <= 0 {
		return nil
	}

	return value
}

func resolveParentID(albumPath string, albumIDByPath map[string]int64) (int64, bool) {
	parentPath, ok := parentAlbumPath(albumPath)
	if !ok {
		return 0, false
	}

	parentID, ok := albumIDByPath[parentPath]
	if !ok {
		return 0, false
	}

	return parentID, true
}

func parentAlbumPath(albumPath string) (string, bool) {
	if albumPath == "" {
		return "", false
	}

	idx := -1
	for i := len(albumPath) - 1; i >= 0; i-- {
		if albumPath[i] == '/' {
			idx = i
			break
		}
	}

	if idx == -1 {
		return "", false
	}

	return albumPath[:idx], true
}

func albumPathFromAssetPath(assetPath string) string {
	idx := -1
	for i := len(assetPath) - 1; i >= 0; i-- {
		if assetPath[i] == '/' {
			idx = i
			break
		}
	}

	if idx == -1 {
		return ""
	}

	return assetPath[:idx]
}
