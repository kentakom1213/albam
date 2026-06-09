package metadata

import (
	"os"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

type ExifMetadata struct {
	TakenAt     *time.Time
	Latitude    *float64
	Longitude   *float64
	CameraMake  *string
	CameraModel *string
}

func ReadExif(path string) ExifMetadata {
	var meta ExifMetadata

	f, err := os.Open(path)
	if err != nil {
		return meta
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return meta
	}

	if takenAt, err := x.DateTime(); err == nil {
		meta.TakenAt = &takenAt
	}

	if lat, lon, err := x.LatLong(); err == nil {
		meta.Latitude = &lat
		meta.Longitude = &lon
	}

	if tag, err := x.Get(exif.Make); err == nil {
		if value, err := tag.StringVal(); err == nil {
			value = strings.TrimSpace(value)
			if value != "" {
				meta.CameraMake = &value
			}
		}
	}

	if tag, err := x.Get(exif.Model); err == nil {
		if value, err := tag.StringVal(); err == nil {
			value = strings.TrimSpace(value)
			if value != "" {
				meta.CameraModel = &value
			}
		}
	}

	return meta
}
