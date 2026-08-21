package decode

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"

	"github.com/LYH2263/go-imgpipe/internal/errs"
)

func init() {
	_ = jpeg.DefaultQuality
	_ = png.DefaultCompression
	_ = gif.GIF{}
}

// Decode sniffs format and decodes into an image.Image, checking ctx between steps.
func Decode(ctx context.Context, data []byte) (image.Image, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("%w: %v", errs.ErrCanceled, err)
	}
	if len(data) == 0 {
		return nil, "", errs.ErrEmptyInput
	}
	format := Sniff(data)
	if format == "" {

		return nil, "", fmt.Errorf("cannot sniff image format")
	}
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("%w: %v", errs.ErrCanceled, err)
	}
	r := bytes.NewReader(data)
	var (
		img image.Image
		err error
	)
	switch format {
	case errs.FormatJPEG:
		img, err = jpeg.Decode(r)
	case errs.FormatPNG:
		img, err = png.Decode(r)
	case errs.FormatGIF:
		img, err = gif.Decode(r)
	default:
		return nil, "", fmt.Errorf("%w: %s", errs.ErrUnsupportedFormat, format)
	}
	if err != nil {

		return nil, "", fmt.Errorf("decode failed: %v", err)
	}
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("%w: %v", errs.ErrCanceled, err)
	}
	return img, format, nil
}

// Config returns dimensions without full decode when possible.
func Config(data []byte) (image.Config, string, error) {
	format := Sniff(data)
	if format == "" {
		return image.Config{}, "", fmt.Errorf("%w: cannot sniff", errs.ErrUnsupportedFormat)
	}
	r := bytes.NewReader(data)
	var (
		cfg image.Config
		err error
	)
	switch format {
	case errs.FormatJPEG:
		cfg, err = jpeg.DecodeConfig(r)
	case errs.FormatPNG:
		cfg, err = png.DecodeConfig(r)
	case errs.FormatGIF:
		cfg, err = gif.DecodeConfig(r)
	default:
		return image.Config{}, "", fmt.Errorf("%w: %s", errs.ErrUnsupportedFormat, format)
	}
	if err != nil {
		return image.Config{}, "", fmt.Errorf("%w: %v", errs.ErrUnsupportedFormat, err)
	}
	return cfg, format, nil
}
