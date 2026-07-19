package imageproc

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/image/webp"
)

func TestFitSize(t *testing.T) {
	tests := []struct {
		name         string
		srcW, srcH   int
		maxW, maxH   int
		wantW, wantH int
	}{
		{
			name:  "landscape",
			srcW:  4000,
			srcH:  3000,
			maxW:  1600,
			maxH:  1600,
			wantW: 1600,
			wantH: 1200,
		},
		{
			name:  "portrait",
			srcW:  3000,
			srcH:  4000,
			maxW:  1600,
			maxH:  1600,
			wantW: 1200,
			wantH: 1600,
		},
		{
			name:  "small image is not enlarged",
			srcW:  300,
			srcH:  200,
			maxW:  512,
			maxH:  512,
			wantW: 300,
			wantH: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := FitSize(tt.srcW, tt.srcH, tt.maxW, tt.maxH)
			if gotW != tt.wantW || gotH != tt.wantH {
				t.Fatalf("FitSize() = (%d, %d), want (%d, %d)", gotW, gotH, tt.wantW, tt.wantH)
			}
		})
	}
}

func TestEnsureWebPVariant(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src.png")
	dstPath := filepath.Join(dir, "cache", "thumb.webp")

	src := image.NewRGBA(image.Rect(0, 0, 100, 50))
	for y := 0; y < 50; y++ {
		for x := 0; x < 100; x++ {
			src.Set(x, y, color.RGBA{R: 255, A: 255})
		}
	}

	f, err := os.Create(srcPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, src); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	created, err := EnsureWebPVariant(srcPath, dstPath, Options{
		MaxWidth:  25,
		MaxHeight: 25,
		Quality:   80,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("first call should create cache")
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Size() == 0 {
		t.Fatal("cached file is empty")
	}

	variant, err := os.Open(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	config, err := webp.DecodeConfig(variant)
	_ = variant.Close()
	if err != nil {
		t.Fatalf("decode cached webp: %v", err)
	}
	if config.Width != 25 || config.Height != 13 {
		t.Fatalf("cached dimensions = %dx%d, want 25x13", config.Width, config.Height)
	}

	created, err = EnsureWebPVariant(srcPath, dstPath, Options{
		MaxWidth:  25,
		MaxHeight: 25,
		Quality:   80,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created {
		t.Fatal("second call should reuse cache")
	}
}

func TestEnsureWebPVariantAppliesExifOrientation(t *testing.T) {
	dir := t.TempDir()

	srcPath := filepath.Join(dir, "src.jpg")
	dstPath := filepath.Join(dir, "cache", "preview.webp")

	src := image.NewRGBA(image.Rect(0, 0, 80, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			src.Set(x, y, color.RGBA{G: 255, A: 255})
		}
	}

	encoded, err := encodeJPEGWithOrientation(src, 6)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(srcPath, encoded, 0o644); err != nil {
		t.Fatal(err)
	}

	created, err := EnsureWebPVariant(srcPath, dstPath, Options{
		MaxWidth:  100,
		MaxHeight: 100,
		Quality:   80,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("first call should create cache")
	}

	variant, err := os.Open(dstPath)
	if err != nil {
		t.Fatal(err)
	}
	config, err := webp.DecodeConfig(variant)
	_ = variant.Close()
	if err != nil {
		t.Fatalf("decode cached webp: %v", err)
	}
	if config.Width != 40 || config.Height != 80 {
		t.Fatalf("cached dimensions = %dx%d, want 40x80", config.Width, config.Height)
	}
}

func encodeJPEGWithOrientation(img image.Image, orientation byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}

	jpegBytes := buf.Bytes()
	if len(jpegBytes) < 2 || jpegBytes[0] != 0xff || jpegBytes[1] != 0xd8 {
		return jpegBytes, nil
	}

	exifPayload := []byte{
		'E', 'x', 'i', 'f', 0, 0,
		'I', 'I', 42, 0,
		8, 0, 0, 0,
		1, 0,
		0x12, 0x01,
		3, 0,
		1, 0, 0, 0,
		orientation, 0, 0, 0,
		0, 0, 0, 0,
	}

	segment := []byte{0xff, 0xe1, 0, 0}
	binary.BigEndian.PutUint16(segment[2:], uint16(len(exifPayload)+2))
	segment = append(segment, exifPayload...)

	out := make([]byte, 0, len(jpegBytes)+len(segment))
	out = append(out, jpegBytes[:2]...)
	out = append(out, segment...)
	out = append(out, jpegBytes[2:]...)
	return out, nil
}
