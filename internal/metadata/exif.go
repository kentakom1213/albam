package metadata

import (
	"os"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
	"github.com/rwcarlsen/goexif/tiff"
)

type ExifMetadata struct {
	TakenAt             *time.Time
	Latitude            *float64
	Longitude           *float64
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
		meta.CameraMake = stringFromTag(tag)
	}

	if tag, err := x.Get(exif.Model); err == nil {
		meta.CameraModel = stringFromTag(tag)
	}

	if tag, err := x.Get(exif.LensMake); err == nil {
		meta.LensMake = stringFromTag(tag)
	}

	if tag, err := x.Get(exif.LensModel); err == nil {
		meta.LensModel = stringFromTag(tag)
	}

	if tag, err := x.Get(exif.FocalLength); err == nil {
		meta.FocalLengthMM = floatFromRatTag(tag)
	}

	if tag, err := x.Get(exif.FocalLengthIn35mmFilm); err == nil {
		meta.FocalLength35mm = intFromTag(tag)
	}

	if tag, err := x.Get(exif.FNumber); err == nil {
		meta.ApertureFNumber = floatFromRatTag(tag)
	}

	if tag, err := x.Get(exif.ExposureTime); err == nil {
		meta.ExposureTimeSeconds = floatFromRatTag(tag)
	}

	if tag, err := x.Get(exif.ISOSpeedRatings); err == nil {
		meta.ISO = intFromTag(tag)
	}

	if tag, err := x.Get(exif.Orientation); err == nil {
		meta.Orientation = intFromTag(tag)
	}

	return meta
}

func stringFromTag(tag *tiff.Tag) *string {
	value, err := tag.StringVal()
	if err != nil {
		return nil
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}

func floatFromRatTag(tag *tiff.Tag) *float64 {
	numerator, denominator, err := tag.Rat2(0)
	if err != nil || denominator == 0 {
		return nil
	}

	value := float64(numerator) / float64(denominator)
	return &value
}

func intFromTag(tag *tiff.Tag) *int {
	value, err := tag.Int(0)
	if err != nil {
		return nil
	}

	return &value
}
