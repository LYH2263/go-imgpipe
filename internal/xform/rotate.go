package xform

import (
        "image"
        "image/draw"

        "github.com/LYH2263/go-imgpipe/internal/errs"
)

func Rotate(src image.Image, degrees int) (*image.RGBA, error) {
        if src == nil {
                return nil, errs.ErrInvalidFrame
        }
        d := degrees % 360
        if d < 0 {
                d += 360
        }
        if d%90 != 0 {
                return nil, errs.ErrInvalidJob
        }
        b := src.Bounds()
        in := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
        draw.Draw(in, in.Bounds(), src, b.Min, draw.Src)
        switch d {
        case 0:
                return in, nil
        case 90:
                return rotate90(in), nil
        case 180:
                return rotate180(in), nil
        case 270:
                return rotate270(in), nil
        default:
                return nil, errs.ErrInvalidJob
        }
}

func rotate90(src *image.RGBA) *image.RGBA {
        w, h := src.Bounds().Dx(), src.Bounds().Dy()
        dst := image.NewRGBA(image.Rect(0, 0, h, w))
        for y := 0; y < h; y++ {
                for x := 0; x < w; x++ {
                        dst.Set(h-1-y, x, src.At(x, y))
                }
        }
        return dst
}

func rotate180(src *image.RGBA) *image.RGBA {
        w, h := src.Bounds().Dx(), src.Bounds().Dy()
        dst := image.NewRGBA(image.Rect(0, 0, w, h))
        for y := 0; y < h; y++ {
                for x := 0; x < w; x++ {
                        dst.Set(w-1-x, h-1-y, src.At(x, y))
                }
        }
        return dst
}

func rotate270(src *image.RGBA) *image.RGBA {
        w, h := src.Bounds().Dx(), src.Bounds().Dy()
        dst := image.NewRGBA(image.Rect(0, 0, h, w))
        for y := 0; y < h; y++ {
                for x := 0; x < w; x++ {
                        dst.Set(y, w-1-x, src.At(x, y))
                }
        }
        return dst
}

func Flip(src image.Image, flipH, flipV bool) *image.RGBA {
        b := src.Bounds()
        in := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
        draw.Draw(in, in.Bounds(), src, b.Min, draw.Src)
        if !flipH && !flipV {
                return in
        }
        w, h := in.Bounds().Dx(), in.Bounds().Dy()
        dst := image.NewRGBA(image.Rect(0, 0, w, h))
        for y := 0; y < h; y++ {
                for x := 0; x < w; x++ {
                        sx, sy := x, y
                        if flipH {
                                sx = w - 1 - x
                        }
                        if flipV {
                                sy = h - 1 - y
                        }
                        dst.Set(x, y, in.At(sx, sy))
                }
        }
        return dst
}
