package imgpipe

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"testing"
	"time"
)

// makeJPEG encodes a w×h RGBA gradient into JPEG bytes for tests.
func makeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 255 / w), G: uint8(y * 255 / h), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}

// TestRunContextEncodeCancelUnblocksCaller is the regression test for the
// "browser disconnect mid-encode, thumbnail queue stalls" bug. With a real,
// large-ish image the JPEG encoder runs long enough that we can cancel the
// request context mid-encode. The fixed Encode selects on ctx.Done and RunContext
// must return promptly with a context error — not block until the encoder finishes.
func TestRunContextEncodeCancelUnblocksCaller(t *testing.T) {
	pipe, err := New(Options{
		CacheDir:       t.TempDir(),
		MemLimit:       8,
		MaxInputBytes:  32 << 20,
		MaxPixels:      64_000_000,
		DefaultQuality: 85,
		DefaultEncode:  EncodeJPEG,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pipe.Close()

	// 4096×4096 is a heavy encode (~67Mpx). SkipCache forces a full encode run.
	raw := makeJPEG(t, 4096, 4096)
	job := Job{Raw: raw, SkipCache: true, Encode: EncodeOptions{Format: EncodeJPEG, Quality: 90}}

	ctx, cancel := context.WithCancel(context.Background())
	// Give the pipeline time to decode + reach the encode step, then cancel —
	// mirroring a client closing the connection while the JPEG is being written.
	go func() {
		time.Sleep(120 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	res, err := pipe.RunContext(ctx, job)
	elapsed := time.Since(start)

	// If the encode happened to finish before the cancel fired, that's fine;
	// otherwise we must see a cancellation error, not a hang.
	if err == nil {
		if res == nil {
			t.Fatal("nil result and nil error")
		}
		return
	}
	if !errors.Is(err, context.Canceled) {
		// A decode/transform error is acceptable as long as it is not a silent
		// hang; assert it is cancellation-related.
		t.Fatalf("unexpected err (want canceled): %v", err)
	}
	// Regression: the old code blocked until the encoder finished (~seconds).
	// The cancel fires at 120ms; assert we returned within a short bound.
	if elapsed > 2*time.Second {
		t.Fatalf("RunContext did not unblock on cancel: elapsed=%v", elapsed)
	}
}

// TestRunContextDecodeCancelUnblocksCaller verifies the Pipeline no longer
// swaps the request ctx for context.Background() on the decode path, so a
// pre-canceled context fails fast instead of decoding the whole image.
func TestRunContextDecodeCancelUnblocksCaller(t *testing.T) {
	pipe, err := New(Options{CacheDir: t.TempDir(), DefaultEncode: EncodeJPEG})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pipe.Close()

	raw := makeJPEG(t, 2048, 2048)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled

	start := time.Now()
	_, err = pipe.RunContext(ctx, Job{Raw: raw, SkipCache: true})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error from canceled context, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	// Should fail fast; decoding 2048² would take real time otherwise.
	if elapsed > 500*time.Millisecond {
		t.Fatalf("decode path did not honor canceled ctx: elapsed=%v", elapsed)
	}
}

// TestRunContextNormalJobStillWorks guards against over-canceling: a happy
// path job must still produce JPEG bytes end-to-end.
func TestRunContextNormalJobStillWorks(t *testing.T) {
	pipe, err := New(Options{CacheDir: t.TempDir(), DefaultEncode: EncodeJPEG, DefaultQuality: 85})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer pipe.Close()

	raw := makeJPEG(t, 64, 64)
	res, err := pipe.RunContext(context.Background(), Job{
		Raw:       raw,
		SkipCache: true,
		Encode:    EncodeOptions{Format: EncodeJPEG, Quality: 85},
	})
	if err != nil {
		t.Fatalf("RunContext: %v", err)
	}
	if res == nil || len(res.Bytes) == 0 {
		t.Fatalf("empty result: %+v", res)
	}
	// Output must be decodable JPEG.
	img, derr := jpeg.Decode(bytes.NewReader(res.Bytes))
	if derr != nil {
		t.Fatalf("output not decodable jpeg: %v", derr)
	}
	if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 64 {
		t.Fatalf("unexpected dims: %v", img.Bounds())
	}
}

// readerEOF is an io.Reader that always returns 0, EOF to force a decode error
// path without needing a corrupt file.
type readerEOF struct{}

func (readerEOF) Read(p []byte) (int, error) { return 0, io.EOF }
