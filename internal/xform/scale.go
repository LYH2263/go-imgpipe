package xform

import (
        "fmt"
        "image"
        "image/draw"

        "github.com/LYH2263/go-imgpipe/internal/errs"
)

func Scale(src image.Image, width, height int, filter string) (*image.RGBA, error) {
        if src == nil {
                return nil, errs.ErrInvalidFrame
        }
        b := src.Bounds()
        sw, sh := b.Dx(), b.Dy()
        if sw <= 0 || sh <= 0 {
                return nil, errs.ErrBadScale
        }
        if width < 0 || height < 0 {
                return nil, errs.ErrBadScale
        }
        if width == 0 && height == 0 {
                return nil, errs.ErrBadScale
        }
        if width == 0 {
                width = sw * height / sh
                if width < 1 {
                        width = 1
                }
        }
        if height == 0 {
                height = sh * width / sw
                if height < 1 {
                        height = 1
                }
        }
        dst := image.NewRGBA(image.Rect(0, 0, width, height))
	switch filter {
	case "", "bilinear":
		bilinearScale(dst, src)
	case "nearest":
		nearestScale(dst, src)
	default:
		return nil, fmt.Errorf("%w: filter %q", errs.ErrBadScale, filter)
	}
        return dst, nil
}

func nearestScale(dst *image.RGBA, src image.Image) {
        db := dst.Bounds()
        sb := src.Bounds()
        sw, sh := sb.Dx(), sb.Dy()
        dw, dh := db.Dx(), db.Dy()
        for y := 0; y < dh; y++ {
                sy := sb.Min.Y + y*sh/dh
                for x := 0; x < dw; x++ {
                        sx := sb.Min.X + x*sw/dw
                        dst.Set(db.Min.X+x, db.Min.Y+y, src.At(sx, sy))
                }
        }
}

func bilinearScale(dst *image.RGBA, src image.Image) {
        sb := src.Bounds()
        tmp := image.NewRGBA(image.Rect(0, 0, sb.Dx(), sb.Dy()))
        draw.Draw(tmp, tmp.Bounds(), src, sb.Min, draw.Src)
        db := dst.Bounds()
        sw, sh := float64(tmp.Bounds().Dx()), float64(tmp.Bounds().Dy())
        dw, dh := db.Dx(), db.Dy()
        for y := 0; y < dh; y++ {
                fy := (float64(y)+0.5)*sh/float64(dh) - 0.5
                for x := 0; x < dw; x++ {
                        fx := (float64(x)+0.5)*sw/float64(dw) - 0.5
                        dst.SetRGBA(db.Min.X+x, db.Min.Y+y, sampleBilinearRGBA(tmp, fx, fy))
                }
        }
}
