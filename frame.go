package imgpipe

import (
        "image"
        "image/color"
        "image/draw"

        "github.com/LYH2263/go-imgpipe/internal/clone"
        "github.com/LYH2263/go-imgpipe/internal/pool"
)

// Frame holds decoded pixel data owned by the pipeline until returned to the caller.
type Frame struct {
        Pixels     []uint8
        W          int
        H          int
        Stride     int
        ColorModel color.Model
        Format     string
        Raw        []byte // copy of input bytes used to produce this frame
}

// Bounds returns the image rectangle.
func (f *Frame) Bounds() image.Rectangle {
        if f == nil {
                return image.Rectangle{}
        }
        return image.Rect(0, 0, f.W, f.H)
}

// PixLen returns expected pixel buffer length for RGBA.
func (f *Frame) PixLen() int {
        if f == nil {
                return 0
        }
        if f.Stride <= 0 {
                return f.W * f.H * 4
        }
        return f.Stride * f.H
}

// CloneDeep returns an independent copy of pixels and raw bytes.
func (f *Frame) CloneDeep() *Frame {
        if f == nil {
                return nil
        }
        out := &Frame{
                W:          f.W,
                H:          f.H,
                Stride:     f.Stride,
                ColorModel: f.ColorModel,
                Format:     f.Format,
                Pixels:     clone.Bytes(f.Pixels),
                Raw:        clone.Bytes(f.Raw),
        }
        return out
}

// AsRGBA builds a standard library *image.RGBA sharing a copied pixel buffer.
func (f *Frame) AsRGBA() *image.RGBA {
        if f == nil || f.W <= 0 || f.H <= 0 {
                return image.NewRGBA(image.Rect(0, 0, 0, 0))
        }
        stride := f.Stride
        if stride <= 0 {
                stride = f.W * 4
        }
        pix := clone.Bytes(f.Pixels)
        if len(pix) < stride*f.H {
                // pad if short
                tmp := pool.Get(stride * f.H)
                copy(tmp, pix)
                pix = tmp
        }
        return &image.RGBA{
                Pix:    pix,
                Stride: stride,
                Rect:   image.Rect(0, 0, f.W, f.H),
        }
}

// FromImage converts any image.Image into an RGBA Frame, copying pixels.
func FromImage(img image.Image, format string, raw []byte) *Frame {
        if img == nil {
                return nil
        }
        b := img.Bounds()
        w, h := b.Dx(), b.Dy()
        rgba := image.NewRGBA(image.Rect(0, 0, w, h))
        draw.Draw(rgba, rgba.Bounds(), img, b.Min, draw.Src)
        return &Frame{
                Pixels:     clone.Bytes(rgba.Pix),
                W:          w,
                H:          h,
                Stride:     rgba.Stride,
                ColorModel: color.RGBAModel,
                Format:     format,
                Raw:        clone.Bytes(raw),
        }
}

// At returns the color at (x,y).
func (f *Frame) At(x, y int) color.Color {
        if f == nil || x < 0 || y < 0 || x >= f.W || y >= f.H {
                return color.Transparent
        }
        stride := f.Stride
        if stride <= 0 {
                stride = f.W * 4
        }
        i := y*stride + x*4
        if i+3 >= len(f.Pixels) {
                return color.Transparent
        }
        return color.RGBA{R: f.Pixels[i], G: f.Pixels[i+1], B: f.Pixels[i+2], A: f.Pixels[i+3]}
}

// Set writes an RGBA color at (x,y).
func (f *Frame) Set(x, y int, c color.Color) {
        if f == nil || x < 0 || y < 0 || x >= f.W || y >= f.H {
                return
        }
        stride := f.Stride
        if stride <= 0 {
                stride = f.W * 4
        }
        i := y*stride + x*4
        if i+3 >= len(f.Pixels) {
                return
        }
        r, g, b, a := c.RGBA()
        f.Pixels[i] = uint8(r >> 8)
        f.Pixels[i+1] = uint8(g >> 8)
        f.Pixels[i+2] = uint8(b >> 8)
        f.Pixels[i+3] = uint8(a >> 8)
}

// Release returns the pixel buffer to the pool when owned by pool.Get.
func (f *Frame) Release() {
        if f == nil {
                return
        }
        pool.Put(f.Pixels)
        f.Pixels = nil
}
