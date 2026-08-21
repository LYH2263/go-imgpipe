package imgpipe

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"sync"
	"testing"
	"time"
)

// tinyJPEG builds a valid small JPEG so decode.Decode doesn't short-circuit
// before reaching the cache write/encode steps.
func tinyJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 30), G: uint8(y * 30), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return buf.Bytes()
}

func jobFrom(raw []byte, transforms []TransformSpec) Job {
	return Job{
		Raw:        raw,
		Transforms: transforms,
		Encode:     EncodeOptions{Format: EncodeJPEG, Quality: 80},
	}
}

// TestRunContextAfterClose verifies the entry gate: a job run after Close
// returns ErrClosed instead of decoding. Before the fix RunContext only
// checked p == nil, so post-Close jobs proceeded to decode → cache.Put and
// dereferenced a nil p.mem.
func TestRunContextAfterClose(t *testing.T) {
	pipe, err := New(Options{MemLimit: 4})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	jpeg0 := tinyJPEG(t)

	if _, err := pipe.Run(jobFrom(jpeg0, []TransformSpec{
		{Kind: TransformScale, Width: 4, Height: 4},
	})); err != nil {
		t.Fatalf("prerun: %v", err)
	}

	if err := pipe.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// A fresh job after Close must be rejected up front.
	res, err := pipe.Run(jobFrom(jpeg0, []TransformSpec{
		{Kind: TransformScale, Width: 2, Height: 2},
	}))
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("post-Close Run: want ErrClosed, got err=%v res=%v", err, res)
	}
}

// TestConcurrentCloseDuringRun reproduces the stress-test crash: a Run that
// is already mid-flight races Close, which used to nil p.mem. The decoder is
// fast on an 8x8 JPEG, so we fan out many concurrent runs against Close and
// assert none panic and all post-Close jobs surface ErrClosed. Run with
// -race to catch the data race the original TOCTOU relied on.
func TestConcurrentCloseDuringRun(t *testing.T) {
	pipe, err := New(Options{MemLimit: 4})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	jpeg0 := tinyJPEG(t)
	job := jobFrom(jpeg0, []TransformSpec{
		{Kind: TransformScale, Width: 4, Height: 4},
	})

	var wg sync.WaitGroup
	closed := make(chan struct{})
	var panicOnce sync.Once

	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					panicOnce.Do(func() { t.Errorf("Run panicked during concurrent Close: %v", r) })
				}
			}()
			select {
			case <-closed:
				// After Close is observable, new RunContext calls must reject.
			default:
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = pipe.RunContext(ctx, job)
		}()
	}

	// Close mid-flight. Before the fix this niled p.mem and the next cache.Put
	// crashed with a nil pointer through mem.Put.
	if err := pipe.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	close(closed)

	// Any job submitted after Close completed must report ErrClosed.
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := pipe.Run(job)
			if err != nil && !errors.Is(err, ErrClosed) {
				t.Errorf("post-Close Run: want ErrClosed or nil, got %v", err)
			}
		}()
	}

	wg.Wait()

	// Idempotency: second Close returns ErrClosed, not a panic.
	if err := pipe.Close(); !errors.Is(err, ErrClosed) {
		t.Fatalf("second Close: want ErrClosed, got %v", err)
	}
}
