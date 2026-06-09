package storage

import (
	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var schemaSQL string

func (s *Storage) Migrate() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(schemaSQL); err != nil {
		return err
	}

	if err := ensureAssetColumns(tx); err != nil {
		return err
	}

	return tx.Commit()
}

func ensureAssetColumns(tx *sql.Tx) error {
	columns := []struct {
		name       string
		definition string
	}{
		{name: "lens_make", definition: "TEXT"},
		{name: "lens_model", definition: "TEXT"},
		{name: "focal_length_mm", definition: "REAL"},
		{name: "focal_length_35mm", definition: "INTEGER"},
		{name: "aperture_f_number", definition: "REAL"},
		{name: "exposure_time_seconds", definition: "REAL"},
		{name: "iso", definition: "INTEGER"},
		{name: "orientation", definition: "INTEGER"},
	}

	existing, err := assetColumns(tx)
	if err != nil {
		return err
	}

	for _, column := range columns {
		if existing[column.name] {
			continue
		}

		if _, err := tx.Exec("ALTER TABLE assets ADD COLUMN " + column.name + " " + column.definition); err != nil {
			return err
		}
	}

	return nil
}

func assetColumns(tx *sql.Tx) (map[string]bool, error) {
	rows, err := tx.Query(`PRAGMA table_info(assets)`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var dataType string
		var notNull int
		var defaultValue any
		var pk int

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}

		columns[name] = true
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
}
