package xform

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"

	"github.com/LYH2263/go-imgpipe/internal/decode"
	"github.com/LYH2263/go-imgpipe/internal/errs"
)

func Overlay(base image.Image, overlayBytes []byte, x, y int, opacity float64) (*image.RGBA, error) {
	if base == nil {
		return nil, errs.ErrInvalidFrame
	}
	if opacity <= 0 {
		opacity = 1
	}
	if opacity > 1 {
		opacity = 1
	}

	ov, err := decodeOverlay(overlayBytes)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errs.ErrOverlay, err)
	}
	if ov == nil {
		return nil, errs.ErrOverlay
	}
	bb := base.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, bb.Dx(), bb.Dy()))
	draw.Draw(dst, dst.Bounds(), base, bb.Min, draw.Src)

	ob := ov.Bounds()
	if ob.Empty() {
		return nil, errs.ErrOverlay
	}
	tinted := image.NewRGBA(image.Rect(0, 0, ob.Dx(), ob.Dy()))
	for py := 0; py < ob.Dy(); py++ {
		for px := 0; px < ob.Dx(); px++ {
			c := color.RGBAModel.Convert(ov.At(ob.Min.X+px, ob.Min.Y+py)).(color.RGBA)
			c.A = uint8(float64(c.A) * opacity)
			tinted.Set(px, py, c)
		}
	}
	target := image.Rect(x, y, x+ob.Dx(), y+ob.Dy()).Intersect(dst.Bounds())
	if target.Empty() {
		return dst, nil
	}
	draw.Draw(dst, target, tinted, image.Pt(target.Min.X-x, target.Min.Y-y), draw.Over)
	return dst, nil
}

func decodeOverlay(data []byte) (image.Image, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("%w: empty overlay bytes", errs.ErrOverlay)
	}
	format := decode.Sniff(data)
	if format == "" {
		return nil, fmt.Errorf("%w: unrecognized overlay format", errs.ErrOverlay)
	}
	r := bytes.NewReader(data)
	switch format {
	case errs.FormatPNG:
		return png.Decode(r)
	case errs.FormatJPEG:
		return jpeg.Decode(r)
	default:
		return nil, fmt.Errorf("%w: unsupported overlay format %q", errs.ErrOverlay, format)
	}
}
