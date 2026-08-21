package xform

import (
        "image"
        "image/draw"

        "github.com/LYH2263/go-imgpipe/internal/errs"
)

func Crop(src image.Image, x, y, width, height int) (*image.RGBA, error) {
        if src == nil {
                return nil, errs.ErrInvalidFrame
        }
        if width <= 0 || height <= 0 || x < 0 || y < 0 {
                return nil, errs.ErrBadCrop
        }
        b := src.Bounds()
        if x+width > b.Dx() || y+height > b.Dy() {
                return nil, errs.ErrBadCrop
        }
        r := image.Rect(0, 0, width, height)
        dst := image.NewRGBA(r)
        draw.Draw(dst, r, src, b.Min.Add(image.Pt(x, y)), draw.Src)
        return dst, nil
}

func CropCenter(src image.Image, width, height int) (*image.RGBA, error) {
        if src == nil {
                return nil, errs.ErrInvalidFrame
        }
        b := src.Bounds()
        if width <= 0 || height <= 0 || width > b.Dx() || height > b.Dy() {
                return nil, errs.ErrBadCrop
        }
        x := (b.Dx() - width) / 2
        y := (b.Dy() - height) / 2
        return Crop(src, x, y, width, height)
}
