package imgpipe_test

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/LYH2263/go-imgpipe"
)

func bugMakeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 2), B: 40, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestBug02_CacheGetPixelSliceAlias(t *testing.T) {
	p, err := imgpipe.New(imgpipe.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	raw := bugMakeJPEG(t, 20, 20)
	res, err := p.Run(imgpipe.Job{
		Raw: raw,
		Transforms: []imgpipe.TransformSpec{{
			Kind: imgpipe.TransformScale, Width: 10, Height: 10, Filter: imgpipe.FilterNearest,
		}},
		Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodePNG},
	})
	if err != nil {
		t.Fatal(err)
	}
	fr1, ok := p.Cache().Get(res.CacheKey)
	if !ok || fr1 == nil || len(fr1.Pixels) == 0 {
		t.Fatal("missing cache frame")
	}
	before := fr1.Pixels[0]
	fr1.Pixels[0] ^= 0xff
	fr2, ok := p.Cache().Get(res.CacheKey)
	if !ok || fr2 == nil {
		t.Fatal("second get missing")
	}
	if fr2.Pixels[0] != before {
		t.Fatalf("Cache.Get shared pixels with caller: got %#x want %#x", fr2.Pixels[0], before)
	}
}
