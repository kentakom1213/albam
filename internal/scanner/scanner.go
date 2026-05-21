package scanner

import (
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

type AssetFile struct {
	Path     string
	RelPath  string
	Filename string
	Ext      string
	Size     int64
	ModTime  time.Time
}

func Scan(root string) ([]AssetFile, error) {
	files := make([]AssetFile, 0)

	err := filepath.WalkDir(
		root,
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if !IsSupportedImage(path) {
				return nil
			}

			info, err := d.Info()
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}

			files = append(files, AssetFile{
				Path:     path,
				RelPath:  filepath.ToSlash(relPath),
				Filename: filepath.Base(path),
				Ext:      strings.ToLower(filepath.Ext(path)),
				Size:     info.Size(),
				ModTime:  info.ModTime(),
			})

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return files, nil
}

// TODO: 設定ファイルで読み込めるように
func IsSupportedImage(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
