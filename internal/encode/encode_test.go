package encode

import (
	"context"
	"errors"
	"image"
	"image/color"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

// slowEncoder blocks for d, then returns a sentinel byte. It never observes ctx
// itself — mirroring the stdlib jpeg/png encoders, which is exactly why
// Registry.Encode must select on ctx.Done rather than wait on the encoder.
func slowEncoder(d time.Duration, done *int32) Encoder {
	return func(ctx context.Context, img image.Image, quality int) ([]byte, error) {
		// Emulate a context-oblivious stdlib encoder: sleep, then produce.
		time.Sleep(d)
		atomic.StoreInt32(done, 1)
		return []byte{0x01}, nil
	}
}

func newRGBA(w, h int) *image.RGBA {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

// TestRegistryEncodeReturnsImmediatelyOnCancel is the regression test for the
// bug where a client disconnect mid-encode left the serving goroutine blocked
// in <-ch until the encoder finished. After the fix, Encode returns the
// context error as soon as ctx is canceled.
func TestRegistryEncodeReturnsImmediatelyOnCancel(t *testing.T) {
	var done int32
	r := &Registry{by: make(map[string]Encoder)}
	r.Register("jpeg", slowEncoder(300*time.Millisecond, &done))

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel shortly after Encode starts so the encoder is in flight.
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := r.Encode(ctx, "jpeg", newRGBA(64, 64), 85)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	// Must return well before the 300ms encoder finishes. Allow generous
	// scheduling slack but fail the old Sleep/死等 behavior.
	if elapsed > 200*time.Millisecond {
		t.Fatalf("Encode did not unblock on cancel: elapsed=%v (encoder needs 300ms)", elapsed)
	}
}

// TestRegistryEncodeWaitsForEncoderWhenNotCanceled ensures the fix did not
// make Encode race-return on every call — a normal run must still wait for
// and return the encoder's real output.
func TestRegistryEncodeWaitsForEncoderWhenNotCanceled(t *testing.T) {
	var done int32
	r := &Registry{by: make(map[string]Encoder)}
	r.Register("jpeg", slowEncoder(30*time.Millisecond, &done))

	out, err := r.Encode(context.Background(), "jpeg", newRGBA(16, 16), 85)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(out) != 1 || out[0] != 0x01 {
		t.Fatalf("unexpected output %v", out)
	}
	if atomic.LoadInt32(&done) != 1 {
		t.Fatal("encoder did not run to completion")
	}
}

// TestWaitWithContextRespectsCancel proves WaitWithContext no longer uses a
// fixed time.Sleep that ignores ctx.
func TestWaitWithContextRespectsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	err := WaitWithContext(ctx, 5*time.Second)
	elapsed := time.Since(start)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
	if elapsed > 500*time.Millisecond {
		t.Fatalf("WaitWithContext did not honor cancellation: elapsed=%v", elapsed)
	}
}

// TestWaitWithContextReturnsOnTimer proves the happy path still waits the full
// duration when nothing is canceled.
func TestWaitWithContextReturnsOnTimer(t *testing.T) {
	start := time.Now()
	err := WaitWithContext(context.Background(), 40*time.Millisecond)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if elapsed < 30*time.Millisecond {
		t.Fatalf("returned too early: elapsed=%v", elapsed)
	}
}

// TestRegistryEncodeNoGoroutineLeakOnCancel checks that the encoder goroutine,
// left running after a cancel, can still complete and write to the buffered
// channel without blocking forever — i.e. the buffer size of 1 is honored.
func TestRegistryEncodeNoGoroutineLeakOnCancel(t *testing.T) {
	before := runtime.NumGoroutine()
	var done int32
	r := &Registry{by: make(map[string]Encoder)}
	r.Register("jpeg", slowEncoder(50*time.Millisecond, &done))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled

	_, err := r.Encode(ctx, "jpeg", newRGBA(8, 8), 85)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// Give the encoder goroutine time to finish and exit.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&done) == 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if atomic.LoadInt32(&done) != 1 {
		t.Fatal("encoder goroutine never completed (buffered channel should have received)")
	}

	// Allow goroutine teardown, then assert we did not leak.
	time.Sleep(50 * time.Millisecond)
	after := runtime.NumGoroutine()
	if after > before+2 {
		t.Fatalf("possible goroutine leak: before=%d after=%d", before, after)
	}
}

// Ensure the existing stdlib encoders still compile/encode after the refactor.
func TestEncodeJPEGAndPNGStillWork(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	img.Set(0, 0, color.RGBA{R: 255, A: 255})

	jb, err := EncodeJPEG(context.Background(), img, 80)
	if err != nil || len(jb) == 0 {
		t.Fatalf("EncodeJPEG failed: err=%v len=%d", err, len(jb))
	}
	pb, err := EncodePNG(context.Background(), img, 0)
	if err != nil || len(pb) == 0 {
		t.Fatalf("EncodePNG failed: err=%v len=%d", err, len(pb))
	}
}
