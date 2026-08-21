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
//
// The standard library decoders (jpeg/png/gif) do not consult a context, so a
// large 4K JPEG cannot be interrupted mid-scan: once jpeg.Decode begins it runs
// to completion. Decode honors ctx at the gate (before sniffing and before the
// decoder is invoked) and wraps the decoder in a select on ctx.Done() so that a
// cancellation that arrives while the decoder is busy returns promptly with
// errs.ErrCanceled instead of blocking on the (now-orphaned) decoder. The
// decoder goroutine is left to finish on its own; its result is discarded.
func Decode(ctx context.Context, data []byte) (image.Image, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("%w: %v", errs.ErrCanceled, err)
	}
	if len(data) == 0 {
		return nil, "", errs.ErrEmptyInput
	}
	format := Sniff(data)
	if format == "" {
		return nil, "", fmt.Errorf("%w: cannot sniff", errs.ErrUnsupportedFormat)
	}
	if err := ctx.Err(); err != nil {
		return nil, "", fmt.Errorf("%w: %v", errs.ErrCanceled, err)
	}

	// The stdlib decoders ignore context, so run the decode in a goroutine and
	// race it against ctx.Done(). On cancellation we return immediately; the
	// decoder keeps running to completion in the background and its result is
	// dropped. This is the same pattern used by internal/encode.Registry.Encode.
	type result struct {
		img image.Image
		err error
	}
	ch := make(chan result, 1)
	go func() {
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
			err = fmt.Errorf("%w: %s", errs.ErrUnsupportedFormat, format)
		}
		ch <- result{img, err}
	}()
	select {
	case <-ctx.Done():
		return nil, "", fmt.Errorf("%w: %v", errs.ErrCanceled, ctx.Err())
	case res := <-ch:
		if res.err != nil {
			return nil, "", fmt.Errorf("%w: %v", errs.ErrUnsupportedFormat, res.err)
		}
		return res.img, format, nil
	}
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
