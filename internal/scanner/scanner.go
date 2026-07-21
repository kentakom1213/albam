package scanner

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/kentakom1213/albam/internal/metadata"
	_ "golang.org/x/image/webp"
)

type AssetFile struct {
	Path     string
	RelPath  string
	Filename string
	Ext      string
	Size     int64
	ModTime  time.Time

	Width               int
	Height              int
	TakenAt             *time.Time
	GPSLatitude         *float64
	GPSLongitude        *float64
	CameraMake          *string
	CameraModel         *string
	LensMake            *string
	LensModel           *string
	FocalLengthMM       *float64
	FocalLength35mm     *int
	ApertureFNumber     *float64
	ExposureTimeSeconds *float64
	ISO                 *int
	Orientation         *int
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

			width, height, err := readImageSize(path)
			if err != nil {
				return err
			}

			exifMeta := metadata.ReadExif(path)
			width, height = orientedImageSize(width, height, exifMeta.Orientation)

			files = append(files, AssetFile{
				Path:                path,
				RelPath:             filepath.ToSlash(relPath),
				Filename:            filepath.Base(path),
				Ext:                 strings.ToLower(filepath.Ext(path)),
				Size:                info.Size(),
				ModTime:             info.ModTime(),
				Width:               width,
				Height:              height,
				TakenAt:             exifMeta.TakenAt,
				GPSLatitude:         exifMeta.Latitude,
				GPSLongitude:        exifMeta.Longitude,
				CameraMake:          exifMeta.CameraMake,
				CameraModel:         exifMeta.CameraModel,
				LensMake:            exifMeta.LensMake,
				LensModel:           exifMeta.LensModel,
				FocalLengthMM:       exifMeta.FocalLengthMM,
				FocalLength35mm:     exifMeta.FocalLength35mm,
				ApertureFNumber:     exifMeta.ApertureFNumber,
				ExposureTimeSeconds: exifMeta.ExposureTimeSeconds,
				ISO:                 exifMeta.ISO,
				Orientation:         exifMeta.Orientation,
			})

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	// sort files
	sort.Slice(
		files,
		func(i, j int) bool {
			return files[i].RelPath <= files[j].RelPath
		},
	)

	return files, nil
}

func readImageSize(path string) (int, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("open image for metadata: %w", err)
	}
	defer file.Close()

	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("decode image metadata: %w", err)
	}

	return config.Width, config.Height, nil
}

func orientedImageSize(width, height int, orientation *int) (int, int) {
	if orientation == nil {
		return width, height
	}

	switch *orientation {
	case 5, 6, 7, 8:
		return height, width
	default:
		return width, height
	}
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
