package storage

import (
	"context"
	"database/sql"
	"strconv"
)

func (s *Storage) GetPhotoForMediaByID(ctx context.Context, photoID string) (string, error) {
	id, err := strconv.ParseInt(photoID, 10, 64)
	if err != nil {
		return "", sql.ErrNoRows
	}

	var path string

	err = s.db.QueryRowContext(ctx, `
SELECT path
FROM assets
WHERE id = ?
`, id).Scan(&path)

	if err != nil {
		return "", err
	}

	return path, nil
}
