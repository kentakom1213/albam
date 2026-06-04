package imageproc

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

type Options struct {
	MaxWidth  int
	MaxHeight int
	Quality   int
}

func EnsureJPEGVariant(srcPath string, dstPath string, opts Options) (bool, error) {
	opts = normalizeOptions(opts)

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false, fmt.Errorf("stat source image: %w", err)
	}
	if srcInfo.IsDir() {
		return false, fmt.Errorf("source is directory: %s", srcPath)
	}

	if dstInfo, err := os.Stat(dstPath); err == nil {
		if !dstInfo.IsDir() && !dstInfo.ModTime().Before(srcInfo.ModTime()) {
			return false, nil
		}
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat cached image: %w", err)
	}

	img, err := decodeImage(srcPath)
	if err != nil {
		return false, err
	}

	resized := resizeToFit(img, opts.MaxWidth, opts.MaxHeight)

	if err := os.MkdirAll(filepath.Dir(dstPath), 0o755); err != nil {
		return false, fmt.Errorf("create cache dir: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dstPath), ".tmp-"+filepath.Base(dstPath)+"-*")
	if err != nil {
		return false, fmt.Errorf("create temp cache file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if err := jpeg.Encode(tmp, resized, &jpeg.Options{Quality: opts.Quality}); err != nil {
		_ = tmp.Close()
		return false, fmt.Errorf("decode jpeg: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return false, fmt.Errorf("close temp cache file: %w", err)
	}

	if err := os.Rename(tmpPath, dstPath); err != nil {
		return false, fmt.Errorf("rename cache file: %w", err)
	}

	return true, nil
}

func normalizeOptions(opts Options) Options {
	if opts.MaxWidth <= 0 {
		opts.MaxWidth = 512
	}
	if opts.MaxHeight <= 0 {
		opts.MaxHeight = opts.MaxWidth
	}
	if opts.Quality <= 0 {
		opts.Quality = 82
	}
	if opts.Quality >= 100 {
		opts.Quality = 100
	}
	return opts
}

func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open source image: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return img, nil
}

func resizeToFit(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	dstW, dstH := FitSize(srcW, srcH, maxWidth, maxHeight)

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))

	// JPEG は alpha を持てないので，透過 PNG / WebP は白背景に合成
	xdraw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, xdraw.Src)
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)

	return dst
}

// returns (dstW, dstH)
func FitSize(srcW, srcH, maxW, maxH int) (int, int) {
	if srcW <= 0 || srcH <= 0 {
		return 1, 1
	}
	if maxW <= 0 && maxH <= 0 {
		return srcW, srcH
	}
	if maxW <= 0 {
		maxW = srcW
	}
	if maxH <= 0 {
		maxH = srcH
	}
	if srcW <= maxW && srcH <= maxH {
		return srcW, srcH
	}

	widthRatio := float64(maxW) / float64(srcW)
	heightRatio := float64(maxH) / float64(srcH)
	ratio := min(widthRatio, heightRatio)

	dstW := int(float64(srcW)*ratio + 0.5)
	dstH := int(float64(srcH)*ratio + 0.5)

	if dstW < 1 {
		dstW = 1
	}
	if dstH < 1 {
		dstH = 1
	}

	return dstW, dstH
}
