package xform

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/LYH2263/go-imgpipe/internal/errs"
)

func baseImg() *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, 4, 4))
}

// makePNG encodes a tiny solid-color PNG so Overlay has valid bytes to decode.
func makePNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// TestOverlayEmptyBytesNoPanic reproduces the reported nil-deref panic:
// empty watermark bytes used to be decoded with a discarded error, then
// ov.Bounds() was called on a nil image. It must now return an error.
func TestOverlayEmptyBytesNoPanic(t *testing.T) {
	_, err := Overlay(baseImg(), nil, 0, 0, 1)
	if err == nil {
		t.Fatal("expected error for empty overlay bytes, got nil")
	}
	if !errors.Is(err, errs.ErrOverlay) {
		t.Fatalf("expected ErrOverlay, got %v", err)
	}
}

// TestOverlayUnrecognizedBytesNoPanic covers garbage (unsniffable) overlay
// bytes: previously the nil decoded image was dereferenced at Bounds().
func TestOverlayUnrecognizedBytesNoPanic(t *testing.T) {
	_, err := Overlay(baseImg(), []byte("not an image"), 0, 0, 1)
	if err == nil {
		t.Fatal("expected error for unrecognized overlay bytes")
	}
	if !errors.Is(err, errs.ErrOverlay) {
		t.Fatalf("expected ErrOverlay, got %v", err)
	}
}

// TestOverlayValidBytes still composites successfully, guarding against an
// over-broad regression that would reject all overlays.
func TestOverlayValidBytes(t *testing.T) {
	ov := makePNG(t, 2, 2)
	dst, err := Overlay(baseImg(), ov, 0, 0, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst == nil {
		t.Fatal("expected non-nil dst")
	}
	if got := dst.Bounds().Dx(); got != 4 {
		t.Fatalf("dst width = %d, want 4", got)
	}
}

// TestOverlayZeroBoundsImage guards against a decoded-but-empty overlay image
// reaching Bounds() arithmetic with a zero rectangle.
func TestOverlayZeroBoundsImage(t *testing.T) {
	// Hand a nil image straight into the compositing path by calling the
	// internal decodeOverlay to confirm it never yields a nil-without-error.
	ov, err := decodeOverlay(nil)
	if err == nil || ov != nil {
		t.Fatalf("decodeOverlay(nil) = (%v, %v), want (nil, err)", ov, err)
	}
}
