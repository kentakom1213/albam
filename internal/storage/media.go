package storage

import (
	"context"
)

func (s *Storage) GetPhotoForMediaByID(ctx context.Context, photoID string) (string, error) {
	var path string

	err := s.db.QueryRowContext(ctx, `
SELECT path
FROM assets
WHERE slug = ?
`, photoID).Scan(&path)

	if err != nil {
		return "", err
	}

	return path, nil
}
