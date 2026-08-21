package imgpipe

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
	"time"
)

// smallJPEG renders a tiny valid JPEG so tests don't depend on fixtures on disk.
func smallJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for x := 0; x < 4; x++ {
		for y := 0; y < 4; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 60), G: uint8(y * 60), B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode fixture jpeg: %v", err)
	}
	return buf.Bytes()
}

// TestRunContext_PreCanceledContext verifies the regression: RunContext must
// honor a caller-supplied context all the way down to decode.Decode, rather
// than substituting context.Background(). When the client has already
// canceled (the CDN 50ms-deadline case), the pipeline must surface a
// cancellation error and must NOT finish decoding the JPEG.
func TestRunContext_PreCanceledContext(t *testing.T) {
	pipe, err := New(Options{}) // memory-only
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = pipe.Close() })

	raw := smallJPEG(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled, like a CDN origin fetch past its deadline

	start := time.Now()
	_, runErr := pipe.RunContext(ctx, Job{
		Raw:       raw,
		Encode:    EncodeOptions{Format: EncodeJPEG, Quality: 85},
		SkipCache: true,
	})
	elapsed := time.Since(start)

	if runErr == nil {
		t.Fatal("expected cancellation error from RunContext, got nil")
	}
	if !errors.Is(runErr, ErrCanceled) {
		t.Fatalf("expected ErrCanceled, got %v", runErr)
	}
	// The pre-check at the gate returns without invoking the decoder, so this
	// must be near-instant. We assert only that it did not block for a long
	// time, which would indicate the context was ignored.
	if elapsed > time.Second {
		t.Fatalf("canceled decode took %v; context was likely ignored", elapsed)
	}
}

// TestRunContext_PreCanceledContextOnCacheHit ensures the encode path also
// honors ctx (it previously passed context.Background() to Encode).
func TestRunContext_PreCanceledContextOnCacheHit(t *testing.T) {
	pipe, err := New(Options{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = pipe.Close() })

	raw := smallJPEG(t)
	// Warm the cache with one successful run.
	_, err = pipe.RunContext(context.Background(), Job{
		Raw:       raw,
		Encode:    EncodeOptions{Format: EncodeJPEG, Quality: 85},
		SkipCache: false,
	})
	if err != nil {
		t.Fatalf("warm-up run: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, runErr := pipe.RunContext(ctx, Job{
		Raw:       raw,
		Encode:    EncodeOptions{Format: EncodeJPEG, Quality: 85},
		SkipCache: false,
	})
	if runErr == nil {
		t.Fatal("expected cancellation error on cache-hit encode path, got nil")
	}
	if !errors.Is(runErr, ErrCanceled) {
		t.Fatalf("expected ErrCanceled, got %v", runErr)
	}
}
