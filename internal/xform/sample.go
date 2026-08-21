package xform

import (
        "image"
        "image/color"
)

func floor(v float64) float64 {
        i := int(v)
        if float64(i) > v {
                return float64(i - 1)
        }
        return float64(i)
}

func lerp4(c00, c10, c01, c11 uint8, tx, ty float64) uint8 {
        a := float64(c00)*(1-tx) + float64(c10)*tx
        b := float64(c01)*(1-tx) + float64(c11)*tx
        return uint8(a*(1-ty) + b*ty + 0.5)
}

func sampleBilinearRGBA(src *image.RGBA, fx, fy float64) color.RGBA {
        b := src.Bounds()
        x0 := int(floor(fx))
        y0 := int(floor(fy))
        x1 := x0 + 1
        y1 := y0 + 1
        if x0 < b.Min.X {
                x0 = b.Min.X
        }
        if y0 < b.Min.Y {
                y0 = b.Min.Y
        }
        if x1 > b.Max.X-1 {
                x1 = b.Max.X - 1
        }
        if y1 > b.Max.Y-1 {
                y1 = b.Max.Y - 1
        }
        if x0 > x1 {
                x0 = x1
        }
        if y0 > y1 {
                y0 = y1
        }
        tx := fx - float64(x0)
        ty := fy - float64(y0)
        if tx < 0 {
                tx = 0
        }
        if ty < 0 {
                ty = 0
        }
        c00 := src.RGBAAt(x0, y0)
        c10 := src.RGBAAt(x1, y0)
        c01 := src.RGBAAt(x0, y1)
        c11 := src.RGBAAt(x1, y1)
        return color.RGBA{
                R: lerp4(c00.R, c10.R, c01.R, c11.R, tx, ty),
                G: lerp4(c00.G, c10.G, c01.G, c11.G, tx, ty),
                B: lerp4(c00.B, c10.B, c01.B, c11.B, tx, ty),
                A: lerp4(c00.A, c10.A, c01.A, c11.A, tx, ty),
        }
}
