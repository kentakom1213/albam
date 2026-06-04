package storage

import (
	"database/sql"
	"path/filepath"
)

type AssetIdentity struct {
	ID   int64
	Path string
}

type PruneAssetsResult struct {
	Removed []AssetIdentity
}

func (s *Storage) PruneAssetsByPaths(keepPaths []string) (PruneAssetsResult, error) {
	keep := make(map[string]struct{}, len(keepPaths))
	for _, path := range keepPaths {
		keep[normalizeAssetPath(path)] = struct{}{}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return PruneAssetsResult{}, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	existing, err := listAssetIdentitiesTx(tx)
	if err != nil {
		return PruneAssetsResult{}, err
	}

	stmt, err := tx.Prepare(`
DELETE FROM assets
WHERE id = ?
`)
	if err != nil {
		return PruneAssetsResult{}, err
	}
	defer stmt.Close()

	result := PruneAssetsResult{
		Removed: make([]AssetIdentity, 0),
	}

	for _, asset := range existing {
		if _, ok := keep[normalizeAssetPath(asset.Path)]; ok {
			continue
		}

		if _, err := stmt.Exec(asset.ID); err != nil {
			return PruneAssetsResult{}, err
		}

		result.Removed = append(result.Removed, asset)
	}

	if err := tx.Commit(); err != nil {
		return PruneAssetsResult{}, err
	}

	return result, nil
}

func listAssetIdentitiesTx(tx *sql.Tx) ([]AssetIdentity, error) {
	rows, err := tx.Query(`
SELECT
	id,
	path
FROM assets
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]AssetIdentity, 0)
	for rows.Next() {
		var asset AssetIdentity
		if err := rows.Scan(&asset.ID, &asset.Path); err != nil {
			return nil, err
		}

		assets = append(assets, asset)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return assets, nil
}

func normalizeAssetPath(path string) string {
	return filepath.ToSlash(filepath.Clean(path))
}
