package storage

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
)

const maxRetry = 10
const slugLength = 8
const slugAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateSlug(length int) (string, error) {
	b := make([]byte, length)

	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(slugAlphabet))))
		if err != nil {
			return "", err
		}

		b[i] = slugAlphabet[n.Int64()]
	}

	return string(b), nil
}

func (s *Storage) generateUniqueAlbumSlug(tx *sql.Tx) (string, error) {
	for i := 0; i < maxRetry; i++ {
		slug, err := generateSlug(slugLength)
		if err != nil {
			return "", err
		}

		var exists bool
		err = tx.QueryRow(
			`SELECT EXISTS(SELECT 1 FROM albums WHERE slug = ?)`,
			slug,
		).Scan(&exists)
		if err != nil {
			return "", err
		}

		if !exists {
			return slug, nil
		}
	}

	return "", fmt.Errorf("failed to generate unique album slug")
}
