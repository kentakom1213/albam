package storage

import (
	"strings"

	"github.com/kentakom1213/albam/internal/indexer"
)

type AlbumBreadcrumbRow struct {
	ID    int64
	Path  string
	Slug  string
	Title string
}

func (s *Storage) ListAlbumBreadcrumbsBySlug(slug string) ([]AlbumBreadcrumbRow, error) {
	var albumPath string

	err := s.db.QueryRow(`
SELECT path
FROM albums
WHERE slug = ?
`, slug).Scan(&albumPath)
	if err != nil {
		return nil, err
	}

	paths := indexer.ExpandAlbumPaths(albumPath)
	placeholders := make([]string, len(paths))
	args := make([]any, len(paths))
	for i, p := range paths {
		placeholders[i] = "?"
		args[i] = p
	}

	rows, err := s.db.Query(`
SELECT
	id,
	path,
	slug,
	title
FROM albums
WHERE path IN (`+strings.Join(placeholders, ",")+`)
`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byPath := make(map[string]AlbumBreadcrumbRow, len(paths))
	for rows.Next() {
		var row AlbumBreadcrumbRow
		if err := rows.Scan(
			&row.ID,
			&row.Path,
			&row.Slug,
			&row.Title,
		); err != nil {
			return nil, err
		}

		byPath[row.Path] = row
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	breadcrumbs := make([]AlbumBreadcrumbRow, 0, len(paths))
	for _, p := range paths {
		if row, ok := byPath[p]; ok {
			breadcrumbs = append(breadcrumbs, row)
		}
	}

	return breadcrumbs, nil
}
