package imageproc

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/gen2brain/webp"
	"github.com/rwcarlsen/goexif/exif"
	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

type Options struct {
	MaxWidth  int
	MaxHeight int
	Quality   int
}

func EnsureWebPVariant(srcPath string, dstPath string, opts Options) (bool, error) {
	opts = normalizeOptions(opts)

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return false, fmt.Errorf("stat source image: %w", err)
	}
	if srcInfo.IsDir() {
		return false, fmt.Errorf("source is directory: %s", srcPath)
	}

	orientation := readExifOrientation(srcPath)
	srcW, srcH, err := readImageSize(srcPath)
	if err != nil {
		return false, err
	}
	srcW, srcH = orientedImageSize(srcW, srcH, orientation)
	wantW, wantH := FitSize(srcW, srcH, opts.MaxWidth, opts.MaxHeight)

	if dstInfo, err := os.Stat(dstPath); err == nil {
		if !dstInfo.IsDir() && !dstInfo.ModTime().Before(srcInfo.ModTime()) && cachedVariantMatches(dstPath, wantW, wantH) {
			return false, nil
		}
	} else if !os.IsNotExist(err) {
		return false, fmt.Errorf("stat cached image: %w", err)
	}

	img, err := decodeImage(srcPath, orientation)
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

	if err := webp.Encode(tmp, resized, webp.Options{Quality: opts.Quality}); err != nil {
		_ = tmp.Close()
		return false, fmt.Errorf("encode webp: %w", err)
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

func readImageSize(path string) (int, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("open source image: %w", err)
	}
	defer f.Close()

	config, _, err := image.DecodeConfig(f)
	if err != nil {
		return 0, 0, fmt.Errorf("decode image metadata: %w", err)
	}

	return config.Width, config.Height, nil
}

func orientedImageSize(width, height int, orientation int) (int, int) {
	switch orientation {
	case 5, 6, 7, 8:
		return height, width
	default:
		return width, height
	}
}

func cachedVariantMatches(path string, wantW, wantH int) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	config, _, err := image.DecodeConfig(f)
	if err != nil {
		return false
	}

	return config.Width == wantW && config.Height == wantH
}

func decodeImage(path string, orientation int) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open source image: %w", err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	return applyOrientation(img, orientation), nil
}

func resizeToFit(src image.Image, maxWidth, maxHeight int) image.Image {
	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	dstW, dstH := FitSize(srcW, srcH, maxWidth, maxHeight)

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))

	// 配信用画像の背景色を入力形式にかかわらず統一する．
	xdraw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, xdraw.Src)
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)

	return dst
}

func readExifOrientation(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 1
	}
	defer f.Close()

	x, err := exif.Decode(f)
	if err != nil {
		return 1
	}

	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return 1
	}

	orientation, err := tag.Int(0)
	if err != nil {
		return 1
	}
	if orientation < 1 || orientation > 8 {
		return 1
	}

	return orientation
}

func applyOrientation(src image.Image, orientation int) image.Image {
	if orientation == 1 {
		return src
	}

	bounds := src.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	dstW, dstH := srcW, srcH
	if orientation >= 5 && orientation <= 8 {
		dstW, dstH = srcH, srcW
	}

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))
	for y := 0; y < dstH; y++ {
		for x := 0; x < dstW; x++ {
			srcX, srcY := orientedSourcePoint(x, y, srcW, srcH, orientation)
			dst.Set(x, y, src.At(bounds.Min.X+srcX, bounds.Min.Y+srcY))
		}
	}

	return dst
}

func orientedSourcePoint(x, y, srcW, srcH, orientation int) (int, int) {
	switch orientation {
	case 2:
		return srcW - 1 - x, y
	case 3:
		return srcW - 1 - x, srcH - 1 - y
	case 4:
		return x, srcH - 1 - y
	case 5:
		return y, x
	case 6:
		return y, srcH - 1 - x
	case 7:
		return srcW - 1 - y, srcH - 1 - x
	case 8:
		return srcW - 1 - y, x
	default:
		return x, y
	}
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
