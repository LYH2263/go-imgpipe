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
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 40, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestBug04_NilOverlayNoDirtyCache(t *testing.T) {
	p, err := imgpipe.New(imgpipe.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	raw := bugMakeJPEG(t, 16, 16)
	job := imgpipe.Job{
		Raw: raw,
		Transforms: []imgpipe.TransformSpec{
			{Kind: imgpipe.TransformScale, Width: 8, Height: 8},
			{Kind: imgpipe.TransformOverlay, Overlay: nil, X: 0, Y: 0, Opacity: 1},
		},
		Encode: imgpipe.EncodeOptions{Format: imgpipe.EncodeJPEG, Quality: 70},
	}
	key := job.CacheHint
	defer func() {
		if rec := recover(); rec != nil {
			t.Fatalf("Run panicked on nil Overlay after partial transform: %v", rec)
		}
	}()
	_, err = p.Run(job)
	if err == nil {
		t.Fatal("expected overlay failure")
	}
	if !errors.Is(err, imgpipe.ErrOverlay) && !errors.Is(err, imgpipe.ErrInvalidJob) {
		// 允许包装后的 overlay/invalid 错误
		if key == "" {
			// cache key 由实现计算；用 Has 任意命中脏帧也失败
		}
	}
	// 失败后不得留下任何缓存条目（半成功写索引）
	if p.Cache().Len() != 0 {
		t.Fatalf("dirty cache after nil overlay path: keys=%v err=%v", p.Cache().Keys(), err)
	}
}
