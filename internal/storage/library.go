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
		_ = album
		panic("TODO")
	}

	_ = albumIDByPath
	panic("TODO")
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

	panic("TODO")
}
