package imgpipe_test

import (
	"bytes"
	"errors"
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

func TestBug03_RunAfterCloseNoPanic(t *testing.T) {
	p, err := imgpipe.New(imgpipe.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Close(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Run after Close panicked: %v", r)
		}
	}()
	_, err = p.Run(imgpipe.Job{
		Raw:    bugMakeJPEG(t, 12, 12),
		Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodeJPEG, Quality: 70},
	})
	if !errors.Is(err, imgpipe.ErrClosed) {
		t.Fatalf("want ErrClosed, got %v", err)
	}
}
