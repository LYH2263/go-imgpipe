package imgpipe_test

import (
	"bytes"
	"context"
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

func TestBug07_RunContextHonorsCancel(t *testing.T) {
	p, err := imgpipe.New(imgpipe.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = p.RunContext(ctx, imgpipe.Job{
		Raw:       bugMakeJPEG(t, 48, 48),
		SkipCache: true,
		Encode:    imgpipe.EncodeOptions{Format: imgpipe.EncodeJPEG, Quality: 75},
	})
	if err == nil {
		t.Fatal("expected cancel error")
	}
	if !errors.Is(err, imgpipe.ErrCanceled) && !errors.Is(err, context.Canceled) {
		t.Fatalf("want canceled, got %v", err)
	}
}
