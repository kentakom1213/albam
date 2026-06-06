package storage

import "fmt"

func (s *Storage) Migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS albums (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	parent_id INTEGER,
	path TEXT NOT NULL UNIQUE,
	slug TEXT NOT NULL,
	title TEXT NOT NULL,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (parent_id) REFERENCES albums(id)
);

CREATE TABLE IF NOT EXISTS assets (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	slug TEXT NOT NULL UNIQUE,
	album_id INTEGER NOT NULL,
	path TEXT NOT NULL UNIQUE,
	filename TEXT NOT NULL,
	ext TEXT NOT NULL,
	size_bytes INTEGER NOT NULL,
	file_mtime DATETIME NOT NULL,
	width INTEGER,
	height INTEGER,
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (album_id) REFERENCES albums(id)
);

CREATE INDEX IF NOT EXISTS idx_albums_slug ON albums(slug);
CREATE INDEX IF NOT EXISTS idx_albums_path ON albums(path COLLATE BINARY);
CREATE INDEX IF NOT EXISTS idx_assets_album_id ON assets(album_id);
	`)
	if err != nil {
		return err
	}

	if err := s.addColumnIfMissing("assets", "width", "INTEGER"); err != nil {
		return err
	}

	return s.addColumnIfMissing("assets", "height", "INTEGER")
}

func (s *Storage) addColumnIfMissing(tableName, columnName, columnType string) error {
	exists, err := s.columnExists(tableName, columnName)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	_, err = s.db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", tableName, columnName, columnType))
	return err
}

func (s *Storage) columnExists(tableName, columnName string) (bool, error) {
	rows, err := s.db.Query("PRAGMA table_info(" + tableName + ")")
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue any
		var primaryKey int

		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return false, err
		}

		if name == columnName {
			return true, nil
		}
	}

	return false, rows.Err()
}
