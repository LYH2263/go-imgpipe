package xform_test

import (
        "image"
        "image/color"
        "testing"

        "github.com/LYH2263/go-imgpipe/internal/xform"
)

func solid(w, h int, c color.Color) *image.RGBA {
        img := image.NewRGBA(image.Rect(0, 0, w, h))
        for y := 0; y < h; y++ {
                for x := 0; x < w; x++ {
                        img.Set(x, y, c)
                }
        }
        return img
}

func TestScaleNearest(t *testing.T) {
        src := solid(10, 10, color.RGBA{R: 255, A: 255})
        dst, err := xform.Scale(src, 5, 5, "nearest")
        if err != nil {
                t.Fatal(err)
        }
        if dst.Bounds().Dx() != 5 {
                t.Fatal(dst.Bounds())
        }
}

func TestCrop(t *testing.T) {
        src := solid(20, 20, color.RGBA{G: 255, A: 255})
        dst, err := xform.Crop(src, 2, 2, 8, 8)
        if err != nil {
                t.Fatal(err)
        }
        if dst.Bounds().Dx() != 8 || dst.Bounds().Dy() != 8 {
                t.Fatal(dst.Bounds())
        }
}

func TestRotate90(t *testing.T) {
        src := solid(4, 8, color.RGBA{B: 255, A: 255})
        dst, err := xform.Rotate(src, 90)
        if err != nil {
                t.Fatal(err)
        }
        if dst.Bounds().Dx() != 8 || dst.Bounds().Dy() != 4 {
                t.Fatalf("%v", dst.Bounds())
        }
}
