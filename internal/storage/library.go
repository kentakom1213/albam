package storage

import (
	"database/sql"

	"github.com/kentakom1213/go-webapp-tutorial/internal/indexer"
	"github.com/kentakom1213/go-webapp-tutorial/internal/model"
)

func (s *Storage) SaveLibrary(library *indexer.Library) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	albumIDByPath := make(map[string]int64, len(library.Albums))

	for _, album := range library.Albums {
		id, err := upsertAlbum(tx, album, nil)
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

		if err := upsertAsset(tx, asset, albumID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func upsertAlbum(tx *sql.Tx, album model.Album, parentID *int64) (int64, error) {
	_, err := tx.Exec(`
INSERT INTO albums (parent_id, path, slug, title, updated_at)
VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(path) DO UPDATE SET
	slug = excluded.slug,
	title = excluded.title,
	updated_at = CURRENT_TIMESTAMP
	`, parentID, album.Path, album.Slug, album.Title)
	if err != nil {
		return 0, err
	}

	var id int64
	err = tx.QueryRow(`SELECT id FROM albums WHERE path = ?`, album.Path).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func updateAlbumParentID(tx *sql.Tx, albumPath string, parentID int64) error {
	_, err := tx.Exec(`
UPDATE albums
SET parent_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE path = ?
`, parentID, albumPath)

	return err
}

func upsertAsset(tx *sql.Tx, asset model.Asset, albumID int64) error {
	_, err := tx.Exec(`
INSERT INTO assets (
	album_id,
	path,
	filename,
	ext,
	size_bytes,
	file_mtime,
	updated_at
)
VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
ON CONFLICT(path) DO UPDATE SET
	album_id = excluded.album_id,
	filename = excluded.filename,
	ext = excluded.ext,
	size_bytes = excluded.size_bytes,
	file_mtime = excluded.file_mtime,
	updated_at = CURRENT_TIMESTAMP
`, albumID, asset.Path, asset.Filename, asset.Ext, asset.Size, asset.ModTime)

	return err
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
