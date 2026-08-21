package imgpipe

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// newJPEG renders a tiny solid-color JPEG so we have a real decodable input.
func newJPEG(t *testing.T, c color.Color) []byte {
	t.Helper()
	const w, h = 4, 4
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("jpeg encode: %v", err)
	}
	return buf.Bytes()
}

// TestRunResultRawNotAliased reproduces the imgd trial scenario: the same
// upload buffer is handed to Run as Job.Raw; after Run succeeds the buffer is
// zeroed. If clone.Bytes is a no-op, Result.Frame.Raw aliases the buffer and is
// zeroed too — and the in-memory cache entry's Raw aliases it as well, so a
// subsequent cache hit returns an all-zero Raw.
func TestRunResultRawNotAliased(t *testing.T) {
	pipe, err := New(Options{CacheDir: ""}) // memory-only
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pipe.Close()

	upload := newJPEG(t, color.RGBA{R: 200, G: 30, B: 60, A: 255})
	origLen := len(upload)
	// Keep an independent copy of the original bytes so Run#2 can compute the
	// same cache key after the upload buffer is zeroed (a cache-hit path).
	original := append([]byte(nil), upload...)
	job := Job{
		Raw: upload,
		Encode: EncodeOptions{
			Format:  EncodeJPEG,
			Quality: 85,
		},
	}

	res, err := pipe.Run(job)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Zero the upload buffer exactly as the trial harness does.
	for i := range upload {
		upload[i] = 0
	}

	// Result.Frame.Raw must survive — it must be a copy, not the upload buffer.
	if len(res.Frame.Raw) != origLen {
		t.Fatalf("Result.Frame.Raw len = %d, want %d", len(res.Frame.Raw), origLen)
	}
	for i, b := range res.Frame.Raw {
		if b != upload[i] { // upload is now all zero
			// found a non-zero byte => not aliased => good; confirm it matches original
			_ = i
			break
		}
	}
	// Stronger check: Raw must equal the original bytes, which are all zero now
	// in `upload`, so verify against a freshly encoded copy instead.
	fresh := newJPEG(t, color.RGBA{R: 200, G: 30, B: 60, A: 255})
	if !bytes.Equal(res.Frame.Raw, fresh) {
		t.Fatalf("Result.Frame.Raw was mutated by zeroing upload (aliasing bug)\nfirst bytes: %v\nwant:        %v",
			res.Frame.Raw[:min(8, len(res.Frame.Raw))], fresh[:min(8, len(fresh))])
	}

	// Now hit the cache for the same key. The entry's Raw must also be intact.
	got, ok := pipe.cache.Get(res.CacheKey)
	if !ok {
		t.Fatalf("cache miss for %q", res.CacheKey)
	}
	if !bytes.Equal(got.Raw, fresh) {
		t.Fatalf("cached entry Raw was mutated by zeroing upload (aliasing in Put path)\nfirst bytes: %v\nwant:        %v",
			got.Raw[:min(8, len(got.Raw))], fresh[:min(8, len(fresh))])
	}

	// A second Run on the same key must come from cache (CacheHit) and have
	// intact Raw. We feed fresh original bytes (same content → same cache key),
	// NOT the now-zeroed upload buffer — the trial harness re-reads the source.
	job2 := job
	job2.Raw = original
	res2, err := pipe.Run(job2)
	if err != nil {
		t.Fatalf("Run#2: %v", err)
	}
	if !res2.CacheHit {
		t.Fatalf("Run#2 expected cache hit")
	}
	if !bytes.Equal(res2.Frame.Raw, fresh) {
		t.Fatalf("Run#2 Frame.Raw was mutated by zeroing upload\nfirst bytes: %v\nwant:        %v",
			res2.Frame.Raw[:min(8, len(res2.Frame.Raw))], fresh[:min(8, len(fresh))])
	}
}

// TestRunResultPixelsNotAliased ensures the pixel buffer in the returned frame
// and the cache entry are independent copies of the input pixels produced by
// FromImage — i.e. the same aliasing class for the Pixels field.
func TestRunResultPixelsNotAliased(t *testing.T) {
	pipe, err := New(Options{CacheDir: ""})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pipe.Close()

	upload := newJPEG(t, color.RGBA{R: 10, G: 220, B: 30, A: 255})
	res, err := pipe.Run(Job{Raw: upload, Encode: EncodeOptions{Format: EncodeJPEG, Quality: 85}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	// Mutate the frame's pixels and confirm the cache entry is unaffected.
	wantPix := append([]uint8(nil), res.Frame.Pixels...)
	for i := range res.Frame.Pixels {
		res.Frame.Pixels[i] ^= 0xff
	}
	got, ok := pipe.cache.Get(res.CacheKey)
	if !ok {
		t.Fatalf("cache miss")
	}
	if !bytes.Equal(got.Pixels, wantPix) {
		t.Fatalf("cached Pixels mutated via returned Frame (aliasing bug)")
	}
}
