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

func TestBug01_RunRawSliceAlias(t *testing.T) {
	p, err := imgpipe.New(imgpipe.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	raw := bugMakeJPEG(t, 24, 24)
	res, err := p.Run(imgpipe.Job{
		Raw:    raw,
		Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodeJPEG, Quality: 80},
	})
	if err != nil {
		t.Fatal(err)
	}
	for i := range raw {
		raw[i] = 0
	}
	fr, ok := p.Cache().Get(res.CacheKey)
	if !ok || fr == nil || len(fr.Raw) == 0 {
		t.Fatal("cache missing after Run")
	}
	allZero := true
	for _, b := range fr.Raw {
		if b != 0 {
			allZero = false
			break
		}
	}
	if allZero {
		t.Fatal("caller mutated Raw polluted cached frame input bytes")
	}
}
