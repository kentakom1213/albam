package storage

import (
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

	return tx.Commit()
}
