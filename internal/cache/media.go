package cache

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

func RemovePhotoVariantCaches(cacheRoot string, photoIDs []int64) error {
	for _, photoID := range photoIDs {
		id := strconv.FormatInt(photoID, 10)

		paths := []string{
			filepath.Join(cacheRoot, "media", "thumb", id+".webp"),
			filepath.Join(cacheRoot, "media", "preview", id+".webp"),
		}

		for _, path := range paths {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}

	return nil
}
