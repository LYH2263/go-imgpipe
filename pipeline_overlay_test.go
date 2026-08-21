package imgpipe

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// tinyPNG encodes a small solid PNG suitable for running through the pipeline.
func tinyPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 10, G: 20, B: 30, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return buf.Bytes()
}

// TestOverlayFailureLeavesNoDirtyCache reproduces the reported scenario:
// a job chains scale then an overlay whose bytes are empty. Previously the
// scale-only intermediate was cached under the full job key before overlay
// ran, so the failed job left a half-success entry: Cache.Has(key) == true
// and a later retry would serve the unwatermarked frame as a cache hit.
//
// After the fix, transform failures must not poison the cache.
func TestOverlayFailureLeavesNoDirtyCache(t *testing.T) {
	raw := tinyPNG(t, 8, 8)
	pipe, err := New(Options{CacheDir: "", MemLimit: 16})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pipe.Close()
	c := pipe.Cache()

	job := Job{
		Raw: raw,
		Transforms: []TransformSpec{
			{Kind: TransformScale, Width: 4, Height: 4, Filter: FilterBilinear},
			{Kind: TransformOverlay, Overlay: nil}, // empty overlay -> failure
		},
		Encode: EncodeOptions{Format: EncodePNG, Quality: 0},
	}

	// Pre-resolve the key the pipeline would compute so we can inspect the
	// index even though Run swallows the transform error into its result.
	key := pipe.PreviewKey(job)

	if _, err := pipe.Run(job); err == nil {
		t.Fatal("expected Run to fail because overlay bytes are empty")
	}

	if c.Has(key) {
		t.Fatalf("dirty frame for key %s left in cache after failed overlay; "+
			"overlay failure path must not leave a half-success cache entry", key)
	}
}

// TestSuccessfulJobPopulatesCache is the positive control: a job that fully
// succeeds must still populate the cache, so the fix did not simply disable
// caching.
func TestSuccessfulJobPopulatesCache(t *testing.T) {
	raw := tinyPNG(t, 8, 8)
	pipe, err := New(Options{CacheDir: "", MemLimit: 16})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pipe.Close()
	c := pipe.Cache()

	job := Job{
		Raw: raw,
		Transforms: []TransformSpec{
			{Kind: TransformScale, Width: 4, Height: 4, Filter: FilterBilinear},
		},
		Encode: EncodeOptions{Format: EncodePNG, Quality: 0},
	}
	key := pipe.PreviewKey(job)

	if _, err := pipe.Run(job); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !c.Has(key) {
		t.Fatal("expected cache to hold the final frame for a successful job")
	}
}

// TestOverlayFailureReportsErrOverlay confirms the overlay nil-deref is
// converted to a clean error rather than a panic, and is reachable through
// the public pipeline.
func TestOverlayFailureReportsErrOverlay(t *testing.T) {
	raw := tinyPNG(t, 8, 8)
	pipe, err := New(Options{CacheDir: "", MemLimit: 16})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pipe.Close()

	job := Job{
		Raw: raw,
		Transforms: []TransformSpec{
			{Kind: TransformOverlay, Overlay: []byte("garbage")},
		},
		Encode: EncodeOptions{Format: EncodePNG, Quality: 0},
	}
	_, err = pipe.Run(job)
	if err == nil {
		t.Fatal("expected error for undecodable overlay bytes")
	}
	if !errors.Is(err, ErrOverlay) {
		t.Fatalf("expected ErrOverlay, got %v", err)
	}
}
