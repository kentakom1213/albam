package indexer

import (
	"path"
	"sort"
	"strings"

	"github.com/kentakom1213/albam/internal/model"
	"github.com/kentakom1213/albam/internal/scanner"
)

type Library struct {
	Albums []model.Album
	Assets []model.Asset
}

func BuildLibrary(files []scanner.AssetFile) (*Library, error) {
	albumPaths := make(map[string]struct{})

	for _, file := range files {
		albumPath := path.Dir(file.RelPath)
		if albumPath == "." {
			albumPath = ""
		}

		for _, p := range expandAlbumPaths(albumPath) {
			albumPaths[p] = struct{}{}
		}
	}

	albums := make([]model.Album, 0, len(albumPaths))
	for albumPath := range albumPaths {
		albums = append(albums, model.Album{
			Path:  albumPath,
			Slug:  "",
			Title: titleFromPath(albumPath),
		})
	}

	sort.Slice(albums, func(i, j int) bool {
		return albums[i].Path < albums[j].Path
	})

	albumIDPyPath := make(map[string]int64, len(albums))
	for i := range albums {
		albums[i].ID = int64(i + 1)
		albumIDPyPath[albums[i].Path] = albums[i].ID
	}

	for i := range albums {
		parentPath, ok := parentAlbumPath(albums[i].Path)
		if !ok {
			continue
		}

		parentID, ok := albumIDPyPath[parentPath]
		if !ok {
			continue
		}

		albums[i].ParentID = &parentID
	}

	assets := make([]model.Asset, 0, len(files))
	for i, file := range files {
		albumPath := path.Dir(file.RelPath)
		if albumPath == "." {
			albumPath = ""
		}

		albumID := albumIDPyPath[albumPath]

		assets = append(assets, model.Asset{
			ID:       int64(i + 1),
			Slug:     "",
			AlbumID:  albumID,
			Path:     file.RelPath,
			Filename: file.Filename,
			Ext:      file.Ext,
			Size:     file.Size,
			ModTime:  file.ModTime,
		})
	}

	return &Library{
		Albums: albums,
		Assets: assets,
	}, nil
}

func expandAlbumPaths(albumPath string) []string {
	if albumPath == "" {
		return []string{""}
	}

	parts := strings.Split(albumPath, "/")
	paths := make([]string, 0, len(parts))

	for i := range parts {
		paths = append(paths, strings.Join(parts[:i+1], "/"))
	}

	return paths
}

func parentAlbumPath(albumPath string) (string, bool) {
	if albumPath == "" {
		return "", false
	}

	parent := path.Dir(albumPath)
	if parent == "." {
		return "", false
	}

	return parent, true
}

func titleFromPath(albumPath string) string {
	if albumPath == "" {
		return "Root"
	}

	return path.Base(albumPath)
}
